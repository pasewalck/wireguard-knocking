package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, err
}

type Config struct {
	DataPath         string   `toml:"data_path"`
	WgInterface      string   `toml:"wg_interface"`
	RemoveIpCmds     []string `toml:"remove_ip_cmds"`
	AddIpCmds        []string `toml:"add_ip_cmds"`
	KeepAliveSeconds int64    `toml:"keep_alive_seconds"`
}

func main() {

	configPath := "/etc/wireguard-knocking/config.toml"

	exists, _ := fileExists(configPath)

	if exists == false {
		fmt.Println("config at", configPath, "missing")
		return
	}

	var conf Config
	md, errDecode := toml.DecodeFile(configPath, &conf)
	if errDecode != nil {
		fmt.Println("error:", errDecode)
		return
	}
	if undecoded := md.Undecoded(); len(undecoded) > 0 {
		fmt.Println("unknown config keys: ", undecoded)
		return
	}

	exists, _ = fileExists(conf.DataPath)

	// We only use the keys, essentially using this as a Set
	lastPeerIpMap := map[string]int64{}
	newPeerIpMap := map[string]int64{}

	if exists == true {
		b, err := os.ReadFile(conf.DataPath)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		var items []string
		err = json.Unmarshal(b, &items)
		if err != nil {
			fmt.Println("error:", err)
			return
		}
		for _, it := range items {
			lastPeerIpMap[it] = 1
		}
	}

	cmd := exec.Command("wg", "show", conf.WgInterface, "dump")
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	if string(out) == "" {
		fmt.Printf("error: 'wg show %s dump' didn't return parsable result!\n", conf.WgInterface)
		return
	}

	peersDump := strings.Split(strings.TrimSpace(string(out)), "\n")[1:]

	currentTimeS := time.Now().Unix()

	for _, peerDump := range peersDump {
		peerDumpSplit := strings.Fields(peerDump)
		if len(peerDumpSplit) < 5 {
			fmt.Printf("error: 'wg show %s dump' didn't return fully parsable result!\n", conf.WgInterface)
			continue
		}
		peerIp := peerDumpSplit[3]
		peerLatestHandshakeS, err := strconv.ParseInt(peerDumpSplit[4], 10, 64)
		if err != nil {
			fmt.Println("error:", err)
			return
		}

		if peerLatestHandshakeS != 0 && currentTimeS-peerLatestHandshakeS < conf.KeepAliveSeconds {
			newPeerIpMap[peerIp] = peerLatestHandshakeS
		}
	}

	for peerIp := range newPeerIpMap {
		_, ok1 := lastPeerIpMap[peerIp]
		val, _ := newPeerIpMap[peerIp]

		if ok1 == false {
			fmt.Printf("adding %s (last handshake %d seconds ago)\n", peerIp, currentTimeS-val)

			for _, addIpCmd := range conf.AddIpCmds {
				formattedCmd := strings.ReplaceAll(addIpCmd, "<ip>", peerIp)
				fmt.Println("running", formattedCmd)
				cmd = exec.Command("bash", "-c", formattedCmd)
				out, err = cmd.CombinedOutput()
				if err != nil {
					fmt.Println("error:", err)
				}
			}
		}
	}
	for peerIp := range lastPeerIpMap {
		val, ok := newPeerIpMap[peerIp]
		if ok == false {
			fmt.Println("removing", peerIp)

			for _, removeIpCmd := range conf.RemoveIpCmds {
				formattedCmd := strings.ReplaceAll(removeIpCmd, "<ip>", peerIp)
				fmt.Println("running ", formattedCmd)
				cmd := exec.Command("bash", "-c", formattedCmd)
				out, err = cmd.CombinedOutput()
				if err != nil {
					fmt.Println("error:", err)
				}
			}
		} else {
			fmt.Printf("persisting %s (last handshake %d seconds ago)\n", peerIp, currentTimeS-val)
		}
	}

	data, _ := json.Marshal(slices.Collect(maps.Keys(newPeerIpMap)))
	err = os.WriteFile(conf.DataPath, data, 0o644)
	if err != nil {
		fmt.Println("error:", err)
	}
}

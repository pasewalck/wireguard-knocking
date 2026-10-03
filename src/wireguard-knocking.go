package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"regexp"
	"strings"

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
	DataPath     string   `toml:"data_path"`
	WgInterface  string   `toml:"wg_interface"`
	RemoveIpCmds []string `toml:"remove_ip_cmds"`
	AddIpCmds    []string `toml:"add_ip_cmds"`
}

func main() {

	configPath := "/etc/wireguard-knocking/config.toml"

	exists, _ := fileExists(configPath)

	if exists == false {
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

	var items []string

	exists, _ = fileExists(conf.DataPath)

	if exists == true {
		b, _ := os.ReadFile(conf.DataPath)
		if err := json.Unmarshal(b, &items); err != nil {
			fmt.Println("error:", err)
		}
	}

	cmd := exec.Command("bash", "-c", fmt.Sprintf("wg show %s endpoints", conf.WgInterface))

	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("error:", err)
	}
	re := regexp.MustCompile(`\d{1,3}.\d{1,3}.\d{1,3}.\d{1,3}`)
	newItems := re.FindAllString(string(out), -1)

	itemsMap := map[string]int{}
	newItemsMap := map[string]int{}

	for _, it := range newItems {
		newItemsMap[it] = 1
	}
	for _, it := range items {
		itemsMap[it] = 1
	}

	for _, it := range newItems {
		_, ok := itemsMap[it]
		if ok == false {
			fmt.Println("adding", it)

			for _, addIpCmd := range conf.AddIpCmds {
				cmd := exec.Command("bash", "-c", strings.ReplaceAll(addIpCmd, "<ip>", it))
				out, err = cmd.CombinedOutput()
				if err != nil {
					fmt.Println("error:", err)
				}
			}
		}
	}
	for _, it := range items {
		_, ok := newItemsMap[it]
		if ok == false {
			fmt.Println("removing", it)

			for _, removeIpCmd := range conf.RemoveIpCmds {
				cmd := exec.Command("bash", "-c", strings.ReplaceAll(removeIpCmd, "<ip>", it))
				out, err = cmd.CombinedOutput()
				if err != nil {
					fmt.Println("error:", err)
				}
			}
		}
	}

	data, _ := json.Marshal(newItems)
	os.WriteFile(conf.DataPath, data, 0o644)

}

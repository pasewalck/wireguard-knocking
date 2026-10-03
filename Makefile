SERVICE = ./system/wireguard-knocking.service
TIMER = ./system/wireguard-knocking.timer
CONFIG = default.config.toml

SERVICE_TARGET = /etc/systemd/system/wireguard-knocking.service
TIMER_TARGET = /etc/systemd/system/wireguard-knocking.timer
CONFIG_TARGET = /etc/wireguard-knocking/config.toml
BINARY_TARGET = /usr/local/bin/wireguard-knocking

.PHONY: all
all: install

.PHONY: install
install:
	@echo "Installing ..."
	@sudo mkdir -p /etc/wireguard-knocking
	@sudo install -m 644 $(SERVICE) $(SERVICE_TARGET)
	@sudo install -m 644 $(TIMER) $(TIMER_TARGET)
	@sudo install -m 644 $(CONFIG) $(CONFIG_TARGET)
	@sudo systemctl daemon-reload
	go -C src mod download
	go -C src build -o /tmp/wireguard-knocking .
	@sudo install -m 755 /tmp/wireguard-knocking $(BINARY_TARGET)
	@rm -fr /tmp/wireguard-knocking
	go -C src clean
	@echo "Installation complete!"

.PHONY: uninstall
uninstall:
	@echo "Uninstalling ..."
	@sudo systemctl disable --now wireguard-knocking.timer
	@sudo rm -f $(SERVICE_TARGET)
	@sudo rm -f $(TIMER_TARGET)
	@sudo rm -fr /etc/wireguard-knocking/
	@sudo rm -f $(BINARY_TARGET)
	@sudo systemctl daemon-reload
	@echo "Uninstallation complete!"

.PHONY: enable
enable:
	@sudo systemctl enable wireguard-knocking.timer

.PHONY: disable
disable:
	@sudo systemctl disable wireguard-knocking.timer

.PHONY: start
start:
	@sudo systemctl start wireguard-knocking.timer

.PHONY: stop
stop:
	@sudo systemctl stop wireguard-knocking.timer

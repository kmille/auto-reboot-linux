export STATE_FILE := /run/user/1001/i3/unlocked.txt
export TIMEOUT := 20s
export COMMAND := date >> /tmp/inactivity.txt
export DEBUG := 1

.PHONY: buildArchPackage clean

run:
	go run ./main.go

build:
	go build -o inactivityd ./main.go

buildInstallArchPackage:
	makepkg --syncdeps --install --force
	rm -rf inactivityd pkg src *.tar.zst auto-reboot-linux

clean:
	rm -rf inactivityd pkg src *.tar.zst auto-reboot-linux

install: build
	sudo install -m 755 inactivityd /usr/local/bin/
	sudo install -m 644 inactivityd.service /etc/systemd/system
	sudo systemctl daemon-reload

uninstall: stop disable
	sudo rm /usr/local/bin/inactivityd
	sudo rm /etc/systemd/system/inactivityd.service
	sudo systemctl daemon-reload

enable:
	sudo systemctl enable inactivityd

disable:
	sudo systemctl disable inactivityd

start:
	sudo systemctl start inactivityd

stop:
	sudo systemctl stop inactivityd

log:
	journalctl --boot --unit=inactivityd

restart:
	sudo systemctl restart inactivityd

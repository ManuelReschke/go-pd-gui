fmt: ## format the go source files
	go fmt ./...
.PHONY: fmt

setup:
	curl -sSL https://git.io/g-install | sh -s
	sudo g install latest
	sudo apt-get -y install gcc libgl1-mesa-dev xorg-dev
	go install github.com/fyne-io/fyne-cross@latest
	fyne-cross linux --pull
.PHONY: setup

build:
	#env GOOS=linux GOARCH=amd64 go build -o bin/linux/go-pd-gui
	#env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o bin/windows/go-pd-gui.exe
	fyne-cross windows -arch=amd64,386
	fyne-cross linux
	fyne-cross android -arch=arm,arm64
.PHONY: build
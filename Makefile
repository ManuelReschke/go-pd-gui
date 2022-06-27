fmt: ## format the go source files
	go fmt ./...
.PHONY: fmt

build:
	#env GOOS=linux GOARCH=amd64 go build -o bin/linux/go-pd-gui
	#env GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o bin/windows/go-pd-gui.exe
	fyne-cross windows -arch=amd64,386
	fyne-cross linux
	fyne-cross android -arch=arm,arm64
.PHONY: build
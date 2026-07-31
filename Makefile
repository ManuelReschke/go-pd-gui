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
	fyne-cross windows -arch=amd64,386
	fyne-cross linux -arch=amd64
	fyne-cross android -arch=arm,arm64
.PHONY: build

# Stage the same asset names as CI (for local inspection / dry-runs)
stage-release-assets: build
	mkdir -p release-assets
	cp fyne-cross/dist/windows-amd64/DRAINY.exe.zip release-assets/DRAINY_Windows_x64.zip
	cp fyne-cross/dist/windows-386/DRAINY.exe.zip   release-assets/DRAINY_Windows_x86.zip
	cp fyne-cross/dist/linux-amd64/DRAINY.tar.xz    release-assets/DRAINY_Linux_x64.tar.xz
	cp fyne-cross/dist/android-arm64/DRAINY.apk     release-assets/DRAINY_Android_arm64.apk
	cp fyne-cross/dist/android-arm/DRAINY.apk       release-assets/DRAINY_Android_arm.apk
	@ls -lah release-assets/
.PHONY: stage-release-assets
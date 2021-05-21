default: linux windows macos

client: linux-client linux-client-arm macos-client macos-client-arm windows-client
server: linux-server macos-server windows-server

linux: linux-client linux-client-arm linux-server

linux-server:
	mkdir -p bin/linux
	GOOS=linux GOARCH=amd64 go get -d -v -x ./cmd/askgod-server
	cd bin/linux ; GOOS=linux GOARCH=amd64 go build ../../cmd/askgod-server

linux-client:
	mkdir -p bin/linux
	GOOS=linux GOARCH=amd64 go get -d -v -x ./cmd/askgod
	cd bin/linux ; CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build ../../cmd/askgod

linux-client-arm:
	mkdir -p bin/linux-arm
	GOOS=linux GOARCH=arm64 go get -d -v -x ./cmd/askgod
	cd bin/linux-arm ; CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build ../../cmd/askgod

windows: windows-client windows-server

windows-server:
	mkdir -p bin/windows
	GOOS=windows GOARCH=amd64 go get -d -v -x ./cmd/askgod-server
	cd bin/windows ; GOOS=windows GOARCH=amd64 go build ../../cmd/askgod-server

windows-client:
	mkdir -p bin/windows
	GOOS=windows GOARCH=amd64 go get -d -v -x ./cmd/askgod
	cd bin/windows ; CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build ../../cmd/askgod

macos: macos-client macos-client-arm macos-server

macos-server:
	mkdir -p bin/macos
	GOOS=darwin GOARCH=amd64 go get -d -v -x ./cmd/askgod
	cd bin/macos ; GOOS=darwin GOARCH=amd64 go build ../../cmd/askgod-server

macos-client:
	mkdir -p bin/macos
	GOOS=darwin GOARCH=amd64 go get -d -v -x ./cmd/askgod
	cd bin/macos ; CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build ../../cmd/askgod

macos-client-arm:
	mkdir -p bin/macos-arm
	GOOS=darwin GOARCH=arm64 go get -d -v -x ./cmd/askgod
	cd bin/macos-arm ; CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build ../../cmd/askgod

check:
	go get -t -v -d -u ./...
	go mod tidy
	golint ./...
	go vet ./...
	go fmt ./...
	gofmt -s -w ./

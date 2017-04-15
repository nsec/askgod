default: linux windows macos

linux:
	mkdir -p bin/linux
	GOOS=linux GOARCH=amd64 go get -d -v -x ./cmd/askgod
	GOOS=linux GOARCH=amd64 go get -d -v -x ./cmd/askgod-server
	cd bin/linux ; GOOS=linux GOARCH=amd64 go build ../../cmd/askgod-server
	cd bin/linux ; GOOS=linux GOARCH=amd64 go build ../../cmd/askgod

windows:
	mkdir -p bin/windows
	GOOS=windows GOARCH=amd64 go get -d -v -x ./cmd/askgod
	GOOS=windows GOARCH=amd64 go get -d -v -x ./cmd/askgod-server
	cd bin/windows ; GOOS=windows GOARCH=amd64 go build ../../cmd/askgod-server
	cd bin/windows ; GOOS=windows GOARCH=amd64 go build ../../cmd/askgod

macos:
	mkdir -p bin/macos
	GOOS=macos GOARCH=amd64 go get -d -v -x ./cmd/askgod
	GOOS=macos GOARCH=amd64 go get -d -v -x ./cmd/askgod-server
	cd bin/macos ; GOOS=darwin GOARCH=amd64 go build ../../cmd/askgod-server
	cd bin/macos ; GOOS=darwin GOARCH=amd64 go build ../../cmd/askgod

check:
	golint ./...
	go vet ./...
	go fmt ./...
	gofmt -s -w ./

default:
	mkdir -p bin/linux bin/macos bin/windows
	cd bin/linux ; GOOS=linux GOARCH=amd64 go build ../../cmd/askgod-server
	cd bin/windows ; GOOS=windows GOARCH=amd64 go build ../../cmd/askgod-server
	cd bin/macos ; GOOS=darwin GOARCH=amd64 go build ../../cmd/askgod-server
	
	cd bin/linux ; GOOS=linux GOARCH=amd64 go build ../../cmd/askgod
	cd bin/windows ; GOOS=windows GOARCH=amd64 go build ../../cmd/askgod
	cd bin/macos ; GOOS=darwin GOARCH=amd64 go build ../../cmd/askgod

check:
	golint ./...
	go vet ./...
	go fmt ./...
	gofmt -s -w ./

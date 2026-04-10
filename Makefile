build-all:
	mkdir -p dist
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o dist/wapclip-windows.exe ./cmd/
	GOOS=darwin  GOARCH=arm64 go build -ldflags="-s -w" -o dist/wapclip-mac-arm    ./cmd/
	GOOS=darwin  GOARCH=amd64 go build -ldflags="-s -w" -o dist/wapclip-mac-intel  ./cmd/
	GOOS=linux   GOARCH=amd64 go build -ldflags="-s -w" -o dist/wapclip-linux      ./cmd/

run-init:
	go run ./cmd/ init

run-daemon:
	go run ./cmd/ daemon

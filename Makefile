.PHONY: all

all: windows mac linux

windows:
	mkdir -p win
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/trixie.exe
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/trixie-srvcfg.exe
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/trixie-srvdb.exe
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/trixie-vmware.exe
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/trixie-fix.exe

mac:
	mkdir -p mac
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/trixie
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/trixie-srvcfg
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/trixie-srvdb
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/trixie-vmware
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/trixie-fix

linux:
	mkdir -p linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/trixie
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/trixie-srvcfg
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/trixie-srvdb
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/trixie-vmware
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/trixie-fix

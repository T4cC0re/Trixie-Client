.PHONY: all win mac linux clean

all: clean win mac linux

clean:
	rm -rf linux/ mac/ win/

win: clean
	mkdir -p win
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/trixie.exe

mac: clean
	mkdir -p mac
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/trixie

linux: clean
	mkdir -p linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/trixie

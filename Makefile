.PHONY: all win mac linux clean

all: clean win mac linux

clean:
	rm -rf linux/ mac/ win/

win: clean
	mkdir -p win
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/trixie.exe
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/tsrvcfg.exe
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/tsrvdb.exe
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/tvmware.exe
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o win/tfix.exe

mac: clean
	mkdir -p mac
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/trixie
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/tsrvcfg
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/tsrvdb
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/tvmware
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w" -o mac/tfix

linux: clean
	mkdir -p linux
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/trixie
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/tsrvcfg
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/tsrvdb
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/tvmware
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o linux/tfix

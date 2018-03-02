.PHONY: all win mac linux clean linux_trace mac_trace win_trace

all: clean win mac linux linux_trace mac_trace win_trace

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

linux_trace: clean
	mkdir -p linux
	GOOS=linux GOARCH=amd64 go build -i -tags trace -o linux/trixie_trace

mac_trace: clean
	mkdir -p mac
	GOOS=darwin GOARCH=amd64 go build -i -tags trace -o mac/trixie_trace

win_trace: clean
	mkdir -p win
	GOOS=windows GOARCH=amd64 go build -i -tags trace -o win/trixie_trace.exe

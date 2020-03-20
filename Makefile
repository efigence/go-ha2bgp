# generate version number
version=$(shell git describe --tags --long --always|sed 's/^v//')
binfile=ha2bgp

all:
	go build -ldflags "-X main.version=$(version)"  -o $(binfile)
	-@go fmt

static:
	go build -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).static

arm:
	GOARCH=arm go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).arm
	GOARCH=arm64 go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).arm64
clean:
	rm -rf vendor
	rm -rf _vendor
version:
	@echo $(version)

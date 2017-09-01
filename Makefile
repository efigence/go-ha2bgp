# generate version number
version=$(shell git describe --tags --long --always|sed 's/^v//')
binfile=ha2bgp

all: vendor | glide.lock
	go build -ldflags "-X main.version=$(version)"  -o $(binfile)
	-@go fmt

static: glide.lock vendor
	go build -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).static

arm:
	GOARCH=arm go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).arm
	GOARCH=arm64 go build  -ldflags "-X main.version=$(version) -extldflags \"-static\"" -o $(binfile).arm64
clean:
	rm -rf vendor
	rm -rf _vendor
vendor: glide.lock
	glide install && touch vendor
glide.lock: glide.yaml
	glide update && touch glide.lock
glide.yaml:
version:
	@echo $(version)

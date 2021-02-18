default: build

test:
	go test -v ./...

test_integration:
	PACKER_ACC=1 go test -count 1 -v ./...  -timeout=120m

build:
	go build -v

install: build
	mkdir -p ~/.packer.d/plugins
	install ./upcloud-packer ~/.packer.d/plugins/packer-builder-upcloud

.PHONY: default test test_integration build install

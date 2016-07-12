package main

import (
	"github.com/jalle19/packer-builder-upcloud/builder/upcloud"
	"github.com/mitchellh/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	server.RegisterBuilder(new(upcloud.Builder))
	server.Serve()
}

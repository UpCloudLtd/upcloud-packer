package main

import (
	"github.com/UpCloudLtd/upcloud-packer/builder/upcloud"
	"github.com/hashicorp/packer/packer/plugin"
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		panic(err)
	}
	if err := server.RegisterBuilder(new(upcloud.Builder)); err != nil {
		panic(err)
	}
	server.Serve()
}

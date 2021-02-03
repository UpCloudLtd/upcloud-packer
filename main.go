package main

import (
	"fmt"
	"os"

	upcloud "github.com/UpCloudLtd/upcloud-packer/builder/upcloud"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/version"
)

var (
	Version           = "4.0.0"
	VersionPrerelease = "dev"
	PluginVersion     = version.InitializePluginVersion(Version, VersionPrerelease)
)

func main() {
	server, err := plugin.Server()
	if err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Error starting plugin server: %s", err))
		os.Exit(1)
	}

	if err := server.RegisterBuilder(new(upcloud.Builder)); err != nil {
		fmt.Fprintln(os.Stderr, fmt.Sprintf("Error register builder: %s", err))
		os.Exit(1)
	}
	server.Serve()
}

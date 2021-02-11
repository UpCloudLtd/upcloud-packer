package upcloud

import (
	"github.com/hashicorp/packer-plugin-sdk/multistep"
)

// sshHostCallback retrieves the public IPv4 address of the server
func sshHostCallback(state multistep.StateBag) (string, error) {
	return state.Get("server_ip").(string), nil
}

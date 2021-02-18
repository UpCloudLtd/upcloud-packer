package upcloud

import (
	"fmt"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

// wraps error logic
func StepHaltWithError(state multistep.StateBag, err error) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	state.Put("error", err)
	ui.Error(err.Error())
	return multistep.ActionHalt
}

// parse public ip from server details
func GetServerIp(details *upcloud.ServerDetails) (string, error) {
	for _, ipAddress := range details.IPAddresses {
		if ipAddress.Access == upcloud.IPAddressAccessPublic && ipAddress.Family == upcloud.IPAddressFamilyIPv4 {
			return ipAddress.Address, nil
		}
	}
	return "", fmt.Errorf("Unable to find the public IPv4 address of the server")
}

func GetNowString() string {
	return time.Now().Format("20060102-150405")
}

// SshHostCallback retrieves the public IPv4 address of the server
func SshHostCallback(state multistep.StateBag) (string, error) {
	return state.Get("server_ip").(string), nil
}

// for config type convertion
type NetworkInterface struct {
	IPAddresses []IPAddress `mapstructure:"ip_addresses"`
	Type        string      `mapstructure:"type"`
	Network     string      `mapstructure:"network,omitempty"`
}

type IPAddress struct {
	Family  string `mapstructure:"family"`
	Address string `mapstructure:"address,omitempty"`
}

func ConvertNetworkTypes(rawNetworking []NetworkInterface) []request.CreateServerInterface {
	networking := []request.CreateServerInterface{}
	for _, iface := range rawNetworking {
		ips := []request.CreateServerIPAddress{}
		for _, ip := range iface.IPAddresses {
			ips = append(ips, request.CreateServerIPAddress{Family: ip.Family, Address: ip.Address})
		}
		networking = append(networking, request.CreateServerInterface{
			IPAddresses: ips,
			Type:        iface.Type,
			Network:     iface.Network,
		})
	}
	return networking
}

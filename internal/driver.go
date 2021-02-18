package upcloud

import (
	"fmt"
	"strings"
	"time"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/request"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
)

const (
	DefaultPlan = "1xCPU-2GB"
)

type (
	Driver interface {
		CreateServer(string, string, string, string, int) (*upcloud.ServerDetails, error)
		DeleteServer(string) error
		StopServer(string) error
		GetStorage(string, string) (*upcloud.Storage, error)
		GetServerStorage(string) (*upcloud.ServerStorageDevice, error)
		CloneStorage(string, string, string) (*upcloud.Storage, error)
		CreateTemplate(string, string) (*upcloud.Storage, error)
		DeleteTemplate(string) error
	}

	driver struct {
		svc    *service.Service
		config *DriverConfig
	}

	DriverConfig struct {
		Username    string
		Password    string
		Timeout     time.Duration
		SSHUsername string
	}
)

func NewDriver(c *DriverConfig) Driver {
	client := client.New(c.Username, c.Password)
	svc := service.New(client)
	return &driver{
		svc:    svc,
		config: c,
	}
}

func (d *driver) CreateServer(storageUuid, zone, prefix, sshKeyPublic string, storageSize int) (*upcloud.ServerDetails, error) {
	// Create server
	request := d.prepareCreateRequest(storageUuid, zone, prefix, sshKeyPublic, storageSize)
	response, err := d.svc.CreateServer(request)
	if err != nil {
		return nil, fmt.Errorf("Error creating server: %s", err)
	}

	// Wait for server to start
	err = d.waitDesiredState(response.UUID, upcloud.ServerStateStarted)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (d *driver) DeleteServer(serverUuid string) error {
	return d.svc.DeleteServerAndStorages(&request.DeleteServerAndStoragesRequest{
		UUID: serverUuid,
	})
}

func (d *driver) StopServer(serverUuid string) error {
	// Ensure the instance is not in maintenance state
	err := d.waitUndesiredState(serverUuid, upcloud.ServerStateMaintenance)
	if err != nil {
		return err
	}

	// Check current server state and do nothing if already stopped
	response, err := d.getServerDetails(serverUuid)
	if err != nil {
		return err
	}

	if response.State == upcloud.ServerStateStopped {
		return nil
	}

	// Stop server
	_, err = d.svc.StopServer(&request.StopServerRequest{
		UUID: serverUuid,
	})
	if err != nil {
		return fmt.Errorf("Failed to stop server: %s", err)
	}

	// Wait for server to stop
	err = d.waitDesiredState(serverUuid, upcloud.ServerStateStopped)
	if err != nil {
		return err
	}
	return nil
}

func (d *driver) CreateTemplate(serverStorageUuid, prefix string) (*upcloud.Storage, error) {
	// create image
	templateTitle := fmt.Sprintf("%s-%s", prefix, GetNowString())
	response, err := d.svc.TemplatizeStorage(&request.TemplatizeStorageRequest{
		UUID:  serverStorageUuid,
		Title: templateTitle,
	})
	if err != nil {
		return nil, fmt.Errorf("Error creating image: %s", err)
	}
	return d.waitStorageOnline(response.UUID)
}

func (d *driver) waitStorageOnline(storageUuid string) (*upcloud.Storage, error) {
	template, err := d.svc.WaitForStorageState(&request.WaitForStorageStateRequest{
		UUID:         storageUuid,
		DesiredState: upcloud.StorageStateOnline,
		Timeout:      d.config.Timeout,
	})
	if err != nil {
		return nil, fmt.Errorf("Error while waiting for storage to change state to 'online': %s", err)
	}
	return &template.Storage, nil
}

// fetch storage by uuid or name
func (d *driver) GetStorage(storageUuid, storageName string) (*upcloud.Storage, error) {
	if storageUuid != "" {
		storage, err := d.getStorageByUuid(storageUuid)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving storage by uuid %q: %s", storageUuid, err)
		}
		return storage, nil
	}

	if storageName != "" {
		storage, err := d.getStorageByName(storageName)
		if err != nil {
			return nil, fmt.Errorf("Error retrieving storage by name %q: %s", storageName, err)
		}
		return storage, nil

	}
	return nil, fmt.Errorf("Error retrieving storage")
}

func (d *driver) DeleteTemplate(templateUuid string) error {
	return d.svc.DeleteStorage(&request.DeleteStorageRequest{
		UUID: templateUuid,
	})
}

func (d *driver) CloneStorage(storageUuid string, zone string, title string) (*upcloud.Storage, error) {
	response, err := d.svc.CloneStorage(&request.CloneStorageRequest{
		UUID:  storageUuid,
		Zone:  zone,
		Title: title,
	})
	if err != nil {
		return nil, err
	}
	return d.waitStorageOnline(response.UUID)
}

func (d *driver) getStorageByUuid(storageUuid string) (*upcloud.Storage, error) {
	response, err := d.svc.GetStorageDetails(&request.GetStorageDetailsRequest{
		UUID: storageUuid,
	})

	if err != nil {
		return nil, fmt.Errorf("Error fetching storages: %s", err)
	}
	return &response.Storage, nil
}

func (d *driver) getStorageByName(storageName string) (*upcloud.Storage, error) {
	response, err := d.svc.GetStorages(&request.GetStoragesRequest{
		Type: upcloud.StorageTypeTemplate,
	})

	if err != nil {
		return nil, fmt.Errorf("Error fetching storages: %s", err)
	}

	var found bool
	var storage upcloud.Storage
	for _, s := range response.Storages {
		if strings.Contains(strings.ToLower(s.Title), strings.ToLower(storageName)) {
			found = true
			storage = s
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("Failed to find storage by name %q", storageName)
	}
	return &storage, nil
}

func (d *driver) waitDesiredState(serverUuid string, state string) error {
	request := &request.WaitForServerStateRequest{
		UUID:         serverUuid,
		DesiredState: state,
		Timeout:      d.config.Timeout,
	}
	if _, err := d.svc.WaitForServerState(request); err != nil {
		return fmt.Errorf("Error while waiting for server to change state to %q: %s", state, err)
	}
	return nil
}

func (d *driver) waitUndesiredState(serverUuid string, state string) error {
	request := &request.WaitForServerStateRequest{
		UUID:           serverUuid,
		UndesiredState: state,
		Timeout:        d.config.Timeout,
	}
	if _, err := d.svc.WaitForServerState(request); err != nil {
		return fmt.Errorf("Error while waiting for server to change state from %q: %s", state, err)
	}
	return nil
}

func (d *driver) getServerDetails(serverUuid string) (*upcloud.ServerDetails, error) {
	response, err := d.svc.GetServerDetails(&request.GetServerDetailsRequest{
		UUID: serverUuid,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to get details for server: %s", err)
	}
	return response, nil
}

func (d *driver) GetServerStorage(serverUuid string) (*upcloud.ServerStorageDevice, error) {
	details, err := d.getServerDetails(serverUuid)
	if err != nil {
		return nil, err
	}

	var found bool
	var storage upcloud.ServerStorageDevice
	for _, s := range details.StorageDevices {
		if s.Type == upcloud.StorageTypeDisk {
			found = true
			storage = s
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("Failed to find storage type disk for server %q", serverUuid)
	}
	return &storage, nil
}

func (d *driver) prepareCreateRequest(storageUuid, zone, prefix, sshKeyPublic string, storageSize int) *request.CreateServerRequest {
	title := fmt.Sprintf("packer-%s-%s", prefix, GetNowString())
	hostname := prefix
	titleDisk := fmt.Sprintf("%s-disk1", title)

	request := request.CreateServerRequest{
		Title:            title,
		Hostname:         hostname,
		Zone:             zone,
		PasswordDelivery: request.PasswordDeliveryNone,
		Plan:             DefaultPlan,
		StorageDevices: []request.CreateServerStorageDevice{
			{
				Action:  request.CreateServerStorageDeviceActionClone,
				Storage: storageUuid,
				Title:   titleDisk,
				Size:    storageSize,
				Tier:    upcloud.StorageTierMaxIOPS,
			},
		},
		Networking: &request.CreateServerNetworking{
			Interfaces: []request.CreateServerInterface{
				{
					IPAddresses: []request.CreateServerIPAddress{
						{
							Family: upcloud.IPAddressFamilyIPv4,
						},
					},
					Type: upcloud.IPAddressAccessPublic,
				},
			},
		},
		LoginUser: &request.LoginUser{
			CreatePassword: "no",
		},
	}
	if sshKeyPublic != "" {
		request.LoginUser.Username = d.config.SSHUsername
		request.LoginUser.SSHKeys = []string{sshKeyPublic}
	}
	return &request
}

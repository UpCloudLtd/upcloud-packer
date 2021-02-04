package upcloud

import (
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/client"
	"github.com/UpCloudLtd/upcloud-go-api/upcloud/service"
)

func getApiService(c *Config) *service.Service {
	client := client.New(c.Username, c.Password)
	return service.New(client)
}

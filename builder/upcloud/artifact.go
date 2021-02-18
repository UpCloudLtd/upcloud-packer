package upcloud

import (
	"fmt"
	"strings"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	internal "github.com/UpCloudLtd/upcloud-packer/internal"
)

// packersdk.Artifact implementation
type Artifact struct {
	config    *Config
	driver    internal.Driver
	Templates []*upcloud.Storage

	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (*Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	return []string{}
}

func (a *Artifact) Id() string {
	result := []string{}
	for _, t := range a.Templates {
		result = append(result, t.UUID)
	}
	return strings.Join(result, ",")
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Storage template created, UUID: %s", a.Id())
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	for _, t := range a.Templates {
		err := a.driver.DeleteTemplate(t.UUID)
		if err != nil {
			return err
		}
	}
	return nil
}

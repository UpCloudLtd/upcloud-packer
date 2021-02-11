package upcloud

import (
	"fmt"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
)

// packersdk.Artifact implementation
type Artifact struct {
	config   *Config
	driver   Driver
	Template *upcloud.Storage

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
	return a.Template.UUID
}

func (a *Artifact) String() string {
	return fmt.Sprintf("Storage template created, UUID: %q, Title: %q", a.Template.UUID, a.Template.Title)
}

func (a *Artifact) State(name string) interface{} {
	return a.StateData[name]
}

func (a *Artifact) Destroy() error {
	return a.driver.DeleteTemplate(a.Template.UUID)
}

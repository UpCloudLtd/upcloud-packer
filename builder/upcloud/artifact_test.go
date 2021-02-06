package upcloud

import (
	"testing"

	"github.com/UpCloudLtd/upcloud-go-api/upcloud"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
)

func TestArtifact_impl(t *testing.T) {
	var _ packersdk.Artifact = new(Artifact)
}

func TestArtifact_Id(t *testing.T) {
	expected := "some-uuid"
	template := &upcloud.Storage{
		UUID: expected,
	}
	a := &Artifact{Template: template}
	result := a.Id()

	if result != expected {
		t.Errorf("Expected: %q, got: %q", expected, result)
	}
}

func TestArtifact_String(t *testing.T) {
	expected := `Storage template created, UUID: "some-uuid", Title: "some-title"`
	template := &upcloud.Storage{
		UUID:  "some-uuid",
		Title: "some-title",
	}
	a := &Artifact{Template: template}
	result := a.String()

	if result != expected {
		t.Errorf("Expected: %q, got: %q", expected, result)
	}
}

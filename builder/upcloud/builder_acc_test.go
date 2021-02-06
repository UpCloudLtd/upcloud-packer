package upcloud

import (
	"os"
	"testing"

	builderT "github.com/hashicorp/packer-plugin-sdk/acctest"
)

// Run tests: PACKER_ACC=1 go test -count 1 -v ./...  -timeout=120m
func TestBuilderAcc_templateUuid(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccTemplateUuid,
	})
}

func TestBuilderAcc_templateName(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAccTemplateName,
	})
}

func TestBuilderAcc_backwardCompatibility(t *testing.T) {
	builderT.Test(t, builderT.TestCase{
		PreCheck: func() { testAccPreCheck(t) },
		Builder:  &Builder{},
		Template: testBuilderAcc_backwardCompatibility,
	})
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("UPCLOUD_API_USER"); v == "" {
		t.Fatal("UPCLOUD_API_USER must be set for acceptance tests")
	}
	if v := os.Getenv("UPCLOUD_API_PASSWORD"); v == "" {
		t.Fatal("UPCLOUD_API_PASSWORD must be set for acceptance tests")
	}
}

const testBuilderAccTemplateUuid = `
{
	"builders": [{
            "type": "test",
            "zone": "nl-ams1",
            "template_uuid": "01000000-0000-4000-8000-000050010400",
            "ssh_username": "root",
            "image_name": "test-builder",
            "storage_size": "20"
	}]
}
`

const testBuilderAccTemplateName = `
{
	"builders": [{
            "type": "test",
            "zone": "nl-ams1",
            "template_name": "ubuntu server 20.04",
            "ssh_username": "root",
            "image_name": "test-builder",
            "storage_size": "20"
	}]
}
`

const testBuilderAcc_backwardCompatibility = `
{
	"builders": [{
            "type": "test",
            "zone": "nl-ams1",
            "storage_uuid": "01000000-0000-4000-8000-000030060200",
            "ssh_username": "root",
            "template_prefix": "test-builder",
            "storage_size": "20"
	}]
}
`

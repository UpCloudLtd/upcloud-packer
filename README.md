# packer-builder-upcloud

This is a Packer builder which can be used to generate storage templates on UpCloud.

## Installation

You can download pre-built binaries of the plugin from 
https://packer-builder-upcloud-builds.s3.eu-central-1.amazonaws.com/index.html. Once you've downloaded the binary for 
your operating system and architecture, rename the file to `packer-builder-upcloud`, move it to `~/.packer.d/plugins` 
and make it executable.

To install from source, use something like this:

```
go get github.com/jalle19/packer-builder-upcloud
cd $GOPATH/src/github.com/jalle19/packer-builder-upcloud
go install
cp $GOPATH/bin/packer-builder-upcloud ~/.packer.d/plugins
```

If you attempt to build it from source you'll most likely bump into some vendored dependency versioning issues. You can 
usually solve these by removing the vendored packages Go complains about from `$GOPATH/src/github.com/packer`.

## Usage

The builder will automatically generate a temporary SSH key pair for the `root` user which is used for provisioning. 
This means that if you don't provision a user during the process you will not be able to gain access to your server.

Here is a sample template (you can find this one and a few others in the `examples/` directory). It reads your UpCloud 
API credentials from the environment and creates an Ubuntu 14.04 server in the `fi-hel1` region.

```json
{
  "variables": {
    "UPCLOUD_USERNAME": "{{ env `UPCLOUD_GO_SDK_TEST_USER` }}",
    "UPCLOUD_PASSWORD": "{{ env `UPCLOUD_GO_SDK_TEST_PASSWORD` }}"
  },
  "builders": [
    {
      "type": "upcloud",
      "username": "{{ user `UPCLOUD_USERNAME` }}",
      "password": "{{ user `UPCLOUD_PASSWORD` }}",
      "zone": "fi-hel1",
      "storage_uuid": "01000000-0000-4000-8000-000030040200"
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": ["apt-get update"]
    }
  ]
}
```

If everything goes according to plan, you should see something like this:

```
$ packer build examples/basic_plan.json 
upcloud output will be in this color.

==> upcloud: Creating temporary SSH key ...
==> upcloud: Creating server "packer-builder-upcloud-1468327456" ...
==> upcloud: Waiting for server "packer-builder-upcloud-1468327456" to enter the "started" state ...
==> upcloud: Server "packer-builder-upcloud-1468327456" is now in "started" state
==> upcloud: Waiting for SSH to become available...
==> upcloud: Connected to SSH!
==> upcloud: Provisioning with shell script: scripts/update.sh
...
==> upcloud: Stopping server "packer-builder-upcloud-1468327456" ...
==> upcloud: Waiting for server "packer-builder-upcloud-1468327456" to enter the "stopped" state ...
==> upcloud: Server "packer-builder-upcloud-1468327456" is now in "stopped" state
==> upcloud: Templatizing storage device "packer-builder-upcloud-1468327456-disk1" ...
==> upcloud: Waiting for storage "packer-builder-upcloud-1468327456-disk1-template-1468327515" to enter the "online" state
==> upcloud: Waiting for server "packer-builder-upcloud-1468327456" to exit the "maintenance" state ...
==> upcloud: Deleting server "packer-builder-upcloud-1468327456" ...
Build 'upcloud' finished.

==> Builds finished. The artifacts of successful builds are:
--> upcloud: Private template (UUID: 01875f67-4eb5-4d90-982c-d7a164646fcb, Title: packer-builder-upcloud-1468327456-disk1-template-1468327515, Zone: fi-hel1)
```

### Configuration reference

This section describes the available configuration options for the builder. Please not that since the purpose of the 
builder is to build a storage template that can be used as a source when cloning new servers, the server used when 
building the template is irrelevant and thus not configurable. 

#### Required values

* `username` (string) The username to use when interfacing with the API
* `password` (string) The password to use when interfacing with the API
* `zone` (string) The zone in which the server and template should be created (e.g. `fi-hel1`)
* `storage_uuid` (string) The UUID of the storage you want to use as a template when creating the server

#### Optional values

* `storage_size` (int) The storage size in gigabytes. Defaults to `30`. Changing this value is useful if you aim to build 
a template for larger server configurations where the disk size is larger than 30 GB.
* `state_timeout_duration` (string) The amount of time to wait for resource state changes. Defaults to `5m`.

## License

This project is distributed under the [MIT License](https://opensource.org/licenses/MIT), see LICENSE.txt for more 
information.

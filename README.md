# UpCloud Packer builder

[![Build Status](https://travis-ci.org/UpCloudLtd/upcloud-packer.svg?branch=master)](https://travis-ci.org/UpCloudLtd/upcloud-packer)

This is a Packer builder which can be used to generate storage templates on UpCloud. It uses the [UpCloud Go SDK](https://github.com/UpCloudLtd/upcloud-go-sdk) to interface with the UpCloud API.

## Installation

### Pre-built binaries

You can download pre-built binaries of the plugin from the [GitHub releases page](https://github.com/UpCloudLtd/upcloud-packer/releases). Just download the archive for your operating system and architecture, unpack it, then place the binary in `~/.packer.d/plugins`. Make sure the file is executable.

### Installing from source

#### Prerequisites

You will need to have the [Go](https://golang.org/) programming language, the [Glide](https://github.com/Masterminds/glide) package manager, and the [Packer](https://www.packer.io/) itself installed. You can find instructions to install each of the prerequisites at their documentation.

#### Building and installing

Run the following commands:

```
go get github.com/UpCloudLtd/upcloud-packer
cd $GOPATH/src/github.com/UpCloudLtd/upcloud-packer
glide install --strip-vendor
go build
cp upcloud-packer ~/.packer.d/plugins/packer-builder-upcloud
```

## Usage

The builder will automatically generate a temporary SSH key pair for the `root` user which is used for provisioning. 
This means that if you do not provision a user during the process you will not be able to gain access to your server.

Here is a sample template, which you can also find in the `examples/` directory. It reads your UpCloud API credentials from the environment and creates an Ubuntu 14.04 server in the `fi-hel1` region.

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

If everything goes according to plan, you should see something like the example output below.

```
$ packer build examples/basic_plan.json 
upcloud output will be in this color.

==> upcloud: Creating temporary SSH key ...
==> upcloud: Creating server "packer-builder-upcloud-1468327456" ...
==> upcloud: Waiting for server "packer-builder-upcloud-1468327456" to enter the "started" state ...
==> upcloud: Server "packer-builder-upcloud-1468327456" is now in "started" state
==> upcloud: Waiting for SSH to become available...
==> upcloud: Connected to SSH!
==> upcloud: Provisioning with shell script: /tmp/packer-shell667235875
    upcloud: Ign http://fi.archive.ubuntu.com trusty InRelease
    upcloud: Get:1 http://fi.archive.ubuntu.com trusty-updates InRelease [65.9 kB]
    upcloud: Get:2 http://fi.archive.ubuntu.com trusty-backports InRelease [65.9 kB]
...
    upcloud: Fetched 33.0 MB in 9s (3,622 kB/s)
    upcloud: Reading package lists...
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

## Configuration reference

This section describes the available configuration options for the builder. Please note that since the purpose of the 
builder is to build a storage template that can be used as a source when cloning new servers, the server used when 
building the template is fairly irrelevant and thus not configurable. 

### Required values

* `username` (string) The username to use when interfacing with the UpCloud API
* `password` (string) The password to use when interfacing with the UpCloud API
* `zone` (string) The zone in which the server and template should be created (e.g. `fi-hel1`)
* `storage_uuid` (string) The UUID of the storage you want to use as a template when creating the server

### Optional values

* `storage_size` (int) The storage size in gigabytes. Defaults to `30`. Changing this value is useful if you aim to build a template for larger server configurations where the preconfigured server disk is larger than 30 GB. Even the operation system disk can be later extended if needed.
* `state_timeout_duration` (string) The amount of time to wait for resource state changes. Defaults to `5m`.
* `template_prefix` (string) The prefix to use for the generated template title. Defaults to an empty string, meaning the prefix will be the storage title. You can use this option to easily differentiate between different templates.

## License

This project is distributed under the [MIT License](https://opensource.org/licenses/MIT), see LICENSE.txt for more 
information.

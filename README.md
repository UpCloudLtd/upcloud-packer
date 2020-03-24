# UpCloud Packer builder

[![Build Status](https://travis-ci.org/UpCloudLtd/upcloud-packer.svg?branch=master)](https://travis-ci.org/UpCloudLtd/upcloud-packer)

This is a builder plugin for Packer which can be used to generate storage templates on UpCloud. It utilises the [UpCloud Go API](https://github.com/UpCloudLtd/upcloud-go-api) to interface with the UpCloud API.

## Installation

### Pre-built binaries

You can download the pre-built binaries of the plugin from the [GitHub releases page](https://github.com/UpCloudLtd/upcloud-packer/releases). Just download the archive for your operating system and architecture, unpack it, and place the binary in the appropriate location, e.g. on Linux `~/.packer.d/plugins`. Make sure the file is executable, then install [Packer](https://www.packer.io/).

### Installing from source

#### Prerequisites

You will need to have the [Go](https://golang.org/) programming language and the [Packer](https://www.packer.io/) itself installed. You can find instructions to install each of the prerequisites at their documentation.

#### Building and installing

Run the following commands to download and install the plugin from the source.

```
git clone https://github.com/UpCloudLtd/upcloud-packer
cd upcloud-packer
go build
cp upcloud-packer ~/.packer.d/plugins/packer-builder-upcloud
```

## Usage

The builder will automatically generate a temporary SSH key pair for the `root` user which is used for provisioning. This means that if you do not provision a user during the process you will not be able to gain access to your server.

If you want to login to a server deployed with the template, you might want to include an SSH key to your `root` user by replacing the `<ssh-rsa_key>` in the below example with your public key.

Here is a sample template, which you can also find in the `examples/` directory. It reads your UpCloud API credentials from the environment variables and creates an Ubuntu 16.04 server in the `nl-ams1` region.

```json
{
  "variables": {
    "UPCLOUD_USERNAME": "{{ env `UPCLOUD_API_USER` }}",
    "UPCLOUD_PASSWORD": "{{ env `UPCLOUD_API_PASSWORD` }}"
  },
  "builders": [
    {
      "type": "upcloud",
      "username": "{{ user `UPCLOUD_USERNAME` }}",
      "password": "{{ user `UPCLOUD_PASSWORD` }}",
      "zone": "nl-ams1",
      "storage_uuid": "01000000-0000-4000-8000-000030060200"
    }
  ],
  "provisioners": [
    {
      "type": "shell",
      "inline": [
        "apt-get update",
        "apt-get upgrade -y",
        "echo '<ssh-rsa_key>' | tee /root/.ssh/authorized_keys"
      ]
    }
  ]
}
```

You will need to provide a username and a password with the access rights to the API functions to authenticate. We recommend setting up a subaccount with only the API privileges for security purposes. You can do this at your [UpCloud Control Panel](https://my.upcloud.com/account) in the My Account menu under the User Accounts tab.

Enter the API user credentials in your terminal with the following commands. Replace the `<API_username>` and `<API_password>` with your user details.

```json
export UPCLOUD_API_USER=<API_username>
export UPCLOUD_API_PASSWORD=<API_password>
```
Then run Packer using the example template with the command underneath.
```
packer build examples/basic_example.json
```
If everything goes according to plan, you should see something like the example output below.

```json
upcloud output will be in this color.

==> upcloud: Creating temporary SSH key ...
==> upcloud: Creating server "packer-builder-upcloud-1502364156" ...
==> upcloud: Waiting for server "packer-builder-upcloud-1502364156" to enter the "started" state ...
==> upcloud: Server "packer-builder-upcloud-1502364156" is now in "started" state
==> upcloud: Waiting for SSH to become available...
==> upcloud: Connected to SSH!
==> upcloud: Provisioning with shell script: /tmp/packer-shell888337462
	upcloud: Get:1 http://security.ubuntu.com/ubuntu xenial-security InRelease [102 kB]
    upcloud: Hit:2 http://fi.archive.ubuntu.com/ubuntu xenial InRelease
    upcloud: Get:3 http://fi.archive.ubuntu.com/ubuntu xenial-updates InRelease [102 kB]
...
==> upcloud: Stopping server "packer-builder-upcloud-1502364156" ...
==> upcloud: Waiting for server "packer-builder-upcloud-1502364156" to enter the "stopped" state ...
==> upcloud: Server "packer-builder-upcloud-1502364156" is now in "stopped" state
==> upcloud: Templatizing storage device "packer-builder-upcloud-1502364156-disk1" ...
==> upcloud: Waiting for storage "packer-builder-upcloud-1502364156-disk1-template-1502364398" to enter the "online" state
==> upcloud: Waiting for server "packer-builder-upcloud-1502364156" to exit the "maintenance" state ...
==> upcloud: Deleting server "packer-builder-upcloud-1502364156" ...
==> upcloud: Deleting disk "packer-builder-upcloud-1502364156-disk1" ...
Build 'upcloud' finished.

==> Builds finished. The artifacts of successful builds are:
--> upcloud: Private template (UUID: 013399d9-5308-46b1-9f89-bcbe0c4b983d, Title: packer-builder-upcloud-1502364156-disk1-template-1502364398, Zone: nl-ams1)
```

## Configuration reference

This section describes the available configuration options for the builder. Please note that the purpose of the builder is to create a storage template that can be used as a source for deploying new servers, therefore the temporary server used for building the template is not configurable.

### Required values

* `username` (string) The username to use when interfacing with the UpCloud API.
* `password` (string) The password to use when interfacing with the UpCloud API.
* `zone` (string) The zone in which the server and template should be created (e.g. `nl-ams1`).
* `storage_uuid` (string) The UUID of the storage you want to use as a template when creating the server.

### Optional values

* `storage_size` (int) The storage size in gigabytes. Defaults to `30`. Changing this value is useful if you aim to build a template for larger server configurations where the preconfigured server disk is larger than 30 GB. The operating system disk can also be later extended if needed.
* `state_timeout_duration` (string) The amount of time to wait for resource state changes. Defaults to `5m`.
* `template_prefix` (string) The prefix to use for the generated template title. Defaults to an empty string, meaning the prefix will be the storage title. You can use this option to easily differentiate between different templates.

## License

This project is distributed under the [MIT License](https://opensource.org/licenses/MIT), see LICENSE.txt for more information.
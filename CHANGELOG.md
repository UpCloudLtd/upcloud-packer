# Change log

## 4.1.0

* Added ability to provide ssh keys
* Added ability to setup network interfaces
* Added ability to create templates into multiple zones
* Changed default template size to 25Gb

## 4.0.0

* Switched to use `packer-plugin-sdk`
* Bumped up dependencies (`upcloud-go-api`, `golang.org/x/crypto` and etc)
* Encapsulated Upcloud API interaction in the driver module
* Added new `storage_name` config parameter

## 3.0.0

* Moved the project to UpCloud's GitHub organization
* Adapted to the fact that the UpCloud Go SDK was moved to UpCloud's GitHub organization
* Renamed the repository from `packer-builder-upcloud` to `upcloud-packer`. The released binaries still use the 
`packer-$TYPE-$NAME` naming scheme though.
* Bumped the Packer dependency to v1.0.0

## 2.0.5

* Fix some issues with the Travis releases
* Remove everything related to the old build host

## 2.0.0

* Fixed package name to use uppercase in imports (github.com/Jalle19 instead of github.com/jalle19)
* Have Travis CI build tags and upload them to GitHub

## 1.0.4

* Validate the specified storage device during `packer validate` (prevents accidentally trying to use CD-ROM devices 
as templates)

## 1.0.3

* Add `template_prefix` option to more easily distinguish between created templates
* Host the builds on an UpCloud server instead of S3
* Build the server hosting the pre-built binaries using this very tool

## 1.0.2

* Validate the specified API credentials during `packer validate`

## 1.0.1

* Add a page containing pre-built binaries for easier installation
* Don't accidentally remove the generated template during the cleanup stage
* Delete the disk used to make the template during cleanup

## 1.0.0

Initial release

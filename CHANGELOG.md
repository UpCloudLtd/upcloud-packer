# Change log

## 2.0.4

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

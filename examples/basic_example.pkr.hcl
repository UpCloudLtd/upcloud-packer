variable "username" {
  type = string
  default = "${env("UPCLOUD_API_USER")}"
}

variable "password" {
  type = string
  default = "${env("UPCLOUD_API_PASSWORD")}"
}

source "upcloud" "test" {
  username = "${var.username}"
  password = "${var.password}"
  zone = "nl-ams1"
  storage_name = "ubuntu server 20.04"
  template_prefix = "ubuntu-server"
}

build {
  sources = ["source.upcloud.test"]

  provisioner "shell" {
    inline = [
      "apt-get update",
      "apt-get upgrade -y",
      "echo '<ssh-rsa_key>' | tee /root/.ssh/authorized_keys"
    ]
  }
}

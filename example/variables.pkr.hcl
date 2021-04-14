variable "bastion_host" {
  type    = string
  default = ""
}
variable "bastion_user" {
  type    = string
  default = ""
}
variable "datacenter_name" {
  type    = string
  default = ""
}
variable "esxi_host" {
  type    = string
  default = ""
}
variable "esxi_password" {
  type    = string
  default = ""
}
variable "esxi_user" {
  type    = string
  default = ""
}
variable "vcenter_endpoint" {
  type    = string
  default = ""
}
variable "vcenter_password" {
  type    = string
  default = ""
}
variable "vcenter_user" {
  type    = string
  default = ""
}

variable "gateway_ip" {
  type    = string
  default = ""
}
variable "vm_ip" {
  type    = string
  default = ""
}
variable "alpine_password" {
  type = string
  default = "alpine"
}


locals {
  iso_url_alpine_312             = "http://dl-cdn.alpinelinux.org/alpine/v3.12/releases/x86_64/alpine-virt-3.12.0-x86_64.iso"
  iso_checksum_url_alpine_312    = "http://dl-cdn.alpinelinux.org/alpine/v3.12/releases/x86_64/alpine-virt-3.12.0-x86_64.iso.sha256"
  floppy_files_alpine_vsphere = [
    "./http/alpine-vsphere-answers",
    "./http/alpine-setup.sh"
  ]

  alpine_312_floppy_boot_command_vsphere = [
    "root<enter><wait1s>",
    "mount -t vfat /dev/fd0 /media/floppy<enter><wait1s>",
    "setup-alpine -f /media/floppy/alpine-vsphere-answers<enter><wait3s>",
    "${var.alpine_password}<enter>",
    "${var.alpine_password}<enter>",
    "<wait6s>",
    "y<enter>",
    "<wait12s>",
    "reboot<enter>",
    "<wait12s>",
    "root<enter>",
    "${var.alpine_password}<enter><wait>",
    "mount -t vfat /dev/fd0 /media/floppy<enter><wait>",
    "/media/floppy/alpine-setup.sh<enter>",
    "<wait55s>",
  ]
}



# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# source blocks are analogous to the "builders" in json templates. They are used
# in build blocks. A build block runs provisioners and post-processors on a
# source. Read the documentation for source blocks here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/source
source "vsphere-iso" "example" {
  CPUs            = 1
  RAM             = 1024
  RAM_reserve_all = true
  boot_command = ["<enter><wait><f6><wait><esc><wait>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs>", "/install/vmlinuz",
    " initrd=/install/initrd.gz", " priority=critical",
    " locale=en_US", " file=/media/preseed.cfg",
  "<enter>"]
  disk_controller_type = ["pvscsi"]
  floppy_files         = ["${path.root}/preseed.cfg"]
  guest_os_type        = "ubuntu64Guest"
  host                 = "esxi-01.example.com"
  insecure_connection  = true
  iso_paths            = ["[datastore1] ISO/ubuntu-16.04.3-server-amd64.iso"]
  network_adapters {
    network_card = "vmxnet3"
  }
  password     = "VMw@re1!"
  ssh_password = "VMw@re1!"
  ssh_username = "packer"
  storage {
    disk_size             = 32768
    disk_thin_provisioned = true
  }
  username       = "administrator@vsphere.local"
  vcenter_server = "vcenter.example.com"
  vm_name        = "example-ubuntu"
}

# a build block invokes sources and runs provisioning steps on them. The
# documentation for build blocks can be found here:
# https://www.packer.io/docs/templates/hcl_templates/blocks/build
build {
  sources = ["source.vsphere-iso.example"]

  provisioner "shell" {
    inline = ["ls /"]
  }
}

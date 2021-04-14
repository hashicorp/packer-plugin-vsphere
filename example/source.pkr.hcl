source "vsphere-iso" "example" {
  vcenter_server      = var.vcenter_endpoint
  username            = var.vcenter_user
  password            = var.vcenter_password
  host                = var.esxi_host
  insecure_connection = true

  datacenter = var.datacenter_name
  datastore  = "datastore1"

  ssh_username = "root"
  ssh_password = var.alpine_password

  CPUs            = 1
  RAM             = 512 * 2
  RAM_reserve_all = true

  guest_os_type = "otherLinux64Guest"
  floppy_files = local.floppy_files_alpine_vsphere

  network_adapters {
    network      = "VM Network"
    network_card = "vmxnet3"
  }

  storage {
    disk_size             = 32768
    disk_thin_provisioned = true
  }

  vm_name      = "alpine-3.12"
  iso_url      = local.iso_url_alpine_312
  iso_checksum = "file:${local.iso_checksum_url_alpine_312}"
  boot_command = local.alpine_312_floppy_boot_command_vsphere
  boot_wait    = "10s"
}
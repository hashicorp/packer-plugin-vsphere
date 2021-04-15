source "vsphere-iso" "example" {
  datacenter        = var.datacenter_name
  vcenter_server    = var.vcenter_endpoint
  username          = var.vcenter_user
  password          = var.vcenter_password
  host              = var.esxi_host
  insecure_connection  = true

  vm_name = "example-ubuntu"
  guest_os_type = "ubuntu64Guest"

  ssh_username = "vagrant"
  ssh_password = "vagrant"

  CPUs =             1
  RAM =              1024
  RAM_reserve_all = true

  disk_controller_type =  ["pvscsi"]
  datastore = "datastore1"
  storage {
    disk_size =        32768
    disk_thin_provisioned = true
  }

   iso_urls = [
     "http://releases.ubuntu.com/16.04/ubuntu-16.04.7-server-amd64.iso"
   ]
   iso_checksum = "sha256:b23488689e16cad7a269eb2d3a3bf725d3457ee6b0868e00c8762d3816e25848"

  network_adapters {
    network =  "VM Network"
    network_card = "vmxnet3"
  }

  floppy_files = [
    "./preseed_hardcoded_ip.cfg"
  ]

  boot_command = [
    "<enter><wait><f6><wait><esc><wait>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs><bs><bs><bs><bs><bs><bs><bs>",
    "<bs><bs><bs>",
    "/install/vmlinuz",
    " initrd=/install/initrd.gz",
    " priority=critical",
    " locale=en_US",
    " file=/media/preseed_hardcoded_ip.cfg",
    " netcfg/get_ipaddress=${var.vm_ip}",
    " netcfg/get_gateway=${var.gateway_ip}",
    "<enter>"
  ]
}
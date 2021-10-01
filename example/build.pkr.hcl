//packer {
//  required_plugins {
//    vsphere = {
//      version = ">= v1.0.0"
//      source  = "github.com/hashicorp/vsphere"
//    }
//  }
//}

build {
  hcp_packer_registry {
    bucket_name = "vsphere-ubuntu-test"
    description = <<EOT
vSphere ubuntu-16.04.7-server-amd64 test!
    EOT
  }

  sources  = [
    "source.vsphere-iso.example"
  ]

//  provisioner "shell-local" {
//    inline  = ["echo the address is: $PACKER_HTTP_ADDR and build name is: $PACKER_BUILD_NAME"]
//  }
}
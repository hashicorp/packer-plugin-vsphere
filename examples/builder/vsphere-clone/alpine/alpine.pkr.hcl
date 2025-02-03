# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

packer {
  required_plugins {
    vsphere = {
      version = "~> 1"
      source  = "github.com/hashicorp/vsphere"
    }
  }
}

variable "vm_name_prefix" {
  type        = string
  default     = "alpine"
  description = "Prefix for naming the virtual machine."
}

variable "guest_os_type" {
  type        = string
  default     = "other5xLinux64Guest"
  description = "The type of guest OS to configure for the virtual machine."
}

variable "username" {
  type        = string
  default     = "administrator@vsphere.local"
  sensitive   = true
  description = "The username for authenticating with the vCenter instance."
}

variable "password" {
  type        = string
  default     = "VMw@re1!"
  sensitive   = true
  description = "The password for authenticating with the vCenter instance."
}

variable "vcenter_server" {
  type        = string
  default     = "vcenter.example.com"
  description = "The vCenter instance used for managing the ESX host."
}

variable "host" {
  type        = string
  default     = "esx-01.example.com"
  description = "The ESX host where the virtual machine will be built."
}

variable "insecure_connection" {
  type        = bool
  default     = true
  description = "Set to true to allow insecure connections to the vCenter instance."
}

variable "communicator"
  type        = string
  default     = "none"
}

locals { timestamp = regex_replace(timestamp(), "[- TZ:]", "") }

source "vsphere-clone" "example" {
  communicator        = var.communicator
  username            = var.username
  password            = var.password
  vcenter_server      = var.vcenter_server
  host                = var.host
  insecure_connection = var.insecure_connection
  vm_name             = "${var.vm_name_prefix}-${local.timestamp}"
  template            = var.vm_name_prefix
}

build {
  sources = ["source.vsphere-clone.example"]

}

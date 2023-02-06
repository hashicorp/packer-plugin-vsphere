# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

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



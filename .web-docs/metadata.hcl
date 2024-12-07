# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# For full specification on the configuration of this file visit:
# https://github.com/hashicorp/integration-template#metadata-configuration
integration {
  name = "VMware vSphere"
  description = "The vSphere plugin is able to create vSphere virtual machines for use with any VMware product."
  identifier = "packer/hashicorp/vsphere"
  flags = ["hcp-ready"]
  component {
    type = "builder"
    name = "vSphere ISO"
    slug = "vsphere-iso"
  }
  component {
    type = "builder"
    name = "vSphere Clone"
    slug = "vsphere-clone"
  }
  component {
    type = "builder"
    name = "vSphere Supervisor"
    slug = "vsphere-supervisor"
  }
  component {
    type = "post-processor"
    name = "vSphere"
    slug = "vsphere"
  }
  component {
    type = "post-processor"
    name = "vSphere Template"
    slug = "vsphere-template"
  }
  component {
    type = "data-source"
    name = "vSphere Virtual Machine"
    slug = "vsphere-virtualmachine"
  }
}

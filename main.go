// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"log"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/clone"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/iso"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/supervisor"
	"github.com/hashicorp/packer-plugin-vsphere/post-processor/vsphere"
	vsphereTemplate "github.com/hashicorp/packer-plugin-vsphere/post-processor/vsphere-template"
	"github.com/hashicorp/packer-plugin-vsphere/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("iso", new(iso.Builder))
	pps.RegisterBuilder("clone", new(clone.Builder))
	pps.RegisterBuilder("supervisor", new(supervisor.Builder))
	pps.RegisterPostProcessor(plugin.DEFAULT_NAME, new(vsphere.PostProcessor))
	pps.RegisterPostProcessor("template", new(vsphereTemplate.PostProcessor))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		log.Fatal(err)
	}
}

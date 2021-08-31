package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/clone"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/iso"
	"github.com/hashicorp/packer-plugin-vsphere/post-processor/vsphere"
	vsphere_template "github.com/hashicorp/packer-plugin-vsphere/post-processor/vsphere-template"
	"github.com/hashicorp/packer-plugin-vsphere/version"
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("iso", new(iso.Builder))
	pps.RegisterBuilder("clone", new(clone.Builder))
	pps.RegisterPostProcessor(plugin.DEFAULT_NAME, new(vsphere.PostProcessor))
	pps.RegisterPostProcessor("template", new(vsphere_template.PostProcessor))
	pps.SetVersion(version.PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

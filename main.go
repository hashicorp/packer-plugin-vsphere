package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/plugin"
	"github.com/hashicorp/packer-plugin-sdk/version"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/clone"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/iso"
	"github.com/hashicorp/packer-plugin-vsphere/post-processor/vsphere"
	vsphere_template "github.com/hashicorp/packer-plugin-vsphere/post-processor/vsphere-template"
)

var (
	// Version is the main version number that is being run at the moment.
	Version = "1.0.1"

	// VersionPrerelease is A pre-release marker for the Version. If this is ""
	// (empty string) then it means that it is a final release. Otherwise, this
	// is a pre-release such as "dev" (in development), "beta", "rc1", etc.
	VersionPrerelease = "dev"

	// PluginVersion is used by the plugin set to allow Packer to recognize
	// what version this plugin is.
	PluginVersion = version.InitializePluginVersion(Version, VersionPrerelease)
)

func main() {
	pps := plugin.NewSet()
	pps.RegisterBuilder("iso", new(iso.Builder))
	pps.RegisterBuilder("clone", new(clone.Builder))
	pps.RegisterPostProcessor(plugin.DEFAULT_NAME, new(vsphere.PostProcessor))
	pps.RegisterPostProcessor("template", new(vsphere_template.PostProcessor))
	pps.SetVersion(PluginVersion)
	err := pps.Run()
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

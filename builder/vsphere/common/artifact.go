package common

import (
	"fmt"
	"os"

	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/vmware/govmomi/object"
)

const BuilderId = "jetbrains.vsphere"

type Artifact struct {
	Outconfig            *OutputConfig
	Name                 string
	Location             LocationConfig
	Datacenter           *object.Datacenter
	VM                   *driver.VirtualMachineDriver
	ContentLibraryConfig *ContentLibraryDestinationConfig
	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
}

func (a *Artifact) BuilderId() string {
	return BuilderId
}

func (a *Artifact) Files() []string {
	if a.Outconfig != nil {
		files, _ := a.Outconfig.ListFiles()
		return files
	}
	return []string{}
}

func (a *Artifact) Id() string {
	return a.Name
}

func (a *Artifact) String() string {
	return a.Name
}

func (a *Artifact) State(name string) interface{} {
	if name == registryimage.ArtifactStateURI {
		return a.stateHCPPackerRegistryMetadata()
	}
	return a.StateData[name]
}

// stateHCPPackerRegistryMetadata will write the metadata as an hcpRegistryImage
func (a *Artifact) stateHCPPackerRegistryMetadata() interface{} {
	labels := make(map[string]interface{})

	floppyPath, ok := a.StateData["uploaded_floppy_path"].(string)
	if ok && floppyPath != "" {
		labels["uploaded_floppy_path"] = floppyPath
	}
	metadata, ok := a.StateData["metadata"].(map[string]string)
	if ok {
		for label, data := range metadata {
			labels[label] = data
		}
	}
	if a.Location.Cluster != "" {
		labels["cluster"] = a.Location.Cluster
	}
	if a.Location.Host != "" {
		labels["host"] = a.Location.Host
	}
	if a.ContentLibraryConfig != nil {
		labels["content_library_destination"] = fmt.Sprintf("%s/%s", a.ContentLibraryConfig.Library, a.ContentLibraryConfig.Name)
	}

	img, _ := registryimage.FromArtifact(a,
		registryimage.WithID(a.Name),
		registryimage.WithRegion(a.Datacenter.Name()),
		registryimage.WithProvider("vsphere"),
		registryimage.SetLabels(labels),
	)
	return img
}

func (a *Artifact) Destroy() error {
	if a.Outconfig != nil {
		os.RemoveAll(a.Outconfig.OutputDir)
	}
	return a.VM.Destroy()
}

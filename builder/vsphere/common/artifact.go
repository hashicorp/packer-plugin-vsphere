package common

import (
	"fmt"
	"log"
	"os"

	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

const BuilderId = "jetbrains.vsphere"

type Artifact struct {
	Outconfig            *OutputConfig
	Name                 string
	Location             LocationConfig
	VM                   *driver.VirtualMachineDriver
	ContentLibraryConfig *ContentLibraryDestinationConfig
	// StateData should store data such as GeneratedData
	// to be shared with post-processors
	StateData map[string]interface{}
	labels    map[string]interface{}
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
	floppyPath, ok := a.StateData["uploaded_floppy_path"].(string)
	if ok && floppyPath != "" {
		a.labels["uploaded_floppy_path"] = floppyPath
	}

	img, _ := registryimage.FromArtifact(a,
		registryimage.WithID(a.Name),
		registryimage.WithRegion(a.Location.String()),
		registryimage.WithProvider("vsphere"),
		registryimage.SetLabels(a.labels),
	)
	return img
}

func (a *Artifact) WriteVMInfoIntoLabels() {
	if a.labels == nil {
		a.labels = make(map[string]interface{})
	}
	info, err := a.VM.Info("config.annotation", "config.hardware", "runtime.host", "resourcePool", "datastore", "network", "summary")
	if err != nil || info == nil {
		log.Printf("[TRACE] error extracting VM metadata: %s", err)
		return
	}
	if info.Config != nil {
		if info.Config.Annotation != "" {
			// VM description
			a.labels["annotation"] = info.Config.Annotation
		}
		// Hardware
		a.labels["num_cpu"] = fmt.Sprintf("%d", info.Config.Hardware.NumCPU)
		a.labels["memory_mb"] = fmt.Sprintf("%d", info.Config.Hardware.MemoryMB)
	}

	if info.Runtime.Host != nil {
		h := a.VM.NewHost(info.Runtime.Host)
		hostInfo, err := h.Info("name")
		if err == nil && hostInfo.Name != "" {
			a.labels["host"] = hostInfo.Name
		}
	}

	if info.ResourcePool != nil {
		p := a.VM.NewResourcePool(info.ResourcePool)
		poolPath, err := p.Path()
		if err == nil && poolPath != "" {
			a.labels["resurce_pool"] = poolPath
		}
	}

	for i, datastore := range info.Datastore {
		dsr := datastore.Reference()
		ds := a.VM.NewDatastore(&dsr)
		dsInfo, err := ds.Info("name")
		if err == nil && dsInfo.Name != "" {
			if i == 0 {
				a.labels["datastore"] = dsInfo.Name
				continue
			}
			key := fmt.Sprintf("datastore_%d", i)
			a.labels[key] = dsInfo.Name
		}
	}

	for i, network := range info.Network {
		net := network.Reference()
		n := a.VM.NewNetwork(&net)
		networkInfo, err := n.Info("name")
		if err == nil && networkInfo.Name != "" {
			if i == 0 {
				a.labels["network"] = networkInfo.Name
				continue
			}
			key := fmt.Sprintf("network_%d", i)
			a.labels[key] = network.String()
		}
	}

	if a.ContentLibraryConfig != nil {
		a.labels["content_library_destination"] = fmt.Sprintf("%s/%s", a.ContentLibraryConfig.Library, a.ContentLibraryConfig.Name)
	}
}

func (a *Artifact) Destroy() error {
	if a.Outconfig != nil {
		os.RemoveAll(a.Outconfig.OutputDir)
	}
	return a.VM.Destroy()
}

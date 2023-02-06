// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	registryimage "github.com/hashicorp/packer-plugin-sdk/packer/registry/image"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/vmware/govmomi/simulator"
	"github.com/vmware/govmomi/vim25/types"
)

func TestArtifactHCPPackerMetadata(t *testing.T) {
	sim, err := NewVCenterSimulator()
	if err != nil {
		t.Fatalf("should not fail: %s", err.Error())
	}
	defer sim.Close()

	vm, vmSim := sim.ChooseSimulatorPreCreatedVM()
	confSpec := types.VirtualMachineConfigSpec{Annotation: "simple vm description"}
	if err := vm.Reconfigure(confSpec); err != nil {
		t.Fatalf("unexpected error %s", err.Error())
	}
	datastore := simulator.Map.Get(vmSim.Datastore[0]).(*simulator.Datastore)
	host := simulator.Map.Get(*vmSim.Runtime.Host).(*simulator.HostSystem)

	expectedLabels := map[string]string{
		"annotation":                  vmSim.Config.Annotation,
		"num_cpu":                     fmt.Sprintf("%d", vmSim.Config.Hardware.NumCPU),
		"memory_mb":                   fmt.Sprintf("%d", vmSim.Config.Hardware.MemoryMB),
		"host":                        host.Name,
		"datastore":                   datastore.Name,
		"content_library_destination": "Library-Name/Item-Name",
		"network":                     "DC0_DVPG0",
		"vsphere_uuid":                vmSim.Config.Uuid,
	}
	artifact := &Artifact{
		Outconfig:  nil,
		Name:       vmSim.Name,
		Datacenter: vm.Datacenter(),
		Location: LocationConfig{
			Host:      host.Name,
			Datastore: datastore.Name,
		},
		ContentLibraryConfig: &ContentLibraryDestinationConfig{
			Library: "Library-Name",
			Name:    "Item-Name",
		},
		VM: vm.(*driver.VirtualMachineDriver),
		StateData: map[string]interface{}{
			"metadata": expectedLabels,
		},
	}

	metadata, ok := artifact.State(registryimage.ArtifactStateURI).(*registryimage.Image)
	if !ok {
		t.Fatalf("expecting a metadata of time registryimage.Image")
	}
	if metadata.ImageID != vmSim.Name {
		t.Fatalf("unexpected image id: %s", metadata.ImageID)
	}
	if metadata.ProviderName != "vsphere" {
		t.Fatalf("unexpected provider name: %s", metadata.ProviderName)
	}
	if metadata.ProviderRegion != vm.Datacenter().Name() {
		t.Fatalf("unexpected provider region: %s", metadata.ProviderRegion)
	}
	if diff := cmp.Diff(expectedLabels, metadata.Labels); diff != "" {
		t.Fatalf("wrong labels: %s", diff)
	}
}

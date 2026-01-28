// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/vmware/packer-plugin-vsphere/builder/vsphere/driver"
)

func GetVMMetadata(vm *driver.VirtualMachineDriver, state multistep.StateBag) map[string]string {
	labels := make(map[string]string)
	info, err := vm.Info("config.uuid", "config.annotation", "config.hardware", "resourcePool", "datastore", "network", "summary")
	if err != nil || info == nil {
		log.Printf("[TRACE] error extracting virtual machine metadata: %s", err)
		return labels
	}
	if info.Config != nil {
		// Saved the virtual machine UUID.
		// If destroyed after import to content library, the UUID is not saved.
		destroyAfterImport, ok := state.Get("destroy_vm").(bool)
		if !ok || !destroyAfterImport {
			labels["vsphere_uuid"] = info.Config.Uuid
		}

		// If the content library is used, save the content library item UUID.
		if itemUuid, ok := state.Get("content_library_item_uuid").(string); ok {
			labels["content_library_item_uuid"] = itemUuid
		}

		// Save the virtual machine annotation, if exists.
		if info.Config.Annotation != "" {
			labels["annotation"] = info.Config.Annotation
		}

		// Save the basic virtual machine hardware summary.
		labels["num_cpu"] = fmt.Sprintf("%d", info.Config.Hardware.NumCPU)
		labels["memory_mb"] = fmt.Sprintf("%d", info.Config.Hardware.MemoryMB)
	}

	// Save the virtual machine resource pool, if exists.
	if info.ResourcePool != nil {
		p := vm.NewResourcePool(info.ResourcePool)
		poolPath, err := p.Path()
		if err == nil && poolPath != "" {
			labels["resource_pool"] = poolPath
		}
	}

	// Save the virtual machine datastore.
	for i, datastore := range info.Datastore {
		dsr := datastore.Reference()
		ds := vm.NewDatastore(&dsr)
		dsInfo, err := ds.Info("name")
		if err == nil && dsInfo.Name != "" {
			if i == 0 {
				labels["datastore"] = dsInfo.Name
				continue
			}
			key := fmt.Sprintf("datastore_%d", i)
			labels[key] = dsInfo.Name
		}
	}

	// Save the virtual machine network.
	for i, network := range info.Network {
		net := network.Reference()
		n := vm.NewNetwork(&net)
		networkInfo, err := n.Info("name")
		if err == nil && networkInfo.Name != "" {
			if i == 0 {
				labels["network"] = networkInfo.Name
				continue
			}
			key := fmt.Sprintf("network_%d", i)
			labels[key] = network.String()
		}
	}

	// Save the virtual machine content library datastore.
	if datastores, ok := state.Get("content_library_datastore").([]string); ok {
		for i, ds := range datastores {
			if i == 0 {
				labels["template_datastore"] = ds
				continue
			}
			key := fmt.Sprintf("template_datastore_%d", i)
			labels[key] = ds
		}
	}

	return labels
}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

func GetVMMetadata(vm *driver.VirtualMachineDriver, state multistep.StateBag) map[string]string {
	labels := make(map[string]string)

	info, err := vm.Info("config.uuid", "config.annotation", "config.hardware", "resourcePool", "datastore", "network", "summary")
	if err != nil || info == nil {
		log.Printf("[TRACE] error extracting VM metadata: %s", err)
		return labels
	}
	if info.Config != nil {
		labels["vsphere_uuid"] = info.Config.Uuid

		// VM description
		if info.Config.Annotation != "" {
			labels["annotation"] = info.Config.Annotation
		}

		// Hardware
		labels["num_cpu"] = fmt.Sprintf("%d", info.Config.Hardware.NumCPU)
		labels["memory_mb"] = fmt.Sprintf("%d", info.Config.Hardware.MemoryMB)
	}

	if info.ResourcePool != nil {
		p := vm.NewResourcePool(info.ResourcePool)
		poolPath, err := p.Path()
		if err == nil && poolPath != "" {
			labels["resurce_pool"] = poolPath
		}
	}

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

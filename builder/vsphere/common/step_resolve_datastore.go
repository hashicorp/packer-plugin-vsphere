// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"log"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/vmware/packer-plugin-vsphere/builder/vsphere/driver"
)

// DatastoreClusterDriver is an interface for drivers that support datastore cluster operations.
type DatastoreClusterDriver interface {
	driver.Driver
	SelectDatastoreFromCluster(clusterName string) (driver.Datastore, string, error)
}

// StepResolveDatastore resolves a datastore from either a direct datastore
// name or a datastore cluster name. When a datastore cluster is specified,
// it uses Storage DRS to select the optimal datastore.
type StepResolveDatastore struct {
	// Datastore is the name of a specific datastore to use.
	Datastore string
	// DatastoreCluster is the name of a datastore cluster to use.
	// When specified, Storage DRS will select the optimal datastore.
	DatastoreCluster string
	// DiskCount is the number of disks that will be created.
	// When using a datastore cluster with multiple disks, this step will be skipped
	// to avoid redundant DRS calls (per-disk DRS calls will be made later).
	DiskCount int
}

// Run resolves a datastore from either a direct datastore name or a datastore cluster.
// When using a datastore cluster, Storage DRS selects the optimal datastore.
// For multi-disk configurations with datastore clusters, this initial selection is used
// for non-disk operations (ISO uploads, etc.), while per-disk DRS calls are made separately.
func (s *StepResolveDatastore) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	d := state.Get("driver").(driver.Driver)

	if s.Datastore == "" && s.DatastoreCluster == "" {
		return multistep.ActionContinue
	}

	var ds driver.Datastore
	var err error
	var selectionMethod string

	if s.DatastoreCluster != "" {
		clusterDriver, ok := d.(DatastoreClusterDriver)
		if !ok {
			state.Put("error", fmt.Errorf("driver does not support datastore cluster operations"))
			return multistep.ActionHalt
		}

		ds, selectionMethod, err = clusterDriver.SelectDatastoreFromCluster(s.DatastoreCluster)
		if err != nil {
			state.Put("error", fmt.Errorf("error resolving datastore from cluster '%s': %s", s.DatastoreCluster, err))
			return multistep.ActionHalt
		}

		if s.DiskCount > 1 {
			log.Printf("[INFO] Selected datastore '%s' from cluster '%s' for non-disk operations (per-disk placement will be requested separately)", ds.Name(), s.DatastoreCluster)
		} else {
			if selectionMethod == driver.SelectionMethodDRS {
				log.Printf("[INFO] Storage DRS selected datastore '%s' from cluster '%s'", ds.Name(), s.DatastoreCluster)
			} else {
				log.Printf("[INFO] Selected datastore '%s' from cluster '%s' (first available)", ds.Name(), s.DatastoreCluster)
			}
		}
	} else {
		log.Printf("[INFO] Using datastore '%s'", s.Datastore)
		ds, err = d.FindDatastore(s.Datastore, "")
		if err != nil {
			state.Put("error", fmt.Errorf("error finding datastore '%s': %s", s.Datastore, err))
			return multistep.ActionHalt
		}
		selectionMethod = "direct"
	}

	state.Put("datastore", ds)
	state.Put("datastore_selection_method", selectionMethod)

	return multistep.ActionContinue
}

// Cleanup performs any necessary cleanup.
func (s *StepResolveDatastore) Cleanup(state multistep.StateBag) {}

// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"fmt"
	"strings"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

// DatastoreCluster represents a vSphere datastore cluster (storage pod).
type DatastoreCluster interface {
	// Name returns the name of the datastore cluster.
	Name() string

	// Reference returns the managed object reference.
	Reference() types.ManagedObjectReference

	// ListDatastores returns all datastores in the cluster.
	ListDatastores() ([]Datastore, error)
}

// DatastoreClusterDriver implements the DatastoreCluster interface.
type DatastoreClusterDriver struct {
	cluster *object.StoragePod
	driver  *VCenterDriver
}

// Name returns the name of the datastore cluster.
func (dsc *DatastoreClusterDriver) Name() string {
	return dsc.cluster.Name()
}

// Reference returns the managed object reference of the datastore cluster.
func (dsc *DatastoreClusterDriver) Reference() types.ManagedObjectReference {
	return dsc.cluster.Reference()
}

// ListDatastores returns all datastores in the cluster.
func (dsc *DatastoreClusterDriver) ListDatastores() ([]Datastore, error) {
	datastores, err := dsc.cluster.Children(dsc.driver.ctx)
	if err != nil {
		return nil, fmt.Errorf("error listing datastores in cluster '%s': %s", dsc.Name(), err)
	}

	var result []Datastore
	for _, dsRef := range datastores {
		// Try to get the datastore via finder first for proper initialization.
		ds, err := dsc.driver.finder.Datastore(dsc.driver.ctx, dsRef.Reference().Value)
		if err != nil {
			// If finder fails, create from reference and fetch properties.
			datastoreObj := object.NewDatastore(dsc.driver.client.Client, dsRef.Reference())
			dsDriver := &DatastoreDriver{
				ds:     datastoreObj,
				driver: dsc.driver,
			}
			info, err := dsDriver.Info("name")
			if err != nil {
				return nil, fmt.Errorf("error getting datastore info: %s", err)
			}
			if info.Name == "" {
				return nil, fmt.Errorf("datastore has empty name")
			}
			result = append(result, dsDriver)
		} else {
			result = append(result, &DatastoreDriver{
				ds:     ds,
				driver: dsc.driver,
			})
		}
	}

	return result, nil
}

// FindDatastoreCluster locates a datastore cluster by name.
// Returns a DatastoreCluster object or an error if the cluster is not found.
func (d *VCenterDriver) FindDatastoreCluster(name string) (DatastoreCluster, error) {
	storagePod, err := d.finder.DatastoreCluster(d.ctx, name)
	if err != nil {
		clusters, listErr := d.finder.DatastoreClusterList(d.ctx, "*")
		if listErr == nil && len(clusters) > 0 {
			var names []string
			for _, c := range clusters {
				names = append(names, c.Name())
			}
			return nil, fmt.Errorf("datastore cluster '%s' not found; available clusters: %s", name, strings.Join(names, ", "))
		}
		return nil, fmt.Errorf("error finding datastore cluster with name '%s': %s", name, err)
	}

	return &DatastoreClusterDriver{
		cluster: storagePod,
		driver:  d,
	}, nil
}

// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"github.com/vmware/govmomi/vim25/types"
)

// DatastoreClusterMock is a mock implementation of the DatastoreCluster interface for testing.
type DatastoreClusterMock struct {
	NameReturn           string
	ReferenceReturn      types.ManagedObjectReference
	ListDatastoresReturn []Datastore
	ListDatastoresErr    error
	ListDatastoresCalled bool
}

// Name returns the mock name of the datastore cluster.
func (dsc *DatastoreClusterMock) Name() string {
	if dsc.NameReturn == "" {
		return "datastore-cluster-mock"
	}
	return dsc.NameReturn
}

// Reference returns the mock managed object reference.
func (dsc *DatastoreClusterMock) Reference() types.ManagedObjectReference {
	if dsc.ReferenceReturn.Type == "" {
		return types.ManagedObjectReference{
			Type:  "StoragePod",
			Value: "datastore-cluster-mock-ref",
		}
	}
	return dsc.ReferenceReturn
}

// ListDatastores returns the mock list of datastores.
func (dsc *DatastoreClusterMock) ListDatastores() ([]Datastore, error) {
	dsc.ListDatastoresCalled = true
	if dsc.ListDatastoresErr != nil {
		return nil, dsc.ListDatastoresErr
	}
	if dsc.ListDatastoresReturn == nil {
		// Return a default mock datastore if none specified
		return []Datastore{&DatastoreMock{NameReturn: "datastore-1"}}, nil
	}
	return dsc.ListDatastoresReturn, nil
}

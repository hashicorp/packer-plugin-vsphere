// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package common

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

func TestStepResolveDatastore_Run(t *testing.T) {
	tc := []struct {
		name              string
		step              *StepResolveDatastore
		driverMock        *VCenterDriverMock
		expectedAction    multistep.StepAction
		expectedDatastore string
		expectedMethod    string
		expectError       bool
		errorContains     string
	}{
		{
			name: "Resolve from direct datastore",
			step: &StepResolveDatastore{
				Datastore: "test-datastore",
			},
			driverMock: func() *VCenterDriverMock {
				m := NewVCenterDriverMock()
				m.DatastoreMock = &driver.DatastoreMock{
					NameReturn: "test-datastore",
				}
				return m
			}(),
			expectedAction:    multistep.ActionContinue,
			expectedDatastore: "test-datastore",
			expectedMethod:    "direct",
			expectError:       false,
		},
		{
			name: "Resolve from datastore cluster with DRS",
			step: &StepResolveDatastore{
				DatastoreCluster: "test-cluster",
			},
			driverMock: func() *VCenterDriverMock {
				m := NewVCenterDriverMock()
				m.SelectDatastoreReturn = &driver.DatastoreMock{
					NameReturn: "cluster-datastore-1",
				}
				m.SelectDatastoreMethod = driver.SelectionMethodDRS
				return m
			}(),
			expectedAction:    multistep.ActionContinue,
			expectedDatastore: "cluster-datastore-1",
			expectedMethod:    driver.SelectionMethodDRS,
			expectError:       false,
		},
		{
			name: "Resolve from datastore cluster with fallback",
			step: &StepResolveDatastore{
				DatastoreCluster: "test-cluster",
			},
			driverMock: func() *VCenterDriverMock {
				m := NewVCenterDriverMock()
				m.SelectDatastoreReturn = &driver.DatastoreMock{
					NameReturn: "cluster-datastore-1",
				}
				m.SelectDatastoreMethod = driver.SelectionMethodFallback
				return m
			}(),
			expectedAction:    multistep.ActionContinue,
			expectedDatastore: "cluster-datastore-1",
			expectedMethod:    driver.SelectionMethodFallback,
			expectError:       false,
		},
		{
			name: "Error finding direct datastore",
			step: &StepResolveDatastore{
				Datastore: "missing-datastore",
			},
			driverMock: func() *VCenterDriverMock {
				m := NewVCenterDriverMock()
				m.FindDatastoreErr = fmt.Errorf("datastore not found")
				return m
			}(),
			expectedAction: multistep.ActionHalt,
			expectError:    true,
			errorContains:  "error finding datastore 'missing-datastore'",
		},
		{
			name: "Error selecting from datastore cluster",
			step: &StepResolveDatastore{
				DatastoreCluster: "missing-cluster",
			},
			driverMock: func() *VCenterDriverMock {
				m := NewVCenterDriverMock()
				m.SelectDatastoreErr = fmt.Errorf("cluster not found")
				return m
			}(),
			expectedAction: multistep.ActionHalt,
			expectError:    true,
			errorContains:  "error resolving datastore from cluster 'missing-cluster'",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			state := basicStateBag(nil)
			state.Put("driver", c.driverMock)

			action := c.step.Run(context.TODO(), state)
			if action != c.expectedAction {
				t.Fatalf("unexpected action: expected '%#v', but returned '%#v'", c.expectedAction, action)
			}

			if c.expectError {
				err, ok := state.Get("error").(error)
				if !ok {
					t.Fatal("expected error in state bag, but none found")
				}
				if !strings.Contains(err.Error(), c.errorContains) {
					t.Fatalf("unexpected error: expected to contain '%s', but got '%s'", c.errorContains, err.Error())
				}
			} else {
				if _, ok := state.GetOk("error"); ok {
					t.Fatal("unexpected error in state bag")
				}

				ds, ok := state.Get("datastore").(driver.Datastore)
				if !ok {
					t.Fatal("expected datastore in state bag, but none found")
				}
				if ds.Name() != c.expectedDatastore {
					t.Fatalf("unexpected datastore: expected '%s', but got '%s'", c.expectedDatastore, ds.Name())
				}

				method, ok := state.Get("datastore_selection_method").(string)
				if !ok {
					t.Fatal("expected datastore_selection_method in state bag, but none found")
				}
				if method != c.expectedMethod {
					t.Fatalf("unexpected selection method: expected '%s', but got '%s'", c.expectedMethod, method)
				}
			}

			// Verify mock was called correctly
			if c.step.Datastore != "" && !c.driverMock.FindDatastoreCalled {
				t.Fatal("expected FindDatastore to be called, but it wasn't")
			}
			if c.step.DatastoreCluster != "" && !c.driverMock.SelectDatastoreCalled {
				t.Fatal("expected SelectDatastoreFromCluster to be called, but it wasn't")
			}
		})
	}
}

func TestStepResolveDatastore_Cleanup(t *testing.T) {
	step := &StepResolveDatastore{}
	state := basicStateBag(nil)

	// Cleanup should be a no-op
	step.Cleanup(state)
}

// VCenterDriverMock embeds DriverMock and adds VCenterDriver-specific methods for testing
type VCenterDriverMock struct {
	*driver.DriverMock

	SelectDatastoreCalled bool
	SelectDatastoreReturn driver.Datastore
	SelectDatastoreMethod string
	SelectDatastoreErr    error
}

// NewVCenterDriverMock creates a new VCenterDriverMock
func NewVCenterDriverMock() *VCenterDriverMock {
	return &VCenterDriverMock{
		DriverMock: driver.NewDriverMock(),
	}
}

// SelectDatastoreFromCluster mocks the VCenterDriver method
func (d *VCenterDriverMock) SelectDatastoreFromCluster(clusterName string) (driver.Datastore, string, error) {
	d.SelectDatastoreCalled = true
	return d.SelectDatastoreReturn, d.SelectDatastoreMethod, d.SelectDatastoreErr
}

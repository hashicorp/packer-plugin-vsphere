// © Broadcom. All Rights Reserved.
// The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
// SPDX-License-Identifier: MPL-2.0

package clone

import (
	"bytes"
	"context"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/vmware/packer-plugin-vsphere/builder/vsphere/common"
	"github.com/vmware/packer-plugin-vsphere/builder/vsphere/driver"
)

func TestCreateConfig_Prepare(t *testing.T) {
	tc := []struct {
		name           string
		config         *CloneConfig
		fail           bool
		expectedErrMsg string
	}{
		{
			name: "Valid config",
			config: &CloneConfig{
				Template: "template name",
				StorageConfig: common.StorageConfig{
					DiskControllerType: []string{"test"},
					Storage: []common.DiskConfig{
						{
							DiskSize: 0,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "storage[0].'disk_size' is required",
		},
		{
			name: "Storage validate disk_size",
			config: &CloneConfig{
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize:            0,
							DiskThinProvisioned: true,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "storage[0].'disk_size' is required",
		},
		{
			name: "Storage validate disk_controller_index",
			config: &CloneConfig{
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize:            32768,
							DiskControllerIndex: 3,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "storage[0].'disk_controller_index' references an unknown disk controller",
		},
		{
			name: "Validate template is set",
			config: &CloneConfig{
				StorageConfig: common.StorageConfig{
					DiskControllerType: []string{"test"},
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "'template' is required",
		},
		{
			name: "Validate LinkedClone and DiskSize set at the same time",
			config: &CloneConfig{
				Template:    "template name",
				LinkedClone: true,
				DiskSize:    32768,
				StorageConfig: common.StorageConfig{
					DiskControllerType: []string{"test"},
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "'linked_clone' and 'disk_size' cannot be used together",
		},
		{
			name: "Validate MacAddress and Network not set at the same time",
			config: &CloneConfig{
				Template:   "template name",
				MacAddress: "some mac address",
				StorageConfig: common.StorageConfig{
					DiskControllerType: []string{"test"},
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "'network' is required when 'mac_address' is specified",
		},
	}

	for _, c := range tc {
		t.Run(c.name, func(t *testing.T) {
			errs := c.config.Prepare()
			if c.fail {
				if len(errs) == 0 {
					t.Fatal("unexpected success: expected failure")
				}
				if errs[0].Error() != c.expectedErrMsg {
					t.Fatalf("unexpected error: expected '%s', but returned '%s'", c.expectedErrMsg, errs[0])
				}
			} else {
				if len(errs) != 0 {
					t.Fatalf("unexpected failure: expected success, but failed: %s", errs[0])
				}
			}
		})
	}
}

func TestStepCreateVM_Run(t *testing.T) {
	state := new(multistep.BasicStateBag)
	state.Put("ui", &packersdk.BasicUi{
		Reader: new(bytes.Buffer),
		Writer: new(bytes.Buffer),
	})
	driverMock := driver.NewDriverMock()
	state.Put("driver", driverMock)
	step := basicStepCloneVM()
	step.Force = true
	vmPath := path.Join(step.Location.Folder, step.Location.VMName)
	vmMock := new(driver.VirtualMachineMock)
	driverMock.VM = vmMock

	if action := step.Run(context.TODO(), state); action == multistep.ActionHalt {
		t.Fatalf("unexpected action: expected '%#v', but returned '%#v'", multistep.ActionContinue, action)
	}

	// Find VM
	if !driverMock.FindVMCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "FindVM")
	}

	// Pre clean VM
	if !driverMock.PreCleanVMCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "PreCleanVM")
	}
	if driverMock.PreCleanForce != step.Force {
		t.Fatalf("unexpected result: expected '%t', but returned '%t'", step.Force, driverMock.PreCleanForce)
	}
	if driverMock.PreCleanVMPath != vmPath {
		t.Fatalf("unexpected result: expected '%s', but returned '%s'", vmPath, driverMock.PreCleanVMPath)
	}

	// Clone VM
	if !vmMock.CloneCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "Clone")
	}
	if diff := cmp.Diff(vmMock.CloneConfig, driverCreateConfig(step.Config, step.Location)); diff != "" {
		t.Fatalf("unexpected result: '%s'", diff)
	}
	vm, ok := state.GetOk("vm")
	if !ok {
		t.Fatalf("unexpected state: '%s' not found", "vm")
	}
	if vm != driverMock.VM {
		t.Fatalf("unexpected result: expected '%s', but returned '%s'", driverMock.VM, vm)
	}
}

func basicStepCloneVM() *StepCloneVM {
	step := &StepCloneVM{
		Config:   createConfig(),
		Location: basicLocationConfig(),
	}
	return step
}

func basicLocationConfig() *common.LocationConfig {
	return &common.LocationConfig{
		VMName:       "test-vm",
		Folder:       "test-folder",
		Cluster:      "test-cluster",
		Host:         "test-host",
		ResourcePool: "test-resource-pool",
		Datastore:    "test-datastore",
	}
}

func createConfig() *CloneConfig {
	return &CloneConfig{
		Template: "template name",
		StorageConfig: common.StorageConfig{
			DiskControllerType: []string{"pvscsi"},
			Storage: []common.DiskConfig{
				{
					DiskSize:            32768,
					DiskThinProvisioned: true,
				},
			},
		},
	}
}

func driverCreateConfig(config *CloneConfig, location *common.LocationConfig) *driver.CloneConfig {
	var disks []driver.Disk
	for _, disk := range config.StorageConfig.Storage {
		disks = append(disks, driver.Disk{
			DiskSize:            disk.DiskSize,
			DiskEagerlyScrub:    disk.DiskEagerlyScrub,
			DiskThinProvisioned: disk.DiskThinProvisioned,
			ControllerIndex:     disk.DiskControllerIndex,
		})
	}

	return &driver.CloneConfig{
		StorageConfig: driver.StorageConfig{
			DiskControllerType: config.StorageConfig.DiskControllerType,
			Storage:            disks,
		},
		Annotation:      config.Notes,
		Name:            location.VMName,
		Folder:          location.Folder,
		Cluster:         location.Cluster,
		Host:            location.Host,
		ResourcePool:    location.ResourcePool,
		Datastore:       location.Datastore,
		LinkedClone:     config.LinkedClone,
		Network:         config.Network,
		MacAddress:      strings.ToLower(config.MacAddress),
		VAppProperties:  config.VAppConfig.Properties,
		PrimaryDiskSize: config.DiskSize,
	}
}

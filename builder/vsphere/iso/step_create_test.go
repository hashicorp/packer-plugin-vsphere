// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iso

import (
	"context"
	"errors"
	"io"
	"path"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

func TestCreateConfig_Prepare(t *testing.T) {
	// Empty config - check defaults
	config := &CreateConfig{
		// Storage is required
		StorageConfig: common.StorageConfig{
			Storage: []common.DiskConfig{
				{
					DiskSize: 32768,
				},
			},
		},
	}
	if errs := config.Prepare(); len(errs) != 0 {
		t.Fatalf("unexpected failure: expected success, but failed: %s", errs[0])
	}
	if config.GuestOSType != "otherGuest" {
		t.Fatalf("unexpected result: expected '%s', but returned '%s'", "otherGuest", config.GuestOSType)
	}
	if len(config.StorageConfig.DiskControllerType) != 1 {
		t.Fatalf("unexpected result: expected '%d', but returned '%d'", 1, len(config.StorageConfig.DiskControllerType))
	}

	// Data validation
	tc := []struct {
		name           string
		config         *CreateConfig
		fail           bool
		expectedErrMsg string
	}{
		{
			name: "Storage validate disk_size",
			config: &CreateConfig{
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
			config: &CreateConfig{
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
			name: "USBController validate 'usb' and 'xhci' can be set together",
			config: &CreateConfig{
				USBController: []string{"usb", "xhci"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail: false,
		},
		{
			name: "USBController validate '1' and '0' can be set together",
			config: &CreateConfig{
				USBController: []string{"1", "0"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail: false,
		},
		{
			name: "USBController validate 'true' and 'false' can be set together",
			config: &CreateConfig{
				USBController: []string{"true", "false"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail: false,
		},
		{
			name: "USBController validate 'true' and 'usb' cannot be set together",
			config: &CreateConfig{
				USBController: []string{"true", "usb"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "there can only be one usb controller and one xhci controller",
		},
		{
			name: "USBController validate '1' and 'usb' cannot be set together",
			config: &CreateConfig{
				USBController: []string{"1", "usb"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "there can only be one usb controller and one xhci controller",
		},
		{
			name: "USBController validate 'xhci' cannot be set more that once",
			config: &CreateConfig{
				USBController: []string{"xhci", "xhci"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "there can only be one usb controller and one xhci controller",
		},
		{
			name: "USBController validate unknown value cannot be set",
			config: &CreateConfig{
				USBController: []string{"unknown"},
				StorageConfig: common.StorageConfig{
					Storage: []common.DiskConfig{
						{
							DiskSize: 32768,
						},
					},
				},
			},
			fail:           true,
			expectedErrMsg: "usb_controller[0] references an unknown usb controller",
		},
	}

	for _, c := range tc {
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
	}
}

func TestStepCreateVM_Run(t *testing.T) {
	state := basicStateBag()
	driverMock := driver.NewDriverMock()
	state.Put("driver", driverMock)
	step := basicStepCreateVM()
	step.Force = true
	vmPath := path.Join(step.Location.Folder, step.Location.VMName)

	if action := step.Run(context.TODO(), state); action == multistep.ActionHalt {
		t.Fatalf("unexpected action: expected '%#v', but returned '%#v'", multistep.ActionContinue, action)
	}

	// Pre clean VM
	if !driverMock.PreCleanVMCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "PreCleanVM")
	}
	if driverMock.PreCleanForce != step.Force {
		t.Fatalf("unexpected result: expected '%t', but returned '%t'.", step.Force, driverMock.PreCleanForce)
	}
	if driverMock.PreCleanVMPath != vmPath {
		t.Fatalf("unexpected result: expected %s, but returned %s", vmPath, driverMock.PreCleanVMPath)
	}

	if !driverMock.CreateVMCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "CreateVM")
	}
	if diff := cmp.Diff(driverMock.CreateConfig, driverCreateConfig(step.Config, step.Location)); diff != "" {
		t.Fatalf("unexpected result: %s", diff)
	}
	vm, ok := state.GetOk("vm")
	if !ok {
		t.Fatalf("unexpected result: expected '%s' to be in state", "vm")
	}
	if vm != driverMock.VM {
		t.Fatalf("unexpected result: expected '%s', but returned '%s'", driverMock.VM, vm)
	}
}

func TestStepCreateVM_RunHalt(t *testing.T) {
	state := basicStateBag()
	step := basicStepCreateVM()

	// PreCleanVM fails
	driverMock := driver.NewDriverMock()
	driverMock.PreCleanShouldFail = true
	state.Put("driver", driverMock)
	if action := step.Run(context.TODO(), state); action != multistep.ActionHalt {
		t.Fatalf("unexpected action: expected '%#v', but returned '%#v'", multistep.ActionHalt, action)
	}
	if !driverMock.PreCleanVMCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "PreCleanVM")
	}

	// CreateVM fails
	driverMock = driver.NewDriverMock()
	driverMock.CreateVMShouldFail = true
	state.Put("driver", driverMock)
	if action := step.Run(context.TODO(), state); action != multistep.ActionHalt {
		t.Fatalf("unexpected action: expected '%#v', but returned '%#v'", multistep.ActionHalt, action)
	}
	if !driverMock.PreCleanVMCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "PreCleanVM")
	}
	if !driverMock.CreateVMCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "CreateVM")
	}
	if _, ok := state.GetOk("vm"); ok {
		t.Fatalf("unexpected result: expected '%s' not to be in state", "vm")
	}
}

func TestStepCreateVM_Cleanup(t *testing.T) {
	state := basicStateBag()
	step := basicStepCreateVM()
	vm := new(driver.VirtualMachineMock)
	state.Put("vm", vm)

	// Clean up when state is cancelled
	state.Put(multistep.StateCancelled, true)
	step.Cleanup(state)
	if !vm.DestroyCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "Destroy")
	}
	vm.DestroyCalled = false
	state.Remove(multistep.StateCancelled)

	// Clean up when state is halted
	state.Put(multistep.StateHalted, true)
	step.Cleanup(state)
	if !vm.DestroyCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "Destroy")
	}
	vm.DestroyCalled = false
	state.Remove(multistep.StateHalted)

	// Clean up when state is destroy_vm is set
	state.Put("destroy_vm", true)
	step.Cleanup(state)
	if !vm.DestroyCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "Destroy")
	}
	vm.DestroyCalled = false
	state.Remove("destroy_vm")

	// Don't clean up if state is not set with previous values
	step.Cleanup(state)
	if vm.DestroyCalled {
		t.Fatalf("unexpected result: expected '%s' not to be called", "Destroy")
	}

	// Destroy fail
	errorBuffer := &strings.Builder{}
	ui := &packersdk.BasicUi{
		Reader:      strings.NewReader(""),
		Writer:      io.Discard,
		ErrorWriter: errorBuffer,
	}
	state.Put("ui", ui)
	state.Put(multistep.StateCancelled, true)
	vm.DestroyError = errors.New("destroy failed")

	step.Cleanup(state)
	if !vm.DestroyCalled {
		t.Fatalf("unexpected result: expected '%s' to be called", "Destroy")
	}
	if !strings.Contains(errorBuffer.String(), vm.DestroyError.Error()) {
		t.Fatalf("unexpected result: expected '%s', but returned '%s'", vm.DestroyError.Error(), errorBuffer.String())
	}
	vm.DestroyCalled = false
	state.Remove(multistep.StateCancelled)

	// Should not destroy if VM is not set
	state.Remove("vm")
	state.Put(multistep.StateCancelled, true)
	step.Cleanup(state)
	if vm.DestroyCalled {
		t.Fatalf("unexpected result: expected '%s' not to be called", "Destroy")
	}
}

func basicStepCreateVM() *StepCreateVM {
	step := &StepCreateVM{
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

func createConfig() *CreateConfig {
	return &CreateConfig{
		Version:     1,
		GuestOSType: "ubuntu64Guest",
		StorageConfig: common.StorageConfig{
			DiskControllerType: []string{"pvscsi"},
			Storage: []common.DiskConfig{
				{
					DiskSize:            32768,
					DiskThinProvisioned: true,
				},
			},
		},
		NICs: []NIC{
			{
				Network:     "VM Network",
				NetworkCard: "vmxnet3",
			},
		},
	}
}

func driverCreateConfig(config *CreateConfig, location *common.LocationConfig) *driver.CreateConfig {
	var networkCards []driver.NIC
	for _, nic := range config.NICs {
		networkCards = append(networkCards, driver.NIC{
			Network:     nic.Network,
			NetworkCard: nic.NetworkCard,
			MacAddress:  strings.ToLower(nic.MacAddress),
			Passthrough: nic.Passthrough,
		})
	}

	var disks []driver.Disk
	for _, disk := range config.StorageConfig.Storage {
		disks = append(disks, driver.Disk{
			DiskSize:            disk.DiskSize,
			DiskEagerlyScrub:    disk.DiskEagerlyScrub,
			DiskThinProvisioned: disk.DiskThinProvisioned,
			ControllerIndex:     disk.DiskControllerIndex,
		})
	}

	return &driver.CreateConfig{
		StorageConfig: driver.StorageConfig{
			DiskControllerType: config.StorageConfig.DiskControllerType,
			Storage:            disks,
		},
		Annotation:    config.Notes,
		Name:          location.VMName,
		Folder:        location.Folder,
		Cluster:       location.Cluster,
		Host:          location.Host,
		ResourcePool:  location.ResourcePool,
		Datastore:     location.Datastore,
		GuestOS:       config.GuestOSType,
		NICs:          networkCards,
		USBController: config.USBController,
		Version:       config.Version,
	}
}

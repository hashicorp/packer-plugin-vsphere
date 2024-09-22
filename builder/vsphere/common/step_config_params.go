// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ConfigParamsConfig

package common

import (
	"context"
	"fmt"
	"log"

	"github.com/vmware/govmomi/vim25/types"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type ConfigParamsConfig struct {
	// A map of key-value pairs to sent to the [`extraConfig`](https://dp-downloads.broadcom.com/api-content/apis/API_VWSA_001/8.0U3/html/ReferenceGuides/vim.vm.ConfigSpec.html#extraConfig).
	// in the vSphere API's `VirtualMachineConfigSpec`.
	//
	// HCL Example:
	//
	// ```hcl
	//   configuration_parameters = {
	//     "disk.EnableUUID" = "TRUE"
	//     "svga.autodetect" = "TRUE"
	//     "log.keepOld"     = "15"
	//   }
	// ```
	//
	// JSON Example:
	//
	// ```json
	//   "configuration_parameters": {
	//     "disk.EnableUUID": "TRUE",
	//     "svga.autodetect": "TRUE",
	//     "log.keepOld": "15"
	//   }
	// ```
	//
	// ~> **Note:** Configuration keys that would conflict with parameters that
	// are explicitly configurable through other fields in the `ConfigSpec`` object
	// are silently ignored. Refer to the [`VirtualMachineConfigSpec`](https://dp-downloads.broadcom.com/api-content/apis/API_VWSA_001/8.0U3/html/ReferenceGuides/vim.vm.ConfigSpec.html)
	// in the vSphere API documentation.
	ConfigParams map[string]string `mapstructure:"configuration_parameters"`
	// Enable time synchronization with the ESXi host where the virtual machine
	// is running. Defaults to `false`.
	ToolsSyncTime bool `mapstructure:"tools_sync_time"`
	// Automatically check for and upgrade VMware Tools after a virtual machine
	// power cycle. Defaults to `false`.
	ToolsUpgradePolicy bool `mapstructure:"tools_upgrade_policy"`
}

type StepConfigParams struct {
	Config *ConfigParamsConfig
}

func (s *StepConfigParams) Run(_ context.Context, state multistep.StateBag) multistep.StepAction {
	ui := state.Get("ui").(packersdk.Ui)
	vm := state.Get("vm").(*driver.VirtualMachineDriver)
	configParams := make(map[string]string)

	if s.Config.ConfigParams != nil {
		configParams = s.Config.ConfigParams
	}

	var info *types.ToolsConfigInfo
	if s.Config.ToolsSyncTime || s.Config.ToolsUpgradePolicy {
		info = &types.ToolsConfigInfo{}

		if s.Config.ToolsSyncTime {
			info.SyncTimeWithHost = &s.Config.ToolsSyncTime
		}

		if s.Config.ToolsUpgradePolicy {
			info.ToolsUpgradePolicy = "UpgradeAtPowerCycle"
		}
	}

	ui.Say("Adding configuration parameters...")

	// Iterate over the map and log each key-value pair.
	for key, value := range configParams {
		log.Printf("[INFO] Adding: %s = %v", key, value)
	}

	if err := vm.AddConfigParams(configParams, info); err != nil {
		state.Put("error", fmt.Errorf("error adding configuration parameters: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepConfigParams) Cleanup(state multistep.StateBag) {}

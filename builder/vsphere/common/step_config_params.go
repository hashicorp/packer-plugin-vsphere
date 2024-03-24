// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type ConfigParamsConfig

package common

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/vim25/types"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

type ConfigParamsConfig struct {
	// configuration_parameters is a direct passthrough to the vSphere API's
	// [VirtualMachineConfigSpec](https://developer.broadcom.com/xapis/virtual-infrastructure-json-api/8.0.2.0/data-structures/VirtualMachineConfigSpec/)
	ConfigParams map[string]string `mapstructure:"configuration_parameters"`
	// Enable time synchronization with the ESXi host where the virtual machine
	// is running. Defaults to `false`.
	ToolsSyncTime bool `mapstructure:"tools_sync_time"`
	// Automatically check for and upgrade VMware Tools after a virtual
	// machine power cycle. Defaults to `false`.
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
	if err := vm.AddConfigParams(configParams, info); err != nil {
		state.Put("error", fmt.Errorf("error adding configuration parameters: %v", err))
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepConfigParams) Cleanup(state multistep.StateBag) {}

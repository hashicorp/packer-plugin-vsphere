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
	// ConfigSpec: https://vdc-download.vmware.com/vmwb-repository/dcr-public/bf660c0a-f060-46e8-a94d-4b5e6ffc77ad/208bc706-e281-49b6-a0ce-b402ec19ef82/SDK/vsphere-ws/docs/ReferenceGuide/vim.vm.ConfigSpec.html
	ConfigParams map[string]string `mapstructure:"configuration_parameters"`

	// Enables time synchronization with the host. Defaults to false.
	ToolsSyncTime bool `mapstructure:"tools_sync_time"`

	// If sets to true, vSphere will automatically check and upgrade VMware Tools upon a system power cycle.
	// If not set, defaults to manual upgrade.
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

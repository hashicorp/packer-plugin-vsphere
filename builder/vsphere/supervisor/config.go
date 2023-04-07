// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config

package supervisor

import (
	"errors"

	packercommon "github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/communicator"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-sdk/template/interpolate"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common"
)

const (
	DefaultSSHUsername = "packer"
)

type Config struct {
	packercommon.PackerConfig `mapstructure:",squash"`
	CommunicatorConfig        communicator.Config `mapstructure:",squash"`
	ConnectSupervisorConfig   `mapstructure:",squash"`
	CreateSourceConfig        `mapstructure:",squash"`
	WatchSourceConfig         `mapstructure:",squash"`

	ctx interpolate.Context
}

func (c *Config) Prepare(raws ...interface{}) ([]string, error) {
	err := config.Decode(c, &config.DecodeOpts{
		PluginType:         common.BuilderId,
		Interpolate:        true,
		InterpolateContext: &c.ctx,
		InterpolateFilter: &interpolate.RenderFilter{
			Exclude: []string{
				"boot_command",
			},
		},
	}, raws...)

	if err != nil {
		return nil, err
	}

	// Set a default value to "ssh_username" as it's required for the SSH communicator.
	// This must call before the CommunicatorConfig.Prepare to avoid it erroring out.
	if c.CommunicatorConfig.SSHUsername == "" {
		c.CommunicatorConfig.SSHUsername = DefaultSSHUsername
	}

	errs := new(packersdk.MultiError)
	errs = packersdk.MultiErrorAppend(errs, c.CommunicatorConfig.Prepare(&c.ctx)...)
	errs = packersdk.MultiErrorAppend(errs, c.ConnectSupervisorConfig.Prepare()...)
	errs = packersdk.MultiErrorAppend(errs, c.CreateSourceConfig.Prepare()...)
	errs = packersdk.MultiErrorAppend(errs, c.WatchSourceConfig.Prepare()...)

	// Verify that SSH communicator is used for connecting to the source VM.
	// This must call after the CommunicatorConfig.Prepare to get the value properly.
	if c.CommunicatorConfig.Type != "ssh" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("only SSH communicator is supported"))
	}

	if len(errs.Errors) > 0 {
		return nil, errs
	}

	return nil, nil
}

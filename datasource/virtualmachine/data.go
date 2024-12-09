// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Tag,DatasourceOutput
package virtualmachine

import (
	"errors"
	"fmt"

	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	vsphere "github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common"
	"github.com/hashicorp/packer-plugin-vsphere/datasource/common/driver"
	"github.com/zclconf/go-cty/cty"
)

// HCL Example:
//
// ```hcl
//
//	tag {
//	  category = "team"
//	  name = "operations"
//	}
//	tag {
//	  category = "sla"
//	  name = "gold"
//	}
//
// ```
type Tag struct {
	// Name of the tag added to virtual machine which must pass the `tag`
	// filter.
	Name string `mapstructure:"name" required:"true"`
	// Name of the tag category that contains the tag.
	//
	// -> **Note:** Both `name` and `category` must be specified in the `tag`
	// filter.
	Category string `mapstructure:"category" required:"true"`
}

type Config struct {
	common.PackerConfig   `mapstructure:",squash"`
	vsphere.ConnectConfig `mapstructure:",squash"`

	// Basic filter with glob support (e.g. `ubuntu_basic*`). Defaults to `*`.
	// Using strict globs will not reduce execution time because vSphere API
	// returns the full inventory. But can be used for better readability over
	// regular expressions.
	Name string `mapstructure:"name"`
	// Extended name filter with regular expressions support
	// (e.g. `ubuntu[-_]basic[0-9]*`). Default is empty. The match of the
	// regular expression is checked by substring. Use `^` and `$` to define a
	// full string. For example, the `^[^_]+$` filter will search names
	// without any underscores. The expression must use
	// [Go Regex Syntax](https://pkg.go.dev/regexp/syntax).
	NameRegex string `mapstructure:"name_regex"`
	// Filter to return only objects that are virtual machine templates.
	// Defaults to `false` and returns all virtual machines.
	Template bool `mapstructure:"template"`
	// Filter to search virtual machines only on the specified ESX host.
	Host string `mapstructure:"host"`
	// Filter to return only that virtual machines that have attached all
	// specifies tags. Specify one or more `tag` blocks to define list of tags
	//  for the filter. Multiple blocks can be created programmatically in HCL2 with the [`dynamic_block`](/packer/docs/templates/hcl_templates/expressions#dynamic-blocks).
	Tags []Tag `mapstructure:"tag"`
	// This filter determines how to handle multiple machines that were
	// matched with all previous filters. Machine creation time is being used
	// to find latest. By default, multiple matching machines results in an
	// error.
	Latest bool `mapstructure:"latest"`
}

type Datasource struct {
	config Config
}

type DatasourceOutput struct {
	// Name of the found virtual machine.
	VmName string `mapstructure:"vm_name"`
}

func (d *Datasource) ConfigSpec() hcldec.ObjectSpec {
	return d.config.FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Configure(raws ...interface{}) error {
	err := config.Decode(&d.config, nil, raws...)
	if err != nil {
		return err
	}

	if d.config.Name == "" {
		d.config.Name = "*"
	}

	var errs *packersdk.MultiError
	if d.config.VCenterServer == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("'vcenter_server' is required"))
	}
	if d.config.Username == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("'username' is required"))
	}
	if d.config.Password == "" {
		errs = packersdk.MultiErrorAppend(errs, errors.New("'password' is required"))
	}
	if len(d.config.Tags) > 0 {
		for _, tag := range d.config.Tags {
			if tag.Name == "" || tag.Category == "" {
				errs = packersdk.MultiErrorAppend(errs, errors.New("both name and category are required for tag"))
			}
		}
	}

	if errs != nil && len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (d *Datasource) OutputSpec() hcldec.ObjectSpec {
	return (&DatasourceOutput{}).FlatMapstructure().HCL2Spec()
}

func (d *Datasource) Execute() (cty.Value, error) {
	dr, err := driver.NewDriver(d.config.ConnectConfig)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to initialize driver: %w", err)
	}

	// This is the first level of filters
	// (the finder with glob will return filtered list or drop an error if found nothing).
	filteredVms, err := dr.Finder.VirtualMachineList(dr.Ctx, d.config.Name)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to retrieve virtual machines list: %w", err)
	}

	// Chain of other filters that will be executed only when defined
	// and previous filter in chain left some virtual machines in the list.
	if d.config.NameRegex != "" {
		filteredVms = filterByNameRegex(filteredVms, d.config.NameRegex)
	}

	if len(filteredVms) > 0 && d.config.Template {
		filteredVms, err = filterByTemplate(dr, filteredVms)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to filter by template attribute: %w", err)
		}
	}

	if len(filteredVms) > 0 && d.config.Host != "" {
		filteredVms, err = filterByHost(dr, d.config, filteredVms)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to filter by host attribute: %w", err)
		}
	}

	if len(filteredVms) > 0 && d.config.Tags != nil {
		filteredVms, err = filterByTags(dr, d.config.Tags, filteredVms)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to filter by tags: %w", err)
		}
	}

	// No VMs passed the filter chain. Nothing to return.
	if len(filteredVms) == 0 {
		return cty.NullVal(cty.EmptyObject), errors.New("no virtual machine matches the filters")
	}

	if len(filteredVms) > 1 {
		if d.config.Latest {
			filteredVms, err = filterByLatest(dr, filteredVms)
			if err != nil {
				return cty.NullVal(cty.EmptyObject), fmt.Errorf("failed to find the latest virtual machine: %w", err)
			}
		} else {
			// Too many machines passed the filter chain. Cannot decide which machine to return.
			return cty.NullVal(cty.EmptyObject), errors.New("more than one virtual machine matched the filters")
		}
	}

	output := DatasourceOutput{
		VmName: filteredVms[0].Name(),
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

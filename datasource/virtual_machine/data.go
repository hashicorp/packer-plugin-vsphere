// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type Config,Tag,DatasourceOutput
package virtual_machine

import (
	"github.com/hashicorp/hcl/v2/hcldec"
	"github.com/hashicorp/packer-plugin-sdk/common"
	"github.com/hashicorp/packer-plugin-sdk/hcl2helper"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	vsCommon "github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common"
	"github.com/pkg/errors"
	"github.com/zclconf/go-cty/cty"
)

// Example of multiple vm_tags blocks in HCL format:
// ```
//
//	vm_tags {
//	  category = "team"
//	  name = "operations"
//	}
//	vm_tags {
//	  category = "SLA"
//	  name = "gold"
//	}
//
// ```
type Tag struct {
	// Tag with this name must be attached to virtual machine which should pass the Tags Filter.
	Name string `mapstructure:"name" required:"true"`
	// Name of the category that contains this tag. Both tag and category must be specified.
	Category string `mapstructure:"category" required:"true"`
}

type Config struct {
	common.PackerConfig    `mapstructure:",squash"`
	vsCommon.ConnectConfig `mapstructure:",squash"`

	// Basic filter with glob support (e.g. `nginx_basic*`). Defaults to `*`.
	// Using strict globs will not reduce execution time because vSphere API returns the full inventory.
	// But can be used for better readability over regular expressions.
	Name string `mapstructure:"name"`
	// Extended name filter with regular expressions support (e.g. `nginx[-_]basic[0-9]*`). Default is empty.
	// The match of the regular expression is checked by substring. Use `^` and `$` to define a full string.
	// E.g. the `^[^_]+$` filter will search names without any underscores.
	// The expression must use [Go Regex Syntax](https://pkg.go.dev/regexp/syntax).
	NameRegex string `mapstructure:"name_regex"`
	// Filter to return only objects that are virtual machine templates.
	// Defaults to `false` and returns all VMs.
	Template bool `mapstructure:"template"`
	// Filter to search virtual machines only on the specified node.
	Node string `mapstructure:"node"`
	// Filter to return only that virtual machines that have attached all specifies tags.
	// Specify one or more `vm_tags` blocks to define list of tags that will make up the filter.
	// Should work since vCenter 6.7. To avoid incompatibility, REST client is being
	// initialized only when at least one tag has been defined in the config.
	VmTags []Tag `mapstructure:"vm_tags"`
	// This filter determines how to handle multiple machines that were matched with all
	// previous filters. Machine creation time is being used to find latest.
	// By default, multiple matching machines results in an error.
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
	if len(d.config.VmTags) > 0 {
		for _, tag := range d.config.VmTags {
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
	driver, err := newDriver(d.config)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), errors.Wrap(err, "failed to initialize driver")
	}

	// This is the first level of filters
	// (the finder with glob will return filtered list or drop an error if found nothing).
	filteredVms, err := driver.finder.VirtualMachineList(driver.ctx, d.config.Name)
	if err != nil {
		return cty.NullVal(cty.EmptyObject), errors.Wrap(err, "failed to retrieve virtual machines list")
	}

	// Chain of other filters that will be executed only when defined
	// and previous filter in chain left some virtual machines in the list.
	if d.config.NameRegex != "" {
		filteredVms = filterByNameRegex(filteredVms, d.config.NameRegex)
	}

	if len(filteredVms) > 0 && d.config.Template {
		filteredVms, err = filterByTemplate(driver, filteredVms)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), errors.Wrap(err, "failed to filter by template attribute")
		}
	}

	if len(filteredVms) > 0 && d.config.Node != "" {
		filteredVms, err = filterByNode(driver, d.config, filteredVms)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), errors.Wrap(err, "failed to filter by node attribute")
		}
	}

	if len(filteredVms) > 0 && d.config.VmTags != nil {
		filteredVms, err = filterByTags(driver, d.config.VmTags, filteredVms)
		if err != nil {
			return cty.NullVal(cty.EmptyObject), errors.Wrap(err, "failed to filter by tags")
		}
	}

	// No VMs passed the filter chain. Nothing to return.
	if len(filteredVms) == 0 {
		return cty.NullVal(cty.EmptyObject), errors.New("not a single VM matches the configured filters")
	}

	if len(filteredVms) > 1 {
		if d.config.Latest {
			filteredVms, err = filterByLatest(driver, filteredVms)
			if err != nil {
				return cty.NullVal(cty.EmptyObject), errors.Wrap(err, "failed to find the latest VM")
			}
		} else {
			// Too many machines passed the filter chain. Cannot decide which machine to return.
			return cty.NullVal(cty.EmptyObject), errors.New("multiple VMs match the configured filters")
		}
	}

	output := DatasourceOutput{
		VmName: filteredVms[0].Name(),
	}

	return hcl2helper.HCL2ValueFromConfig(output, d.OutputSpec()), nil
}

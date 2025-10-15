// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package virtualmachine

import (
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
)

// filterVms removes virtual machines from vmList that do not match the datasource config filters.
func filterVms(vmList []*object.VirtualMachine, c Config, d *driver.VCenterDriver) ([]*object.VirtualMachine, error) {
	filterFuncs := make([]func(*object.VirtualMachine) (bool, error), 0)

	if c.NameRegex != "" {
		re := regexp.MustCompile(c.NameRegex)
		filterFuncs = append(filterFuncs, func(vm *object.VirtualMachine) (bool, error) {
			return re.MatchString(vm.Name()), nil
		})
	}

	if c.Template {
		filterFuncs = append(filterFuncs, func(vm *object.VirtualMachine) (bool, error) {
			isTemplate, err := vm.IsTemplate(d.Ctx)
			if err != nil {
				return false, fmt.Errorf("error checking if virtual machine is a template: %w", err)
			}
			return isTemplate, nil
		})
	}

	if c.Host != "" {
		hostVms, err := getHostVms(d, c.Host)
		if err != nil {
			return nil, err
		}

		filterFuncs = append(filterFuncs, func(vm *object.VirtualMachine) (bool, error) {
			vmName := vm.Name()
			for _, hostVm := range hostVms {
				if vmName == hostVm.Name {
					return true, nil
				}
			}
			return false, nil
		})
	}

	if c.Tags != nil {
		filterFuncs = append(filterFuncs, func(vm *object.VirtualMachine) (bool, error) {
			result, err := configTagsMatchHostTags(d, vm, c.Tags)
			if err != nil {
				return false, err
			}
			return result, nil
		})
	}

	result := make([]*object.VirtualMachine, 0)
	for _, vm := range vmList {
		var ok bool
		var err error
		if len(filterFuncs) == 0 {
			ok = true
		}
		for _, vmPassedFilter := range filterFuncs {
			ok, err = vmPassedFilter(vm)
			if err != nil {
				return nil, fmt.Errorf("failed to filter vm: %w", err)
			}
			if !ok {
				break
			}
		}
		if ok {
			result = append(result, vm)
		}
	}

	return result, nil
}

// findLatestVM returns the most recently created virtual machine from vmList.
func findLatestVM(d *driver.VCenterDriver, vmList []*object.VirtualMachine) ([]*object.VirtualMachine, error) {
	var latestVM *object.VirtualMachine
	var latestTimestamp time.Time
	for _, elementVM := range vmList {
		var vmConfig mo.VirtualMachine
		err := elementVM.Properties(d.Ctx, elementVM.Reference(), []string{"config"}, &vmConfig)
		if err != nil {
			return nil, fmt.Errorf("error retrieving config properties for the virtual machine: %w", err)
		}
		if vmConfig.Config.CreateDate.After(latestTimestamp) {
			latestVM = elementVM
			latestTimestamp = *vmConfig.Config.CreateDate
		}
	}
	result := []*object.VirtualMachine{latestVM}
	return result, nil
}

// getHostVms retrieves all virtual machines on the specified host.
func getHostVms(d *driver.VCenterDriver, hostName string) ([]mo.VirtualMachine, error) {
	pc := property.DefaultCollector(d.Client.Client)
	obj, err := d.Finder.HostSystem(d.Ctx, hostName)
	if err != nil {
		return nil, fmt.Errorf("error finding defined host system: %w", err)
	}

	var host mo.HostSystem
	err = pc.RetrieveOne(d.Ctx, obj.Reference(), []string{"vm"}, &host)
	if err != nil {
		return nil, fmt.Errorf("error retrieving properties of host system: %w", err)
	}

	var hostVms []mo.VirtualMachine
	err = pc.Retrieve(d.Ctx, host.Vm, []string{"name"}, &hostVms)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties for the virtual machine: %w", err)
	}
	return hostVms, nil
}

// configTagsMatchHostTags returns true if the virtual machine has all tags specified in tagList.
func configTagsMatchHostTags(d *driver.VCenterDriver, vm *object.VirtualMachine, tagList []Tag) (bool, error) {
	err := d.RestClient.Login(d.Ctx)
	if err != nil {
		return false, fmt.Errorf("failed to login to REST API: %w", err)
	}

	tagMan := tags.NewManager(d.RestClient.Client())
	realTagsList, err := tagMan.GetAttachedTags(d.Ctx, vm.Reference())
	if err != nil {
		return false, fmt.Errorf("failed return tags for the virtual machine: %w", err)
	}
	matchedTagsCount := 0
	for _, configTag := range tagList {
		configTagMatched := false
		for _, realTag := range realTagsList {
			if configTag.Name == realTag.Name {
				category, err := tagMan.GetCategory(d.Ctx, realTag.CategoryID)
				if err != nil {
					return false, fmt.Errorf("failed to return tag category for tag: %w", err)
				}
				if configTag.Category == category.Name {
					configTagMatched = true
					break
				}
			}
		}
		if configTagMatched {
			matchedTagsCount++
		} else {
			break
		}
	}
	return matchedTagsCount == len(tagList), nil
}

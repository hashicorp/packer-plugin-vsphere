// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package virtualmachine

import (
	"fmt"
	"regexp"
	"time"

	"github.com/hashicorp/packer-plugin-vsphere/datasource/common/driver"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
)

// filterByNameRegex filters machines by matching their names against defined regular expression.
func filterByNameRegex(vmList []*object.VirtualMachine, nameRegex string) []*object.VirtualMachine {
	re, _ := regexp.Compile(nameRegex)
	result := make([]*object.VirtualMachine, 0)
	for _, i := range vmList {
		if re.MatchString(i.Name()) {
			result = append(result, i)
		}
	}
	return result
}

// filterByTemplate filters machines by template attribute. Only templates will pass the filter.
func filterByTemplate(driver *driver.VCenterDriver, vmList []*object.VirtualMachine) ([]*object.VirtualMachine, error) {
	result := make([]*object.VirtualMachine, 0)
	for _, i := range vmList {
		isTemplate, err := i.IsTemplate(driver.Ctx)
		if err != nil {
			return nil, fmt.Errorf("error checking if virtual machine is a template: %w", err)
		}

		if isTemplate {
			result = append(result, i)
		}
	}
	return result, nil
}

// filterByHost filters machines by ESX host placement.
// Only machines that are stored on the defined host will pass the filter.
func filterByHost(driver *driver.VCenterDriver, config Config, vmList []*object.VirtualMachine) ([]*object.VirtualMachine, error) {
	pc := property.DefaultCollector(driver.Client.Client)
	obj, err := driver.Finder.HostSystem(driver.Ctx, config.Host)
	if err != nil {
		return nil, fmt.Errorf("error finding defined host system: %w", err)
	}

	var host mo.HostSystem
	err = pc.RetrieveOne(driver.Ctx, obj.Reference(), []string{"vm"}, &host)
	if err != nil {
		return nil, fmt.Errorf("error retrieving properties of host system: %w", err)
	}

	var hostVms []mo.VirtualMachine
	err = pc.Retrieve(driver.Ctx, host.Vm, []string{"name"}, &hostVms)
	if err != nil {
		return nil, fmt.Errorf("failed to get properties for the virtual machine: %w", err)
	}

	result := make([]*object.VirtualMachine, 0)
	for _, filteredVm := range vmList {
		vmName := filteredVm.Name()
		for _, hostVm := range hostVms {
			if vmName == hostVm.Name {
				result = append(result, filteredVm)
			}
		}
	}

	return result, nil
}

// filterByTags filters machines by tags. Only machines that has all the tags from list will pass the filter.
func filterByTags(driver *driver.VCenterDriver, vmTags []Tag, vmList []*object.VirtualMachine) ([]*object.VirtualMachine, error) {
	result := make([]*object.VirtualMachine, 0)
	tagMan := tags.NewManager(driver.RestClient)
	for _, filteredVm := range vmList {
		realTagsList, err := tagMan.GetAttachedTags(driver.Ctx, filteredVm.Reference())
		if err != nil {
			return nil, fmt.Errorf("failed return tags for the virtual machine: %w", err)
		}
		matchedTagsCount := 0
		for _, configTag := range vmTags {
			configTagMatched := false
			for _, realTag := range realTagsList {
				if configTag.Name == realTag.Name {
					category, err := tagMan.GetCategory(driver.Ctx, realTag.CategoryID)
					if err != nil {
						return nil, fmt.Errorf("failed to return tag category for tag: %w", err)
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
				// If a single requested tag from config not matched then no need to proceed.
				// Fail early.
				break
			}
		}
		if matchedTagsCount == len(vmTags) {
			result = append(result, filteredVm)
		}
	}

	return result, nil
}

// filterByLatest filters machines by creation date. This filter returns list with one element.
func filterByLatest(driver *driver.VCenterDriver, vmList []*object.VirtualMachine) ([]*object.VirtualMachine, error) {
	var latestVM *object.VirtualMachine
	var latestTimestamp time.Time
	for _, elementVM := range vmList {
		var vmConfig mo.VirtualMachine
		err := elementVM.Properties(driver.Ctx, elementVM.Reference(), []string{"config"}, &vmConfig)
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

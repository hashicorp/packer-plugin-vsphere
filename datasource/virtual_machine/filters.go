package virtual_machine

import (
	"regexp"
	"time"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/tags"
	"github.com/vmware/govmomi/vim25/mo"
)

// Filter machines by matching their names against defined regular expression.
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

// Filter machines by template attribute. Only templates will pass the filter.
func filterByTemplate(driver *VCenterDriver, vmList []*object.VirtualMachine) ([]*object.VirtualMachine, error) {
	result := make([]*object.VirtualMachine, 0)
	for _, i := range vmList {
		isTemplate, err := i.IsTemplate(driver.ctx)
		if err != nil {
			return nil, errors.Wrap(err, "error checking if VM is a tempalte")
		}

		if isTemplate {
			result = append(result, i)
		}
	}
	return result, nil
}

// Filter machines by node placement. Only machines that are stored on the defined node will pass the filter.
func filterByNode(driver *VCenterDriver, config Config, vmList []*object.VirtualMachine) ([]*object.VirtualMachine, error) {
	pc := property.DefaultCollector(driver.client.Client)
	obj, err := driver.finder.HostSystem(driver.ctx, config.Node)
	if err != nil {
		return nil, errors.Wrap(err, "error finding defined host system")
	}

	var host mo.HostSystem
	err = pc.RetrieveOne(driver.ctx, obj.Reference(), []string{"vm"}, &host)
	if err != nil {
		return nil, errors.Wrap(err, "error retrieving properties of host system")
	}

	var nodeVms []mo.VirtualMachine
	err = pc.Retrieve(driver.ctx, host.Vm, []string{"name"}, &nodeVms)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get properties for VM")
	}

	result := make([]*object.VirtualMachine, 0)
	for _, filteredVm := range vmList {
		vmName := filteredVm.Name()
		for _, nodeVm := range nodeVms {
			if vmName == nodeVm.Name {
				result = append(result, filteredVm)
			}
		}
	}

	return result, nil
}

// Filter machines by tags. Tags are stored in the driver as list of flatTag elements.
// Only machines that has all the tags from list will pass the filter.
func filterByTags(driver *VCenterDriver, vmTags []Tag, vmList []*object.VirtualMachine) ([]*object.VirtualMachine, error) {
	result := make([]*object.VirtualMachine, 0)
	tagMan := tags.NewManager(driver.restClient)
	for _, filteredVm := range vmList {
		realTagsList, err := tagMan.GetAttachedTags(driver.ctx, filteredVm.Reference())
		if err != nil {
			return nil, errors.Wrap(err, "failed to get attached tags for vm")
		}
		matchedTagsCount := 0
		for _, configTag := range vmTags {
			configTagMatched := false
			for _, realTag := range realTagsList {
				if configTag.Name == realTag.Name {
					category, err := tagMan.GetCategory(driver.ctx, realTag.CategoryID)
					if err != nil {
						return nil, errors.Wrap(err, "failed to get attached category for tag")
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

func filterByLatest(driver *VCenterDriver, vmList []*object.VirtualMachine) ([]*object.VirtualMachine, error) {
	var latestVM *object.VirtualMachine
	var latestTimestamp time.Time
	for _, elementVM := range vmList {
		var vmConfig mo.VirtualMachine
		err := elementVM.Properties(driver.ctx, elementVM.Reference(), []string{"config"}, &vmConfig)
		if err != nil {
			return nil, errors.Wrap(err, "error retrieving config properties of VM")
		}
		if vmConfig.Config.CreateDate.After(latestTimestamp) {
			latestVM = elementVM
			latestTimestamp = *vmConfig.Config.CreateDate
		}
	}
	result := []*object.VirtualMachine{latestVM}
	return result, nil
}

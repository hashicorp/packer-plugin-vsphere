// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testing

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/tags"
)

// MarkSimulatedVmAsTemplate powers off the virtual machine before converting it to a template (because the simulator
// creates all virtual machines in an online state).
func MarkSimulatedVmAsTemplate(ctx context.Context, vm *object.VirtualMachine) error {
	task, err := vm.PowerOff(ctx)
	if err != nil {
		return fmt.Errorf("failed to issue powering off command to the machine: %w", err)
	}
	err = task.Wait(ctx)
	if err != nil {
		return fmt.Errorf("failed to power off the machine: %w", err)
	}
	err = vm.MarkAsTemplate(ctx)
	if err != nil {
		return fmt.Errorf("failed to mark virtual machine as a template: %w", err)
	}
	return nil
}

// FindOrCreateCategory tries to find category passed by name, creates category if not found and returns category ID.
// Category will be created with "MULTIPLE" constraint.
func FindOrCreateCategory(ctx context.Context, man *tags.Manager, catName string) (string, error) {
	categoryList, err := man.GetCategories(ctx)
	if err != nil {
		return "", fmt.Errorf("cannot return categories from cluster: %w", err)
	}
	for _, category := range categoryList {
		if category.Name == catName {
			return category.ID, nil
		}
	}
	newCategoryID, err := man.CreateCategory(ctx, &tags.Category{Name: catName, Cardinality: "MULTIPLE"})
	if err != nil {
		return "", fmt.Errorf("cannot create category: %w", err)
	}
	return newCategoryID, nil
}

// FindOrCreateTag tries to find the tagName in category with catID, creates if not found and returns tag ID.
func FindOrCreateTag(ctx context.Context, man *tags.Manager, catID string, tagName string) (string, error) {
	tagsInCategory, err := man.GetTagsForCategory(ctx, catID)
	if err != nil {
		return "", fmt.Errorf("cannot return tags for category: %w", err)
	}
	for _, tag := range tagsInCategory {
		if tag.Name == tagName {
			return tag.ID, nil
		}
	}
	newTagID, err := man.CreateTag(ctx, &tags.Tag{Name: tagName, CategoryID: catID})
	if err != nil {
		return "", fmt.Errorf("cannot create tag: %w", err)
	}
	return newTagID, nil
}

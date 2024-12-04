package testing

import (
	"context"

	"github.com/pkg/errors"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vapi/tags"
)

func MarkSimulatedVmAsTemplate(ctx context.Context, vm *object.VirtualMachine) error {
	task, err := vm.PowerOff(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to issue powering off command to the machine")
	}
	err = task.Wait(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to power off the machine")
	}
	err = vm.MarkAsTemplate(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to mark VM as a template")
	}
	return nil
}

// Try to find category passed by name, create category if not found and return category ID.
// Category will be created with "MULTIPLE" constraint.
func FindOrCreateCategory(ctx context.Context, man *tags.Manager, catName string) (string, error) {
	categoryList, err := man.GetCategories(ctx)
	if err != nil {
		return "", errors.Wrap(err, "cannot get categories from cluster")
	}
	for _, category := range categoryList {
		if category.Name == catName {
			return category.ID, nil
		}
	}
	newCategoryID, err := man.CreateCategory(ctx, &tags.Category{Name: catName, Cardinality: "MULTIPLE"})
	if err != nil {
		return "", errors.Wrap(err, "cannot create category")
	}
	return newCategoryID, nil
}

// Try to find the tagName in category with catID, create if not found and return tag ID.
func FindOrCreateTag(ctx context.Context, man *tags.Manager, catID string, tagName string) (string, error) {
	tagsInCategory, err := man.GetTagsForCategory(ctx, catID)
	if err != nil {
		return "", errors.Wrap(err, "cannot get tags for category")
	}
	for _, tag := range tagsInCategory {
		if tag.Name == tagName {
			return tag.ID, nil
		}
	}
	newTagID, err := man.CreateTag(ctx, &tags.Tag{Name: tagName, CategoryID: catID})
	if err != nil {
		return "", errors.Wrap(err, "cannot create tag")
	}
	return newTagID, nil
}

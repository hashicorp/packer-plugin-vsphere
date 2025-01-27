// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"fmt"
	"path"
	"strings"

	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type Folder struct {
	driver *VCenterDriver
	folder *object.Folder
}

func (d *VCenterDriver) NewFolder(ref *types.ManagedObjectReference) *Folder {
	return &Folder{
		folder: object.NewFolder(d.client.Client, *ref),
		driver: d,
	}
}

// FindFolder locates or creates a folder structure within a datastore based
// on the provided folder name. Returns a pointer to the Folder object or an
// error if the operation fails.
func (d *VCenterDriver) FindFolder(name string) (*Folder, error) {
	if name != "" {
		// If the folder does not exist, create it.
		parent := ""
		parentFolder, err := d.finder.Folder(d.ctx, path.Join(d.datacenter.InventoryPath, "vm"))
		if err != nil {
			return nil, err
		}
		folders := strings.Split(name, "/")
		for _, folder := range folders {
			parent = path.Join(parent, folder)
			f, err := d.finder.Folder(d.ctx, path.Join(d.datacenter.InventoryPath, "vm", parent))
			if _, ok := err.(*find.NotFoundError); ok {
				f, err = parentFolder.CreateFolder(d.ctx, folder)
			}
			if err != nil {
				return nil, err
			}
			parentFolder = f
		}
	}

	f, err := d.finder.Folder(d.ctx, path.Join(d.datacenter.InventoryPath, "vm", name))
	if err != nil {
		return nil, err
	}

	return &Folder{
		folder: f,
		driver: d,
	}, nil
}

// Info retrieves properties of the folder object with optional filters specified
// as parameters. If no parameters are provided, all properties are returned.
func (f *Folder) Info(params ...string) (*mo.Folder, error) {
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var info mo.Folder
	err := f.folder.Properties(f.driver.ctx, f.folder.Reference(), p, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// Path constructs and returns the full hierarchical path of the folder,
// starting from the datacenter level or an error.
func (f *Folder) Path() (string, error) {
	info, err := f.Info("name", "parent")
	if err != nil {
		return "", err
	}
	if info.Parent.Type == "Datacenter" {
		return "", nil
	} else {
		parent := f.driver.NewFolder(info.Parent)
		path, err := parent.Path()
		if err != nil {
			return "", err
		}
		if path == "" {
			return info.Name, nil
		} else {
			return fmt.Sprintf("%v/%v", path, info.Name), nil
		}
	}
}

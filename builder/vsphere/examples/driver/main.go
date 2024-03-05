// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

func main() {
	d, err := driver.NewDriver(&driver.ConnectConfig{
		VCenterServer:      "vcenter.example.com",
		Username:           "root",
		Password:           "jetbrains",
		InsecureConnection: true,
	})
	if err != nil {
		panic(err)
	}

	ds, err := d.FindDatastore("", "esxi-01.example.com")
	if err != nil {
		panic(err)
	}

	fmt.Println(ds.Name())
}

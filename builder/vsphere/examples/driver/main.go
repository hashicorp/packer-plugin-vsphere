// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"
	"os"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

func main() {
	username := os.Getenv("VSPHERE_USERNAME")
	if username == "" {
		username = "administrator@vsphere.local"
	}

	password := os.Getenv("VSPHERE_PASSWORD")
	if password == "" {
		password = "VMw@re1!"
	}

	vcenter := os.Getenv("VSPHERE_VCENTER_SERVER")
	if vcenter == "" {
		vcenter = "vcenter.example.com"
	}

	host := os.Getenv("VSPHERE_HOST")
	if host == "" {
		host = "esxi-01.example.com"
	}

	d, err := driver.NewDriver(&driver.ConnectConfig{
		VCenterServer:      vcenter,
		Username:           username,
		Password:           password,
		InsecureConnection: true,
	})
	if err != nil {
		panic(err)
	}

	ds, err := d.FindDatastore("", host)
	if err != nil {
		panic(err)
	}

	fmt.Println(ds.Name())
}

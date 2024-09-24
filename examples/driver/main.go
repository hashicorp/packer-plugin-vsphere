// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"fmt"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common/utils"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

func main() {
	vcenter := utils.GetenvOrDefault(utils.EnvVcenterServer, utils.DefaultVcenterServer)
	username := utils.GetenvOrDefault(utils.EnvVsphereUsername, utils.DefaultVsphereUsername)
	password := utils.GetenvOrDefault(utils.EnvVspherePassword, utils.DefaultVspherePassword)
	host := utils.GetenvOrDefault(utils.EnvVsphereHost, utils.DefaultVsphereHost)

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

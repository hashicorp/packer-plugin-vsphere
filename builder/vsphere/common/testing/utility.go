// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package testing

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common/utils"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

func NewVMName() string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("test-%v", r.Intn(1000))
}

func RenderConfig(builderType string, config map[string]interface{}) string {
	t := map[string][]map[string]interface{}{
		"builders": {
			{"type": builderType},
		},
	}
	for k, v := range config {
		t["builders"][0][k] = v
	}

	j, _ := json.Marshal(t)
	return string(j)
}

func TestConn() (driver.Driver, error) {
	vcenter := utils.GetenvOrDefault(utils.EnvVcenterServer, utils.DefaultVcenterServer)
	username := utils.GetenvOrDefault(utils.EnvVsphereUsername, utils.DefaultVsphereUsername)
	password := utils.GetenvOrDefault(utils.EnvVspherePassword, utils.DefaultVspherePassword)

	d, err := driver.NewDriver(&driver.ConnectConfig{
		VCenterServer:      vcenter,
		Username:           username,
		Password:           password,
		InsecureConnection: true,
	})
	if err != nil {
		return nil, fmt.Errorf("error connecting to endpoint: %v", err)
	}
	return d, nil
}

func GetVM(d driver.Driver, name string) (driver.VirtualMachine, error) {
	vm, err := d.FindVM(name)
	if err != nil {
		return nil, fmt.Errorf("error finding virtual machine: %v", err)
	}
	return vm, nil
}

func CleanupVM(d driver.Driver, name string) error {
	vm, err := GetVM(d, name)
	if err != nil {
		return fmt.Errorf("error finding virtual machine: %v", err)
	}
	return vm.Destroy()
}

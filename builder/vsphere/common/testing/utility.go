package testing

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
)

func NewVMName() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("test-%v", rand.Intn(1000))
}

func RenderConfig(builderType string, config map[string]interface{}) string {
	t := map[string][]map[string]interface{}{
		"builders": {
			map[string]interface{}{
				"type": builderType,
			},
		},
	}
	for k, v := range config {
		t["builders"][0][k] = v
	}

	j, _ := json.Marshal(t)
	return string(j)
}

func TestConn() (driver.Driver, error) {
	username := os.Getenv("VSPHERE_USERNAME")
	if username == "" {
		username = "root"
	}
	password := os.Getenv("VSPHERE_PASSWORD")
	if password == "" {
		password = "jetbrains"
	}

	d, err := driver.NewDriver(&driver.ConnectConfig{
		VCenterServer:      "vcenter.vsphere65.test",
		Username:           username,
		Password:           password,
		InsecureConnection: true,
	})
	if err != nil {
		return nil, fmt.Errorf("Cannot connect: ", err)
	}
	return d, nil
}

func GetVM(d driver.Driver, name string) (driver.VirtualMachine, error) {
	vm, err := d.FindVM(name)
	if err != nil {
		return nil, fmt.Errorf("Cannot find VM: %v", err)
	}

	return vm, nil
}

func CleanupVM(d driver.Driver, name string) error {
	vm, err := GetVM(d, name)
	if err != nil {
		return fmt.Errorf("Cannot find VM: %v", err)
	}
	return vm.Destroy()
}

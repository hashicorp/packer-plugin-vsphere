// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iso

import (
	"fmt"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	commonT "github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common/testing"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common/utils"
	"github.com/vmware/govmomi/vim25/types"
)

func TestAccISOBuilderAcc_default(t *testing.T) {
	config := defaultConfig()
	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_basic_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return checkDefault(config["vm_name"].(string), config["host"].(string), "datastore1")
		},
	}
	acctest.TestPlugin(t, testCase)
}

func defaultConfig() map[string]interface{} {
	vcenter := utils.GetenvOrDefault(utils.EnvVcenterServer, utils.DefaultVcenterServer)
	username := utils.GetenvOrDefault(utils.EnvVsphereUsername, utils.DefaultVsphereUsername)
	password := utils.GetenvOrDefault(utils.EnvVspherePassword, utils.DefaultVspherePassword)
	host := utils.GetenvOrDefault(utils.EnvVsphereHost, utils.DefaultVsphereHost)

	config := map[string]interface{}{
		"vcenter_server":      vcenter,
		"username":            username,
		"password":            password,
		"host":                host,
		"insecure_connection": true,

		"ssh_username": "packer",
		"ssh_password": "VMw@re1!",

		"vm_name": commonT.NewVMName(),
		"storage": map[string]interface{}{
			"disk_size": 2048,
		},

		"communicator": "none", // do not start the VM without any bootable devices
	}

	return config
}

func checkDefault(name string, host string, datastore string) error {
	d, err := commonT.TestConn()
	if err != nil {
		return fmt.Errorf("cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("cannot find VM: %v", err)
	}

	vmInfo, err := vm.Info("name", "parent", "runtime.host", "resourcePool", "datastore", "layoutEx.disk", "config.firmware")
	if err != nil {
		return fmt.Errorf("cannot read VM properties: %v", err)
	}

	if vmInfo.Name != name {
		return fmt.Errorf("unexpected virtual machine name: expected '%v', but returned '%v'", name, vmInfo.Name)
	}

	f := d.NewFolder(vmInfo.Parent)
	folderPath, err := f.Path()
	if err != nil {
		return fmt.Errorf("cannot read folder name: %v", err)
	}
	if folderPath != "" {
		return fmt.Errorf("unexpected folder: expected '/', but returned '%v'", folderPath)
	}

	h := d.NewHost(vmInfo.Runtime.Host)
	hostInfo, err := h.Info("name")
	if err != nil {
		return fmt.Errorf("cannot read host properties: %#v", err)
	}
	if hostInfo.Name != host {
		return fmt.Errorf("unexpected host name: expected '%v', but returned '%v'", host, hostInfo.Name)
	}

	p := d.NewResourcePool(vmInfo.ResourcePool)
	poolPath, err := p.Path()
	if err != nil {
		return fmt.Errorf("cannot read resource pool name: %v", err)
	}
	if poolPath != "" {
		return fmt.Errorf("unexpected resource pool: expected '/', but returned '%v'", poolPath)
	}

	dsr := vmInfo.Datastore[0].Reference()
	ds := d.NewDatastore(&dsr)
	dsInfo, err := ds.Info("name")
	if err != nil {
		return fmt.Errorf("cannot read datastore properties: %#v", err)
	}
	if dsInfo.Name != datastore {
		return fmt.Errorf("unexpected datastore name: expected '%v', but returned '%v'", datastore, dsInfo.Name)
	}

	fw := vmInfo.Config.Firmware
	if fw != "bios" {
		return fmt.Errorf("unexpected firmware: expected 'bios', but returned '%v'", fw)
	}
	return nil
}

func TestAccISOBuilderAcc_notes(t *testing.T) {
	config := defaultConfig()
	config["notes"] = "test"

	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_notes_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return checkNotes(config["vm_name"].(string))
		},
	}
	acctest.TestPlugin(t, testCase)
}

func checkNotes(name string) error {
	d, err := commonT.TestConn()
	if err != nil {
		return fmt.Errorf("cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("cannot find VM: %v", err)
	}
	vmInfo, err := vm.Info("config.annotation")
	if err != nil {
		return fmt.Errorf("cannot read VM properties: %v", err)
	}

	notes := vmInfo.Config.Annotation
	if notes != "test" {
		return fmt.Errorf("notes should be 'test'")
	}

	return nil
}

func TestAccISOBuilderAcc_hardware(t *testing.T) {
	config := defaultConfig()
	config["CPUs"] = 2
	config["cpu_cores"] = 2
	config["CPU_reservation"] = 1000
	config["CPU_limit"] = 1500
	config["RAM"] = 2048
	config["RAM_reservation"] = 1024
	config["NestedHV"] = true
	config["firmware"] = "efi"
	config["video_ram"] = 8192

	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_hardware_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return checkHardware(config["vm_name"].(string))
		},
	}
	acctest.TestPlugin(t, testCase)
}

func checkHardware(name string) error {
	d, err := commonT.TestConn()
	if err != nil {
		return fmt.Errorf("cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("cannot find VM: %v", err)
	}
	vmInfo, err := vm.Info("config")
	if err != nil {
		return fmt.Errorf("cannot read VM properties: %v", err)
	}

	cpuSockets := vmInfo.Config.Hardware.NumCPU
	if cpuSockets != 2 {
		return fmt.Errorf("VM should have 2 CPU sockets, but returned %v", cpuSockets)
	}

	cpuCores := vmInfo.Config.Hardware.NumCoresPerSocket
	if cpuCores != 2 {
		return fmt.Errorf("VM should have 2 CPU cores per socket, but returned %v", cpuCores)
	}

	cpuReservation := *vmInfo.Config.CpuAllocation.Reservation
	if cpuReservation != 1000 {
		return fmt.Errorf("VM should have CPU reservation for 1000 Mhz, but returned %v", cpuReservation)
	}

	cpuLimit := *vmInfo.Config.CpuAllocation.Limit
	if cpuLimit != 1500 {
		return fmt.Errorf("VM should have CPU reservation for 1500 Mhz, but returned %v", cpuLimit)
	}

	ram := vmInfo.Config.Hardware.MemoryMB
	if ram != 2048 {
		return fmt.Errorf("VM should have 2048 MB of RAM, but returned %v", ram)
	}

	ramReservation := *vmInfo.Config.MemoryAllocation.Reservation
	if ramReservation != 1024 {
		return fmt.Errorf("VM should have RAM reservation for 1024 MB, but returned %v", ramReservation)
	}

	nestedHV := vmInfo.Config.NestedHVEnabled
	if !*nestedHV {
		return fmt.Errorf("VM should have NestedHV enabled, but returned %v", nestedHV)
	}

	fw := vmInfo.Config.Firmware
	if fw != "efi" {
		return fmt.Errorf("unexpected firmware: expected 'efi', but returned '%v'", fw)
	}

	l, err := vm.Devices()
	if err != nil {
		return fmt.Errorf("cannot read VM devices: %v", err)
	}
	c := l.PickController((*types.VirtualIDEController)(nil))
	if c == nil {
		return fmt.Errorf("VM should have IDE controller")
	}
	s := l.PickController((*types.VirtualAHCIController)(nil))
	if s != nil {
		return fmt.Errorf("VM should have no SATA controllers")
	}

	v := l.SelectByType((*types.VirtualMachineVideoCard)(nil))
	if len(v) != 1 {
		return fmt.Errorf("virtual machine should have one video card")
	}
	if v[0].(*types.VirtualMachineVideoCard).VideoRamSizeInKB != 8192 {
		return fmt.Errorf("video memory should be equal 8192")
	}

	return nil
}

func TestAccISOBuilderAcc_limit(t *testing.T) {
	config := defaultConfig()
	config["CPUs"] = 1 // hardware is customized, but CPU limit is not specified explicitly

	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_limit_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return checkLimit(config["vm_name"].(string))
		},
	}
	acctest.TestPlugin(t, testCase)
}

func checkLimit(name string) error {
	d, err := commonT.TestConn()
	if err != nil {
		return fmt.Errorf("cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("cannot find VM: %v", err)
	}
	vmInfo, err := vm.Info("config.cpuAllocation")
	if err != nil {
		return fmt.Errorf("cannot read VM properties: %v", err)
	}

	limit := *vmInfo.Config.CpuAllocation.Limit
	if limit != -1 { // must be unlimited
		return fmt.Errorf("unexpected CPU limit: expected '%v', but returned '%v'", -1, limit)
	}

	return nil
}

func TestAccISOBuilderAcc_sata(t *testing.T) {
	config := defaultConfig()
	config["cdrom_type"] = "sata"

	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_sata_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return checkSata(config["vm_name"].(string))
		},
	}
	acctest.TestPlugin(t, testCase)
}

func checkSata(name string) error {
	d, err := commonT.TestConn()
	if err != nil {
		return fmt.Errorf("cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("cannot find VM: %v", err)
	}

	l, err := vm.Devices()
	if err != nil {
		return fmt.Errorf("cannot read VM devices: %v", err)
	}

	c := l.PickController((*types.VirtualAHCIController)(nil))
	if c == nil {
		return fmt.Errorf("vm has no SATA controllers")
	}

	return nil
}

func TestAccISOBuilderAcc_cdrom(t *testing.T) {
	config := defaultConfig()
	config["iso_paths"] = []string{
		"[datastore1] test0.iso",
		"[datastore1] test1.iso",
	}
	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_cdrom_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccISOBuilderAcc_networkCard(t *testing.T) {
	config := defaultConfig()
	config["network_adapters"] = map[string]interface{}{
		"network_card": "vmxnet3",
	}
	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_networkCard_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return checkNetworkCard(config["vm_name"].(string))
		},
	}
	acctest.TestPlugin(t, testCase)
}

func checkNetworkCard(name string) error {
	d, err := commonT.TestConn()
	if err != nil {
		return fmt.Errorf("cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("cannot find VM: %v", err)
	}
	devices, err := vm.Devices()
	if err != nil {
		return fmt.Errorf("cannot read VM properties: %v", err)
	}

	netCards := devices.SelectByType((*types.VirtualEthernetCard)(nil))
	if len(netCards) == 0 {
		return fmt.Errorf("cannot find the network card")
	}
	if len(netCards) > 1 {
		return fmt.Errorf("found more than one network card")
	}
	if _, ok := netCards[0].(*types.VirtualVmxnet3); !ok {
		return fmt.Errorf("unexpected network card type: %s", netCards[0])
	}

	return nil
}

func TestAccISOBuilderAcc_createFloppy(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "packer-vsphere-iso-test")
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	_, err = fmt.Fprint(tmpFile, "Hello, World!")
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	err = tmpFile.Close()
	if err != nil {
		t.Fatalf("unexpected error: '%s'", err)
	}
	config := defaultConfig()
	config["floppy_files"] = []string{tmpFile.Name()}
	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_createFloppy_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("unexpected error: expected 'nil', but returned '%s'", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestAccISOBuilderAcc_full(t *testing.T) {
	config := fullConfig()
	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_full_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return checkFull(config["vm_name"].(string))
		},
	}
	acctest.TestPlugin(t, testCase)
}

func fullConfig() map[string]interface{} {
	vcenter := utils.GetenvOrDefault(utils.EnvVcenterServer, utils.DefaultVcenterServer)
	username := utils.GetenvOrDefault(utils.EnvVsphereUsername, utils.DefaultVsphereUsername)
	password := utils.GetenvOrDefault(utils.EnvVspherePassword, utils.DefaultVspherePassword)
	host := utils.GetenvOrDefault(utils.EnvVsphereHost, utils.DefaultVsphereHost)

	config := map[string]interface{}{
		"vcenter_server":      vcenter,
		"username":            username,
		"password":            password,
		"host":                host,
		"insecure_connection": true,

		"vm_name": commonT.NewVMName(),

		"RAM": 512,
		"disk_controller_type": []string{
			"pvscsi",
		},
		"storage": map[string]interface{}{
			"disk_size":             1024,
			"disk_thin_provisioned": true,
		},
		"network_adapters": map[string]interface{}{
			"network_card": "vmxnet3",
		},
		"guest_os_type": "other3xLinux64Guest",

		"iso_paths": []string{
			"[datastore1] ISO/alpine-standard-3.8.2-x86_64.iso",
		},
		"floppy_files": []string{
			"../examples/alpine/answerfile",
			"../examples/alpine/setup.sh",
		},

		"boot_wait": "20s",
		"boot_command": []string{
			"root<enter><wait>",
			"mount -t vfat /dev/fd0 /media/floppy<enter><wait>",
			"setup-alpine -f /media/floppy/answerfile<enter>",
			"<wait5>",
			"jetbrains<enter>",
			"jetbrains<enter>",
			"<wait5>",
			"y<enter>",
			"<wait10><wait10><wait10><wait10>",
			"reboot<enter>",
			"<wait10><wait10><wait10>",
			"root<enter>",
			"jetbrains<enter><wait>",
			"mount -t vfat /dev/fd0 /media/floppy<enter><wait>",
			"/media/floppy/SETUP.SH<enter>",
		},

		"ssh_username": "packer",
		"ssh_password": "VMw@re1!",
	}

	return config
}

func checkFull(name string) error {
	d, err := commonT.TestConn()
	if err != nil {
		return fmt.Errorf("cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("cannot find VM: %v", err)
	}
	vmInfo, err := vm.Info("config.bootOptions")
	if err != nil {
		return fmt.Errorf("cannot read VM properties: %v", err)
	}

	order := vmInfo.Config.BootOptions.BootOrder
	if order != nil {
		return fmt.Errorf("boot order must be empty")
	}

	devices, err := vm.Devices()
	if err != nil {
		return fmt.Errorf("cannot read devices: %v", err)
	}
	cdroms := devices.SelectByType((*types.VirtualCdrom)(nil))
	for _, cd := range cdroms {
		_, ok := cd.(*types.VirtualCdrom).Backing.(*types.VirtualCdromRemotePassthroughBackingInfo)
		if !ok {
			return fmt.Errorf("wrong cdrom backing")
		}
	}

	return nil
}

func TestAccISOBuilderAcc_bootOrder(t *testing.T) {
	config := fullConfig()
	config["boot_order"] = "disk,cdrom,floppy"

	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_bootOrder_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return checkBootOrder(config["vm_name"].(string))
		},
	}
	acctest.TestPlugin(t, testCase)
}

func checkBootOrder(name string) error {
	d, err := commonT.TestConn()
	if err != nil {
		return fmt.Errorf("cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("cannot find VM: %v", err)
	}

	vmInfo, err := vm.Info("config.bootOptions")
	if err != nil {
		return fmt.Errorf("cannot read VM properties: %v", err)
	}

	order := vmInfo.Config.BootOptions.BootOrder
	if order == nil {
		return fmt.Errorf("boot order must not be empty")
	}

	return nil
}

func TestISOBuilderAcc_cluster(t *testing.T) {
	config := defaultConfig()
	config["cluster"] = "cluster1"
	config["host"] = "esxi-02.example.com"
	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_bootOrder_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

func TestISOBuilderAcc_clusterDRS(t *testing.T) {
	config := defaultConfig()
	config["cluster"] = "cluster2"
	config["host"] = ""
	config["datastore"] = "datastore3" // bug #183
	config["network_adapters"] = map[string]interface{}{
		"network": "VM Network",
	}

	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_bootOrder_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("bad exit code; logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

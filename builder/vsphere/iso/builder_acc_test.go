package iso

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/hashicorp/packer-plugin-sdk/acctest"
	commonT "github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/common/testing"
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
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return checkDefault(config["vm_name"].(string), config["host"].(string), "datastore1")
		},
	}
	acctest.TestPlugin(t, testCase)
}

func defaultConfig() map[string]interface{} {
	username := os.Getenv("VSPHERE_USERNAME")
	if username == "" {
		username = "root"
	}
	password := os.Getenv("VSPHERE_PASSWORD")
	if password == "" {
		password = "jetbrains"
	}

	vcenter := os.Getenv("VSPHERE_VCENTER_SERVER")
	if vcenter == "" {
		vcenter = "vcenter.vsphere65.test"
	}

	host := os.Getenv("VSPHERE_HOST")
	if host == "" {
		host = "esxi-1.vsphere65.test"
	}

	config := map[string]interface{}{
		"vcenter_server":      vcenter,
		"username":            username,
		"password":            password,
		"host":                host,
		"insecure_connection": true,

		"ssh_username": "root",
		"ssh_password": "jetbrains",

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
		return fmt.Errorf("Cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("Cannot find VM: %v", err)
	}

	vmInfo, err := vm.Info("name", "parent", "runtime.host", "resourcePool", "datastore", "layoutEx.disk", "config.firmware")
	if err != nil {
		return fmt.Errorf("Cannot read VM properties: %v", err)
	}

	if vmInfo.Name != name {
		return fmt.Errorf("Invalid VM name: expected '%v', got '%v'", name, vmInfo.Name)
	}

	f := d.NewFolder(vmInfo.Parent)
	folderPath, err := f.Path()
	if err != nil {
		return fmt.Errorf("Cannot read folder name: %v", err)
	}
	if folderPath != "" {
		return fmt.Errorf("Invalid folder: expected '/', got '%v'", folderPath)
	}

	h := d.NewHost(vmInfo.Runtime.Host)
	hostInfo, err := h.Info("name")
	if err != nil {
		return fmt.Errorf("Cannot read host properties: ", err)
	}
	if hostInfo.Name != host {
		return fmt.Errorf("Invalid host name: expected '%v', got '%v'", host, hostInfo.Name)
	}

	p := d.NewResourcePool(vmInfo.ResourcePool)
	poolPath, err := p.Path()
	if err != nil {
		return fmt.Errorf("Cannot read resource pool name: %v", err)
	}
	if poolPath != "" {
		return fmt.Errorf("Invalid resource pool: expected '/', got '%v'", poolPath)
	}

	dsr := vmInfo.Datastore[0].Reference()
	ds := d.NewDatastore(&dsr)
	dsInfo, err := ds.Info("name")
	if err != nil {
		return fmt.Errorf("Cannot read datastore properties: ", err)
	}
	if dsInfo.Name != datastore {
		return fmt.Errorf("Invalid datastore name: expected '%v', got '%v'", datastore, dsInfo.Name)
	}

	fw := vmInfo.Config.Firmware
	if fw != "bios" {
		return fmt.Errorf("Invalid firmware: expected 'bios', got '%v'", fw)
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
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
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
		return fmt.Errorf("Cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("Cannot find VM: %v", err)
	}
	vmInfo, err := vm.Info("config.annotation")
	if err != nil {
		return fmt.Errorf("Cannot read VM properties: %v", err)
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
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
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
		return fmt.Errorf("Cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("Cannot find VM: %v", err)
	}
	vmInfo, err := vm.Info("config")
	if err != nil {
		return fmt.Errorf("Cannot read VM properties: %v", err)
	}

	cpuSockets := vmInfo.Config.Hardware.NumCPU
	if cpuSockets != 2 {
		return fmt.Errorf("VM should have 2 CPU sockets, got %v", cpuSockets)
	}

	cpuCores := vmInfo.Config.Hardware.NumCoresPerSocket
	if cpuCores != 2 {
		return fmt.Errorf("VM should have 2 CPU cores per socket, got %v", cpuCores)
	}

	cpuReservation := *vmInfo.Config.CpuAllocation.Reservation
	if cpuReservation != 1000 {
		return fmt.Errorf("VM should have CPU reservation for 1000 Mhz, got %v", cpuReservation)
	}

	cpuLimit := *vmInfo.Config.CpuAllocation.Limit
	if cpuLimit != 1500 {
		return fmt.Errorf("VM should have CPU reservation for 1500 Mhz, got %v", cpuLimit)
	}

	ram := vmInfo.Config.Hardware.MemoryMB
	if ram != 2048 {
		return fmt.Errorf("VM should have 2048 MB of RAM, got %v", ram)
	}

	ramReservation := *vmInfo.Config.MemoryAllocation.Reservation
	if ramReservation != 1024 {
		return fmt.Errorf("VM should have RAM reservation for 1024 MB, got %v", ramReservation)
	}

	nestedHV := vmInfo.Config.NestedHVEnabled
	if !*nestedHV {
		return fmt.Errorf("VM should have NestedHV enabled, got %v", nestedHV)
	}

	fw := vmInfo.Config.Firmware
	if fw != "efi" {
		return fmt.Errorf("Invalid firmware: expected 'efi', got '%v'", fw)
	}

	l, err := vm.Devices()
	if err != nil {
		return fmt.Errorf("Cannot read VM devices: %v", err)
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
		return fmt.Errorf("VM should have one video card")
	}
	if v[0].(*types.VirtualMachineVideoCard).VideoRamSizeInKB != 8192 {
		return fmt.Errorf("Video RAM should be equal 8192")
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
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
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
		return fmt.Errorf("Cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("Cannot find VM: %v", err)
	}
	vmInfo, err := vm.Info("config.cpuAllocation")
	if err != nil {
		return fmt.Errorf("Cannot read VM properties: %v", err)
	}

	limit := *vmInfo.Config.CpuAllocation.Limit
	if limit != -1 { // must be unlimited
		return fmt.Errorf("Invalid CPU limit: expected '%v', got '%v'", -1, limit)
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
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
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
		return fmt.Errorf("Cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("Cannot find VM: %v", err)
	}

	l, err := vm.Devices()
	if err != nil {
		return fmt.Errorf("Cannot read VM devices: %v", err)
	}

	c := l.PickController((*types.VirtualAHCIController)(nil))
	if c == nil {
		return fmt.Errorf("VM has no SATA controllers")
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
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
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
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
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
		return fmt.Errorf("Cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("Cannot find VM: %v", err)
	}
	devices, err := vm.Devices()
	if err != nil {
		return fmt.Errorf("Cannot read VM properties: %v", err)
	}

	netCards := devices.SelectByType((*types.VirtualEthernetCard)(nil))
	if len(netCards) == 0 {
		return fmt.Errorf("Cannot find the network card")
	}
	if len(netCards) > 1 {
		return fmt.Errorf("Found several network catds")
	}
	if _, ok := netCards[0].(*types.VirtualVmxnet3); !ok {
		return fmt.Errorf("The network card type is not the expected one (vmxnet3)")
	}

	return nil
}

func TestAccISOBuilderAcc_createFloppy(t *testing.T) {
	tmpFile, err := ioutil.TempFile("", "packer-vsphere-iso-test")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	_, err = fmt.Fprint(tmpFile, "Hello, World!")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	err = tmpFile.Close()
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	config := defaultConfig()
	config["floppy_files"] = []string{tmpFile.Name()}
	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_createFloppy_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
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
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return checkFull(config["vm_name"].(string))
		},
	}
	acctest.TestPlugin(t, testCase)
}

func fullConfig() map[string]interface{} {
	username := os.Getenv("VSPHERE_USERNAME")
	if username == "" {
		username = "root"
	}
	password := os.Getenv("VSPHERE_PASSWORD")
	if password == "" {
		password = "jetbrains"
	}

	config := map[string]interface{}{
		"vcenter_server":      "vcenter.vsphere65.test",
		"username":            username,
		"password":            password,
		"insecure_connection": true,

		"vm_name": commonT.NewVMName(),
		"host":    "esxi-1.vsphere65.test",

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

		"ssh_username": "root",
		"ssh_password": "jetbrains",
	}

	return config
}

func checkFull(name string) error {
	d, err := commonT.TestConn()
	if err != nil {
		return fmt.Errorf("Cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("Cannot find VM: %v", err)
	}
	vmInfo, err := vm.Info("config.bootOptions")
	if err != nil {
		return fmt.Errorf("Cannot read VM properties: %v", err)
	}

	order := vmInfo.Config.BootOptions.BootOrder
	if order != nil {
		return fmt.Errorf("Boot order must be empty")
	}

	devices, err := vm.Devices()
	if err != nil {
		return fmt.Errorf("Cannot read devices: %v", err)
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
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
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
		return fmt.Errorf("Cannot connect %v", err)
	}
	vm, err := d.FindVM(name)
	if err != nil {
		return fmt.Errorf("Cannot find VM: %v", err)
	}

	vmInfo, err := vm.Info("config.bootOptions")
	if err != nil {
		return fmt.Errorf("Cannot read VM properties: %v", err)
	}

	order := vmInfo.Config.BootOptions.BootOrder
	if order == nil {
		return fmt.Errorf("Boot order must not be empty")
	}

	return nil
}

func TestISOBuilderAcc_cluster(t *testing.T) {
	config := defaultConfig()
	config["cluster"] = "cluster1"
	config["host"] = "esxi-2.vsphere65.test"
	testCase := &acctest.PluginTestCase{
		Name:     "vsphere-iso_bootOrder_test",
		Template: commonT.RenderConfig("vsphere-iso", config),
		Teardown: func() error {
			d, err := commonT.TestConn()
			if err != nil {
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
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
				return fmt.Errorf("Cannot connect %v", err)
			}
			return commonT.CleanupVM(d, config["vm_name"].(string))
		},
		Check: func(buildCommand *exec.Cmd, logfile string) error {
			if buildCommand.ProcessState != nil {
				if buildCommand.ProcessState.ExitCode() != 0 {
					return fmt.Errorf("Bad exit code. Logfile: %s", logfile)
				}
			}
			return nil
		},
	}
	acctest.TestPlugin(t, testCase)
}

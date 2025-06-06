// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package driver

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"

	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/nfc"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/ovf"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vapi/vcenter"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type VirtualMachine interface {
	Info(params ...string) (*mo.VirtualMachine, error)
	Devices() (object.VirtualDeviceList, error)
	CdromDevices() (object.VirtualDeviceList, error)
	FloppyDevices() (object.VirtualDeviceList, error)
	Clone(ctx context.Context, config *CloneConfig) (VirtualMachine, error)
	updateVAppConfig(ctx context.Context, newProps map[string]string) (*types.VmConfigSpec, error)
	AddPublicKeys(ctx context.Context, publicKeys string) error
	Properties(ctx context.Context) (*mo.VirtualMachine, error)
	Destroy() error
	Configure(config *HardwareConfig) error
	Reconfigure(spec types.VirtualMachineConfigSpec) error
	Customize(spec types.CustomizationSpec) error
	ResizeDisk(diskSize int64) ([]types.BaseVirtualDeviceConfigSpec, error)
	WaitForIP(ctx context.Context, ipNet *net.IPNet) (string, error)
	PowerOn() error
	PowerOff() error
	IsPoweredOff() (bool, error)
	StartShutdown() error
	WaitForShutdown(ctx context.Context, timeout time.Duration) error
	CreateSnapshot(name string) error
	ConvertToTemplate() error
	IsTemplate() (bool, error)
	ConvertToVirtualMachine(vsphereCluster string, vsphereHost string, vsphereResourcePool string) error
	ImportOvfToContentLibrary(ovf vcenter.OVF) error
	ImportToContentLibrary(template vcenter.Template) error
	GetDir() (string, error)
	AddFloppy(imgPath string) error
	SetBootOrder(order []string) error
	RemoveDevice(keepFiles bool, device ...types.BaseVirtualDevice) error
	addDevice(device types.BaseVirtualDevice) error
	AddConfigParams(params map[string]string, info *types.ToolsConfigInfo) error
	AddFlag(ctx context.Context, info *types.VirtualMachineFlagInfo) error
	Export() (*nfc.Lease, error)
	CreateDescriptor(m *ovf.Manager, cdp types.OvfCreateDescriptorParams) (*types.OvfCreateDescriptorResult, error)
	NewOvfManager() *ovf.Manager
	GetOvfExportOptions(m *ovf.Manager) ([]types.OvfOptionInfo, error)
	Datacenter() *object.Datacenter

	AddCdrom(controllerType string, datastoreIsoPath string) error
	CreateCdrom(c *types.VirtualController) (*types.VirtualCdrom, error)
	RemoveCdroms() error
	RemoveNCdroms(nCdroms int) error
	EjectCdroms() error
	AddSATAController() error
	FindSATAController() (*types.VirtualAHCIController, error)

	RemoveNetworkAdapters() error
}

type VirtualMachineDriver struct {
	vm     *object.VirtualMachine
	driver *VCenterDriver
}

type CloneConfig struct {
	Name            string
	Folder          string
	Cluster         string
	Host            string
	ResourcePool    string
	Datastore       string
	LinkedClone     bool
	Network         string
	MacAddress      string
	Annotation      string
	VAppProperties  map[string]string
	PrimaryDiskSize int64
	StorageConfig   StorageConfig
}

type PCIPassthroughAllowedDevice struct {
	VendorId    string
	DeviceId    string
	SubVendorId string
	SubDeviceId string
}

type HardwareConfig struct {
	CPUs                  int32
	CpuCores              int32
	CPUReservation        int64
	CPULimit              int64
	RAM                   int64
	RAMReservation        int64
	RAMReserveAll         bool
	NestedHV              bool
	CpuHotAddEnabled      bool
	MemoryHotAddEnabled   bool
	VideoRAM              int64
	Displays              int32
	AllowedDevices        []PCIPassthroughAllowedDevice
	VGPUProfile           string
	Firmware              string
	ForceBIOSSetup        bool
	VTPMEnabled           bool
	VirtualPrecisionClock string
}

type NIC struct {
	Network     string
	NetworkCard string
	MacAddress  string
	Passthrough *bool
}

type CreateConfig struct {
	Annotation    string
	Name          string
	Folder        string
	Cluster       string
	Host          string
	ResourcePool  string
	Datastore     string
	GuestOS       string
	NICs          []NIC
	USBController []string
	Version       uint
	StorageConfig StorageConfig
}

// NewVM creates a new virtual machine object.
func (d *VCenterDriver) NewVM(ref *types.ManagedObjectReference) VirtualMachine {
	return &VirtualMachineDriver{
		vm:     object.NewVirtualMachine(d.client.Client, *ref),
		driver: d,
	}
}

// FindVM locates a virtual machine by its name.
func (d *VCenterDriver) FindVM(name string) (VirtualMachine, error) {
	vm, err := d.finder.VirtualMachine(d.ctx, name)
	if err != nil {
		return nil, err
	}
	return &VirtualMachineDriver{
		vm:     vm,
		driver: d,
	}, nil
}

// PreCleanVM checks for an existing virtual machine at the specified path and optionally forces its removal.
func (d *VCenterDriver) PreCleanVM(ui packersdk.Ui, vmPath string, force bool, vsphereCluster string, vsphereHost string, vsphereResourcePool string) error {
	vm, err := d.FindVM(vmPath)
	if err != nil {
		var notFoundError *find.NotFoundError
		if !errors.As(err, &notFoundError) {
			return fmt.Errorf("error looking up existing virtual machine: %v", err)
		}
	}
	if force && vm != nil {
		ui.Sayf("Removing the existing virtual machine at %s based on use of the '-force' option...", vmPath)

		// Power off the virtual machine if it is powered on.
		_ = vm.PowerOff()

		// Check if the virtual machine is a template and convert it back to a
		// virtual machine if necessary.
		isTemplate, err := vm.IsTemplate()
		if err != nil {
			return fmt.Errorf("error determining if the virtual machine is a template%s: %v", vmPath, err)
		} else if isTemplate {
			ui.Sayf("Attempting to convert the template at %s to a virtual machine...", vmPath)
			err := vm.ConvertToVirtualMachine(vsphereCluster, vsphereHost, vsphereResourcePool)
			if err != nil {
				return fmt.Errorf("error converting template back to virtual machine for cleanup %s: %v", vmPath, err)
			}
		}

		err = vm.Destroy()
		if err != nil {
			return fmt.Errorf("error destroying %s: %v", vmPath, err)
		}
	}
	if !force && vm != nil {
		return fmt.Errorf("%s already exists, you can use -force flag to destroy it: %v", vmPath, err)
	}

	return nil
}

// CreateVM creates a new virtual machine based on the provided configuration
// specification.
func (d *VCenterDriver) CreateVM(config *CreateConfig) (VirtualMachine, error) {
	createSpec := types.VirtualMachineConfigSpec{
		Name:       config.Name,
		Annotation: config.Annotation,
		GuestId:    config.GuestOS,
	}
	if config.Version != 0 {
		createSpec.Version = fmt.Sprintf("%s%d", "vmx-", config.Version)
	}

	folder, err := d.FindFolder(config.Folder)
	if err != nil {
		return nil, err
	}

	resourcePool, err := d.FindResourcePool(config.Cluster, config.Host, config.ResourcePool)
	if err != nil {
		return nil, err
	}

	var host *object.HostSystem
	if config.Cluster != "" && config.Host != "" {
		h, err := d.FindHost(config.Host)
		if err != nil {
			return nil, err
		}
		host = h.host
	}

	datastore, err := d.FindDatastore(config.Datastore, config.Host)
	if err != nil {
		return nil, err
	}

	devices := object.VirtualDeviceList{}
	storageConfigSpec, err := config.StorageConfig.AddStorageDevices(devices)
	if err != nil {
		return nil, err
	}
	createSpec.DeviceChange = append(createSpec.DeviceChange, storageConfigSpec...)

	devices, err = addNetwork(d, devices, config)
	if err != nil {
		return nil, err
	}

	t := true
	for _, usbType := range config.USBController {
		var usb types.BaseVirtualDevice
		switch usbType {
		// handle "true" and "1" for backwards compatibility
		case "usb", "true", "1":
			usb = &types.VirtualUSBController{
				EhciEnabled: &t,
			}
		case "xhci":
			usb = new(types.VirtualUSBXHCIController)
		default:
			continue
		}

		devices = append(devices, usb)
	}

	devicesConfigSpec, err := devices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return nil, err
	}
	createSpec.DeviceChange = append(createSpec.DeviceChange, devicesConfigSpec...)

	createSpec.Files = &types.VirtualMachineFileInfo{
		VmPathName: fmt.Sprintf("[%s]", datastore.Name()),
	}

	task, err := folder.folder.CreateVM(d.ctx, createSpec, resourcePool.pool, host)
	if err != nil {
		return nil, err
	}
	taskInfo, err := task.WaitForResult(d.ctx, nil)
	if err != nil {
		return nil, err
	}

	vmRef, ok := taskInfo.Result.(types.ManagedObjectReference)
	if !ok {
		return nil, fmt.Errorf("something went wrong when creating the VM")
	}

	return d.NewVM(&vmRef), nil
}

// Info retrieves properties of the virtual machine object with optional filters
// specified  as parameters. If no parameters are provided, all properties are
// returned.
func (vm *VirtualMachineDriver) Info(params ...string) (*mo.VirtualMachine, error) {
	var p []string
	if len(params) == 0 {
		p = []string{"*"}
	} else {
		p = params
	}
	var info mo.VirtualMachine
	err := vm.vm.Properties(vm.driver.ctx, vm.vm.Reference(), p, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// Devices returns a list of devices attached to the virtual machine.
func (vm *VirtualMachineDriver) Devices() (object.VirtualDeviceList, error) {
	vmInfo, err := vm.Info("config.hardware.device")
	if err != nil {
		return nil, err
	}

	return vmInfo.Config.Hardware.Device, nil
}

// FloppyDevices returns a list of floppy devices attached to the virtual
// machine.
func (vm *VirtualMachineDriver) FloppyDevices() (object.VirtualDeviceList, error) {
	device, err := vm.Devices()
	if err != nil {
		return device, err
	}
	floppies := device.SelectByType((*types.VirtualFloppy)(nil))
	return floppies, nil
}

// CdromDevices returns a list of all CD-ROM devices attached to the virtual
// machine.
func (vm *VirtualMachineDriver) CdromDevices() (object.VirtualDeviceList, error) {
	device, err := vm.Devices()
	if err != nil {
		return device, err
	}
	floppies := device.SelectByType((*types.VirtualCdrom)(nil))
	return floppies, nil
}

// Clone creates a new virtual machine by cloning an existing one.
func (vm *VirtualMachineDriver) Clone(ctx context.Context, config *CloneConfig) (VirtualMachine, error) {
	folder, err := vm.driver.FindFolder(config.Folder)
	if err != nil {
		return nil, fmt.Errorf("error finding folder: %s", err)
	}

	var relocateSpec types.VirtualMachineRelocateSpec

	pool, err := vm.driver.FindResourcePool(config.Cluster, config.Host, config.ResourcePool)
	if err != nil {
		return nil, fmt.Errorf("error finding resource pool: %s", err)
	}
	poolRef := pool.pool.Reference()
	relocateSpec.Pool = &poolRef

	datastore, err := vm.driver.FindDatastore(config.Datastore, config.Host)
	if err != nil {
		return nil, fmt.Errorf("error finding datastore: %s", err)
	}
	datastoreRef := datastore.Reference()
	relocateSpec.Datastore = &datastoreRef

	if config.Cluster != "" && config.Host != "" {
		h, err := vm.driver.FindHost(config.Host)
		if err != nil {
			return nil, err
		}
		hostRef := h.host.Reference()
		relocateSpec.Host = &hostRef
	}

	var cloneSpec types.VirtualMachineCloneSpec
	cloneSpec.Location = relocateSpec
	cloneSpec.PowerOn = false

	if config.LinkedClone {
		cloneSpec.Location.DiskMoveType = "createNewChildDiskBacking"

		tpl, err := vm.Info("snapshot")
		if err != nil {
			return nil, fmt.Errorf("error getting snapshot info for virtual machine: %s", err)
		}
		if tpl.Snapshot == nil {
			err = errors.New("`linked_clone=true`, but template has no snapshots")
			return nil, err
		}
		cloneSpec.Snapshot = tpl.Snapshot.CurrentSnapshot
	}

	var configSpec types.VirtualMachineConfigSpec
	cloneSpec.Config = &configSpec

	if config.Annotation != "" {
		configSpec.Annotation = config.Annotation
	}

	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return nil, err
	}

	if config.PrimaryDiskSize > 0 {
		deviceResizeSpec, err := vm.ResizeDisk(config.PrimaryDiskSize)
		if err != nil {
			return nil, fmt.Errorf("failed to resize primary disk: %s", err)
		}
		configSpec.DeviceChange = append(configSpec.DeviceChange, deviceResizeSpec...)
	}

	virtualDisks := devices.SelectByType((*types.VirtualDisk)(nil))
	virtualControllers := devices.SelectByType((*types.VirtualController)(nil))

	// Use existing devices to avoid overlapping configuration.
	existingDevices := object.VirtualDeviceList{}
	existingDevices = append(existingDevices, virtualDisks...)
	existingDevices = append(existingDevices, virtualControllers...)

	storageConfigSpec, err := config.StorageConfig.AddStorageDevices(existingDevices)
	if err != nil {
		return nil, fmt.Errorf("failed to add storage devices: %s", err)
	}
	configSpec.DeviceChange = append(configSpec.DeviceChange, storageConfigSpec...)

	if config.Network != "" {
		net, err := vm.driver.FindNetwork(config.Network)
		if err != nil {
			return nil, fmt.Errorf("error finding network: %s", err)
		}
		backing, err := net.network.EthernetCardBackingInfo(ctx)
		if err != nil {
			return nil, fmt.Errorf("error finding ethernet card backing info: %s", err)
		}

		devices, err := vm.vm.Device(ctx)
		if err != nil {
			return nil, fmt.Errorf("error finding virtual machine devices: %s", err)
		}

		adapter, err := findNetworkAdapter(devices)
		if err != nil {
			return nil, fmt.Errorf("error finding network adapter: %s", err)
		}

		current := adapter.GetVirtualEthernetCard()
		current.Backing = backing

		if config.MacAddress != "" {
			current.AddressType = string(types.VirtualEthernetCardMacTypeManual)
			current.MacAddress = config.MacAddress
		}

		config := &types.VirtualDeviceConfigSpec{
			Device:    adapter.(types.BaseVirtualDevice),
			Operation: types.VirtualDeviceConfigSpecOperationEdit,
		}

		configSpec.DeviceChange = append(configSpec.DeviceChange, config)
	}

	vAppConfig, err := vm.updateVAppConfig(ctx, config.VAppProperties)
	if err != nil {
		return nil, fmt.Errorf("error updating VAppConfig: %s", err)
	}
	configSpec.VAppConfig = vAppConfig

	task, err := vm.vm.Clone(vm.driver.ctx, folder.folder, config.Name, cloneSpec)
	if err != nil {
		return nil, fmt.Errorf("error calling vm.vm.Clone task: %s", err)
	}

	info, err := task.WaitForResult(ctx, nil)
	if err != nil {
		if ctx.Err() == context.Canceled {
			err = task.Cancel(context.TODO())
			return nil, err
		}

		return nil, fmt.Errorf("error waiting for virtual machine clone to complete: %s", err)
	}

	vmRef, ok := info.Result.(types.ManagedObjectReference)
	if !ok {
		log.Printf("[ERROR] unexpected result during cloning operation: %s", info.Result)
		return nil, fmt.Errorf("error occured while cloning the virtual machine")
	}

	created := vm.driver.NewVM(&vmRef)
	return created, nil
}

// updateVAppConfig updates the vApp configuration of a virtual machine with new
// properties.
func (vm *VirtualMachineDriver) updateVAppConfig(ctx context.Context, newProps map[string]string) (*types.VmConfigSpec, error) {
	if len(newProps) == 0 {
		return nil, nil
	}

	vProps, _ := vm.Properties(ctx)
	if vProps.Config.VAppConfig == nil {
		return nil, fmt.Errorf("no vApp configuration found; cannot set vApp propertie")
	}

	allProperties := vProps.Config.VAppConfig.GetVmConfigInfo().Property

	var props []types.VAppPropertySpec
	for _, p := range allProperties {
		userValue, setByUser := newProps[p.Id]
		if !setByUser {
			continue
		}

		if !*p.UserConfigurable {
			return nil, fmt.Errorf("vApp property with userConfigurable=false specified in vapp.properties: %+v", reflect.ValueOf(newProps).MapKeys())
		}

		prop := types.VAppPropertySpec{
			ArrayUpdateSpec: types.ArrayUpdateSpec{
				Operation: types.ArrayUpdateOperationEdit,
			},
			Info: &types.VAppPropertyInfo{
				Key:              p.Key,
				Id:               p.Id,
				Value:            userValue,
				UserConfigurable: p.UserConfigurable,
			},
		}
		props = append(props, prop)

		delete(newProps, p.Id)
	}

	if len(newProps) > 0 {
		return nil, fmt.Errorf("unsupported vApp properties in vapp.properties: %+v", reflect.ValueOf(newProps).MapKeys())
	}

	return &types.VmConfigSpec{
		Property: props,
	}, nil
}

// AddPublicKeys adds public keys to the virtual machine.
func (vm *VirtualMachineDriver) AddPublicKeys(ctx context.Context, publicKeys string) error {
	newProps := map[string]string{"public-keys": publicKeys}
	config, err := vm.updateVAppConfig(ctx, newProps)
	if err != nil {
		return fmt.Errorf("not possible to save temporary public key: %s", err)
	}

	confSpec := types.VirtualMachineConfigSpec{VAppConfig: config}
	task, err := vm.vm.Reconfigure(vm.driver.ctx, confSpec)
	if err != nil {
		return err
	}

	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

// Properties retrieves the properties of a virtual machine.
func (vm *VirtualMachineDriver) Properties(ctx context.Context) (*mo.VirtualMachine, error) {
	log.Printf("fetching properties for VM %q", vm.vm.InventoryPath)
	var props mo.VirtualMachine
	if err := vm.vm.Properties(ctx, vm.vm.Reference(), nil, &props); err != nil {
		return nil, err
	}
	return &props, nil
}

// Destroy removes the virtual machine.
func (vm *VirtualMachineDriver) Destroy() error {
	task, err := vm.vm.Destroy(vm.driver.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

// Configure modifies the configuration of an existing virtual machine based on
// the provided configuration specification.
func (vm *VirtualMachineDriver) Configure(config *HardwareConfig) error {
	var confSpec types.VirtualMachineConfigSpec
	confSpec.NumCPUs = config.CPUs
	confSpec.NumCoresPerSocket = config.CpuCores
	confSpec.MemoryMB = config.RAM

	var cpuSpec types.ResourceAllocationInfo
	cpuSpec.Reservation = &config.CPUReservation
	if config.CPULimit != 0 {
		cpuSpec.Limit = &config.CPULimit
	}
	confSpec.CpuAllocation = &cpuSpec

	var ramSpec types.ResourceAllocationInfo
	ramSpec.Reservation = &config.RAMReservation
	confSpec.MemoryAllocation = &ramSpec

	confSpec.MemoryReservationLockedToMax = &config.RAMReserveAll
	confSpec.NestedHVEnabled = &config.NestedHV

	confSpec.CpuHotAddEnabled = &config.CpuHotAddEnabled
	confSpec.MemoryHotAddEnabled = &config.MemoryHotAddEnabled

	if config.Displays == 0 {
		config.Displays = 1
	}

	if config.VideoRAM != 0 || config.Displays != 0 {
		devices, err := vm.vm.Device(vm.driver.ctx)
		if err != nil {
			return err
		}
		l := devices.SelectByType((*types.VirtualMachineVideoCard)(nil))
		if len(l) != 1 {
			return err
		}
		card := l[0].(*types.VirtualMachineVideoCard)
		card.VideoRamSizeInKB = config.VideoRAM
		card.NumDisplays = config.Displays
		spec := &types.VirtualDeviceConfigSpec{
			Device:    card,
			Operation: types.VirtualDeviceConfigSpecOperationEdit,
		}
		confSpec.DeviceChange = append(confSpec.DeviceChange, spec)
	}

	if config.VGPUProfile != "" {
		devices, err := vm.vm.Device(vm.driver.ctx)
		if err != nil {
			return err
		}

		pciDevices := devices.SelectByType((*types.VirtualPCIPassthrough)(nil))
		vGPUDevices := pciDevices.SelectByBackingInfo((*types.VirtualPCIPassthroughVmiopBackingInfo)(nil))
		var operation types.VirtualDeviceConfigSpecOperation
		if len(vGPUDevices) > 1 {
			return err
		} else if len(pciDevices) == 1 {
			operation = types.VirtualDeviceConfigSpecOperationEdit
		} else if len(pciDevices) == 0 {
			operation = types.VirtualDeviceConfigSpecOperationAdd
		}

		vGPUProfile := newVGPUProfile(config.VGPUProfile)
		spec := &types.VirtualDeviceConfigSpec{
			Device:    &vGPUProfile,
			Operation: operation,
		}
		log.Printf("Adding vGPU device with profile '%s'", config.VGPUProfile)
		confSpec.DeviceChange = append(confSpec.DeviceChange, spec)
	}

	if len(config.AllowedDevices) > 0 {
		VirtualPCIPassthroughAllowedDevice, err := newVirtualPCIPassthroughAllowedDevice(config.AllowedDevices)
		if err != nil {
			log.Printf("Failed to create VirtualPCIPassthrough: %s", err)
			return err
		}
		spec := &types.VirtualDeviceConfigSpec{
			Device:    &VirtualPCIPassthroughAllowedDevice,
			Operation: types.VirtualDeviceConfigSpecOperationAdd,
		}
		confSpec.DeviceChange = append(confSpec.DeviceChange, spec)
	}

	efiSecureBootEnabled := false
	firmware := config.Firmware

	if firmware == "efi-secure" {
		firmware = "efi"
		efiSecureBootEnabled = true
	}

	confSpec.Firmware = firmware
	confSpec.BootOptions = &types.VirtualMachineBootOptions{
		EnterBIOSSetup:       types.NewBool(config.ForceBIOSSetup),
		EfiSecureBootEnabled: types.NewBool(efiSecureBootEnabled),
	}

	task, err := vm.vm.Reconfigure(vm.driver.ctx, confSpec)
	if err != nil {
		return err
	}

	_, err = task.WaitForResult(vm.driver.ctx, nil)
	if err != nil {
		return err
	}

	devices, err := vm.Devices()
	if err != nil {
		return err
	}
	TPMs := devices.SelectByType((*types.VirtualTPM)(nil))
	hasTPM := len(TPMs) > 0
	if config.VTPMEnabled != hasTPM {
		if !hasTPM {
			device := &types.VirtualTPM{}
			err = vm.addDevice(device)
		} else {
			err = vm.RemoveDevice(false, TPMs...)
		}
	}
	if err != nil {
		return err
	}

	if config.VirtualPrecisionClock != "" && config.VirtualPrecisionClock != "none" {
		device := &types.VirtualPrecisionClock{
			VirtualDevice: types.VirtualDevice{
				Backing: &types.VirtualPrecisionClockSystemClockBackingInfo{
					Protocol: config.VirtualPrecisionClock,
				},
			},
		}
		err = vm.addDevice(device)
		if err != nil {
			return err
		}
	}

	return err
}

// Reconfigure modifies the configuration of an existing virtual machine based
// on the provided configuration specification.
func (vm *VirtualMachineDriver) Reconfigure(confSpec types.VirtualMachineConfigSpec) error {
	task, err := vm.vm.Reconfigure(vm.driver.ctx, confSpec)
	if err != nil {
		return err
	}

	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

// Customize applies the given CustomizationSpec to the virtual machine.
func (vm *VirtualMachineDriver) Customize(spec types.CustomizationSpec) error {
	task, err := vm.vm.Customize(vm.driver.ctx, spec)
	if err != nil {
		return err
	}
	return task.Wait(vm.driver.ctx)
}

// ResizeDisk adjusts the size of the virtual disk to the specified diskSize in
// KB. Returns a slice of configuration specifications to apply the change or
// an error if the operation fails.
// TODO: This method should be refactored to support resizing multiple disks.
func (vm *VirtualMachineDriver) ResizeDisk(diskSize int64) ([]types.BaseVirtualDeviceConfigSpec, error) {
	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return nil, err
	}

	disk, err := findDisk(devices)
	if err != nil {
		return nil, err
	}

	disk.CapacityInKB = diskSize * 1024
	disk.CapacityInBytes = disk.CapacityInKB * 1024

	return []types.BaseVirtualDeviceConfigSpec{
		&types.VirtualDeviceConfigSpec{
			Device:    disk,
			Operation: types.VirtualDeviceConfigSpecOperationEdit,
		},
	}, nil
}

// PowerOn starts the virtual machine and waits for the operation to complete.
func (vm *VirtualMachineDriver) PowerOn() error {
	task, err := vm.vm.PowerOn(vm.driver.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

// WaitForIP waits for the virtual machine to obtain an IP address.
func (vm *VirtualMachineDriver) WaitForIP(ctx context.Context, ipNet *net.IPNet) (string, error) {
	netIP, err := vm.vm.WaitForNetIP(ctx, false)
	if err != nil {
		return "", err
	}

	for _, ips := range netIP {
		for _, ip := range ips {
			parseIP := net.ParseIP(ip)
			if ipNet != nil && !ipNet.Contains(parseIP) {
				// IP address is not in the expected range.
				continue
			}
			// Default to IPv4 if no IPNet is provided.
			if ipNet == nil && parseIP.To4() == nil {
				continue
			}
			return ip, nil
		}
	}

	// Unable to find an IP address.
	return "", nil
}

// PowerOff stops the virtual machine and waits for the operation to complete.
func (vm *VirtualMachineDriver) PowerOff() error {
	state, err := vm.vm.PowerState(vm.driver.ctx)
	if err != nil {
		return err
	}

	if state == types.VirtualMachinePowerStatePoweredOff {
		return nil
	}

	task, err := vm.vm.PowerOff(vm.driver.ctx)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

// IsPoweredOff checks if the virtual machine is powered off.
func (vm *VirtualMachineDriver) IsPoweredOff() (bool, error) {
	state, err := vm.vm.PowerState(vm.driver.ctx)
	if err != nil {
		return false, err
	}

	return state == types.VirtualMachinePowerStatePoweredOff, nil
}

// StartShutdown initiates a guest shutdown operation.
func (vm *VirtualMachineDriver) StartShutdown() error {
	err := vm.vm.ShutdownGuest(vm.driver.ctx)
	return err
}

// WaitForShutdown waits for the virtual machine to power off.
func (vm *VirtualMachineDriver) WaitForShutdown(ctx context.Context, timeout time.Duration) error {
	shutdownTimer := time.After(timeout)
	for {
		off, err := vm.IsPoweredOff()
		if err != nil {
			return err
		}
		if off {
			break
		}

		select {
		case <-shutdownTimer:
			err := errors.New("timeout while waiting for machine to shutdown")
			return err
		case <-ctx.Done():
			return nil
		default:
			time.Sleep(1 * time.Second)
		}
	}
	return nil
}

// CreateSnapshot creates a snapshot of the virtual machine.
func (vm *VirtualMachineDriver) CreateSnapshot(name string) error {
	task, err := vm.vm.CreateSnapshot(vm.driver.ctx, name, "", false, false)
	if err != nil {
		return err
	}
	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

// ConvertToTemplate converts the virtual machine to a template.
func (vm *VirtualMachineDriver) ConvertToTemplate() error {
	return vm.vm.MarkAsTemplate(vm.driver.ctx)
}

// IsTemplate checks if the virtual machine is a template.
func (vm *VirtualMachineDriver) IsTemplate() (bool, error) {
	state, err := vm.vm.IsTemplate(vm.driver.ctx)
	if err != nil {
		return false, err
	}

	return state, nil
}

// ConvertToVirtualMachine converts the template to a virtual machine.
func (vm *VirtualMachineDriver) ConvertToVirtualMachine(vsphereCluster string, vsphereHost string, vsphereResourcePool string) error {
	var host *object.HostSystem
	if vsphereCluster != "" && vsphereHost != "" {
		h, err := vm.driver.FindHost(vsphereHost)
		if err != nil {
			return err
		}
		host = h.host
	}

	resourcePool, err := vm.driver.FindResourcePool(vsphereCluster, vsphereHost, vsphereResourcePool)
	if err != nil {
		return err
	}

	return vm.vm.MarkAsVirtualMachine(vm.driver.ctx, *resourcePool.pool, host)
}

// ImportOvfToContentLibrary imports the OVF to the content library.
func (vm *VirtualMachineDriver) ImportOvfToContentLibrary(ovf vcenter.OVF) error {
	err := vm.driver.restClient.Login(vm.driver.ctx)
	if err != nil {
		return err
	}

	l, err := vm.driver.FindContentLibraryByName(ovf.Target.LibraryID)
	if err != nil {
		log.Printf("cannot find content library: %v", err)
		vm.logout()
		return err
	}
	if l.library.Type != "LOCAL" {
		return fmt.Errorf("cannot deploy a VM to the content library %s of type %s; "+
			"the content library must be of type LOCAL", ovf.Target.LibraryID, l.library.Type)
	}

	item, err := vm.driver.FindContentLibraryItem(l.library.ID, ovf.Spec.Name)
	if err == nil {
		// Update the content library item, if it exists.
		ovf.Target.LibraryItemID = item.ID
		if item.Description != nil && ovf.Spec.Description != *item.Description {
			err = vm.driver.UpdateContentLibraryItem(item, ovf.Spec.Name, ovf.Spec.Description)
			if err != nil {
				log.Printf("cannot update content library: %v", err)
				vm.logout()
				return err
			}
		}
	}

	ovf.Target.LibraryID = l.library.ID
	ovf.Source.Value = vm.vm.Reference().Value
	ovf.Source.Type = "VirtualMachine"

	vcm := vcenter.NewManager(vm.driver.restClient.client)
	_, err = vcm.CreateOVF(vm.driver.ctx, ovf)
	if err != nil {
		return err
	}

	return vm.driver.restClient.Logout(vm.driver.ctx)
}

// ImportToContentLibrary imports the virtual machine to the content library.
func (vm *VirtualMachineDriver) ImportToContentLibrary(template vcenter.Template) error {
	err := vm.driver.restClient.Login(vm.driver.ctx)
	if err != nil {
		return err
	}

	l, err := vm.driver.FindContentLibraryByName(template.Library)
	if err != nil {
		log.Printf("cannot find content library: %v", err)
		vm.logout()
		return err
	}
	if l.library.Type != "LOCAL" {
		return fmt.Errorf("cannot deploy a VM to the content library %s of type %s; "+
			"the content library must be of type LOCAL", template.Library, l.library.Type)
	}

	template.Library = l.library.ID
	template.SourceVM = vm.vm.Reference().Value

	if template.Placement.ResourcePool != "" {
		rp, err := vm.driver.FindResourcePool(template.Placement.Cluster, template.Placement.Host, template.Placement.ResourcePool)
		if err != nil {
			log.Printf("cannot find resource pool: %v", err)
			vm.logout()
			return err
		}
		template.Placement.ResourcePool = rp.pool.Reference().Value
	}
	if template.VMHomeStorage != nil {
		d, err := vm.driver.FindDatastore(template.VMHomeStorage.Datastore, template.Placement.Host)
		if err != nil {
			log.Printf("cannot find datastore: %v", err)
			vm.logout()
			return err
		}
		template.VMHomeStorage.Datastore = d.Reference().Value
	}

	if template.Placement.Cluster != "" {
		c, err := vm.driver.FindCluster(template.Placement.Cluster)
		if err != nil {
			return err
		}
		template.Placement.Cluster = c.cluster.Reference().Value
	}
	if template.Placement.Folder != "" {
		f, err := vm.driver.FindFolder(template.Placement.Folder)
		if err != nil {
			log.Printf("cannot find folder: %v", err)
			vm.logout()
			return err
		}
		template.Placement.Folder = f.folder.Reference().Value
	}
	if template.Placement.Host != "" {
		h, err := vm.driver.FindHost(template.Placement.Host)
		if err != nil {
			log.Printf("cannot find host: %v", err)
			vm.logout()
			return err
		}
		template.Placement.Host = h.host.Reference().Value
	}

	vcm := vcenter.NewManager(vm.driver.restClient.client)
	_, err = vcm.CreateTemplate(vm.driver.ctx, template)
	if err != nil {
		log.Printf("cannot create template: %v", err)
		vm.logout()
		return err
	}

	return vm.driver.restClient.Logout(vm.driver.ctx)
}

// GetDir returns the directory of the virtual machine. Returns an error if the
// operation fails.
func (vm *VirtualMachineDriver) GetDir() (string, error) {
	vmInfo, err := vm.Info("name", "layoutEx.file")
	if err != nil {
		return "", err
	}

	vmxName := fmt.Sprintf("/%s.vmx", vmInfo.Name)
	for _, file := range vmInfo.LayoutEx.File {
		if strings.Contains(file.Name, vmInfo.Name) {
			return RemoveDatastorePrefix(file.Name[:len(file.Name)-len(vmxName)]), nil
		}
	}
	return "", fmt.Errorf("cannot find '%s'", vmxName)
}

// addNetwork adds a network to the virtual machine. Returns a list of devices
// with the network added or an error if the  operation fails.
func addNetwork(d *VCenterDriver, devices object.VirtualDeviceList, config *CreateConfig) (object.VirtualDeviceList, error) {
	for _, nic := range config.NICs {
		network, err := findNetwork(nic.Network, config.Host, d)
		if err != nil {
			return nil, err
		}

		backing, err := network.EthernetCardBackingInfo(d.ctx)
		if err != nil {
			return nil, err
		}

		device, err := object.EthernetCardTypes().CreateEthernetCard(nic.NetworkCard, backing)
		if err != nil {
			return nil, err
		}

		card := device.(types.BaseVirtualEthernetCard).GetVirtualEthernetCard()
		if nic.MacAddress != "" {
			card.AddressType = string(types.VirtualEthernetCardMacTypeManual)
			card.MacAddress = nic.MacAddress
		}
		card.UptCompatibilityEnabled = nic.Passthrough

		devices = append(devices, device)
	}
	return devices, nil
}

// findNetwork finds a network based on the network name and host.
func findNetwork(network string, host string, d *VCenterDriver) (object.NetworkReference, error) {
	if network != "" {
		var err error
		networks, err := d.FindNetworks(network)
		if err != nil {
			return nil, err
		}
		if len(networks) == 1 {
			return networks[0].network, nil
		}

		// If there are multiple networks then try to match the host.
		if host != "" {
			h, err := d.FindHost(host)
			if err != nil {
				return nil, &MultipleNetworkFoundError{network, fmt.Sprintf("unable to match a network to the host %s: %s", host, err)}
			}
			for _, n := range networks {
				info, err := n.Info("host")
				if err != nil {
					continue
				}
				for _, host := range info.Host {
					if h.host.Reference().Value == host.Reference().Value {
						return n.network, nil
					}
				}
			}
			return nil, &MultipleNetworkFoundError{network, fmt.Sprintf("unable to match a network to the host %s", host)}
		}

		return nil, &MultipleNetworkFoundError{network, "specify the inventory path or id of the network"}
	}

	if host != "" {
		h, err := d.FindHost(host)
		if err != nil {
			return nil, err
		}

		i, err := h.Info("network")
		if err != nil {
			return nil, err
		}

		if len(i.Network) > 1 {
			return nil, fmt.Errorf("more than one network found on host %s, specify the network", host)
		}

		return object.NewNetwork(d.client.Client, i.Network[0]), nil
	}

	return nil, fmt.Errorf("error finding network; 'host' and 'network' not specified. at least one of the two must be specified")
}

// newVirtualPCIPassthroughAllowedDevice creates a virtual PCI passthrough device.
func newVirtualPCIPassthroughAllowedDevice(devices []PCIPassthroughAllowedDevice) (types.VirtualPCIPassthrough, error) {
	allowedDevices := make([]types.VirtualPCIPassthroughAllowedDevice, len(devices))
	for i, device := range devices {
		deviceId, err := strconv.ParseInt(device.DeviceId, 16, 32)
		if err != nil {
			return types.VirtualPCIPassthrough{}, fmt.Errorf("error parsing DeviceId %s: %s", device.DeviceId, err)
		}
		vendorId, err := strconv.ParseInt(device.VendorId, 16, 32)
		if err != nil {
			return types.VirtualPCIPassthrough{}, fmt.Errorf("error parsing VendorId %s: %s", device.VendorId, err)
		}
		subVendorId, err := strconv.ParseInt(device.SubVendorId, 16, 32)
		if err != nil {
			return types.VirtualPCIPassthrough{}, fmt.Errorf("error parsing SubVendorId %s: %s", device.SubVendorId, err)
		}
		subDeviceId, err := strconv.ParseInt(device.SubDeviceId, 16, 32)
		if err != nil {
			return types.VirtualPCIPassthrough{}, fmt.Errorf("error parsing SubDeviceId %s: %s", device.SubDeviceId, err)
		}

		allowedDevices[i] = types.VirtualPCIPassthroughAllowedDevice{
			DeviceId:    int32(deviceId),
			VendorId:    int32(vendorId),
			SubVendorId: int32(subVendorId),
			SubDeviceId: int32(subDeviceId),
		}

		log.Printf("adding pci dynamic direct i/o passthrough device with device_id '%s',vendor_id '%s',subsystem_id '%s',subsystem_vendor_id '%s'",
			device.DeviceId,
			device.VendorId,
			device.SubDeviceId,
			device.SubVendorId)
	}

	return types.VirtualPCIPassthrough{
		VirtualDevice: types.VirtualDevice{
			DeviceInfo: &types.Description{
				Summary: "",
				Label:   "New PCI device",
			},
			Backing: &types.VirtualPCIPassthroughDynamicBackingInfo{
				AllowedDevice: allowedDevices,
			},
		},
	}, nil
}

// newVGPUProfile creates a vGPU profile.
func newVGPUProfile(vGPUProfile string) types.VirtualPCIPassthrough {
	return types.VirtualPCIPassthrough{
		VirtualDevice: types.VirtualDevice{
			DeviceInfo: &types.Description{
				Summary: "",
				Label:   fmt.Sprintf("New vGPU %v PCI device", vGPUProfile),
			},
			Backing: &types.VirtualPCIPassthroughVmiopBackingInfo{
				Vgpu: vGPUProfile,
			},
		},
	}
}

// MountCdrom mounts a CD-ROM to the virtual machine.
func (vm *VirtualMachineDriver) MountCdrom(controllerType string, datastoreIsoPath string, _cdrom types.BaseVirtualDevice) error {
	cdrom := _cdrom.(*types.VirtualCdrom)
	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return err
	}

	ds := &DatastoreIsoPath{path: datastoreIsoPath}
	if !ds.Validate() {
		return fmt.Errorf("%s is not a valid iso path", datastoreIsoPath)
	}
	if libPath, err := vm.driver.FindContentLibraryFileDatastorePath(ds.GetFilePath()); err == nil {
		datastoreIsoPath = libPath
	} else {
		log.Printf("Using %s as the datastore path", datastoreIsoPath)
	}

	devices.InsertIso(cdrom, datastoreIsoPath)

	err = devices.Connect(cdrom)
	if err != nil {
		return err
	}
	return nil
}

// AddCdrom adds a CD-ROM to the virtual machine.
func (vm *VirtualMachineDriver) AddCdrom(controllerType string, datastoreIsoPath string) error {
	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return err
	}

	var controller *types.VirtualController
	if controllerType == "sata" {
		c, err := vm.FindSATAController()
		if err != nil {
			return err
		}
		controller = c.GetVirtualController()
	} else {
		c, err := devices.FindIDEController("")
		if err != nil {
			return err
		}
		controller = c.GetVirtualController()
	}

	cdrom, err := vm.CreateCdrom(controller)
	if err != nil {
		return err
	}

	if datastoreIsoPath == "" {
		cdrom.Backing = &types.VirtualCdromRemotePassthroughBackingInfo{}
		cdrom.Connectable = &types.VirtualDeviceConnectInfo{}
	} else {
		err := vm.MountCdrom(controllerType, datastoreIsoPath, cdrom)
		if err != nil {
			return err
		}
	}

	log.Printf("Creating CD-ROM on controller '%v' with iso '%v'", controller, datastoreIsoPath)
	return vm.addDevice(cdrom)
}

// AddFloppy adds a floppy disk to the virtual machine.
func (vm *VirtualMachineDriver) AddFloppy(imgPath string) error {
	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return err
	}

	floppy, err := devices.CreateFloppy()
	if err != nil {
		return err
	}

	if imgPath != "" {
		floppy = devices.InsertImg(floppy, imgPath)
	}

	return vm.addDevice(floppy)
}

// SetBootOrder sets the boot order of the virtual machine.
func (vm *VirtualMachineDriver) SetBootOrder(order []string) error {
	devices, err := vm.vm.Device(vm.driver.ctx)
	if err != nil {
		return err
	}

	bootOptions := types.VirtualMachineBootOptions{
		BootOrder: devices.BootOrder(order),
	}

	return vm.vm.SetBootOptions(vm.driver.ctx, &bootOptions)
}

// RemoveDevice removes a device from the virtual machine.
func (vm *VirtualMachineDriver) RemoveDevice(keepFiles bool, device ...types.BaseVirtualDevice) error {
	return vm.vm.RemoveDevice(vm.driver.ctx, keepFiles, device...)
}

// addDevice adds a device to the virtual machine.
func (vm *VirtualMachineDriver) addDevice(device types.BaseVirtualDevice) error {
	newDevices := object.VirtualDeviceList{device}
	confSpec := types.VirtualMachineConfigSpec{}
	var err error
	confSpec.DeviceChange, err = newDevices.ConfigSpec(types.VirtualDeviceConfigSpecOperationAdd)
	if err != nil {
		return err
	}

	task, err := vm.vm.Reconfigure(vm.driver.ctx, confSpec)
	if err != nil {
		return err
	}

	_, err = task.WaitForResult(vm.driver.ctx, nil)
	return err
}

// AddConfigParams adds configuration parameters to the virtual machine.
func (vm *VirtualMachineDriver) AddConfigParams(params map[string]string, info *types.ToolsConfigInfo) error {
	var confSpec types.VirtualMachineConfigSpec

	var ov []types.BaseOptionValue
	for k, v := range params {
		o := &types.OptionValue{
			Key:   k,
			Value: v,
		}
		ov = append(ov, o)
	}
	confSpec.ExtraConfig = ov

	confSpec.Tools = info

	if len(confSpec.ExtraConfig) > 0 || confSpec.Tools != nil {
		task, err := vm.vm.Reconfigure(vm.driver.ctx, confSpec)
		if err != nil {
			return fmt.Errorf("failed to start reconfiguration task: %w", err)
		}

		_, err = task.WaitForResult(vm.driver.ctx, nil)
		if err != nil {
			return fmt.Errorf("reconfiguration task failed: %w", err)
		}

		log.Println("[INFO] Reconfiguration task completed successfully.")

		// Retrieve the current configuration.
		var moVM mo.VirtualMachine
		err = vm.vm.Properties(vm.driver.ctx, vm.vm.Reference(), []string{"config.extraConfig"}, &moVM)
		if err != nil {
			return fmt.Errorf("failed to retrieve current configuration: %w", err)
		}

		// Check for ignored parameters.
		var ignoredParams []string
		for k, v := range params {
			found := false
			for _, option := range moVM.Config.ExtraConfig {
				if optVal, ok := option.(*types.OptionValue); ok && optVal.Key == k && optVal.Value == v {
					found = true
					break
				}
			}
			if !found {
				ignoredParams = append(ignoredParams, fmt.Sprintf("%s = %v", k, v))
			}
		}

		if len(ignoredParams) > 0 {
			log.Printf("[INFO] Ignored the following parameters: [%s]", strings.Join(ignoredParams, " , "))
			log.Printf("[INFO] Some configuration keys were ignored due to conflicts with other fields in the ConfigSpec. Refer to VirtualMachineConfigSpec in the vSphere API documentation.")
		}
	}

	return nil
}

// AddFlag adds a flag to the virtual machine.
func (vm *VirtualMachineDriver) AddFlag(ctx context.Context, flagSpec *types.VirtualMachineFlagInfo) error {
	confSpec := types.VirtualMachineConfigSpec{
		Flags: flagSpec,
	}

	task, err := vm.vm.Reconfigure(ctx, confSpec)
	if err != nil {
		return err
	}

	err = task.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}

// Export exports the virtual machine.
func (vm *VirtualMachineDriver) Export() (*nfc.Lease, error) {
	return vm.vm.Export(vm.driver.ctx)
}

// CreateDescriptor creates a descriptor for the virtual machine used when exporting the virtual machine to an OVF.
func (vm *VirtualMachineDriver) CreateDescriptor(m *ovf.Manager, cdp types.OvfCreateDescriptorParams) (*types.OvfCreateDescriptorResult, error) {
	return m.CreateDescriptor(vm.driver.ctx, vm.vm, cdp)
}

// NewOvfManager creates a new OVF manager instance.
func (vm *VirtualMachineDriver) NewOvfManager() *ovf.Manager {
	return ovf.NewManager(vm.vm.Client())
}

// GetOvfExportOptions retrieves the OVF export options for the virtual machine.
func (vm *VirtualMachineDriver) GetOvfExportOptions(m *ovf.Manager) ([]types.OvfOptionInfo, error) {
	var mgr mo.OvfManager
	err := property.DefaultCollector(vm.vm.Client()).RetrieveOne(vm.driver.ctx, m.Reference(), nil, &mgr)
	if err != nil {
		return nil, err
	}
	return mgr.OvfExportOption, nil
}

// NewHost creates a new host instance.
func (vm *VirtualMachineDriver) NewHost(ref *types.ManagedObjectReference) *Host {
	return vm.driver.NewHost(ref)
}

// NewResourcePool creates a new resource pool instance.
func (vm *VirtualMachineDriver) NewResourcePool(ref *types.ManagedObjectReference) *ResourcePool {
	return vm.driver.NewResourcePool(ref)
}

// NewDatastore creates a new datastore instance.
func (vm *VirtualMachineDriver) NewDatastore(ref *types.ManagedObjectReference) Datastore {
	return vm.driver.NewDatastore(ref)
}

// NewNetwork creates a new network instance.
func (vm *VirtualMachineDriver) NewNetwork(ref *types.ManagedObjectReference) *Network {
	return vm.driver.NewNetwork(ref)
}

// Datacenter returns the datacenter of the virtual machine.
func (vm *VirtualMachineDriver) Datacenter() *object.Datacenter {
	return vm.driver.datacenter
}

// FindContentLibraryItemUUID finds a content library item by name.
func (vm *VirtualMachineDriver) FindContentLibraryItemUUID(library string, name string) (string, error) {
	err := vm.driver.restClient.Login(vm.driver.ctx)
	if err != nil {
		return "", err
	}

	l, err := vm.driver.FindContentLibraryByName(library)
	if err != nil {
		log.Printf("cannot find content library: %v", err)
		vm.logout()
		return "", err
	}

	item, err := vm.driver.FindContentLibraryItemUUID(l.library.ID, name)
	if err != nil {
		log.Printf("cannot find content library item: %v", err)
		vm.logout()
		return "", err
	}

	return item, nil
}

// FindContentLibraryTemplateDatastoreName finds the datastore name of the content library template.
func (vm *VirtualMachineDriver) FindContentLibraryTemplateDatastoreName(library string) ([]string, error) {
	err := vm.driver.restClient.Login(vm.driver.ctx)
	if err != nil {
		return nil, err
	}

	l, err := vm.driver.FindContentLibraryByName(library)
	if err != nil {
		log.Printf("cannot find content library: %v", err)
		vm.logout()
		return nil, err
	}
	var datastores []string
	for _, storage := range l.library.Storage {
		name, err := vm.driver.GetDatastoreName(storage.DatastoreID)
		if err != nil {
			log.Printf("Failed to get Content Library datastore name: %s", err)
			continue
		}
		datastores = append(datastores, name)
	}
	return datastores, vm.driver.restClient.Logout(vm.driver.ctx)
}

// logout logs the user out of the vCenter.
func (vm *VirtualMachineDriver) logout() {
	if vm.driver.restClient == nil {
		return
	}
	if err := vm.driver.restClient.Logout(vm.driver.ctx); err != nil {
		log.Printf("cannot logout: %s ", err)
	}
}

// findNetworkAdapter finds a network adapter in the virtual machine.
func findNetworkAdapter(l object.VirtualDeviceList) (types.BaseVirtualEthernetCard, error) {
	c := l.SelectByType((*types.VirtualEthernetCard)(nil))
	if len(c) == 0 {
		return nil, errors.New("no network adapter device found")
	}

	return c[0].(types.BaseVirtualEthernetCard), nil
}

// RemoveNetworkAdapters removes all network adapters from the virtual machine.
func (vm *VirtualMachineDriver) RemoveNetworkAdapters() error {
	devices, err := vm.Devices()
	if err != nil {
		return fmt.Errorf("error retrieving devices: %s", err)
	}

	networkAdapters := devices.SelectByType((*types.VirtualEthernetCard)(nil))
	if len(networkAdapters) == 0 {
		return nil
	}

	for _, adapter := range networkAdapters {
		err = vm.RemoveDevice(false, adapter)
		if err != nil {
			return fmt.Errorf("error removing network adapter: %s", err)
		}
	}

	return nil
}

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CustomizeConfig,LinuxOptions,WindowsOptions,WindowsOptionsGuiUnattended,WindowsOptionsUserData,WindowsOptionsGuiRunOnce,WindowsOptionsIdentification,WindowsOptionsLicenseFilePrintData,NetworkInterfaces,NetworkInterface,GlobalDnsSettings,GlobalRoutingSettings
package clone

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/vmware/govmomi/vim25/types"
)

// A cloned virtual machine can be [customized](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vm_admin.doc/GUID-58E346FF-83AE-42B8-BE58-253641D257BC.html)
// to configure host, network, or licensing settings.
//
// To perform virtual machine customization as a part of the clone process, specify the customize block with the
// respective customization options. Windows guests are customized using Sysprep, which will result in the machine SID being reset.
// Before using customization, check that your source VM meets the [requirements](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vm_admin.doc/GUID-E63B6FAA-8D35-428D-B40C-744769845906.html)
// for guest OS customization on vSphere.
// See the [customization example](#customization-example) for a usage synopsis.
//
// The settings for customize are as follows:
type CustomizeConfig struct {
	// Settings to Linux guest OS customization. See [Linux customization settings](#linux-customization-settings).
	LinuxOptions *LinuxOptions `mapstructure:"linux_options"`
	// Settings to Windows guest OS customization.
	WindowsOptions *WindowsOptions `mapstructure:"windows_options"`
	// Supply your own sysprep.xml file to allow full control of the customization process out-of-band of vSphere.
	WindowsSysPrepFile string `mapstructure:"windows_sysprep_file"`
	// Configure network interfaces on a per-interface basis that should matched up to the network adapters present in the VM.
	// To use DHCP, declare an empty network_interface for each adapter being configured. This field is required.
	// See [Network interface settings](#network-interface-settings).
	NetworkInterfaces     NetworkInterfaces `mapstructure:"network_interface"`
	GlobalRoutingSettings `mapstructure:",squash"`
	GlobalDnsSettings     `mapstructure:",squash"`
}

type LinuxOptions struct {
	// The domain name for this machine. This, along with [host_name](#host_name), make up the FQDN of this virtual machine.
	Domain string `mapstructure:"domain"`
	// The host name for this machine. This, along with [domain](#domain), make up the FQDN of this virtual machine.
	Hostname string `mapstructure:"host_name"`
	// Tells the operating system that the hardware clock is set to UTC. Default: true.
	HWClockUTC config.Trilean `mapstructure:"hw_clock_utc"`
	// Sets the time zone. The default is UTC.
	Timezone string `mapstructure:"time_zone"`
}

type WindowsOptions struct {
	// CustomizationGuiRunOnce
	// A list of commands to run at first user logon, after guest customization.
	RunOnceCommandList *[]string `mapstructure:"run_once_command_list"`
	// CustomizationGuiUnattended
	// Specifies whether or not the VM automatically logs on as Administrator.
	AutoLogon *bool `mapstructure:"auto_logon"`
	// Specifies how many times the VM should auto-logon the Administrator account when auto_logon is true. Default 1
	AutoLogonCount *int32 `mapstructure:"auto_logon_count"`
	// The new administrator password for this virtual machine.
	AdminPassword *string `mapstructure:"admin_password"`
	// The new time zone for the virtual machine. This is a sysprep-dictated timezone code. Default 85 (GMT)
	TimeZone *int32 `mapstructure:"time_zone"`
	// CustomizationIdentification
	// The workgroup for this virtual machine - AD Join is not supported
	Workgroup string `mapstructure:"workgroup"`
	// CustomizationUserData
	// The host name for this virtual machine.
	ComputerName string `mapstructure:"computer_name"`
	// The full name of the user of this virtual machine. Default: "Administrator"
	FullName string `mapstructure:"full_name"`
	// The organization name this virtual machine is being installed for. Default: "Managed by Packer"
	OrganizationName string `mapstructure:"organization_name"`
	// The product key for this virtual machine.
	ProductKey string `mapstructure:"product_key"`
}

type NetworkInterface struct {
	// Network interface-specific DNS server settings for Windows operating systems.
	// Ignored on Linux and possibly other operating systems - for those systems, please see the [global DNS settings](#global-dns-settings) section.
	DnsServerList []string `mapstructure:"dns_server_list"`
	// Network interface-specific DNS search domain for Windows operating systems.
	// Ignored on Linux and possibly other operating systems - for those systems, please see the [global DNS settings](#global-dns-settings) section.
	DnsDomain string `mapstructure:"dns_domain"`
	// The IPv4 address assigned to this network adapter. If left blank or not included, DHCP is used.
	Ipv4Address string `mapstructure:"ipv4_address"`
	// The IPv4 subnet mask, in bits (example: 24 for 255.255.255.0).
	Ipv4NetMask int `mapstructure:"ipv4_netmask"`
	// The IPv6 address assigned to this network adapter. If left blank or not included, auto-configuration is used.
	Ipv6Address string `mapstructure:"ipv6_address"`
	// The IPv6 subnet mask, in bits (example: 32).
	Ipv6NetMask int `mapstructure:"ipv6_netmask"`
}

type NetworkInterfaces []NetworkInterface

// The settings here must match the IP/mask of at least one network_interface supplied to customization.
type GlobalRoutingSettings struct {
	// The IPv4 default gateway when using network_interface customization on the virtual machine.
	Ipv4Gateway string `mapstructure:"ipv4_gateway"`
	// The IPv6 default gateway when using network_interface customization on the virtual machine.
	Ipv6Gateway string `mapstructure:"ipv6_gateway"`
}

// The following settings configure DNS globally, generally for Linux systems. For Windows systems,
// this is done per-interface, see [network interface](#network_interface) settings.
type GlobalDnsSettings struct {
	// The list of DNS servers to configure on a virtual machine.
	DnsServerList []string `mapstructure:"dns_server_list"`
	// A list of DNS search domains to add to the DNS configuration on the virtual machine.
	DnsSuffixList []string `mapstructure:"dns_suffix_list"`
}

type StepCustomize struct {
	Config *CustomizeConfig
}

func (c *CustomizeConfig) Prepare() []error {
	var errs []error

	options_number := 0
	if c.LinuxOptions != nil {
		options_number = options_number + 1
	}
	if c.WindowsOptions != nil {
		options_number = options_number + 1
	}
	if c.WindowsSysPrepFile != "" {
		options_number = options_number + 1
	}

	if options_number > 1 {
		errs = append(errs, fmt.Errorf("Only one of `linux_options`, `windows_options`, `windows_sysprep_file` can be set"))
	} else if options_number == 0 {
		errs = append(errs, fmt.Errorf("One of `linux_options`, `windows_options`, `windows_sysprep_file` must be set"))
	}

	if c.LinuxOptions != nil {
		errs = c.LinuxOptions.prepare(errs)
	}
	if c.WindowsOptions != nil {
		errs = c.WindowsOptions.prepare(errs)
	}

	return errs
}

func (s *StepCustomize) Run(ctx context.Context, state multistep.StateBag) multistep.StepAction {
	vm := state.Get("vm").(*driver.VirtualMachineDriver)
	ui := state.Get("ui").(packersdk.Ui)

	identity, err := s.identitySettings()
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	nicSettingsMap := s.nicSettingsMap()
	globalIpSettings := s.globalIpSettings()

	spec := types.CustomizationSpec{
		Identity:         identity,
		NicSettingMap:    nicSettingsMap,
		GlobalIPSettings: globalIpSettings,
	}
	ui.Say("Customizing VM...")
	err = vm.Customize(spec)
	if err != nil {
		state.Put("error", err)
		return multistep.ActionHalt
	}

	return multistep.ActionContinue
}

func (s *StepCustomize) identitySettings() (types.BaseCustomizationIdentitySettings, error) {
	if s.Config.LinuxOptions != nil {
		return s.Config.LinuxOptions.linuxPrep(), nil
	}

	if s.Config.WindowsOptions != nil {
		return s.Config.WindowsOptions.sysprep(), nil
	}

	if s.Config.WindowsSysPrepFile != "" {
		sysPrep, err := ioutil.ReadFile(s.Config.WindowsSysPrepFile)
		if err != nil {
			return nil, fmt.Errorf("error on reading %s: %s", s.Config.WindowsSysPrepFile, err)
		}
		return &types.CustomizationSysprepText{
			Value: string(sysPrep),
		}, nil
	}

	return nil, fmt.Errorf("no customization identity found")
}

func (s *StepCustomize) nicSettingsMap() []types.CustomizationAdapterMapping {
	result := make([]types.CustomizationAdapterMapping, len(s.Config.NetworkInterfaces))
	var ipv4gwFound, ipv6gwFound bool
	for i := range s.Config.NetworkInterfaces {
		var adapter types.CustomizationIPSettings
		adapter, ipv4gwFound, ipv6gwFound = s.ipSettings(i, !ipv4gwFound, !ipv6gwFound)
		obj := types.CustomizationAdapterMapping{
			Adapter: adapter,
		}
		result[i] = obj
	}
	return result
}

func (s *StepCustomize) ipSettings(n int, ipv4gwAdd bool, ipv6gwAdd bool) (types.CustomizationIPSettings, bool, bool) {
	var v4gwFound, v6gwFound bool
	var obj types.CustomizationIPSettings

	ipv4Address := s.Config.NetworkInterfaces[n].Ipv4Address
	if ipv4Address != "" {
		ipv4mask := s.Config.NetworkInterfaces[n].Ipv4NetMask
		ipv4Gateway := s.Config.Ipv4Gateway
		obj.Ip = &types.CustomizationFixedIp{
			IpAddress: ipv4Address,
		}
		obj.SubnetMask = v4CIDRMaskToDotted(ipv4mask)
		// Check for the gateway
		if ipv4gwAdd && ipv4Gateway != "" && matchGateway(ipv4Address, ipv4mask, ipv4Gateway) {
			obj.Gateway = []string{ipv4Gateway}
			v4gwFound = true
		}
	} else {
		obj.Ip = &types.CustomizationDhcpIpGenerator{}
	}

	obj.DnsServerList = s.Config.NetworkInterfaces[n].DnsServerList
	obj.DnsDomain = s.Config.NetworkInterfaces[n].DnsDomain
	obj.IpV6Spec, v6gwFound = s.IPSettingsIPV6Address(n, ipv6gwAdd)

	return obj, v4gwFound, v6gwFound
}

func v4CIDRMaskToDotted(mask int) string {
	m := net.CIDRMask(mask, 32)
	a := int(m[0])
	b := int(m[1])
	c := int(m[2])
	d := int(m[3])
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

func (s *StepCustomize) IPSettingsIPV6Address(n int, gwAdd bool) (*types.CustomizationIPSettingsIpV6AddressSpec, bool) {
	addr := s.Config.NetworkInterfaces[n].Ipv6Address
	var gwFound bool
	if addr == "" {
		return nil, gwFound
	}
	mask := s.Config.NetworkInterfaces[n].Ipv6NetMask
	gw := s.Config.Ipv6Gateway
	obj := &types.CustomizationIPSettingsIpV6AddressSpec{
		Ip: []types.BaseCustomizationIpV6Generator{
			&types.CustomizationFixedIpV6{
				IpAddress:  addr,
				SubnetMask: int32(mask),
			},
		},
	}
	if gwAdd && gw != "" && matchGateway(addr, mask, gw) {
		obj.Gateway = []string{gw}
		gwFound = true
	}
	return obj, gwFound
}

// matchGateway take an IP, mask, and gateway, and checks to see if the gateway
// is reachable from the IP address.
func matchGateway(a string, m int, g string) bool {
	ip := net.ParseIP(a)
	gw := net.ParseIP(g)
	var mask net.IPMask
	if ip.To4() != nil {
		mask = net.CIDRMask(m, 32)
	} else {
		mask = net.CIDRMask(m, 128)
	}
	if ip.Mask(mask).Equal(gw.Mask(mask)) {
		return true
	}
	return false
}

func (s *StepCustomize) globalIpSettings() types.CustomizationGlobalIPSettings {
	return types.CustomizationGlobalIPSettings{
		DnsServerList: s.Config.DnsServerList,
		DnsSuffixList: s.Config.DnsSuffixList,
	}
}

func (l *LinuxOptions) prepare(errs []error) []error {
	if l.Hostname == "" {
		errs = append(errs, fmt.Errorf("linux options `host_name` is empty"))
	}
	if l.Domain == "" {
		errs = append(errs, fmt.Errorf("linux options `domain` is empty"))
	}

	if l.HWClockUTC == config.TriUnset {
		l.HWClockUTC = config.TriTrue
	}
	if l.Timezone == "" {
		l.Timezone = "UTC"
	}
	return errs
}

func (l *LinuxOptions) linuxPrep() *types.CustomizationLinuxPrep {
	obj := &types.CustomizationLinuxPrep{
		HostName: &types.CustomizationFixedName{
			Name: l.Hostname,
		},
		Domain:     l.Domain,
		TimeZone:   l.Timezone,
		HwClockUTC: l.HWClockUTC.ToBoolPointer(),
	}
	return obj
}

func (w *WindowsOptions) prepare(errs []error) []error {
	if w.ComputerName == "" {
		errs = append(errs, fmt.Errorf("The `computer_name` is required"))
	}
	if w.FullName == "" {
		w.FullName = "Administrator"
	}
	if w.OrganizationName == "" {
		w.OrganizationName = "Managed by Packer"
	}
	return errs
}

func (w *WindowsOptions) sysprep() *types.CustomizationSysprep {
	obj := &types.CustomizationSysprep{
		GuiUnattended:  w.guiUnattended(),
		UserData:       w.userData(),
		GuiRunOnce:     w.guiRunOnce(),
		Identification: w.identification(),
	}
	return obj
}

func (w *WindowsOptions) guiRunOnce() *types.CustomizationGuiRunOnce {
	obj := &types.CustomizationGuiRunOnce{
		CommandList: *w.RunOnceCommandList,
	}
	if len(obj.CommandList) < 1 {
		return nil
	}
	return obj
}

func boolValue(p *bool, fallback bool) bool {
	if p == nil {
		return fallback
	}
	return *p
}

func intValue(p *int32, fallback int32) int32 {
	if p == nil {
		return fallback
	}
	return *p
}

func (w *WindowsOptions) guiUnattended() types.CustomizationGuiUnattended {
	obj := types.CustomizationGuiUnattended{
		TimeZone:       intValue(w.TimeZone, 85),
		AutoLogon:      boolValue(w.AutoLogon, false),
		AutoLogonCount: intValue(w.AutoLogonCount, 1),
	}
	if w.AdminPassword != nil {
		obj.Password = &types.CustomizationPassword{
			Value:     *w.AdminPassword,
			PlainText: true,
		}
	}
	return obj
}

func (w *WindowsOptions) identification() types.CustomizationIdentification {
	obj := types.CustomizationIdentification{
		JoinWorkgroup: w.Workgroup,
	}
	return obj
}

func (w *WindowsOptions) userData() types.CustomizationUserData {
	obj := types.CustomizationUserData{
		FullName: w.FullName,
		OrgName:  w.OrganizationName,
		ComputerName: &types.CustomizationFixedName{
			Name: w.ComputerName,
		},
		ProductId: w.ProductKey,
	}
	return obj
}

func (s *StepCustomize) Cleanup(_ multistep.StateBag) {}

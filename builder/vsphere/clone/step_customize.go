// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:generate packer-sdc struct-markdown
//go:generate packer-sdc mapstructure-to-hcl2 -type CustomizeConfig,LinuxOptions,WindowsOptions,WindowsOptionsGuiUnattended,WindowsOptionsUserData,WindowsOptionsGuiRunOnce,WindowsOptionsIdentification,WindowsOptionsLicenseFilePrintData,NetworkInterfaces,NetworkInterface,GlobalDnsSettings,GlobalRoutingSettings
package clone

import (
	"context"
	"fmt"
	"net"
	"os"

	"github.com/hashicorp/packer-plugin-sdk/multistep"
	packersdk "github.com/hashicorp/packer-plugin-sdk/packer"
	"github.com/hashicorp/packer-plugin-sdk/template/config"
	"github.com/hashicorp/packer-plugin-vsphere/builder/vsphere/driver"
	"github.com/vmware/govmomi/vim25/types"
)

var (
	errCustomizeOptionMutualExclusive   = fmt.Errorf("only one of `linux_options`, `windows_options`, `windows_sysprep_file` can be set")
	windowsSysprepFileDeprecatedMessage = "`windows_sysprep_file` is deprecated and will be removed in a future release. please use `windows_sysprep_text`."
)

// A cloned virtual machine can be [customized](https://docs.vmware.com/en/VMware-vSphere/8.0/vsphere-vm-administration/GUID-58E346FF-83AE-42B8-BE58-253641D257BC.html)
// to configure host, network, or licensing settings.
//
// To perform virtual machine customization as a part of the clone process,
// specify the customize block with the respective customization options.
// Windows guests are customized using Sysprep, which will result in the machine
// SID being reset. Before using customization, check that your source virtual
// machine meets the [requirements](https://docs.vmware.com/en/VMware-vSphere/8.0/vsphere-vm-administration/GUID-E63B6FAA-8D35-428D-B40C-744769845906.html)
// for guest OS customization on vSphere. Refer to the [customization example](#customization-example) for a usage synopsis.
//
// The settings for guest customization include:
type CustomizeConfig struct {
	// Settings for the guest customization of Linux operating systems.
	// Refer to the [Linux options](#linux-options) section for additional
	// details.
	LinuxOptions *LinuxOptions `mapstructure:"linux_options"`
	// Settings for the guest customization of Windows operating systems.
	// Refer to the [Windows options](#windows-options) section for additional
	// details.
	WindowsOptions *WindowsOptions `mapstructure:"windows_options"`
	// Provide a `sysprep.xml` file to allow control of the customization
	// process independent of vSphere. This option is deprecated, please use
	// `windows_sysprep_text`.
	WindowsSysPrepFile string `mapstructure:"windows_sysprep_file"`
	// Provide the text for the `sysprep.xml` content to allow control of the
	// customization process independent of vSphere.
	//
	// HCL Examples:
	//
	// ```hcl
	// customize {
	//    windows_sysprep_text = file("<path-to-sysprep.xml>")
	// }
	// ```
	//
	// ```hcl
	// customize {
	//    windows_sysprep_text = templatefile("<path-to-sysprep.xml>", {
	//       var1="example"
	//       var2="example-2"
	//    })
	// }
	// ```
	//
	// JSON Examples
	//
	// ```json
	// {
	//   "customize": {
	//     "windows_sysprep_text": "<path-to-sysprep.xml>"
	//   }
	// }
	// ```
	//
	// ```json
	// {
	//   "customize": {
	//     "windows_sysprep_text": "<path-to-sysprep.xml>",
	//     "var1": "example",
	//     "var2": "example-2"
	//   }
	// }
	WindowsSysPrepText string `mapstructure:"windows_sysprep_text"`
	// Set up network interfaces individually to correspond with the network
	// adapters on the virtual machine. To use DHCP, specify an empty
	// `network_interface` for each configured adapter. This field is mandatory.
	// Refer to the [network interface](#network-interface-settings) section for
	// additional details.
	NetworkInterfaces     NetworkInterfaces `mapstructure:"network_interface"`
	GlobalRoutingSettings `mapstructure:",squash"`
	GlobalDnsSettings     `mapstructure:",squash"`
}

type LinuxOptions struct {
	// The domain name for the guest operating system. Used with
	// [host_name](#host_name) to construct the fully qualified domain name
	// (FQDN).
	Domain string `mapstructure:"domain"`
	// The hostname for the guest operating system. Used with [domain](#domain)
	// to construct the fully qualified domain name (FQDN).
	Hostname string `mapstructure:"host_name"`
	// Set the hardware clock to Coordinated Universal Time (UTC).
	// Defaults to `true`.
	HWClockUTC config.Trilean `mapstructure:"hw_clock_utc"`
	// The time zone for the guest operating system.
	Timezone string `mapstructure:"time_zone"`
}

type WindowsOptions struct {
	// A list of commands to run at first logon after the guest operating system
	// is customized.
	RunOnceCommandList []string `mapstructure:"run_once_command_list"`
	// Automatically log on the `Administrator` account after the guest operating
	// system is customized.
	AutoLogon *bool `mapstructure:"auto_logon"`
	// The number of times the guest operating system should auto-logon the
	// `Administrator` account when `auto_logon` is set to `true`.
	// Defaults to `1`.
	AutoLogonCount *int32 `mapstructure:"auto_logon_count"`
	// The password for the guest operating system's `Administrator` account.
	AdminPassword *string `mapstructure:"admin_password"`
	// The time zone for the guest operating system.
	// Defaults to `85` (Pacific Time).
	TimeZone *int32 `mapstructure:"time_zone"`
	// The workgroup for the guest operating system.
	// Joining an Active Directory domain is not supported.
	Workgroup string `mapstructure:"workgroup"`
	// The hostname for the guest operating system.
	ComputerName string `mapstructure:"computer_name"`
	// The full name for the guest operating system's `Administrator` account.
	// Defaults to `Administrator`.
	FullName string `mapstructure:"full_name"`
	// The organization name for the guest operating system.
	// Defaults to `Built by Packer`.
	OrganizationName string `mapstructure:"organization_name"`
	// The product key for the guest operating system.
	ProductKey string `mapstructure:"product_key"`
}

type NetworkInterface struct {
	// The DNS servers for a specific network interface on a Windows guest
	// operating system. Ignored on Linux. Refer to the
	// [global DNS settings](#global-dns-settings) section for additional
	// details.
	DnsServerList []string `mapstructure:"dns_server_list"`
	// The DNS search domain for a specific network interface on a Windows guest
	// operating system. Ignored on Linux. Refer to the
	// [global DNS settings](#global-dns-settings) section for additional
	// details.
	DnsDomain string `mapstructure:"dns_domain"`
	// The IPv4 address assigned to the network adapter. If left blank or not
	// included, DHCP is used.
	Ipv4Address string `mapstructure:"ipv4_address"`
	// The IPv4 subnet mask, in bits, for the network adapter. For example, `24`
	// for a `/24` subnet.
	Ipv4NetMask int `mapstructure:"ipv4_netmask"`
	// The IPv6 address assigned to the network adapter. If left blank or not
	// included, autoconfiguration is used.
	Ipv6Address string `mapstructure:"ipv6_address"`
	// The IPv6 subnet mask, in bits, for the network adapter. For example, `64`
	// for a `/64` subnet.
	Ipv6NetMask int `mapstructure:"ipv6_netmask"`
}

type NetworkInterfaces []NetworkInterface

// The settings must match the IP address and subnet mask of at least one
// `network_interface` for the customization.
type GlobalRoutingSettings struct {
	// The IPv4 default gateway when using `network_interface` customization.
	Ipv4Gateway string `mapstructure:"ipv4_gateway"`
	// The IPv6 default gateway when using `network_interface` customization.
	Ipv6Gateway string `mapstructure:"ipv6_gateway"`
}

// The following settings configure DNS globally for Linux guest operating
// systems. For Windows guest operating systems, this is set for each network
// interface. Refer to the [network interface](#network_interface) section for
// additional details.
type GlobalDnsSettings struct {
	// A list of DNS servers to configure on the guest operating system.
	DnsServerList []string `mapstructure:"dns_server_list"`
	// A list of DNS search domains to add to the DNS configuration on the guest
	// operating system.
	DnsSuffixList []string `mapstructure:"dns_suffix_list"`
}

type StepCustomize struct {
	Config *CustomizeConfig
}

func (c *CustomizeConfig) Prepare() ([]string, []error) {
	var errs []error
	var warnings []string

	if len(c.NetworkInterfaces) == 0 {
		errs = append(errs, fmt.Errorf("one or more `network_interface` must be provided"))
	}

	optionsNumber := 0
	if c.LinuxOptions != nil {
		optionsNumber++
	}
	if c.WindowsOptions != nil {
		optionsNumber++
	}
	if c.WindowsSysPrepFile != "" {
		warnings = append(warnings, windowsSysprepFileDeprecatedMessage)
		optionsNumber++
	}
	if c.WindowsSysPrepText != "" {
		optionsNumber++
	}

	if optionsNumber > 1 {
		errs = append(errs, errCustomizeOptionMutualExclusive)
	} else if optionsNumber == 0 {
		errs = append(errs, fmt.Errorf("one of `linux_options`, `windows_options`, `windows_sysprep_file`, or 'windows_sysprep_text' must be set"))
	}

	if c.LinuxOptions != nil {
		errs = c.LinuxOptions.prepare(errs)
	}
	if c.WindowsOptions != nil {
		errs = c.WindowsOptions.prepare(errs)
	}

	return warnings, errs
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
		sysPrep, err := os.ReadFile(s.Config.WindowsSysPrepFile)
		if err != nil {
			return nil, fmt.Errorf("error on reading %s: %s", s.Config.WindowsSysPrepFile, err)
		}
		return &types.CustomizationSysprepText{
			Value: string(sysPrep),
		}, nil
	}

	if s.Config.WindowsSysPrepText != "" {
		return &types.CustomizationSysprepText{
			Value: s.Config.WindowsSysPrepText,
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
		errs = append(errs, fmt.Errorf("linux options: `host_name` is required"))
	}
	if l.Domain == "" {
		errs = append(errs, fmt.Errorf("linux options: `domain` is required"))
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
		errs = append(errs, fmt.Errorf("windows options: `computer_name` is required"))
	}
	if w.FullName == "" {
		w.FullName = "Administrator"
	}
	if w.OrganizationName == "" {
		w.OrganizationName = "Built by Packer"
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
	if len(w.RunOnceCommandList) == 0 {
		return &types.CustomizationGuiRunOnce{
			CommandList: []string{""},
		}
	}

	return &types.CustomizationGuiRunOnce{
		CommandList: w.RunOnceCommandList,
	}
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

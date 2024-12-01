Type: `vsphere-clone`

Artifact BuilderId: `jetbrains.vsphere`

This builder clones an existing template, modifies the virtual machine image, and saves the result
as a new template using the vSphere API.

-> **Note:** This builder is developed to maintain compatibility with VMware vSphere versions until
their respective End of General Support dates. For detailed information, refer to the
[Broadcom Product Lifecycle](https://support.broadcom.com/group/ecx/productlifecycle).

## Examples

Examples are available in the [examples](https://github.com/hashicorp/packer-plugin-vsphere/tree/main/builder/vsphere/examples/)
directory of the GitHub repository.

## Configuration Reference

There are many configuration options available for this builder. In addition to the items listed
here, you will want to review the general configuration references for [Hardware](#hardware-configuration),
[Output](#output-configuration), [Boot](#boot-configuration), [Run](#run-configuration), [Shutdown](#shutdown-configuration),
[Communicator](#communicator-configuration), and [Export](#export-configuration) configuration
references, which are necessary for a build to succeed and can be found further down the page.

**Optional:**

<!-- Code generated from the comments of the Config struct in builder/vsphere/clone/config.go; DO NOT EDIT MANUALLY -->

- `create_snapshot` (bool) - Create a snapshot of the virtual machine to use as a base for linked
  clones. Defaults to `false`.

- `snapshot_name` (string) - The name of the snapshot when `create_snapshot` is `true`.
  Defaults to `Created By Packer`.

- `convert_to_template` (bool) - Convert the cloned virtual machine to a template after the build is
  complete. Defaults to `false`.
  If set to `true`, the virtual machine can not be imported to a content
  library.

- `export` (\*common.ExportConfig) - The configuration for exporting the virtual machine to an OVF.
  The virtual machine is not exported if [export configuration](#export-configuration)
  is not specified.

- `content_library_destination` (\*common.ContentLibraryDestinationConfig) - The configuration for importing a VM template or OVF template to a
  content library. The template will not be imported if no
  [content library import configuration](#content-library-import-configuration)
  is specified. If set, `convert_to_template` must be set to `false`.

- `customize` (\*CustomizeConfig) - The customization options for the virtual machine.
  Refer to the [customization options](#customization) section for more
  information.

<!-- End of code generated from the comments of the Config struct in builder/vsphere/clone/config.go; -->


### Clone Configuration

**Optional:**

<!-- Code generated from the comments of the CloneConfig struct in builder/vsphere/clone/step_clone.go; DO NOT EDIT MANUALLY -->

- `template` (string) - The name of the source virtual machine to clone.

- `disk_size` (int64) - The size of the primary disk in MiB. Cannot be used with `linked_clone`.
  -> **Note:** Only the primary disk size can be specified. Additional
  disks are not supported.

- `linked_clone` (bool) - Create the virtual machine as a linked clone from the latest snapshot.
  Defaults to `false`. Cannot be used with `disk_size`.`

- `network` (string) - The network to which the virtual machine will connect.
  
  For example:
  
  - Name: `<NetworkName>`
  - Inventory Path: `/<DatacenterName>/<FolderName>/<NetworkName>`
  - Managed Object ID (Port Group): `Network:network-<xxxxx>`
  - Managed Object ID (Distributed Port Group): `DistributedVirtualPortgroup::dvportgroup-<xxxxx>`
  - Logical Switch UUID: `<uuid>`
  - Segment ID: `/infra/segments/<SegmentID>`
  
  ~> **Note:** If more than one network resolves to the same name, either
  the inventory path to network or an ID must be provided.
  
  ~> **Note:** If no network is specified, provide `host` to allow the
  plugin to search for an available network.

- `mac_address` (string) - The network card MAC address. For example `00:50:56:00:00:00`.
  If set, the `network` must be also specified.

- `notes` (string) - The annotations for the virtual machine.

- `destroy` (bool) - Destroy the virtual machine after the build is complete.
  Defaults to `false`.

- `vapp` (vAppConfig) - The vApp Options for the virtual machine. For more information, refer to
  the [vApp Options Configuration](/packer/integrations/hashicorp/vmware/latest/components/builder/vsphere-clone#vapp-options-configuration)
  section.

<!-- End of code generated from the comments of the CloneConfig struct in builder/vsphere/clone/step_clone.go; -->


<!-- Code generated from the comments of the StorageConfig struct in builder/vsphere/common/storage_config.go; DO NOT EDIT MANUALLY -->

- `disk_controller_type` ([]string) - The disk controller type. One of `lsilogic`, `lsilogic-sas`, `pvscsi`,
  `nvme`, `scsi`, or `sata`. Defaults to `lsilogic`. Use a list to define
  additional controllers. Refer to [SCSI, SATA, and NVMe Storage Controller
  Conditions, Limitations, and Compatibility](https://docs.vmware.com/en/VMware-vSphere/8.0/vsphere-vm-administration/GUID-5872D173-A076-42FE-8D0B-9DB0EB0E7362.html)
  for additional information.

- `storage` ([]DiskConfig) - A collection of one or more disks to be provisioned.
  Refer to the [Storage Configuration](#storage-configuration) section for additional information.

<!-- End of code generated from the comments of the StorageConfig struct in builder/vsphere/common/storage_config.go; -->


### Storage Configuration

When cloning a virtual machine, the storage configuration can be used to add additional storage and
disk controllers. The resulting virtual machine will contain the origin virtual machine storage and
disk controller plus the new configured ones.

<!-- Code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; DO NOT EDIT MANUALLY -->

The following example that will create a 15GB and a 20GB disk on the virtual
machine. The second disk will be thin provisioned:

HCL Example:

```hcl

	storage {
	    disk_size = 15000
	}
	storage {
	    disk_size = 20000
	    disk_thin_provisioned = true
	}

```

JSON Example:

```json

	"storage": [
	  {
	    "disk_size": 15000
	  },
	  {
	    "disk_size": 20000,
	    "disk_thin_provisioned": true
	  }
	],

```

The following example will use two PVSCSI controllers and two disks on each
controller.

HCL Example:

```hcl

	 disk_controller_type = ["pvscsi", "pvscsi"]
		storage {
		   disk_size = 15000
		   disk_controller_index = 0
		}
		storage {
		   disk_size = 15000
		   disk_controller_index = 0
		}
		storage {
		   disk_size = 15000
		   disk_controller_index = 1
		}
		storage {
		   disk_size = 15000
		   disk_controller_index = 1
		}

```

JSON Example:

```json

	"disk_controller_type": ["pvscsi", "pvscsi"],
	"storage": [
	  {
	    "disk_size": 15000,
	    "disk_controller_index": 0
	  },
	  {
	    "disk_size": 15000,
	    "disk_controller_index": 0
	  },
	  {
	    "disk_size": 15000,
	    "disk_controller_index": 1
	  },
	  {
	    "disk_size": 15000,
	    "disk_controller_index": 1
	  }
	],

```

<!-- End of code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; -->


**Required:**

<!-- Code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; DO NOT EDIT MANUALLY -->

- `disk_size` (int64) - The size of the disk in MiB.

<!-- End of code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; -->


**Optional:**

<!-- Code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; DO NOT EDIT MANUALLY -->

- `disk_thin_provisioned` (bool) - Enable thin provisioning for the disk.
  Defaults to `false`.

- `disk_eagerly_scrub` (bool) - Enable eager scrubbing for the disk.
  Defaults to `false`.

- `disk_controller_index` (int) - The assigned disk controller for the disk.
  Defaults to the first controller, `(0)`.

<!-- End of code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; -->


### vApp Options Configuration

**Optional:**

<!-- Code generated from the comments of the vAppConfig struct in builder/vsphere/clone/step_clone.go; DO NOT EDIT MANUALLY -->

- `properties` (map[string]string) - The values for the available vApp properties. These are used to supply
  configuration parameters to a virtual machine. This machine is cloned
  from a template that originated from an imported OVF or OVA file.
  
  -> **Note:** The only supported usage path for vApp properties is for
  existing user-configurable keys. These generally come from an existing
  template that was created from an imported OVF or OVA file.
  
  You cannot set values for vApp properties on virtual machines created
  from scratch, on virtual machines that lack a vApp configuration, or on
  property keys that do not exist.
  
  HCL Example:
  ```hcl
    vapp {
      properties = {
        hostname  = var.hostname
        user-data = base64encode(var.user_data)
      }
    }
  ```
  
  JSON Example:
  
  ```json
    "vapp": {
        "properties": {
            "hostname": "{{ user `hostname`}}",
            "user-data": "{{ env `USERDATA`}}"
        }
    }
  ```
  
  A `user-data` field requires the content of a YAML file to be encoded
  with base64. This can be done using an environment variable:
  
  ```console
  export USERDATA=$(gzip -c9 <userdata.yaml | { base64 -w0 2>/dev/null || base64; })
  ```

<!-- End of code generated from the comments of the vAppConfig struct in builder/vsphere/clone/step_clone.go; -->


### Extra Configuration Parameters

**Optional:**

<!-- Code generated from the comments of the ConfigParamsConfig struct in builder/vsphere/common/step_config_params.go; DO NOT EDIT MANUALLY -->

- `configuration_parameters` (map[string]string) - A map of key-value pairs to sent to the [`extraConfig`](https://dp-downloads.broadcom.com/api-content/apis/API_VWSA_001/8.0U3/html/ReferenceGuides/vim.vm.ConfigSpec.html#extraConfig).
  in the vSphere API's `VirtualMachineConfigSpec`.
  
  HCL Example:
  
  ```hcl
    configuration_parameters = {
      "disk.EnableUUID" = "TRUE"
      "svga.autodetect" = "TRUE"
      "log.keepOld"     = "15"
    }
  ```
  
  JSON Example:
  
  ```json
    "configuration_parameters": {
      "disk.EnableUUID": "TRUE",
      "svga.autodetect": "TRUE",
      "log.keepOld": "15"
    }
  ```
  
  ~> **Note:** Configuration keys that would conflict with parameters that
  are explicitly configurable through other fields in the `ConfigSpec`` object
  are silently ignored. Refer to the [`VirtualMachineConfigSpec`](https://dp-downloads.broadcom.com/api-content/apis/API_VWSA_001/8.0U3/html/ReferenceGuides/vim.vm.ConfigSpec.html)
  in the vSphere API documentation.

- `tools_sync_time` (bool) - Enable time synchronization with the ESXi host where the virtual machine
  is running. Defaults to `false`.

- `tools_upgrade_policy` (bool) - Automatically check for and upgrade VMware Tools after a virtual machine
  power cycle. Defaults to `false`.

<!-- End of code generated from the comments of the ConfigParamsConfig struct in builder/vsphere/common/step_config_params.go; -->


### Customization

<!-- Code generated from the comments of the CustomizeConfig struct in builder/vsphere/clone/step_customize.go; DO NOT EDIT MANUALLY -->

A cloned virtual machine can be [customized](https://docs.vmware.com/en/VMware-vSphere/8.0/vsphere-vm-administration/GUID-58E346FF-83AE-42B8-BE58-253641D257BC.html)
to configure host, network, or licensing settings.

To perform virtual machine customization as a part of the clone process,
specify the customize block with the respective customization options.
Windows guests are customized using Sysprep, which will result in the machine
SID being reset. Before using customization, check that your source virtual
machine meets the [requirements](https://docs.vmware.com/en/VMware-vSphere/8.0/vsphere-vm-administration/GUID-E63B6FAA-8D35-428D-B40C-744769845906.html)
for guest OS customization on vSphere. Refer to the [customization example](#customization-example) for a usage synopsis.

The settings for guest customization include:

<!-- End of code generated from the comments of the CustomizeConfig struct in builder/vsphere/clone/step_customize.go; -->


**Optional:**

<!-- Code generated from the comments of the CustomizeConfig struct in builder/vsphere/clone/step_customize.go; DO NOT EDIT MANUALLY -->

- `linux_options` (\*LinuxOptions) - Settings for the guest customization of Linux operating systems.
  Refer to the [Linux options](#linux-options) section for additional
  details.

- `windows_options` (\*WindowsOptions) - Settings for the guest customization of Windows operating systems.
  Refer to the [Windows options](#windows-options) section for additional
  details.

- `windows_sysprep_file` (string) - Provide a `sysprep.xml` file to allow control of the customization
  process independent of vSphere. This option is deprecated, please use
  `windows_sysprep_text`.

- `windows_sysprep_text` (string) - Provide the text for the `sysprep.xml` content to allow control of the
  customization process independent of vSphere.
  
  HCL Examples:
  
  ```hcl
  customize {
     windows_sysprep_text = file("<path-to-sysprep.xml>")
  }
  ```
  
  ```hcl
  customize {
     windows_sysprep_text = templatefile("<path-to-sysprep.xml>", {
        var1="example"
        var2="example-2"
     })
  }
  ```
  
  JSON Examples
  
  ```json
  {
    "customize": {
      "windows_sysprep_text": "<path-to-sysprep.xml>"
    }
  }
  ```
  
  ```json
  {
    "customize": {
      "windows_sysprep_text": "<path-to-sysprep.xml>",
      "var1": "example",
      "var2": "example-2"
    }
  }

- `network_interface` (NetworkInterfaces) - Set up network interfaces individually to correspond with the network
  adapters on the virtual machine. To use DHCP, specify an empty
  `network_interface` for each configured adapter. This field is mandatory.
  Refer to the [network interface](#network-interface-settings) section for
  additional details.

<!-- End of code generated from the comments of the CustomizeConfig struct in builder/vsphere/clone/step_customize.go; -->


#### Network Interface Settings

**Optional:**

<!-- Code generated from the comments of the NetworkInterface struct in builder/vsphere/clone/step_customize.go; DO NOT EDIT MANUALLY -->

- `dns_server_list` ([]string) - The DNS servers for a specific network interface on a Windows guest
  operating system. Ignored on Linux. Refer to the
  [global DNS settings](#global-dns-settings) section for additional
  details.

- `dns_domain` (string) - The DNS search domain for a specific network interface on a Windows guest
  operating system. Ignored on Linux. Refer to the
  [global DNS settings](#global-dns-settings) section for additional
  details.

- `ipv4_address` (string) - The IPv4 address assigned to the network adapter. If left blank or not
  included, DHCP is used.

- `ipv4_netmask` (int) - The IPv4 subnet mask, in bits, for the network adapter. For example, `24`
  for a `/24` subnet.

- `ipv6_address` (string) - The IPv6 address assigned to the network adapter. If left blank or not
  included, autoconfiguration is used.

- `ipv6_netmask` (int) - The IPv6 subnet mask, in bits, for the network adapter. For example, `64`
  for a `/64` subnet.

<!-- End of code generated from the comments of the NetworkInterface struct in builder/vsphere/clone/step_customize.go; -->


<!-- Code generated from the comments of the RemoveNetworkConfig struct in builder/vsphere/common/step_remove_network.go; DO NOT EDIT MANUALLY -->

- `remove_network_adapter` (bool) - Remove all network adapters from template. Defaults to `false`.

<!-- End of code generated from the comments of the RemoveNetworkConfig struct in builder/vsphere/common/step_remove_network.go; -->


#### Global Routing Settings

<!-- Code generated from the comments of the GlobalRoutingSettings struct in builder/vsphere/clone/step_customize.go; DO NOT EDIT MANUALLY -->

The settings must match the IP address and subnet mask of at least one
`network_interface` for the customization.

<!-- End of code generated from the comments of the GlobalRoutingSettings struct in builder/vsphere/clone/step_customize.go; -->


**Optional:**

<!-- Code generated from the comments of the GlobalRoutingSettings struct in builder/vsphere/clone/step_customize.go; DO NOT EDIT MANUALLY -->

- `ipv4_gateway` (string) - The IPv4 default gateway when using `network_interface` customization.

- `ipv6_gateway` (string) - The IPv6 default gateway when using `network_interface` customization.

<!-- End of code generated from the comments of the GlobalRoutingSettings struct in builder/vsphere/clone/step_customize.go; -->


#### Global DNS Settings

<!-- Code generated from the comments of the GlobalDnsSettings struct in builder/vsphere/clone/step_customize.go; DO NOT EDIT MANUALLY -->

The following settings configure DNS globally for Linux guest operating
systems. For Windows guest operating systems, this is set for each network
interface. Refer to the [network interface](#network_interface) section for
additional details.

<!-- End of code generated from the comments of the GlobalDnsSettings struct in builder/vsphere/clone/step_customize.go; -->


**Optional:**

<!-- Code generated from the comments of the GlobalDnsSettings struct in builder/vsphere/clone/step_customize.go; DO NOT EDIT MANUALLY -->

- `dns_server_list` ([]string) - A list of DNS servers to configure on the guest operating system.

- `dns_suffix_list` ([]string) - A list of DNS search domains to add to the DNS configuration on the guest
  operating system.

<!-- End of code generated from the comments of the GlobalDnsSettings struct in builder/vsphere/clone/step_customize.go; -->


#### Linux Customization Settings

**Optional:**

<!-- Code generated from the comments of the LinuxOptions struct in builder/vsphere/clone/step_customize.go; DO NOT EDIT MANUALLY -->

- `domain` (string) - The domain name for the guest operating system. Used with
  [host_name](#host_name) to construct the fully qualified domain name
  (FQDN).

- `host_name` (string) - The hostname for the guest operating system. Used with [domain](#domain)
  to construct the fully qualified domain name (FQDN).

- `hw_clock_utc` (boolean) - Set the hardware clock to Coordinated Universal Time (UTC).
  Defaults to `true`.

- `time_zone` (string) - The time zone for the guest operating system.

<!-- End of code generated from the comments of the LinuxOptions struct in builder/vsphere/clone/step_customize.go; -->


**Linux Customization Example**

HCL Example:

```hcl
    customize {
      linux_options {
        host_name = "foo"
        domain = "example.com"
      }

      network_interface {
        ipv4_address = "10.0.0.10"
        ipv4_netmask = "24"
      }

      ipv4_gateway = 10.0.0.1
      dns_server_list = ["10.0.0.18"]
    }
```

JSON Example:

```json
    "customize": {
      "linux_options": {
        "host_name": "foo",
        "domain": "example.com"
      },
      "network_interface": {
        "ipv4_address": "10.0.0.10",
        "ipv4_netmask": "24"
      },
      "ipv4_gateway": "10.0.0.1",
      "dns_server_list": ["10.0.0.18"]
    }
```

#### Windows Customization Settings

**Optional:**

<!-- Code generated from the comments of the WindowsOptions struct in builder/vsphere/clone/step_customize.go; DO NOT EDIT MANUALLY -->

- `run_once_command_list` ([]string) - A list of commands to run at first logon after the guest operating system
  is customized.

- `auto_logon` (\*bool) - Automatically log on the `Administrator` account after the guest operating
  system is customized.

- `auto_logon_count` (\*int32) - The number of times the guest operating system should auto-logon the
  `Administrator` account when `auto_logon` is set to `true`.
  Defaults to `1`.

- `admin_password` (\*string) - The password for the guest operating system's `Administrator` account.

- `time_zone` (\*int32) - The time zone for the guest operating system.
  Defaults to `85` (Pacific Time).

- `workgroup` (string) - The workgroup for the guest operating system.
  Joining an Active Directory domain is not supported.

- `computer_name` (string) - The hostname for the guest operating system.

- `full_name` (string) - The full name for the guest operating system's `Administrator` account.
  Defaults to `Administrator`.

- `organization_name` (string) - The organization name for the guest operating system.
  Defaults to `Built by Packer`.

- `product_key` (string) - The product key for the guest operating system.

<!-- End of code generated from the comments of the WindowsOptions struct in builder/vsphere/clone/step_customize.go; -->


**Windows Customization Example**

HCL Example:

```hcl
    customize {
      windows_options {
        computer_name = "foo"
        workgroup = "example"
        product_key = "XXXXX-XXXXX-XXXXX-XXXXX-XXXXX"
        admin_password = "password"
      }
      network_interface {
        ipv4_address = "10.0.0.10"
        ipv4_netmask = "24"
      }
      ipv4_gateway = 10.0.0.1
      dns_server_list = ["10.0.0.18"]
    }
```

JSON Example:

```json
    "customize": {
      "windows_options": {
        "host_name": "foo",
        "workgroup": "example",
        "product_key": "XXXXX-XXXXX-XXXXX-XXXXX-XXXXX",
        "admin_password": "password"
      },
      "network_interface": {
        "ipv4_address": "10.0.0.10",
        "ipv4_netmask": "24"
      },
      "ipv4_gateway": "10.0.0.1",
      "dns_server_list": ["10.0.0.18"]
    }
```

### Boot Configuration

<!-- Code generated from the comments of the BootConfig struct in bootcommand/config.go; DO NOT EDIT MANUALLY -->

The boot configuration is very important: `boot_command` specifies the keys
to type when the virtual machine is first booted in order to start the OS
installer. This command is typed after boot_wait, which gives the virtual
machine some time to actually load.

The boot_command is an array of strings. The strings are all typed in
sequence. It is an array only to improve readability within the template.

There are a set of special keys available. If these are in your boot
command, they will be replaced by the proper key:

-   `<bs>` - Backspace

-   `<del>` - Delete

-   `<enter> <return>` - Simulates an actual "enter" or "return" keypress.

-   `<esc>` - Simulates pressing the escape key.

-   `<tab>` - Simulates pressing the tab key.

-   `<f1> - <f12>` - Simulates pressing a function key.

-   `<up> <down> <left> <right>` - Simulates pressing an arrow key.

-   `<spacebar>` - Simulates pressing the spacebar.

-   `<insert>` - Simulates pressing the insert key.

-   `<home> <end>` - Simulates pressing the home and end keys.

  - `<pageUp> <pageDown>` - Simulates pressing the page up and page down
    keys.

-   `<menu>` - Simulates pressing the Menu key.

-   `<leftAlt> <rightAlt>` - Simulates pressing the alt key.

-   `<leftCtrl> <rightCtrl>` - Simulates pressing the ctrl key.

-   `<leftShift> <rightShift>` - Simulates pressing the shift key.

-   `<leftSuper> <rightSuper>` - Simulates pressing the ⌘ or Windows key.

  - `<wait> <wait5> <wait10>` - Adds a 1, 5 or 10 second pause before
    sending any additional keys. This is useful if you have to generally
    wait for the UI to update before typing more.

  - `<waitXX>` - Add an arbitrary pause before sending any additional keys.
    The format of `XX` is a sequence of positive decimal numbers, each with
    optional fraction and a unit suffix, such as `300ms`, `1.5h` or `2h45m`.
    Valid time units are `ns`, `us` (or `µs`), `ms`, `s`, `m`, `h`. For
    example `<wait10m>` or `<wait1m20s>`.

  - `<XXXOn> <XXXOff>` - Any printable keyboard character, and of these
    "special" expressions, with the exception of the `<wait>` types, can
    also be toggled on or off. For example, to simulate ctrl+c, use
    `<leftCtrlOn>c<leftCtrlOff>`. Be sure to release them, otherwise they
    will be held down until the machine reboots. To hold the `c` key down,
    you would use `<cOn>`. Likewise, `<cOff>` to release.

  - `{{ .HTTPIP }} {{ .HTTPPort }}` - The IP and port, respectively of an
    HTTP server that is started serving the directory specified by the
    `http_directory` configuration parameter. If `http_directory` isn't
    specified, these will be blank!

-   `{{ .Name }}` - The name of the VM.

Example boot command. This is actually a working boot command used to start an
CentOS 6.4 installer:

In JSON:

```json
"boot_command": [

	   "<tab><wait>",
	   " ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/centos6-ks.cfg<enter>"
	]

```

In HCL2:

```hcl
boot_command = [

	   "<tab><wait>",
	   " ks=http://{{ .HTTPIP }}:{{ .HTTPPort }}/centos6-ks.cfg<enter>"
	]

```

The example shown below is a working boot command used to start an Ubuntu
12.04 installer:

In JSON:

```json
"boot_command": [

	"<esc><esc><enter><wait>",
	"/install/vmlinuz noapic ",
	"preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg ",
	"debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
	"hostname={{ .Name }} ",
	"fb=false debconf/frontend=noninteractive ",
	"keyboard-configuration/modelcode=SKIP keyboard-configuration/layout=USA ",
	"keyboard-configuration/variant=USA console-setup/ask_detect=false ",
	"initrd=/install/initrd.gz -- <enter>"

]
```

In HCL2:

```hcl
boot_command = [

	"<esc><esc><enter><wait>",
	"/install/vmlinuz noapic ",
	"preseed/url=http://{{ .HTTPIP }}:{{ .HTTPPort }}/preseed.cfg ",
	"debian-installer=en_US auto locale=en_US kbd-chooser/method=us ",
	"hostname={{ .Name }} ",
	"fb=false debconf/frontend=noninteractive ",
	"keyboard-configuration/modelcode=SKIP keyboard-configuration/layout=USA ",
	"keyboard-configuration/variant=USA console-setup/ask_detect=false ",
	"initrd=/install/initrd.gz -- <enter>"

]
```

For more examples of various boot commands, see the sample projects from our
[community templates page](https://packer.io/community-tools#templates).

<!-- End of code generated from the comments of the BootConfig struct in bootcommand/config.go; -->


**Optional:**

<!-- Code generated from the comments of the BootConfig struct in bootcommand/config.go; DO NOT EDIT MANUALLY -->

- `boot_keygroup_interval` (duration string | ex: "1h5m2s") - Time to wait after sending a group of key pressses. The value of this
  should be a duration. Examples are `5s` and `1m30s` which will cause
  Packer to wait five seconds and one minute 30 seconds, respectively. If
  this isn't specified, a sensible default value is picked depending on
  the builder type.

- `boot_wait` (duration string | ex: "1h5m2s") - The time to wait after booting the initial virtual machine before typing
  the `boot_command`. The value of this should be a duration. Examples are
  `5s` and `1m30s` which will cause Packer to wait five seconds and one
  minute 30 seconds, respectively. If this isn't specified, the default is
  `10s` or 10 seconds. To set boot_wait to 0s, use a negative number, such
  as "-1s"

- `boot_command` ([]string) - This is an array of commands to type when the virtual machine is first
  booted. The goal of these commands should be to type just enough to
  initialize the operating system installer. Special keys can be typed as
  well, and are covered in the section below on the boot command. If this
  is not specified, it is assumed the installer will start itself.

<!-- End of code generated from the comments of the BootConfig struct in bootcommand/config.go; -->


### HTTP Directory Configuration

<!-- Code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; DO NOT EDIT MANUALLY -->

Packer will create an http server serving `http_directory` when it is set, a
random free port will be selected and the architecture of the directory
referenced will be available in your builder.

Example usage from a builder:

```
wget http://{{ .HTTPIP }}:{{ .HTTPPort }}/foo/bar/preseed.cfg
```

<!-- End of code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; -->


**Optional:**

<!-- Code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; DO NOT EDIT MANUALLY -->

- `http_directory` (string) - Path to a directory to serve using an HTTP server. The files in this
  directory will be available over HTTP that will be requestable from the
  virtual machine. This is useful for hosting kickstart files and so on.
  By default this is an empty string, which means no HTTP server will be
  started. The address and port of the HTTP server will be available as
  variables in `boot_command`. This is covered in more detail below.

- `http_content` (map[string]string) - Key/Values to serve using an HTTP server. `http_content` works like and
  conflicts with `http_directory`. The keys represent the paths and the
  values contents, the keys must start with a slash, ex: `/path/to/file`.
  `http_content` is useful for hosting kickstart files and so on. By
  default this is empty, which means no HTTP server will be started. The
  address and port of the HTTP server will be available as variables in
  `boot_command`. This is covered in more detail below.
  Example:
  ```hcl
    http_content = {
      "/a/b"     = file("http/b")
      "/foo/bar" = templatefile("${path.root}/preseed.cfg", { packages = ["nginx"] })
    }
  ```

- `http_port_min` (int) - These are the minimum and maximum port to use for the HTTP server
  started to serve the `http_directory`. Because Packer often runs in
  parallel, Packer will choose a randomly available port in this range to
  run the HTTP server. If you want to force the HTTP server to be on one
  port, make this minimum and maximum port the same. By default the values
  are `8000` and `9000`, respectively.

- `http_port_max` (int) - HTTP Port Max

- `http_bind_address` (string) - This is the bind address for the HTTP server. Defaults to 0.0.0.0 so that
  it will work with any network interface.

<!-- End of code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; -->


- `http_interface` (string) - The network interface (for example, `en0`, `ens192`, etc.) that the
  HTTP server will use to serve the `http_directory`. The plugin will identify the IP address
  associated with this network interface and bind to it.

<!-- Code generated from the comments of the BootConfig struct in builder/vsphere/common/step_boot_command.go; DO NOT EDIT MANUALLY -->

- `http_ip` (string) - The IP address to use for the HTTP server to serve the `http_directory`.

<!-- End of code generated from the comments of the BootConfig struct in builder/vsphere/common/step_boot_command.go; -->


~> **Notes:**
  - The options `http_bind_address` and `http_interface` are mutually exclusive.
  - Both `http_bind_address` and `http_interface` have higher priority than `http_ip`.
  - The `http_bind_address` is matched against the IP addresses of the host's network interfaces. If
    no match is found, the plugin will terminate.
  - Similarly, `http_interface` is compared with the host's network interfaces. If there's no
    corresponding network interface, the plugin will also terminate.
  - If neither `http_bind_address`, `http_interface`, and `http_ip` are provided, the plugin will
    automatically find and use the IP address of the first non-loopback interface for `http_ip`.

### Floppy Configuration

**Optional:**

<!-- Code generated from the comments of the FloppyConfig struct in builder/vsphere/common/step_add_floppy.go; DO NOT EDIT MANUALLY -->

- `floppy_img_path` (string) - Datastore path to a floppy image that will be mounted to the virtual
  machine. Example: `[datastore] iso/foo.flp`.

- `floppy_files` ([]string) - A list of local files to be mounted to the virtual machine's floppy
  drive.

- `floppy_dirs` ([]string) - A list of directories to copy files from.

- `floppy_content` (map[string]string) - Key/Values to add to the floppy disk. The keys represent the paths, and
  the values contents. It can be used alongside `floppy_files` or
  `floppy_dirs`, which is useful to add large files without loading them
  into memory. If any paths are specified by both, the contents in
  `floppy_content` will take precedence.
  
  HCL Example:
  
  ```hcl
  floppy_content = {
    "meta-data" = jsonencode(local.instance_data)
    "user-data" = templatefile("user-data", { packages = ["nginx"] })
  }
  ```

- `floppy_label` (string) - The label to use for the floppy disk that is attached when the virtual
  machine is booted. This is most useful for cloud-init, Kickstart or other
  early initialization tools, which can benefit from labelled floppy disks.
  By default, the floppy label will be 'packer'.

<!-- End of code generated from the comments of the FloppyConfig struct in builder/vsphere/common/step_add_floppy.go; -->


### Connection Configuration

**Optional:**

<!-- Code generated from the comments of the ConnectConfig struct in builder/vsphere/common/step_connect.go; DO NOT EDIT MANUALLY -->

- `vcenter_server` (string) - The fully qualified domain name or IP address of the vCenter Server
  instance.

- `username` (string) - The username to authenticate with the vCenter Server instance.

- `password` (string) - The password to authenticate with the vCenter Server instance.

- `insecure_connection` (bool) - Do not validate the certificate of the vCenter Server instance.
  Defaults to `false`.
  
  -> **Note:** This option is beneficial in scenarios where the certificate
  is self-signed or does not meet standard validation criteria.

- `datacenter` (string) - The name of the datacenter object in the vSphere inventory.
  
  -> **Note:** Required if more than one datacenter object exists in the
  vSphere inventory.

<!-- End of code generated from the comments of the ConnectConfig struct in builder/vsphere/common/step_connect.go; -->


### Hardware Configuration

**Optional:**

<!-- Code generated from the comments of the HardwareConfig struct in builder/vsphere/common/step_hardware.go; DO NOT EDIT MANUALLY -->

- `CPUs` (int32) - The number of virtual CPUs cores for the virtual machine.

- `cpu_cores` (int32) - The number of virtual CPU cores per socket for the virtual machine.

- `CPU_reservation` (int64) - The CPU reservation in MHz.

- `CPU_limit` (int64) - The upper limit of available CPU resources in MHz.

- `CPU_hot_plug` (bool) - Enable CPU hot plug setting for virtual machine. Defaults to `false`

- `RAM` (int64) - The amount of memory for the virtual machine in MB.

- `RAM_reservation` (int64) - The guaranteed minimum allocation of memory for the virtual machine in MB.

- `RAM_reserve_all` (bool) - Reserve all allocated memory. Defaults to `false`.
  
  -> **Note:** May not be used together with `RAM_reservation`.

- `RAM_hot_plug` (bool) - Enable memory hot add setting for virtual machine. Defaults to `false`.

- `video_ram` (int64) - The amount of video memory in KB. Defaults to 4096 KB.
  
  -> **Note:** Refer to the [vSphere documentation](https://docs.vmware.com/en/VMware-vSphere/8.0/vsphere-vm-administration/GUID-789C3913-1053-4850-A0F0-E29C3D32B6DA.html)
  for supported maximums.

- `displays` (int32) - The number of video displays. Defaults to `1`.
  
  `-> **Note:** Refer to the [vSphere documentation](https://docs.vmware.com/en/VMware-vSphere/8.0/vsphere-vm-administration/GUID-789C3913-1053-4850-A0F0-E29C3D32B6DA.html)
  for supported maximums.

- `pci_passthrough_allowed_device` ([]PCIPassthroughAllowedDevice) - Configure Dynamic DirectPath I/O [PCI Passthrough](#pci-passthrough-configuration) for
  virtual machine. Refer to the [vSphere documentation](https://docs.vmware.com/en/VMware-vSphere/8.0/vsphere-vm-administration/GUID-5B3CAB26-5D06-4A99-92A0-3A04C69CE64B.html)

- `vgpu_profile` (string) - vGPU profile for accelerated graphics. Refer to the [NVIDIA GRID vGPU documentation](https://docs.nvidia.com/grid/latest/grid-vgpu-user-guide/index.html#configure-vmware-vsphere-vm-with-vgpu)
  for examples of profile names. Defaults to none.

- `NestedHV` (bool) - Enable nested hardware virtualization for the virtual machine.
  Defaults to `false`.

- `firmware` (string) - The firmware for the virtual machine.
  
  The available options for this setting are: 'bios', 'efi', and
  'efi-secure'.
  
  -> **Note:** Use `efi-secure` for UEFI Secure Boot.

- `force_bios_setup` (bool) - Force entry into the BIOS setup screen during boot. Defaults to `false`.

- `vTPM` (bool) - Enable virtual trusted platform module (TPM) device for the virtual
  machine. Defaults to `false`.

- `precision_clock` (string) - The virtual precision clock device for the virtual machine.
  Defaults to `none`.
  
  The available options for this setting are: `none`, `ntp`, and `ptp`.

<!-- End of code generated from the comments of the HardwareConfig struct in builder/vsphere/common/step_hardware.go; -->


### Location Configuration

**Optional:**

<!-- Code generated from the comments of the LocationConfig struct in builder/vsphere/common/config_location.go; DO NOT EDIT MANUALLY -->

- `vm_name` (string) - The name of the virtual machine.

- `folder` (string) - The virtual machine folder where the virtual machine is created.

- `cluster` (string) - The cluster where the virtual machine is created.
  Refer to the [Working With Clusters And Hosts](#working-with-clusters-and-hosts)
  section for more details.

- `host` (string) - The ESXi host where the virtual machine is created. A full path must be
  specified if the ESXi host is in a folder. For example `folder/host`.
  Refer to the [Working With Clusters And Hosts](#working-with-clusters-and-hosts)
  section for more details.

- `resource_pool` (string) - The resource pool where the virtual machine is created.
  If this is not specified, the root resource pool associated with the
  `host` or `cluster` is used.
  
  ~> **Note:**  The full path to the resource pool must be provided.
  For example, a simple resource pool path might resemble `rp-packer` and
  a nested path might resemble 'rp-packer/rp-linux-images'.

- `datastore` (string) - The datastore where the virtual machine is created.
  Required if `host` is a cluster, or if `host` has multiple datastores.

- `set_host_for_datastore_uploads` (bool) - The ESXI host used for uploading files to the datastore.
  Defaults to `false`.

<!-- End of code generated from the comments of the LocationConfig struct in builder/vsphere/common/config_location.go; -->


### Run Configuration

**Optional:**

<!-- Code generated from the comments of the RunConfig struct in builder/vsphere/common/step_run.go; DO NOT EDIT MANUALLY -->

- `boot_order` (string) - The priority of boot devices. Defaults to `disk,cdrom`.
  
  The available boot devices are: `floppy`, `cdrom`, `ethernet`, and
  `disk`.
  
  -> **Note:** If not set, the boot order is temporarily set to
  `disk,cdrom` for the duration of the build and then cleared upon
  build completion.

<!-- End of code generated from the comments of the RunConfig struct in builder/vsphere/common/step_run.go; -->


### Shutdown Configuration

**Optional:**

<!-- Code generated from the comments of the ShutdownConfig struct in builder/vsphere/common/step_shutdown.go; DO NOT EDIT MANUALLY -->

- `shutdown_command` (string) - Specify a virtual machine guest shutdown command. This command will be run using
  the `communicator`. Otherwise, the VMware Tools are used to gracefully shut down
  the virtual machine.

- `shutdown_timeout` (duration string | ex: "1h5m2s") - Amount of time to wait for graceful shut down of the virtual machine.
  Defaults to `5m` (5 minutes).
  This will likely need to be modified if the `communicator` is 'none'.

- `disable_shutdown` (bool) - Packer normally halts the virtual machine after all provisioners have
  run when no `shutdown_command` is defined. If this is set to `true`, Packer
  *will not* halt the virtual machine but will assume that you will send the stop
  signal yourself through a `preseed.cfg`, a script or the final provisioner.
  Packer will wait for a default of 5 minutes until the virtual machine is shutdown.
  The timeout can be changed using `shutdown_timeout` option.

<!-- End of code generated from the comments of the ShutdownConfig struct in builder/vsphere/common/step_shutdown.go; -->


### Wait Configuration

**Optional:**

<!-- Code generated from the comments of the WaitIpConfig struct in builder/vsphere/common/step_wait_for_ip.go; DO NOT EDIT MANUALLY -->

- `ip_wait_timeout` (duration string | ex: "1h5m2s") - Amount of time to wait for VM's IP, similar to 'ssh_timeout'.
  Defaults to `30m` (30 minutes). Refer to the Golang
  [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
  documentation for full details.

- `ip_settle_timeout` (duration string | ex: "1h5m2s") - Amount of time to wait for VM's IP to settle down, sometimes VM may
  report incorrect IP initially, then it is recommended to set that
  parameter to apx. 2 minutes. Examples `45s` and `10m`.
  Defaults to `5s` (5 seconds). Refer to the Golang
  [ParseDuration](https://golang.org/pkg/time/#ParseDuration)
  documentation for full details.

- `ip_wait_address` (\*string) - Set this to a CIDR address to cause the service to wait for an address that is contained in
  this network range. Defaults to `0.0.0.0/0` for any IPv4 address. Examples include:
  
  * empty string ("") - remove all filters
  * `0:0:0:0:0:0:0:0/0` - allow only ipv6 addresses
  * `192.168.1.0/24` - only allow ipv4 addresses from 192.168.1.1 to 192.168.1.254

<!-- End of code generated from the comments of the WaitIpConfig struct in builder/vsphere/common/step_wait_for_ip.go; -->


### CD-ROM Configuration

<!-- Code generated from the comments of the CDConfig struct in multistep/commonsteps/extra_iso_config.go; DO NOT EDIT MANUALLY -->

An iso (CD) containing custom files can be made available for your build.

By default, no extra CD will be attached. All files listed in this setting
get placed into the root directory of the CD and the CD is attached as the
second CD device.

This config exists to work around modern operating systems that have no
way to mount floppy disks, which was our previous go-to for adding files at
boot time.

<!-- End of code generated from the comments of the CDConfig struct in multistep/commonsteps/extra_iso_config.go; -->


**Optional:**

<!-- Code generated from the comments of the CDConfig struct in multistep/commonsteps/extra_iso_config.go; DO NOT EDIT MANUALLY -->

- `cd_files` ([]string) - A list of files to place onto a CD that is attached when the VM is
  booted. This can include either files or directories; any directories
  will be copied onto the CD recursively, preserving directory structure
  hierarchy. Symlinks will have the link's target copied into the directory
  tree on the CD where the symlink was. File globbing is allowed.
  
  Usage example (JSON):
  
  ```json
  "cd_files": ["./somedirectory/meta-data", "./somedirectory/user-data"],
  "cd_label": "cidata",
  ```
  
  Usage example (HCL):
  
  ```hcl
  cd_files = ["./somedirectory/meta-data", "./somedirectory/user-data"]
  cd_label = "cidata"
  ```
  
  The above will create a CD with two files, user-data and meta-data in the
  CD root. This specific example is how you would create a CD that can be
  used for an Ubuntu 20.04 autoinstall.
  
  Since globbing is also supported,
  
  ```hcl
  cd_files = ["./somedirectory/*"]
  cd_label = "cidata"
  ```
  
  Would also be an acceptable way to define the above cd. The difference
  between providing the directory with or without the glob is whether the
  directory itself or its contents will be at the CD root.
  
  Use of this option assumes that you have a command line tool installed
  that can handle the iso creation. Packer will use one of the following
  tools:
  
    * xorriso
    * mkisofs
    * hdiutil (normally found in macOS)
    * oscdimg (normally found in Windows as part of the Windows ADK)

- `cd_content` (map[string]string) - Key/Values to add to the CD. The keys represent the paths, and the values
  contents. It can be used alongside `cd_files`, which is useful to add large
  files without loading them into memory. If any paths are specified by both,
  the contents in `cd_content` will take precedence.
  
  Usage example (HCL):
  
  ```hcl
  cd_files = ["vendor-data"]
  cd_content = {
    "meta-data" = jsonencode(local.instance_data)
    "user-data" = templatefile("user-data", { packages = ["nginx"] })
  }
  cd_label = "cidata"
  ```

- `cd_label` (string) - CD Label

<!-- End of code generated from the comments of the CDConfig struct in multistep/commonsteps/extra_iso_config.go; -->


<!-- Code generated from the comments of the CDRomConfig struct in builder/vsphere/common/step_add_cdrom.go; DO NOT EDIT MANUALLY -->

- `cdrom_type` (string) - The type of controller to use for the CD-ROM device. Defaults to `ide`.
  
  The available options for this setting are: `ide` and `sata`.

- `iso_paths` ([]string) - A list of paths to ISO files in either a datastore or a content library
  that will be attached to the virtual machine.
  
  HCL Example:
  
  ```hcl
  iso_paths = [
    "[nfs] iso/ubuntu-server-amd64.iso",
    "Example Content Library/ubuntu-server-amd64/ubuntu-server-amd64.iso"
  ]
  ```
  
  JSON Example:
  
  ```json
  "iso_paths": [
    "[nfs] iso/ubuntu-server-amd64.iso",
    "Example Content Library/ubuntu-server-amd64/ubuntu-server-amd64.iso"
  ]
  ```
  
  Two ISOs are referenced:
  
  1. An ISO in the "_iso_" folder of the "_nfs_" datastore with the file
    name of "_ubuntu-server-amd64.iso_". "_ubuntu-server-amd64.iso_".
  2. An ISO in the "_Example Content Library_" content library with the
    item name of "_ubuntu-server-amd64_".
  
  -> **Note:** All files in a content library have an associated item name.
  To determine the file name, view the datastore backing the content
  library or use the `govc` vSphere CLI.

<!-- End of code generated from the comments of the CDRomConfig struct in builder/vsphere/common/step_add_cdrom.go; -->


<!-- Code generated from the comments of the RemoveCDRomConfig struct in builder/vsphere/common/step_remove_cdrom.go; DO NOT EDIT MANUALLY -->

- `remove_cdrom` (bool) - Remove all CD-ROM devices from the virtual machine when the build is
  complete. Defaults to `false`.

<!-- End of code generated from the comments of the RemoveCDRomConfig struct in builder/vsphere/common/step_remove_cdrom.go; -->


### Communicator Configuration

#### Common

**Optional:**

<!-- Code generated from the comments of the Config struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `communicator` (string) - Packer currently supports three kinds of communicators:
  
  -   `none` - No communicator will be used. If this is set, most
      provisioners also can't be used.
  
  -   `ssh` - An SSH connection will be established to the machine. This
      is usually the default.
  
  -   `winrm` - A WinRM connection will be established.
  
  In addition to the above, some builders have custom communicators they
  can use. For example, the Docker builder has a "docker" communicator
  that uses `docker exec` and `docker cp` to execute scripts and copy
  files.

- `pause_before_connecting` (duration string | ex: "1h5m2s") - We recommend that you enable SSH or WinRM as the very last step in your
  guest's bootstrap script, but sometimes you may have a race condition
  where you need Packer to wait before attempting to connect to your
  guest.
  
  If you end up in this situation, you can use the template option
  `pause_before_connecting`. By default, there is no pause. For example if
  you set `pause_before_connecting` to `10m` Packer will check whether it
  can connect, as normal. But once a connection attempt is successful, it
  will disconnect and then wait 10 minutes before connecting to the guest
  and beginning provisioning.

<!-- End of code generated from the comments of the Config struct in communicator/config.go; -->


#### SSH

**Optional:**

<!-- Code generated from the comments of the SSH struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `ssh_host` (string) - The address to SSH to. This usually is automatically configured by the
  builder.

- `ssh_port` (int) - The port to connect to SSH. This defaults to `22`.

- `ssh_username` (string) - The username to connect to SSH with. Required if using SSH.

- `ssh_password` (string) - A plaintext password to use to authenticate with SSH.

- `ssh_ciphers` ([]string) - This overrides the value of ciphers supported by default by Golang.
  The default value is [
    "aes128-gcm@openssh.com",
    "chacha20-poly1305@openssh.com",
    "aes128-ctr", "aes192-ctr", "aes256-ctr",
  ]
  
  Valid options for ciphers include:
  "aes128-ctr", "aes192-ctr", "aes256-ctr", "aes128-gcm@openssh.com",
  "chacha20-poly1305@openssh.com",
  "arcfour256", "arcfour128", "arcfour", "aes128-cbc", "3des-cbc",

- `ssh_clear_authorized_keys` (bool) - If true, Packer will attempt to remove its temporary key from
  `~/.ssh/authorized_keys` and `/root/.ssh/authorized_keys`. This is a
  mostly cosmetic option, since Packer will delete the temporary private
  key from the host system regardless of whether this is set to true
  (unless the user has set the `-debug` flag). Defaults to "false";
  currently only works on guests with `sed` installed.

- `ssh_key_exchange_algorithms` ([]string) - If set, Packer will override the value of key exchange (kex) algorithms
  supported by default by Golang. Acceptable values include:
  "curve25519-sha256@libssh.org", "ecdh-sha2-nistp256",
  "ecdh-sha2-nistp384", "ecdh-sha2-nistp521",
  "diffie-hellman-group14-sha1", and "diffie-hellman-group1-sha1".

- `ssh_certificate_file` (string) - Path to user certificate used to authenticate with SSH.
  The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_pty` (bool) - If `true`, a PTY will be requested for the SSH connection. This defaults
  to `false`.

- `ssh_timeout` (duration string | ex: "1h5m2s") - The time to wait for SSH to become available. Packer uses this to
  determine when the machine has booted so this is usually quite long.
  Example value: `10m`.
  This defaults to `5m`, unless `ssh_handshake_attempts` is set.

- `ssh_disable_agent_forwarding` (bool) - If true, SSH agent forwarding will be disabled. Defaults to `false`.

- `ssh_handshake_attempts` (int) - The number of handshakes to attempt with SSH once it can connect.
  This defaults to `10`, unless a `ssh_timeout` is set.

- `ssh_bastion_host` (string) - A bastion host to use for the actual SSH connection.

- `ssh_bastion_port` (int) - The port of the bastion host. Defaults to `22`.

- `ssh_bastion_agent_auth` (bool) - If `true`, the local SSH agent will be used to authenticate with the
  bastion host. Defaults to `false`.

- `ssh_bastion_username` (string) - The username to connect to the bastion host.

- `ssh_bastion_password` (string) - The password to use to authenticate with the bastion host.

- `ssh_bastion_interactive` (bool) - If `true`, the keyboard-interactive used to authenticate with bastion host.

- `ssh_bastion_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with the
  bastion host. The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_bastion_certificate_file` (string) - Path to user certificate used to authenticate with bastion host.
  The `~` can be used in path and will be expanded to the
  home directory of current user.

- `ssh_file_transfer_method` (string) - `scp` or `sftp` - How to transfer files, Secure copy (default) or SSH
  File Transfer Protocol.
  
  **NOTE**: Guests using Windows with Win32-OpenSSH v9.1.0.0p1-Beta, scp
  (the default protocol for copying data) returns a a non-zero error code since the MOTW
  cannot be set, which cause any file transfer to fail. As a workaround you can override the transfer protocol
  with SFTP instead `ssh_file_transfer_method = "sftp"`.

- `ssh_proxy_host` (string) - A SOCKS proxy host to use for SSH connection

- `ssh_proxy_port` (int) - A port of the SOCKS proxy. Defaults to `1080`.

- `ssh_proxy_username` (string) - The optional username to authenticate with the proxy server.

- `ssh_proxy_password` (string) - The optional password to use to authenticate with the proxy server.

- `ssh_keep_alive_interval` (duration string | ex: "1h5m2s") - How often to send "keep alive" messages to the server. Set to a negative
  value (`-1s`) to disable. Example value: `10s`. Defaults to `5s`.

- `ssh_read_write_timeout` (duration string | ex: "1h5m2s") - The amount of time to wait for a remote command to end. This might be
  useful if, for example, packer hangs on a connection after a reboot.
  Example: `5m`. Disabled by default.

- `ssh_remote_tunnels` ([]string) - 

- `ssh_local_tunnels` ([]string) - 

<!-- End of code generated from the comments of the SSH struct in communicator/config.go; -->


<!-- Code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `temporary_key_pair_type` (string) - `dsa` | `ecdsa` | `ed25519` | `rsa` ( the default )
  
  Specifies the type of key to create. The possible values are 'dsa',
  'ecdsa', 'ed25519', or 'rsa'.
  
  NOTE: DSA is deprecated and no longer recognized as secure, please
  consider other alternatives like RSA or ED25519.

- `temporary_key_pair_bits` (int) - Specifies the number of bits in the key to create. For RSA keys, the
  minimum size is 1024 bits and the default is 4096 bits. Generally, 3072
  bits is considered sufficient. DSA keys must be exactly 1024 bits as
  specified by FIPS 186-2. For ECDSA keys, bits determines the key length
  by selecting from one of three elliptic curve sizes: 256, 384 or 521
  bits. Attempting to use bit lengths other than these three values for
  ECDSA keys will fail. Ed25519 keys have a fixed length and bits will be
  ignored.
  
  NOTE: DSA is deprecated and no longer recognized as secure as specified
  by FIPS 186-5, please consider other alternatives like RSA or ED25519.

<!-- End of code generated from the comments of the SSHTemporaryKeyPair struct in communicator/config.go; -->


- `ssh_keypair_name` (string) - If specified, this is the key that will be used for SSH with the
  machine. The key must match a key pair name loaded up into the remote.
  By default, this is blank, and Packer will generate a temporary keypair
  unless [`ssh_password`](#ssh_password) is used.
  [`ssh_private_key_file`](#ssh_private_key_file) or
  [`ssh_agent_auth`](#ssh_agent_auth) must be specified when
  [`ssh_keypair_name`](#ssh_keypair_name) is utilized.


- `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
  The `~` can be used in path and will be expanded to the home directory
  of current user.


- `ssh_agent_auth` (bool) - If true, the local SSH agent will be used to authenticate connections to
  the source instance. No temporary keypair will be created, and the
  values of [`ssh_password`](#ssh_password) and
  [`ssh_private_key_file`](#ssh_private_key_file) will be ignored. The
  environment variable `SSH_AUTH_SOCK` must be set for this option to work
  properly.


-> **NOTE:** The builder uses vApp Options to inject SSH public keys to the virtual machine. The `temporary_key_pair_name`
will only work if the template being cloned contains the vApp property `public-keys`. If using `ssh_private_key_file`,
provide the public key using the `configuration_parameters` or [vApp Options Configuration](/packer/integrations/hashicorp/vsphere/latest/components/builder/vsphere-clone#vapp-options-configuration) whenever the `guestinto.userdata` is available.
Refer to the [VMware](https://docs.cloud-init.io/en/latest/reference/data-source/vmware.html) datasource in cloud-init 21.3 and later for additional information.

#### Windows Remote Management (WinRM)

**Optional:**

<!-- Code generated from the comments of the WinRM struct in communicator/config.go; DO NOT EDIT MANUALLY -->

- `winrm_username` (string) - The username to use to connect to WinRM.

- `winrm_password` (string) - The password to use to connect to WinRM.

- `winrm_host` (string) - The address for WinRM to connect to.
  
  NOTE: If using an Amazon EBS builder, you can specify the interface
  WinRM connects to via
  [`ssh_interface`](/packer/integrations/hashicorp/amazon/latest/components/builder/ebs#ssh_interface)

- `winrm_no_proxy` (bool) - Setting this to `true` adds the remote
  `host:port` to the `NO_PROXY` environment variable. This has the effect of
  bypassing any configured proxies when connecting to the remote host.
  Default to `false`.

- `winrm_port` (int) - The WinRM port to connect to. This defaults to `5985` for plain
  unencrypted connection and `5986` for SSL when `winrm_use_ssl` is set to
  true.

- `winrm_timeout` (duration string | ex: "1h5m2s") - The amount of time to wait for WinRM to become available. This defaults
  to `30m` since setting up a Windows machine generally takes a long time.

- `winrm_use_ssl` (bool) - If `true`, use HTTPS for WinRM.

- `winrm_insecure` (bool) - If `true`, do not check server certificate chain and host name.

- `winrm_use_ntlm` (bool) - If `true`, NTLMv2 authentication (with session security) will be used
  for WinRM, rather than default (basic authentication), removing the
  requirement for basic authentication to be enabled within the target
  guest. Further reading for remote connection authentication can be found
  [here](https://msdn.microsoft.com/en-us/library/aa384295(v=vs.85).aspx).

<!-- End of code generated from the comments of the WinRM struct in communicator/config.go; -->


### Export Configuration

<!-- Code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; DO NOT EDIT MANUALLY -->

You can export an image in Open Virtualization Format (OVF) to the Packer
host.

HCL Example:

```hcl

	# ...
	vm_name = "example-ubuntu"
	# ...
	export {
	  force = true
	  output_directory = "./output-artifacts"
	}

```

JSON Example:

```json
...

	"vm_name": "example-ubuntu",

...

	"export": {
	  "force": true,
	  "output_directory": "./output-artifacts"
	},

```

The above configuration would create the following files:

```text
./output-artifacts/example-ubuntu-disk-0.vmdk
./output-artifacts/example-ubuntu.mf
./output-artifacts/example-ubuntu.ovf
```

<!-- End of code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; -->


**Optional:**

<!-- Code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; DO NOT EDIT MANUALLY -->

- `name` (string) - The name of the exported image in Open Virtualization Format (OVF).
  
  -> **Note:** The name of the virtual machine with the `.ovf` extension is
  used if this option is not specified.

- `force` (bool) - Forces the export to overwrite existing files. Defaults to `false`.
  If set to `false`, an error is returned if the file(s) already exists.

- `image_files` (bool) - Include additional image files that are  associated with the virtual
  machine. Defaults to `false`. For example, `.nvram` and `.log` files.

- `manifest` (string) - The hash algorithm to use when generating a manifest file. Defaults to
  `sha256`.
  
  The available options for this setting are: 'none', 'sha1', 'sha256', and
  'sha512'.
  
  --> **Tip:** Use `none` to disable the creation of a manifest file.

- `options` ([]string) - Advanced image export options. Available options include:
  * `mac` - MAC address is exported for each Ethernet device.
  * `uuid` - UUID is exported for the virtual machine.
  * `extraconfig` - Extra configuration options are exported for the
    virtual machine.
  * `nodevicesubtypes` - Resource subtypes for CD/DVD drives, floppy
    drives, and SCSI controllers are not exported.
  
  For example, adding the following export configuration option outputs the
  MAC addresses for each Ethernet device in the OVF descriptor:
  
  HCL Example:
  
  ```hcl
  ...
    export {
      options = ["mac"]
    }
  ```
  
  JSON: Example:
  
  ```json
  ...
    "export": {
      "options": ["mac"]
    },
  ```

- `output_format` (string) - The output format for the exported virtual machine image.
  Defaults to `ovf`. Available options include `ovf` and `ova`.
  
  When set to `ova`, the image is first exported using Open Virtualization
  Format (`.ovf`) and then converted to an Open Virtualization Archive
  (`.ova`) using the VMware [Open Virtualization Format Tool](https://developer.broadcom.com/tools/open-virtualization-format-ovf-tool/latest)
  (ovftool). The intermediate files are removed after the conversion.
  
  ~> **Note:** To use the `ova` format option, VMware ovftool must be
  installed on the Packer host and accessible in either the system `PATH`
  or the user's `PATH`.

<!-- End of code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; -->


### Output Configuration

**Optional:**

<!-- Code generated from the comments of the OutputConfig struct in builder/vsphere/common/output_config.go; DO NOT EDIT MANUALLY -->

- `output_directory` (string) - The directory where artifacts from the build, such as the virtual machine
  files and disks, will be output to. The path to the directory may be
  relative or absolute. If relative, the path is relative to the working
  directory Packer is run from. This directory must not exist or, if
  created, must be empty prior to running the builder. By default, this is
  "output-<buildName>" where "buildName" is the name of the build.

- `directory_permission` (os.FileMode) - The permissions to apply to the "output_directory", and to any parent
  directories that get created for output_directory.  By default, this is
  "0750". You should express the permission as quoted string with a
  leading zero such as "0755" in JSON file, because JSON does not support
  octal value. In Unix-like OS, the actual permission may differ from
  this value because of umask.

<!-- End of code generated from the comments of the OutputConfig struct in builder/vsphere/common/output_config.go; -->


### Content Library Configuration

<!-- Code generated from the comments of the ContentLibraryDestinationConfig struct in builder/vsphere/common/step_import_to_content_library.go; DO NOT EDIT MANUALLY -->

Create a content library item in a content library whose content is a VM
template or an OVF template created from the virtual machine image after
the build is complete.

The template is stored in an existing or newly created library item.

<!-- End of code generated from the comments of the ContentLibraryDestinationConfig struct in builder/vsphere/common/step_import_to_content_library.go; -->


**Optional:**

<!-- Code generated from the comments of the ContentLibraryDestinationConfig struct in builder/vsphere/common/step_import_to_content_library.go; DO NOT EDIT MANUALLY -->

- `library` (string) - The name of the content library in which the new content library item
  containing the template will be created or updated. The content library
  must be of type Local to allow deploying virtual machines.

- `name` (string) - The name of the content library item that will be created or updated.
  For VM templates, the name of the item should be different from
  [vm_name](#vm_name) and the default is [vm_name](#vm_name) + timestamp
  when not set. VM templates will always be imported to a new library item.
  For OVF templates, the name defaults to [vm_name](#vm_name) when not set,
  and if an item with the same name already exists it will be then updated
  with the new OVF template, otherwise a new item will be created.
  
  ~> **Note:** It's not possible to update existing content library items
  with a new VM template. If updating an existing content library item is
  necessary, use an OVF template instead by setting the [ovf](#ovf) option
  as `true`.

- `description` (string) - A description for the content library item that will be created.
  Defaults to "Packer imported [vm_name](#vm_name) VM template".

- `cluster` (string) - The cluster where the VM template will be placed.
  If `cluster` and `resource_pool` are both specified, `resource_pool` must
  belong to cluster. If `cluster` and `host` are both specified, the ESXi
  host must be a member of the cluster. This option is not used when
  importing OVF templates. Defaults to [`cluster`](#cluster).

- `folder` (string) - The virtual machine folder where the VM template will be placed.
  This option is not used when importing OVF templates. Defaults to
  the same folder as the source virtual machine.

- `host` (string) - The ESXi host where the virtual machine template will be placed.
  If `host` and `resource_pool` are both specified, `resource_pool` must
  belong to host. If `host` and `cluster` are both specified, `host` must
  be a member of the cluster. This option is not used when importing OVF
  templates. Defaults to [`host`](#host).

- `resource_pool` (string) - The resource pool where the virtual machine template will be placed.
  Defaults to [`resource_pool`](#resource_pool). If [`resource_pool`](#resource_pool)
  is unset, the system will attempt to choose a suitable resource pool
  for the VM template.

- `datastore` (string) - The datastore for the virtual machine template's configuration and log
  files. This option is not used when importing OVF templates.
  Defaults to the storage backing associated with the content library.

- `destroy` (bool) - Destroy the virtual machine after the import to the content library.
  Defaults to `false`.

- `ovf` (bool) - Import an OVF template to the content library item. Defaults to `false`.

- `skip_import` (bool) - Skip the import to the content library item. Useful during a build test
  stage. Defaults to `false`.

- `ovf_flags` ([]string) - Flags to use for OVF package creation. The supported flags can be
  obtained using ExportFlag.list. If unset, no flags will be used.
  Known values: `EXTRA_CONFIG`, `PRESERVE_MAC`.

<!-- End of code generated from the comments of the ContentLibraryDestinationConfig struct in builder/vsphere/common/step_import_to_content_library.go; -->


HCL Example:

```hcl
	content_library_destination {
		library = "Example Content Library"
	}
```

JSON Example:

```json
	"content_library_destination" : {
	    "library": "Example Content Library"
	}
```

## Working with Clusters and Hosts

### Standalone ESXi Hosts

Only use the `host` option. Optionally, specify a `resource_pool`:

HCL Example:

```hcl
  host = "esxi-01.example.com"
  resource_pool = "example_resource_pool"
```

JSON Example:

```json
  "host": "esxi-01.example.com",
  "resource_pool": "example_resource_pool",
```

### Clusters with Distributed Resource Scheduler Enabled

Only use the `cluster` option. Optionally, specify a `resource_pool`:

HCL Example:

```hcl
  cluster = "cluster-01"
  resource_pool = "example_resource_pool"
```

JSON Example:

```json
  "cluster": "cluster-01",
  "resource_pool": "example_resource_pool",
```

### Clusters without Distributed Resource Scheduler Enabled

Use the `cluster` and `host` parameters:

HCL Example:

```hcl
  cluster = "cluster-01"
  host = "esxi-01.example.com"
```

JSON Example:

```json
  "cluster": "cluster-01",
  "host": "esxi-01.example.com",
```

## Privileges

- VM folder (this object and children):

  ```text
  Virtual machine > Inventory
  Virtual machine > Configuration
  Virtual machine > Interaction
  Virtual machine > Snapshot management
  Virtual machine > Provisioning
  ```

- Resource pool, host, or cluster (this object):

  ```text
  Resource -> Assign virtual machine to resource pool
  ```

- Host in clusters without DRS (this object):

  ```text
  Read-only
  ```

- Datastore (this object):

  ```text
  Datastore > Allocate space
  Datastore > Browse datastore
  Datastore > Low level file operations
  ```

- Network (this object):

  ```text
  Network > Assign network
  ```

- Distributed switch (this object):

  ```text
  Read-only
  ```

- Datacenter (this object):

  ```text
  Datastore > Low level file operations
  ```

- Host (this object):

  ```text
  Host > Configuration > System Management
  ```

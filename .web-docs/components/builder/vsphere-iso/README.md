Type: `vsphere-iso`
Artifact BuilderId: `jetbrains.vsphere`

This builder uses the vSphere API, and creates virtual machines remotely. It
starts from an ISO file and creates new VMs from scratch.

- VMware Player is not required.
- It uses the official vCenter Server API, and does not require ESXi host [modification](/packer/integrations/hashicorp/vsphere/latest/components/builder/vsphere-iso#building-on-a-remote-vsphere-hypervisor)
- The builder supports versions following the VMware Product Lifecycle Matrix
  from General Availability to End of General Support. Builds on versions that
  are end of support may work, but configuration options may throw errors if
  they do not exist in the vSphere API for those versions.

## Examples

See complete Ubuntu, Windows, and macOS templates in the [examples folder](https://github.com/hashicorp/packer-plugin-vsphere/tree/main/builder/vsphere/examples/).

# Configuration Reference

There are many configuration options available for this builder. In addition to
the items listed here, you will want to look at the general configuration
references for [HTTP](#http-directory-configuration),
[Floppy](#floppy-configuration),
[Boot](#boot-configuration),
[Hardware](#hardware-configuration),
[Output](#output-configuration),
[Run](#run-configuration),
[Shutdown](#shutdown-configuration),
[Communicator](#communicator-configuration),
[Export](#export-configuration),
configuration references, which are
necessary for this build to succeed and can be found further down the page.

<!-- Code generated from the comments of the Config struct in builder/vsphere/iso/config.go; DO NOT EDIT MANUALLY -->

- `create_snapshot` (bool) - Create a snapshot when set to `true`, so the VM can be used as a base
  for linked clones. Defaults to `false`.

- `snapshot_name` (string) - When `create_snapshot` is `true`, `snapshot_name` determines the name of the snapshot.
  Defaults to `Created By Packer`.

- `convert_to_template` (bool) - Convert VM to a template. Defaults to `false`.

- `export` (\*common.ExportConfig) - Configuration for exporting VM to an ovf file.
  The VM will not be exported if no [Export Configuration](#export-configuration) is specified.

- `content_library_destination` (\*common.ContentLibraryDestinationConfig) - Configuration for importing the VM template to a Content Library.
  The VM template will not be imported if no [Content Library Import Configuration](#content-library-import-configuration) is specified.
  The import doesn't work if [convert_to_template](#convert_to_template) is set to true.

<!-- End of code generated from the comments of the Config struct in builder/vsphere/iso/config.go; -->


## Boot Configuration

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


We send each character to the VM with a default delay of 100ms between groups.
The delay alleviates possible issues with latency and CPU
contention. If you notice missing keys, you can tune this delay by specifying
"boot_keygroup_interval" in your Packer template, for example:

**JSON**

```json
{
  "builders": [
    {
      "type": "vsphere-iso",
      "boot_keygroup_interval": "500ms"
      ...
    }
  ]
}
```

**HCL2**

```hcl
source "vsphere-iso" "example" {
    boot_keygroup_interval = "500ms"
    # ...
}
```


## Optional:

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


<!-- Code generated from the comments of the BootConfig struct in builder/vsphere/common/step_boot_command.go; DO NOT EDIT MANUALLY -->

- `http_ip` (string) - The IP address to use for the HTTP server started to serve the `http_directory`.
  If unset, Packer will automatically discover and assign an IP.

<!-- End of code generated from the comments of the BootConfig struct in builder/vsphere/common/step_boot_command.go; -->


## Http directory configuration

<!-- Code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; DO NOT EDIT MANUALLY -->

Packer will create an http server serving `http_directory` when it is set, a
random free port will be selected and the architecture of the directory
referenced will be available in your builder.

Example usage from a builder:

```
wget http://{{ .HTTPIP }}:{{ .HTTPPort }}/foo/bar/preseed.cfg
```

<!-- End of code generated from the comments of the HTTPConfig struct in multistep/commonsteps/http_config.go; -->


## Optional:

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


## Floppy configuration

<!-- Code generated from the comments of the FloppyConfig struct in builder/vsphere/common/step_add_floppy.go; DO NOT EDIT MANUALLY -->

- `floppy_img_path` (string) - Datastore path to a floppy image that will be mounted to the VM.
  Example: `[datastore1] ISO/pvscsi-Windows8.flp`.

- `floppy_files` ([]string) - List of local files to be mounted to the VM floppy drive. Can be used to
  make Debian preseed or RHEL kickstart files available to the VM.

- `floppy_dirs` ([]string) - List of directories to copy files from.

- `floppy_content` (map[string]string) - Key/Values to add to the floppy disk. The keys represent the paths, and
  the values contents. It can be used alongside `floppy_files` or
  `floppy_dirs`, which is useful to add large files without loading them
  into memory. If any paths are specified by both, the contents in
  `floppy_content` will take precedence.
  
  Usage example (HCL):
  
  ```hcl
  floppy_content = {
    "meta-data" = jsonencode(local.instance_data)
    "user-data" = templatefile("user-data", { packages = ["nginx"] })
  }
  ```

- `floppy_label` (string) - The label to use for the floppy disk that
  is attached when the VM is booted. This is most useful for cloud-init,
  Kickstart or other early initialization tools, which can benefit from labelled floppy disks.
  By default, the floppy label will be 'packer'.

<!-- End of code generated from the comments of the FloppyConfig struct in builder/vsphere/common/step_add_floppy.go; -->


## Connection Configuration

<!-- Code generated from the comments of the ConnectConfig struct in builder/vsphere/common/step_connect.go; DO NOT EDIT MANUALLY -->

- `vcenter_server` (string) - vCenter Server hostname.

- `username` (string) - vSphere username.

- `password` (string) - vSphere password.

- `insecure_connection` (bool) - Do not validate the vCenter Server TLS certificate. Defaults to `false`.

- `datacenter` (string) - vSphere datacenter name. Required if there is more than one datacenter in the vSphere inventory.

<!-- End of code generated from the comments of the ConnectConfig struct in builder/vsphere/common/step_connect.go; -->


## Hardware Configuration

<!-- Code generated from the comments of the HardwareConfig struct in builder/vsphere/common/step_hardware.go; DO NOT EDIT MANUALLY -->

- `CPUs` (int32) - Number of CPU cores.

- `cpu_cores` (int32) - Number of CPU cores per socket.

- `CPU_reservation` (int64) - Amount of reserved CPU resources in MHz.

- `CPU_limit` (int64) - Upper limit of available CPU resources in MHz.

- `CPU_hot_plug` (bool) - Enable CPU hot plug setting for virtual machine. Defaults to `false`.

- `RAM` (int64) - Amount of RAM in MB.

- `RAM_reservation` (int64) - Amount of reserved RAM in MB.

- `RAM_reserve_all` (bool) - Reserve all available RAM. Defaults to `false`. Cannot be used together
  with `RAM_reservation`.

- `RAM_hot_plug` (bool) - Enable RAM hot plug setting for virtual machine. Defaults to `false`.

- `video_ram` (int64) - Amount of video memory in KB. See [vSphere documentation](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vm_admin.doc/GUID-789C3913-1053-4850-A0F0-E29C3D32B6DA.html)
  for supported maximums. Defaults to 4096 KB.

- `displays` (int32) - Number of video displays. See [vSphere documentation](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vm_admin.doc/GUID-789C3913-1053-4850-A0F0-E29C3D32B6DA.html)
  for supported maximums. Defaults to 1.

- `vgpu_profile` (string) - vGPU profile for accelerated graphics. See [NVIDIA GRID vGPU documentation](https://docs.nvidia.com/grid/latest/grid-vgpu-user-guide/index.html#configure-vmware-vsphere-vm-with-vgpu)
  for examples of profile names. Defaults to none.

- `NestedHV` (bool) - Enable nested hardware virtualization for VM. Defaults to `false`.

- `firmware` (string) - Set the Firmware for virtual machine. Supported values: `bios`, `efi` or `efi-secure`. Defaults to `bios`.

- `force_bios_setup` (bool) - During the boot, force entry into the BIOS setup screen. Defaults to `false`.

- `vTPM` (bool) - Add virtual TPM device for virtual machine. Defaults to `false`.

<!-- End of code generated from the comments of the HardwareConfig struct in builder/vsphere/common/step_hardware.go; -->


## Location Configuration

<!-- Code generated from the comments of the LocationConfig struct in builder/vsphere/common/config_location.go; DO NOT EDIT MANUALLY -->

- `vm_name` (string) - Name of the new VM to create.

- `folder` (string) - VM folder to create the VM in.

- `cluster` (string) - vSphere cluster where target VM is created. See the
  [Working With Clusters And Hosts](#working-with-clusters-and-hosts)
  section above for more details.

- `host` (string) - ESXi host where target VM is created. A full path must be specified if
  the host is in a folder. For example `folder/host`. See the
  [Working With Clusters And Hosts](#working-with-clusters-and-hosts)
  section above for more details.

- `resource_pool` (string) - vSphere resource pool. If not set, it will look for the root resource
  pool of the `host` or `cluster`. If a root resource is not found, it
  will then look for a default resource pool.

- `datastore` (string) - vSphere datastore. Required if `host` is a cluster, or if `host` has
  multiple datastores.

- `set_host_for_datastore_uploads` (bool) - Set this to true if packer should use the host for uploading files
  to the datastore. Defaults to false.

<!-- End of code generated from the comments of the LocationConfig struct in builder/vsphere/common/config_location.go; -->


## Run Configuration

<!-- Code generated from the comments of the RunConfig struct in builder/vsphere/common/step_run.go; DO NOT EDIT MANUALLY -->

- `boot_order` (string) - Priority of boot devices. Defaults to `disk,cdrom`

<!-- End of code generated from the comments of the RunConfig struct in builder/vsphere/common/step_run.go; -->


## Shutdown Configuration

<!-- Code generated from the comments of the ShutdownConfig struct in builder/vsphere/common/step_shutdown.go; DO NOT EDIT MANUALLY -->

- `shutdown_command` (string) - Specify a VM guest shutdown command. This command will be executed using
  the `communicator`. Otherwise, the VMware Tools are used to gracefully shutdown
  the VM.

- `shutdown_timeout` (duration string | ex: "1h5m2s") - Amount of time to wait for graceful VM shutdown.
  Defaults to 5m or five minutes.
  This will likely need to be modified if the `communicator` is 'none'.

- `disable_shutdown` (bool) - Packer normally halts the virtual machine after all provisioners have
  run when no `shutdown_command` is defined. If this is set to `true`, Packer
  *will not* halt the virtual machine but will assume that you will send the stop
  signal yourself through a preseed.cfg, a script or the final provisioner.
  Packer will wait for a default of five minutes until the virtual machine is shutdown.
  The timeout can be changed using `shutdown_timeout` option.

<!-- End of code generated from the comments of the ShutdownConfig struct in builder/vsphere/common/step_shutdown.go; -->


## Wait Configuration

<!-- Code generated from the comments of the WaitIpConfig struct in builder/vsphere/common/step_wait_for_ip.go; DO NOT EDIT MANUALLY -->

- `ip_wait_timeout` (duration string | ex: "1h5m2s") - Amount of time to wait for VM's IP, similar to 'ssh_timeout'.
  Defaults to 30m (30 minutes). See the Golang
  [ParseDuration](https://golang.org/pkg/time/#ParseDuration) documentation
  for full details.

- `ip_settle_timeout` (duration string | ex: "1h5m2s") - Amount of time to wait for VM's IP to settle down, sometimes VM may
  report incorrect IP initially, then its recommended to set that
  parameter to apx. 2 minutes. Examples 45s and 10m. Defaults to
  5s(5 seconds). See the Golang
  [ParseDuration](https://golang.org/pkg/time/#ParseDuration) documentation
   for full details.

- `ip_wait_address` (\*string) - Set this to a CIDR address to cause the service to wait for an address that is contained in
  this network range. Defaults to "0.0.0.0/0" for any ipv4 address. Examples include:
  
  * empty string ("") - remove all filters
  * `0:0:0:0:0:0:0:0/0` - allow only ipv6 addresses
  * `192.168.1.0/24` - only allow ipv4 addresses from 192.168.1.1 to 192.168.1.254

<!-- End of code generated from the comments of the WaitIpConfig struct in builder/vsphere/common/step_wait_for_ip.go; -->


## ISO Configuration

<!-- Code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; DO NOT EDIT MANUALLY -->

By default, Packer will symlink, download or copy image files to the Packer
cache into a "`hash($iso_url+$iso_checksum).$iso_target_extension`" file.
Packer uses [hashicorp/go-getter](https://github.com/hashicorp/go-getter) in
file mode in order to perform a download.

go-getter supports the following protocols:

* Local files
* Git
* Mercurial
* HTTP
* Amazon S3

Examples:
go-getter can guess the checksum type based on `iso_checksum` length, and it is
also possible to specify the checksum type.

In JSON:

```json

	"iso_checksum": "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```json

	"iso_checksum": "file:ubuntu.org/..../ubuntu-14.04.1-server-amd64.iso.sum",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```json

	"iso_checksum": "file://./shasums.txt",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```json

	"iso_checksum": "file:./shasums.txt",
	"iso_url": "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

In HCL2:

```hcl

	iso_checksum = "946a6077af6f5f95a51f82fdc44051c7aa19f9cfc5f737954845a6050543d7c2"
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```hcl

	iso_checksum = "file:ubuntu.org/..../ubuntu-14.04.1-server-amd64.iso.sum"
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```hcl

	iso_checksum = "file://./shasums.txt"
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

```hcl

	iso_checksum = "file:./shasums.txt",
	iso_url = "ubuntu.org/.../ubuntu-14.04.1-server-amd64.iso"

```

<!-- End of code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; -->


## Required:

<!-- Code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; DO NOT EDIT MANUALLY -->

- `iso_checksum` (string) - The checksum for the ISO file or virtual hard drive file. The type of
  the checksum is specified within the checksum field as a prefix, ex:
  "md5:{$checksum}". The type of the checksum can also be omitted and
  Packer will try to infer it based on string length. Valid values are
  "none", "{$checksum}", "md5:{$checksum}", "sha1:{$checksum}",
  "sha256:{$checksum}", "sha512:{$checksum}" or "file:{$path}". Here is a
  list of valid checksum values:
   * md5:090992ba9fd140077b0661cb75f7ce13
   * 090992ba9fd140077b0661cb75f7ce13
   * sha1:ebfb681885ddf1234c18094a45bbeafd91467911
   * ebfb681885ddf1234c18094a45bbeafd91467911
   * sha256:ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
   * ed363350696a726b7932db864dda019bd2017365c9e299627830f06954643f93
   * file:http://releases.ubuntu.com/20.04/SHA256SUMS
   * file:file://./local/path/file.sum
   * file:./local/path/file.sum
   * none
  Although the checksum will not be verified when it is set to "none",
  this is not recommended since these files can be very large and
  corruption does happen from time to time.

- `iso_url` (string) - A URL to the ISO containing the installation image or virtual hard drive
  (VHD or VHDX) file to clone.

<!-- End of code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; -->


## Optional:

<!-- Code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; DO NOT EDIT MANUALLY -->

- `iso_urls` ([]string) - Multiple URLs for the ISO to download. Packer will try these in order.
  If anything goes wrong attempting to download or while downloading a
  single URL, it will move on to the next. All URLs must point to the same
  file (same checksum). By default this is empty and `iso_url` is used.
  Only one of `iso_url` or `iso_urls` can be specified.

- `iso_target_path` (string) - The path where the iso should be saved after download. By default will
  go in the packer cache, with a hash of the original filename and
  checksum as its name.

- `iso_target_extension` (string) - The extension of the iso file after download. This defaults to `iso`.

<!-- End of code generated from the comments of the ISOConfig struct in multistep/commonsteps/iso_config.go; -->


## CDRom Configuration

Each iso defined in the CDRom Configuration adds a new drive. If the "iso_url" is defined in
addition to the "iso_paths", the "iso_url" is added to the VM first. This keeps the "iso_url" first in
the boot order by default allowing the boot iso being defined by the iso_url and the vmware tools iso added
from the datastore. Example:

**JSON**

```json
"iso_urls": [
  "win10.iso",
  "http://example.org/isos/win10.iso"
],
"iso_paths": [
    "[] /usr/lib/vmware/isoimages/windows.iso"
],
```

**HCL2**

```hcl
iso_urls = [
  "win10.iso",
  "http://example.org/isos/win10.iso"
]

iso_paths = [
    "[] /usr/lib/vmware/isoimages/windows.iso"
]
```


<!-- Code generated from the comments of the CDRomConfig struct in builder/vsphere/common/step_add_cdrom.go; DO NOT EDIT MANUALLY -->

- `cdrom_type` (string) - Which controller to use. Example: `sata`. Defaults to `ide`.

- `iso_paths` ([]string) - List of Datastore or Content Library paths to ISO files that will be mounted to the VM.
  Here's an HCL2 example:
  ```hcl
  iso_paths = [
    "[datastore1] ISO/ubuntu.iso",
    "Packer Library Test/ubuntu-16.04.6-server-amd64/ubuntu-16.04.6-server-amd64.iso"
  ]
  ```

<!-- End of code generated from the comments of the CDRomConfig struct in builder/vsphere/common/step_add_cdrom.go; -->


<!-- Code generated from the comments of the RemoveCDRomConfig struct in builder/vsphere/common/step_remove_cdrom.go; DO NOT EDIT MANUALLY -->

- `remove_cdrom` (bool) - Remove CD-ROM devices from template. Defaults to `false`.

<!-- End of code generated from the comments of the RemoveCDRomConfig struct in builder/vsphere/common/step_remove_cdrom.go; -->


<!-- Code generated from the comments of the CDConfig struct in multistep/commonsteps/extra_iso_config.go; DO NOT EDIT MANUALLY -->

An iso (CD) containing custom files can be made available for your build.

By default, no extra CD will be attached. All files listed in this setting
get placed into the root directory of the CD and the CD is attached as the
second CD device.

This config exists to work around modern operating systems that have no
way to mount floppy disks, which was our previous go-to for adding files at
boot time.

<!-- End of code generated from the comments of the CDConfig struct in multistep/commonsteps/extra_iso_config.go; -->


## Optional:

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


## Create Configuration

<!-- Code generated from the comments of the CreateConfig struct in builder/vsphere/iso/step_create.go; DO NOT EDIT MANUALLY -->

- `vm_version` (uint) - Set VM hardware version. Defaults to the most current VM hardware
  version supported by the vCenter Server version. See
  [VMware KB article 1003746](https://kb.vmware.com/s/article/1003746) for
  the full list of supported VM hardware versions.

- `guest_os_type` (string) - Set VM OS type. Defaults to `otherGuest`. See [
  here](https://developer.vmware.com/apis/358/vsphere/doc/vim.vm.GuestOsDescriptor.GuestOsIdentifier.html)
  for a full list of possible values.

- `network_adapters` ([]NIC) - Sets the network adapters, if no adapter is specified, every network related task is not applicable.

- `usb_controller` ([]string) - Create USB controllers for the virtual machine. "usb" for a usb 2.0 controller. "xhci" for a usb 3.0 controller. There can only be at most one of each.

- `notes` (string) - VM notes.

- `destroy` (bool) - If set to true, the VM will be destroyed after the builder completes

<!-- End of code generated from the comments of the CreateConfig struct in builder/vsphere/iso/step_create.go; -->


<!-- Code generated from the comments of the StorageConfig struct in builder/vsphere/common/storage_config.go; DO NOT EDIT MANUALLY -->

- `disk_controller_type` ([]string) - Set VM disk controller type. Example `lsilogic`, `lsilogic-sas`, `pvscsi`, `nvme`, or `scsi`. Use a list to define additional controllers.
  Defaults to `lsilogic`. See
  [SCSI, SATA, and NVMe Storage Controller Conditions, Limitations, and Compatibility](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.vm_admin.doc/GUID-5872D173-A076-42FE-8D0B-9DB0EB0E7362.html#GUID-5872D173-A076-42FE-8D0B-9DB0EB0E7362)
  for additional details.

- `storage` ([]DiskConfig) - Configures a collection of one or more disks to be provisioned along with the VM. See the [Storage Configuration](#storage-configuration).

<!-- End of code generated from the comments of the StorageConfig struct in builder/vsphere/common/storage_config.go; -->


## Network Adapter Configuration

<!-- Code generated from the comments of the NIC struct in builder/vsphere/iso/step_create.go; DO NOT EDIT MANUALLY -->

Defines a Network Adapter
If no adapter is defined, network tasks (communicators, most provisioners) won't work, so it's advised to define one.

Example that creates two network adapters:

In JSON:
```json

	"network_adapters": [
	  {
	    "network": "VM Network",
	    "network_card": "vmxnet3"
	  },
	  {
	    "network": "OtherNetwork",
	    "network_card": "vmxnet3"
	  }
	],

```
In HCL2:
```hcl

	network_adapters {
	    network = "VM Network"
	    network_card = "vmxnet3"
	}
	network_adapters {
	    network = "OtherNetwork"
	    network_card = "vmxnet3"
	}

```

<!-- End of code generated from the comments of the NIC struct in builder/vsphere/iso/step_create.go; -->


<!-- Code generated from the comments of the NIC struct in builder/vsphere/iso/step_create.go; DO NOT EDIT MANUALLY -->

- `network_card` (string) - Set VM network card type. Example `vmxnet3`.

<!-- End of code generated from the comments of the NIC struct in builder/vsphere/iso/step_create.go; -->


## Optional

<!-- Code generated from the comments of the NIC struct in builder/vsphere/iso/step_create.go; DO NOT EDIT MANUALLY -->

- `network` (string) - Set the network in which the VM will be connected to. If no network is
  specified, `host` must be specified to allow Packer to look for the
  available network. If the network is inside a network folder in vCenter,
  you need to provide the full path to the network.

- `mac_address` (string) - Set network card MAC address

- `passthrough` (\*bool) - Enable DirectPath I/O passthrough

<!-- End of code generated from the comments of the NIC struct in builder/vsphere/iso/step_create.go; -->


## Storage Configuration

<!-- Code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; DO NOT EDIT MANUALLY -->

Defines the disk storage for a VM.

Example that will create a 15GB and a 20GB disk on the VM. The second disk will be thin provisioned:

In JSON:
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
In HCL2:
```hcl

	storage {
	    disk_size = 15000
	}
	storage {
	    disk_size = 20000
	    disk_thin_provisioned = true
	}

```

Example that creates 2 pvscsi controllers and adds 2 disks to each one:

In JSON:
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

In HCL2:
```hcl

	disk_controller_type = ["pvscsi", "pvscsi"]
	storage {
	   disk_size = 15000,
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

<!-- End of code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; -->


<!-- Code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; DO NOT EDIT MANUALLY -->

- `disk_size` (int64) - The size of the disk in MB.

<!-- End of code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; -->


## Optional

<!-- Code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; DO NOT EDIT MANUALLY -->

- `disk_thin_provisioned` (bool) - Enable VMDK thin provisioning for VM. Defaults to `false`.

- `disk_eagerly_scrub` (bool) - Enable VMDK eager scrubbing for VM. Defaults to `false`.

- `disk_controller_index` (int) - The assigned disk controller. Defaults to the first one (0)

<!-- End of code generated from the comments of the DiskConfig struct in builder/vsphere/common/storage_config.go; -->


## Export Configuration

<!-- Code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; DO NOT EDIT MANUALLY -->

You may optionally export an ovf from vSphere to the instance running Packer.

Example usage:

In JSON:
```json
...

	"vm_name": "example-ubuntu",

...

	"export": {
	  "force": true,
	  "output_directory": "./output_vsphere"
	},

```
In HCL2:
```hcl

	# ...
	vm_name = "example-ubuntu"
	# ...
	export {
	  force = true
	  output_directory = "./output_vsphere"
	}

```
The above configuration would create the following files:

```text
./output_vsphere/example-ubuntu-disk-0.vmdk
./output_vsphere/example-ubuntu.mf
./output_vsphere/example-ubuntu.ovf
```

<!-- End of code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; -->


## Optional:

<!-- Code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; DO NOT EDIT MANUALLY -->

- `name` (string) - Name of the ovf. defaults to the name of the VM

- `force` (bool) - Overwrite ovf if it exists

- `images` (bool) - Deprecated: Images will be removed in a future release. Please see `image_files` for more details on this argument.

- `image_files` (bool) - In exported files, include additional image files that are attached to the VM, such as nvram, iso, img.

- `manifest` (string) - Generate manifest using sha1, sha256, sha512. Defaults to 'sha256'. Use 'none' for no manifest.

- `options` ([]string) - Advanced ovf export options. Options can include:
  * mac - MAC address is exported for all ethernet devices
  * uuid - UUID is exported for all virtual machines
  * extraconfig - all extra configuration options are exported for a virtual machine
  * nodevicesubtypes - resource subtypes for CD/DVD drives, floppy drives, and serial and parallel ports are not exported
  
  For example, adding the following export config option would output the mac addresses for all Ethernet devices in the ovf file:
  
  In JSON:
  ```json
  ...
    "export": {
      "options": ["mac"]
    },
  ```
  In HCL2:
  ```hcl
  ...
    export {
      options = ["mac"]
    }
  ```

<!-- End of code generated from the comments of the ExportConfig struct in builder/vsphere/common/step_export.go; -->


## Output Configuration:

<!-- Code generated from the comments of the OutputConfig struct in builder/vsphere/common/output_config.go; DO NOT EDIT MANUALLY -->

- `output_directory` (string) - This setting specifies the directory that
  artifacts from the build, such as the virtual machine files and disks,
  will be output to. The path to the directory may be relative or
  absolute. If relative, the path is relative to the working directory
  packer is executed from. This directory must not exist or, if
  created, must be empty prior to running the builder. By default this is
  "output-BUILDNAME" where "BUILDNAME" is the name of the build.

- `directory_permission` (os.FileMode) - The permissions to apply to the "output_directory", and to any parent
  directories that get created for output_directory.  By default this is
  "0750". You should express the permission as quoted string with a
  leading zero such as "0755" in JSON file, because JSON does not support
  octal value. In Unix-like OS, the actual permission may differ from
  this value because of umask.

<!-- End of code generated from the comments of the OutputConfig struct in builder/vsphere/common/output_config.go; -->


## Content Library Import Configuration

<!-- Code generated from the comments of the ContentLibraryDestinationConfig struct in builder/vsphere/common/step_import_to_content_library.go; DO NOT EDIT MANUALLY -->

With this configuration Packer creates a library item in a content library whose content is a VM template
or an OVF template created from the just built VM.
The template is stored in a existing or newly created library item.

<!-- End of code generated from the comments of the ContentLibraryDestinationConfig struct in builder/vsphere/common/step_import_to_content_library.go; -->


<!-- Code generated from the comments of the ContentLibraryDestinationConfig struct in builder/vsphere/common/step_import_to_content_library.go; DO NOT EDIT MANUALLY -->

- `library` (string) - Name of the library in which the new library item containing the template should be created/updated.
  The Content Library should be of type Local to allow deploying virtual machines.

- `name` (string) - Name of the library item that will be created or updated.
  For VM templates, the name of the item should be different from [vm_name](#vm_name) and
  the default is [vm_name](#vm_name) + timestamp when not set. VM templates will be always imported to a new library item.
  For OVF templates, the name defaults to [vm_name](#vm_name) when not set, and if an item with the same name already
  exists it will be then updated with the new OVF template, otherwise a new item will be created.
  
  ~> **Note**: It's not possible to update existing library items with a new VM template. If updating an existing library
  item is necessary, use an OVF template instead by setting the [ovf](#ovf) option as `true`.

- `description` (string) - Description of the library item that will be created.
  Defaults to "Packer imported [vm_name](#vm_name) VM template".

- `cluster` (string) - Cluster onto which the virtual machine template should be placed.
  If cluster and resource_pool are both specified, resource_pool must belong to cluster.
  If cluster and host are both specified, host must be a member of cluster.
  This option is not used when importing OVF templates.
  Defaults to [cluster](#cluster).

- `folder` (string) - Virtual machine folder into which the virtual machine template should be placed.
  This option is not used when importing OVF templates.
  Defaults to the same folder as the source virtual machine.

- `host` (string) - Host onto which the virtual machine template should be placed.
  If host and resource_pool are both specified, resource_pool must belong to host.
  If host and cluster are both specified, host must be a member of cluster.
  This option is not used when importing OVF templates.
  Defaults to [host](#host).

- `resource_pool` (string) - Resource pool into which the virtual machine template should be placed.
  Defaults to [resource_pool](#resource_pool). if [resource_pool](#resource_pool) is also unset,
  the system will attempt to choose a suitable resource pool for the virtual machine template.

- `datastore` (string) - The datastore for the virtual machine template's configuration and log files.
  This option is not used when importing OVF templates.
  Defaults to the storage backing associated with the library specified by library.

- `destroy` (bool) - If set to true, the VM will be destroyed after deploying the template to the Content Library.
  Defaults to `false`.

- `ovf` (bool) - When set to true, Packer will import and OVF template to the content library item. Defaults to `false`.

- `skip_import` (bool) - When set to true, the VM won't be imported to the content library item. Useful for setting to `true` during a build test stage. Defaults to `false`.

- `ovf_flags` ([]string) - Flags to use for OVF package creation. The supported flags can be obtained using ExportFlag.list. If unset, no flags will be used. Known values: EXTRA_CONFIG, PRESERVE_MAC

<!-- End of code generated from the comments of the ContentLibraryDestinationConfig struct in builder/vsphere/common/step_import_to_content_library.go; -->


Minimal example of usage to import a VM template:

**JSON**

```json
	"content_library_destination" : {
	    "library": "Packer Library Test"
	}
```

**HCL2**

```hcl
	content_library_destination {
			library = "Packer Library Test"
	}
```


Minimal example of usage to import a OVF template:

**JSON**

```json
	"content_library_destination" : {
	    "library": "Packer Library Test",
	    "ovf": true
	}
```

**HCL2**

```hcl
	content_library_destination {
			library = "Packer Library Test"
			ovf = true
	}
```


## Extra Configuration Parameters

<!-- Code generated from the comments of the ConfigParamsConfig struct in builder/vsphere/common/step_config_params.go; DO NOT EDIT MANUALLY -->

- `configuration_parameters` (map[string]string) - configuration_parameters is a direct passthrough to the vSphere API's
  ConfigSpec: https://vdc-download.vmware.com/vmwb-repository/dcr-public/bf660c0a-f060-46e8-a94d-4b5e6ffc77ad/208bc706-e281-49b6-a0ce-b402ec19ef82/SDK/vsphere-ws/docs/ReferenceGuide/vim.vm.ConfigSpec.html

- `tools_sync_time` (bool) - Enables time synchronization with the host. Defaults to false.

- `tools_upgrade_policy` (bool) - If sets to true, vSphere will automatically check and upgrade VMware Tools upon a system power cycle.
  If not set, defaults to manual upgrade.

<!-- End of code generated from the comments of the ConfigParamsConfig struct in builder/vsphere/common/step_config_params.go; -->


## Communicator configuration

## Optional common fields:

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


## Optional SSH fields:

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
  with SFTP instead `ssh_file_transfer_protocol = "sftp"`.

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


- `ssh_private_key_file` (string) - Path to a PEM encoded private key file to use to authenticate with SSH.
  The `~` can be used in path and will be expanded to the home directory
  of current user.


## Optional WinRM fields:

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


## Working With Clusters And Hosts

### Standalone Hosts

Only use the `host` option. Optionally specify a `resource_pool`:

**JSON**

```json
"host": "esxi-1.vsphere65.test",
"resource_pool": "pool1",
```

**HCL2**

```hcl
host = ""esxi-1.vsphere65.test""
resource_pool = "pool1"
```


### Clusters Without DRS

Use the `cluster` and `host`parameters:

**JSON**

```json
"cluster": "cluster1",
"host": "esxi-2.vsphere65.test",
```

**HCL2**

```hcl
cluster = "cluster1"
host = "esxi-2.vsphere65.test"
```


### Clusters With DRS

Only use the `cluster` option. Optionally specify a `resource_pool`:

**JSON**

```json
"cluster": "cluster2",
"resource_pool": "pool1",
```

**HCL2**

```hcl
cluster = "cluster2"
resource_pool = "pool1"
```


## Required vSphere Privileges

It is recommended to create a custom vSphere role with the required privileges to integrate Packer with vSphere. Accounts or groups can be added to the role to ensure that Packer has **_the least privileged_** access to the infrastructure. For example, a named service account (_e.g._ svc-packer-vsphere@example.com). 

Clone the default **Read-Only** vSphere role and add the following privileges, which are based on the capabilities of the `vsphere-iso` plugin:

Category        | Privilege                                           | Reference
----------------|-----------------------------------------------------|---------
Content Library | Add library item                                    | `ContentLibrary.AddLibraryItem`
 ...            | Update Library Item                                 | `ContentLibrary.UpdateLibraryItem`
Datastore       | Allocate space                                      | `Datastore.AllocateSpace`
...             | Browse datastore                                    | `Datastore.Browse`
...             | Low level file operations                           | `Datastore.FileManagement`
Network         | Assign network                                      | `Network.Assign`
Resource        | Assign virtual machine to resource pool             | `Resource.AssignVMToPool`
vApp            | Export                                              | `vApp.Export`
Virtual Machine | Configuration > Add new disk                        | `VirtualMachine.Config.AddNewDisk`
...             | Configuration > Add or remove device                | `VirtualMachine.Config.AddRemoveDevice`
...             | Configuration > Advanced configuration              | `VirtualMachine.Config.AdvancedConfig`
...             | Configuration > Change CPU count                    | `VirtualMachine.Config.CPUCount`
...             | Configuration > Change memory                       | `VirtualMachine.Config.Memory`
...             | Configuration > Change settings                     | `VirtualMachine.Config.Settings`
...             | Configuration > Change Resource                     | `VirtualMachine.Config.Resource`
...             | Configuration > Set annotation                      | `VirtualMachine.Config.Annotation`
...             | Edit Inventory > Create from existing               | `VirtualMachine.Inventory.CreateFromExisting`
...             | Edit Inventory > Create new                         | `VirtualMachine.Inventory.Create`
...             | Edit Inventory > Remove                             | `VirtualMachine.Inventory.Delete`
...             | Interaction > Configure CD media                    | `VirtualMachine.Interact.SetCDMedia`
...             | Interaction > Configure floppy media                | `VirtualMachine.Interact.SetFloppyMedia`
...             | Interaction > Connect devices                       | `VirtualMachine.Interact.DeviceConnection`
...             | Interaction > Inject USB HID scan codes             | `VirtualMachine.Interact.PutUsbScanCodes`
...             | Interaction > Power off                             | `VirtualMachine.Interact.PowerOff`
...             | Interaction > Power on                              | `VirtualMachine.Interact.PowerOn`
...             | Provisioning > Create template from virtual machine | `VirtualMachine.Provisioning.CreateTemplateFromVM`
...             | Provisioning > Mark as template                     | `VirtualMachine.Provisioning.MarkAsTemplate`
...             | Provisioning > Mark as virtual machine              | `VirtualMachine.Provisioning.MarkAsVM`
...             | State > Create snapshot                             | `VirtualMachine.State.CreateSnapshot`

Global permissions **[are required](https://docs.vmware.com/en/VMware-vSphere/7.0/com.vmware.vsphere.security.doc/GUID-03B36057-B38C-479C-BD78-341CD83A0584.html)** for the content library based on the hierarchical inheritance of permissions. Once the custom vSphere role is created, assign **Global Permissions** in vSphere to the accounts or groups used for the Packer to vSphere integration, if using the content library.

For example:

1. Log in to the vCenter Server at _https://<management_vcenter_server_fqdn>/ui_ as `administrator@vsphere.local`.
2. Select **Menu** > **Administration**.
3. In the left pane, select **Access control** > **Global permissions** and click the **Add permissions** icon.
4. In the **Add permissions** dialog box, enter the service account (_e.g._ svc-packer-vsphere@example.com), select the custom role (_e.g._ Packer to vSphere Integration Role) and the **Propagate to children** check box, and click OK.

In an environment with many vCenter Server instances, such as management and workload, in enhanced linked-mode, you may wish to further reduce the scope of access across the vSphere infrastructure. For example, if you do not want Packer to have access to the management vCenter Server instance, but only allow access to workload vCenter Server instances:

1. From the **Hosts and clusters** inventory, select management vCenter Server to restrict scope, and click the **Permissions** tab.
2. Select the service account with the custom role assigned and click the **Change role** icon.
3. In the **Change role** dialog box, from the **Role** drop-down menu, select **No Access**, select the **Propagate to children** check box, and click **OK**.

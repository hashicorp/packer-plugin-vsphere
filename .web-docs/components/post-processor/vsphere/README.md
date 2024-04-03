Type: `vsphere`
Artifact BuilderId: `packer.post-processor.vsphere`

This post-processor uploads an artifact to a vSphere endpoint.

The artifact must be a VMX, OVA, or OVF file.

## Configuration

The following configuration options are available for the post-processor.

Required:

<!-- Code generated from the comments of the Config struct in post-processor/vsphere/post-processor.go; DO NOT EDIT MANUALLY -->

- `cluster` (string) - Specifies the vSphere cluster or ESXi host to upload the virtual machine.
  This can be either the name of the vSphere cluster or the fully qualified domain name (FQDN)
  or IP address of the ESXi host.

- `datacenter` (string) - Specifies the name of the vSphere datacenter object to place the virtual machine.
  This is _not required_ if `resource_pool` is specified.

- `datastore` (string) - Specifies the name of the vSphere datastore to place the virtual machine.

- `host` (string) - Specifies the fully qualified domain name or IP address of the vCenter Server or ESXi host.

- `password` (string) - Specifies the password to use to authenticate to the vSphere endpoint.

- `username` (string) - Specifies the username to use to authenticate to the vSphere endpoint.

<!-- End of code generated from the comments of the Config struct in post-processor/vsphere/post-processor.go; -->


Optional:

<!-- Code generated from the comments of the Config struct in post-processor/vsphere/post-processor.go; DO NOT EDIT MANUALLY -->

- `disk_mode` (string) - Specifies the disk format of the target virtual machine. One of `thin`, `thick`,

- `esxi_host` (string) - Specifies the fully qualified domain name or IP address of the ESXi host to upload the
  virtual machine. This is _not required_ if `host` is a vCenter Server.

- `insecure` (bool) - Specifies whether to skip the verification of the server certificate. Defaults to `false`.

- `options` ([]string) - Specifies custom options to add in `ovftool`.
  Use `ovftool --help` to list all the options available.

- `overwrite` (bool) - Specifies whether to overwrite the existing files.
  If `true`, forces existing files to to be overwritten. Defaults to `false`.

- `resource_pool` (string) - Specifies the name of the resource pool to place the virtual machine.

- `vm_folder` (string) - Specifies the name of the virtual machine folder path where the virtual machine will be
  placed.

- `vm_name` (string) - Specifies the name of the virtual machine to be created on the vSphere endpoint.

- `vm_network` (string) - Specifies the name of the network in which to place the virtual machine.

- `hardware_version` (string) - Specifies the maximum virtual hardware version for the deployed virtual machine.
  
  It does not upgrade the virtual hardware version of the source VM. Instead, it limits the
  virtual hardware version of the deployed virtual machine  to the specified version.
  If the source virtual machine's hardware version is higher than the specified version, the
  deployed virtual machine's hardware version will be downgraded to the specified version.
  
  If the source virtual machine's hardware version is lower than or equal to the specified
  version, the deployed virtual machine's hardware version will be the same as the source
  virtual machine's.
  
  This option is useful when deploying to vCenter Server instance ot an ESXi host whose
  version is different than the one used to create the artifact.
  
  See [VMware KB 1003746](https://kb.vmware.com/s/article/1003746) for more information on the
  virtual hardware versions supported.

- `max_retries` (int) - Specifies the maximum number of times to retry the upload operation if it fails.
  Defaults to `5`.

<!-- End of code generated from the comments of the Config struct in post-processor/vsphere/post-processor.go; -->


- `keep_input_artifact` (boolean) - Specifies to preserve the local virtual machines files, even
  after importing them to the vSphere endpoint. Defaults to `false`.

# Example Usage

The following is an example of the vSphere post-processor being used in conjunction with the `null`
builder to upload a VMX to a vSphere cluster. You can also use this post-processor with the VMX
artifact from a build.

An example is shown below, showing only the post-processor configuration:

In HCL2:

```hcl
source "null" "example" {
    communicator = "none"
}

build {
    sources = [
        "source.null.example"
    ]

    post-processors {
      post-processor "vsphere"{
          vm_name             = "packer"
          host                = "vcenter.example.com"
          username            = "administrator@vsphere.local"
          password            = "VMw@re1!"
          datacenter          = "dc-01"
          cluster             = "cluster-01"
          datastore           = "datastore-01"
          vm_network          = "VM Network"
          keep_input_artifact = true
      }
    }
}
```

In JSON:

```json
{
  "builders": [
    {
      "type": "null",
      "communicator": "none"
    }
  ],
  "post-processors": [
    [
      {
        "type": "vsphere",
        "vm_name": "packer",
        "host": "vcenter.example.com",
        "username": "administrator@vsphere.local",
        "password": "VMw@re1!",
        "datacenter": "dc-01",
        "cluster": "cluster-01",
        "datastore": "datastore-01",
        "vm_network": "VM Network",
        "keep_input_artifact": true,
      }
    ]
  ]
}
```

# Privileges

The post-processor uses `ovftool` and tneeds several privileges to be able to run `ovftool`.

Rather than giving full administrator access, you can create a role to give the post-processor the
privileges necessary to run.

Below is an example role that will work. Please note that this is a user-supplied list so there may
be a few extraneous privileges that are not strictly required.

For vSphere, the role needs the following privileges:

- `Datastore.AllocateSpace`
- `Host.Config.AdvancedConfig`
- `Host.Config.NetService`
- `Host.Config.Network`
- `Network.Assign`
- `System.Anonymous`
- `System.Read`
- `System.View`
- `VApp.Import`
- `VirtualMachine.Config.AddNewDisk`
- `VirtualMachine.Config.AdvancedConfig`
- `VirtualMachine.Inventory.Delete`

The role must be authorized on the:

- Cluster of the host.
- The destination folder.
- The destination datastore.
- The network to be assigned.

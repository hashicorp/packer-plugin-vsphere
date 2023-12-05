Type: `vsphere`
Artifact BuilderId: `packer.post-processor.vsphere`

The Packer vSphere post-processor takes an artifact and uploads it to a vSphere endpoint.
The artifact must have a vmx/ova/ovf image.

## Configuration

There are many configuration options available for the post-processor. They are
segmented below into two categories: required and optional parameters. Within
each category, the available configuration keys are alphabetized.

Required:

- `cluster` (string) - The vSphere cluster or ESXi host to upload the VM. This can be
  either the name of the vSphere cluster or the FQDN/IP address of an ESXi host.

- `datacenter` (string) - The name of the datacenter within the vSphere environemnt
   to add the VM.

- `datastore` (string) - The name of the datastore to place the VM. This is
  _not required_ if `resource_pool` is specified.

- `host` (string) - The vSphere endpoint that will be contacted to perform
  the VM upload.

- `password` (string) - The password to use to authenticate to the endpoint.

- `username` (string) - The username to use to authenticate to the endpoint.

- `vm_name` (string) - The name of the VM after upload.

Optional:

- `esxi_host` (string) - Target ESXi host. Used to assign specific ESXi
  host to upload the resulting VM, when a vCenter Server is used as
  `host`. Can be either an FQDN (e.g., "esxi-01.example.com", requires proper DNS
  setup and/or correct DNS search domain setting) or an IPv4 address.

- `disk_mode` (string) - Target disk format. See `ovftool` manual for
  available options. Default: `thick`.

- `insecure` (boolean) - Whether or not the connection can be done
  over an insecure connection. Default: `false`

- `keep_input_artifact` (boolean) - When `true`, preserve the local VM files,
  even after importing them to the vSphere endpoint. Default: `false`.

- `resource_pool` (string) - The resource pool in which to upload the VM.

- `vm_folder` (string) - The folder within the datastore to place the VM.

- `vm_network` (string) - The name of the network in which to place the VM.

- `overwrite` (boolean) - If `true`, force the system to overwrite the
  existing files instead create new ones. Default: `false`

- `hardware_version` (string) - This option sets the maximum virtual hardware version
  for the deployed VM. It does not upgrade the virtual hardware version of the source VM.
  Instead, it limits the virtual hardware version of the deployed VM to the specified
  version. If the source VM's hardware version is higher than the specified version,
  the deployed VM's hardware version will be downgraded to the specified version.
  If the source VM's hardware version is lower than or equal to the specified version,
  the deployed VM's hardware version will be the same as the source VM's.
  This option is useful when deploying to a vSphere / ESXi host whose version is different
  than the one used to create the artifact.
  
  See [VMware KB 1003746](https://kb.vmware.com/s/article/1003746) for more information
  on the virtual hardware versions supported for each vSphere / ESXi version.

- `options` (array of strings) - Custom options to add in `ovftool`. See
  `ovftool --help` to list all the options

# Example

The following is an example of the vSphere post-processor being used in
conjunction with the null builder to upload a vmx to a vSphere cluster.

You can also use this post-processor with the vmx artifact from a build.

**HCL2**

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

**JSON**

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

The post-processor uses `ovftool` and therefore needs the same privileges
as `ovftool`. Rather than giving full administrator access, you can create a role
to give the post-processor the permissions necessary to run. Below is an example
role. Please note that this is a user-supplied list so there may be a few
extraneous permissions that are not strictly required.

For vSphere the role needs the following privileges:

    Datastore.AllocateSpace
    Host.Config.AdvancedConfig
    Host.Config.NetService
    Host.Config.Network
    Network.Assign
    System.Anonymous
    System.Read
    System.View
    VApp.Import
    VirtualMachine.Config.AddNewDisk
    VirtualMachine.Config.AdvancedConfig
    VirtualMachine.Inventory.Delete

And this role must be authorized on the:

    Cluster of the host
    The destination folder (not on Datastore, on the vSphere logical view)
    The network to be assigned
    The destination datastore.

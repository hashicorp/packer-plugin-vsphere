Type: `vsphere-template`
Artifact BuilderId: `packer.post-processor.vsphere`

This post-processor uses an artifact from the `vmware-iso` builder with an ESXi host or an artifact
from the [vSphere](/packer/integrations/hashicorp/vsphere/latest/components/post-processor/vsphere) post-processor. It then marks
the virtual machine as a template and moves it to your specified path.

## Configuration

The following configuration options are available for the post-processor.

Required:

<!-- Code generated from the comments of the Config struct in post-processor/vsphere-template/post-processor.go; DO NOT EDIT MANUALLY -->

- `host` (string) - Specifies the fully qualified domain name or IP address of the vSphere endpoint.

- `username` (string) - Specifies the username to use to authenticate to the vSphere endpoint.

- `password` (string) - Specifies the password to use to authenticate to the vSphere endpoint.

<!-- End of code generated from the comments of the Config struct in post-processor/vsphere-template/post-processor.go; -->


Optional:

<!-- Code generated from the comments of the Config struct in post-processor/vsphere-template/post-processor.go; DO NOT EDIT MANUALLY -->

- `insecure` (bool) - Specifies whether to skip the verification of the server certificate. Defaults to `false`.

- `datacenter` (string) - Specifies the name of the datacenter to use.
  Required when the vCenter Server instance endpoint has more than one datacenter.

- `folder` (string) - Specifies the name of the virtual machine folder path where the template will be created.

- `snapshot_enable` (bool) - Specifies whether to create a snapshot before marking as a template. Defaults to `false`.\

- `snapshot_name` (string) - Specifies the name of the snapshot. Required when `snapshot_enable` is `true`.

- `snapshot_description` (string) - Specifies a description for the snapshot. Required when `snapshot_enable` is `true`.

- `reregister_vm` (boolean) - Specifies to keep the virtual machine registered after marking as a template.

<!-- End of code generated from the comments of the Config struct in post-processor/vsphere-template/post-processor.go; -->


- `keep_input_artifact` (boolean) - This option is not applicable to `vsphere-template`. For a
  template to function, the original virtual machine from which it was generated cannot be deleted.
  Therefore, the vSphere Template post-processor always preserves the original virtual machine.

  ~> **Note**: If you are getting permission denied errors when trying to mark as a template, but it
  works in the vSphere UI, set this to `false`. Default is `true`.

## Example Usage

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
      post-processor "vsphere-template"{
          host                = "vcenter.example.com"
          insecure            = false
          username            = "administrator@vsphere.local"
          password            = "VMw@re1!"
          datacenter          = "dc-01"
          folder              = "/templates/os/distro"
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
        "type": "vsphere-template",
        "host": "vcenter.example.com",
        "insecure": true,
        "username": "administrator@vsphere.local",
        "password": "VMw@re1!",
        "datacenter": "dc-01",
        "folder": "/templates/os/distro"
      }
    ]
  ]
}
```

## Using the vSphere Template with Local Builders

Once the [vSphere](/packer/integrations/hashicorp/vsphere/latest/components/post-processor/vsphere) post-processor takes an artifact
from the builder and uploads it to a vSphere endpoint, you may want the virtual machine to be marked
as a template.

In HCL2:

```hcl
build {
    sources = [
        "source.null.example"
    ]

  post-processors {
    post-processor "vsphere" {
      # ...
    }

    post-processor "vsphere-template" {
      # ...
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
         ...
      },
      {
        "type": "vsphere-template",
         ...
      }
    ],
    {
      "type": "...",
      ...
    }
  ]
}
```

In the example above, the result of each builder is passed through the defined sequence of
post-processors starting with the `vsphere` post-processor which will upload the artifact to a
vSphere endpoint. The resulting artifact is then passed on to the `vsphere-template` post-processor
which handles marking a VM as a template. In JSON, note that the `vsphere` and `vsphere-template`
post-processors can be paired together in their own array.

## Permissions

The post processor needs several privileges to be able to mark the virtual as a template.

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

  and either (if `reregister_vm` is `false`):

  - `VirtualMachine.Provisioning.MarkAsTemplate`

  or (if `reregister_vm` is `true` or unset):

  - `VirtualMachine.Inventory.Register`
  - `VirtualMachine.Inventory.Unregister`

The role must be authorized on the:

- Cluster of the host.
- The destination folder.
- The destination datastore.
- The network to be assigned.

# Troubleshooting

Some users have reported that vSphere templates created from local vSphere builds get their boot
order reset to CD-ROM only instead of the original boot order defined by the template. If this issue
affects you, the solution is to set `"bios.hddOrder": "scsi0:0"` in your builder's `vmx_data`.

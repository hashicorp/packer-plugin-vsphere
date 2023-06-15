The vSphere plugin is able to create vSphere virtual machines for use with any VMware product. 
To achieve this, the plugin comes with three builders, and two post-processors
to build the VM depending on the strategy you want to use.

### Installation
To install this plugin add this code into your Packer configuration and run [packer init](/packer/docs/commands/init)

```hcl
packer {
  required_plugins {
    vsphere = {
      version = "~> 1"
      source  = "github.com/hashicorp/vsphere"
    }
  }
}
```

Alternatively, you can use `packer plugins install` to manage installation of this plugin.

```sh
packer plugins install github.com/hashicorp/vsphere
```

### Components
#### Builders:
- [vsphere-iso](/packer/integrations/hashicorp/vsphere/latest/components/builder/vsphere-iso) - This builder starts from an
  ISO file and utilizes the vSphere API to build on a remote esx instance.
  This allows you to build vms even if you do not have SSH access to your vSphere cluster.

- [vsphere-clone](/packer/integrations/hashicorp/vsphere/latest/components/builder/vsphere-clone) - This builder clones a
  vm from an existing template, then modifies it and saves it as a new
  template. It uses the vSphere API to build on a remote esx instance.
  This allows you to build vms even if you do not have SSH access to your vSphere cluster.

- [vsphere-supervisor](/packer/integrations/hashicorp/vsphere/latest/components/builder/vsphere-supervisor) - This builder deploys a
  vm to a vSphere Supervisor cluster, using the VM-Service API. This allows you to build
  vms without spec yaml files and configure them after using the Packer provisioners.

#### Post-Processors
- [vsphere](/packer/integrations/hashicorp/vsphere/latest/components/post-processor/vsphere) - The Packer vSphere post-processor takes an artifact 
  and uploads it to a vSphere endpoint.

- [vsphere-template](/packer/integrations/hashicorp/vsphere/latest/components/post-processor/vsphere-template) - The Packer vSphere Template post-processor takes an 
  artifact from the vmware-iso builder, built on an ESXi host (i.e. remote) or an artifact from the 
  [vSphere](/packer/integrations/hashicorp/vsphere/latest/components/post-processor/vsphere) post-processor, marks the VM as a template, and leaves it in the path of 
  your choice.

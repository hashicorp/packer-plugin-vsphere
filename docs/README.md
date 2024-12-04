<!-- markdownlint-disable first-line-h1 no-inline-html -->

The vSphere plugin is able to create vSphere virtual machines for use with VMware products.

To achieve this, the plugin comes with three builders, and two post-processors to build the virtual
machine depending on the strategy you want to use.

The Packer Plugin for VMware vSphere is a multi-component plugin can be used with HashiCorp Packer
to create virtual machine images for VMware vSphere.

The plugin includes three builders which are able to create images, depending on your desired
strategy:

### Installation

To install this plugin add this code into your Packer configuration and run
[packer init](/packer/docs/commands/init)

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

#### Builders

- [vsphere-iso](/packer/integrations/hashicorp/vsphere/latest/components/builder/vsphere-iso) - This
  builder starts from an ISO file and uses the vSphere API to build a virtual machine image on
  an ESXi host.

- [vsphere-clone](/packer/integrations/hashicorp/vsphere/latest/components/builder/vsphere-clone) -
  This builder clones a virtual machine from an existing template using the uses the vSphere API and
  then modifies and saves it as a new template.

- [vsphere-supervisor](/packer/integrations/hashicorp/vsphere/latest/components/builder/vsphere-supervisor) -
  This builder deploys and publishes new virtual machine to a vSphere Supervisor cluster using VM
  Service.

#### Data Sources

- [vsphere-virtual_machine](/packer/integrations/hashicorp/vsphere/latest/components/data-source/vsphere-virtual_machine) -
  This datasource returns name of existing virtual machine that matches all defined filters to use
  it as a builder source for `vsphere-clone`.

#### Post-Processors

- [vsphere](/packer/integrations/hashicorp/vsphere/latest/components/post-processor/vsphere) -
  This post-processor uploads an artifact to a vSphere endpoint. The artifact must be a VMX, OVA,
  or OVF file.

- [vsphere-template](/packer/integrations/hashicorp/vsphere/latest/components/post-processor/vsphere-template) - This post-processor uses an artifact from the `vmware-iso` builder with an ESXi host or an
    artifact from the [vSphere](/packer/plugins/post-processors/vsphere/vsphere) post-processor. It
    then marks the virtual machine as a template and moves it to your specified path.

### Differences from the Packer Plugin for VMware

While both this plugin and the [`packer-plugin-vmware`](packer/integrations/hashicorp/vmware) are
designed to create virtual machine images, there are some key differences:

- **Platforms**: This plugin is specifically developed to utilize the VMware vSphere API,
  facilitating the creation of virtual machine images by integrating with VMware vCenter Server and
  the VMware vSphere Hypervisor. On the other hand, `packer-plugin-vmware` supports a variety of
  platforms including VMware vSphere Hypervisor and desktop virtualization products such as VMware
  Fusion, VMware Workstation, and VMware Player, though it does not utilize the vSphere API for its
  operations.

- **Focus**: This plugin is purpose-built with a focus on VMware vSphere, offering capabilities such
  as creating virtual machine images, cloning and modifying base virtual machine images, and
  exporting artifacts in specified locations and formats. In contrast, `packer-plugin-vmware`
  includes builders that operate on both VMware vSphere Hypervisor and the aforementioned desktop
  virtualization products, providing a different set of functionalities, including support for
  Vagrant.

Please refer to the documentation for each plugin to understand the specific capabilities and
configuration options.

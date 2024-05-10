# Packer Plugin for VMware vSphere

The Packer Plugin for VMware vSphere is a multi-component plugin can be used with
[HashiCorp Packer][packer] to create virtual machine images for [VMware vSphere][docs-vsphere]®.

The plugin includes three builders and two post-processors which are able to create images,
depending on your desired strategy:

**Builders**

- `vsphere-iso` - This
  builder starts from an ISO file and uses the vSphere API to build a virtual machine image on
  an ESXi host.

- `vsphere-clone` -
  This builder clones a virtual machine from an existing template using the uses the vSphere API and
  then modifies and saves it as a new template.

- `vsphere-supervisor` -
  This builder deploys and publishes new virtual machine to a vSphere Supervisor cluster using VM
  Service.

**Post-Processors**

- `vsphere` - This post-processor uploads an artifact to a vSphere endpoint. The artifact must be a
  VMX, OVA, or OVF file.

- `vsphere-template` - This post-processor uses an artifact from the `vmware-iso` builder with an
  ESXi host or an artifact from the [vSphere](/packer/plugins/post-processors/vsphere/vsphere)
  post-processor. It then marks the virtual machine as a template and moves it to your specified
  path.

## Differences from the Packer Plugin for VMware

While both this plugin and the `packer-plugin-vmware` are designed to create virtual machine images,
there are some key differences:

- **Platforms**: This plugin is specifically developed to utilize the VMware vSphere API,
  facilitating the creation of virtual machine images by integrating with VMware vCenter Server and the
  VMware vSphere Hypervisor. On the other hand, `packer-plugin-vmware` supports a variety of
  platforms including VMware vSphere Hypervisor and desktop virtualization products such as VMware
  Fusion, VMware Workstation, and VMware Player, though it does not utilize the vSphere API for its
  operations.

- **Focus**: This plugin is purpose-built with a focus on VMware vSphere, offering capabilities such
  as creating virtual machine images, cloning and modifying base virtual machine images, and
  exporting artifacts in specified locations and formats. In contrast, `packer-plugin-vmware`
  includes builders that operate on both VMware vSphere Hypervisor and the aforementioned desktop
  virtualization products, providing a different set of functionalities, including support for
  Vagrant.

Please refer to the documentation for each plugin to understand the specific capabilities and configuration options.

## Requirements

- [VMware vSphere][docs-vsphere]

    The plugin supports versions in accordance with the [Broadcom Product Lifecycle][product-lifecycle].

- [Go 1.19][golang-install]

    Required if building the plugin.

## Installation

### Using Pre-built Releases

#### Automatic Installation

Packer v1.7.0 and later supports the `packer init` command which enables the automatic installation
of Packer plugins. For more information, see the [Packer documentation][docs-packer-init].

To install this plugin, copy and paste this code (HCL2) into your Packer configuration and run
`packer init`.

```hcl
packer {
  required_version = ">= 1.7.0"
  required_plugins {
    vsphere = {
      version = ">= 1.2.7"
      source  = "github.com/hashicorp/vsphere"
    }
  }
}
```

#### Manual Installation

You can download [pre-built binary releases][releases-vsphere-plugin] of the plugin on GitHub. Once
you have downloaded the latest release archive for your target operating system and architecture,
extract the release archive to retrieve the plugin binary file for your platform.

To install the downloaded plugin, please follow the Packer documentation on [installing a plugin][docs-packer-plugin-install].

### Using the Source

If you prefer to build the plugin from sources, clone the GitHub repository locally and run the
command `go build` from the repository root directory. Upon successful compilation, a
`packer-plugin-vsphere` plugin binary file can be found in the root directory.

To install the compiled plugin, please follow the Packer documentation on [installing a plugin][docs-packer-plugin-install].

### Configuration

For more information on how to configure the plugin, please see the [plugin documentation][docs-vsphere-plugin].

- `vsphere-iso` [builder documentation][docs-vsphere-iso]

- `vsphere-clone` [builder documentation][docs-vsphere-clone]

- `vsphere-supervisor` [builder documentation][docs-vsphere-supervisor]

## Contributing

- If you think you've found a bug in the code or you have a question regarding the usage of this
  software, please reach out to us by opening an issue in this GitHub repository.

- Contributions to this project are welcome: if you want to add a feature or a fix a bug, please do
  so by opening a pull request in this GitHub repository. In case of feature contribution, we kindly
  ask you to open an issue to discuss it beforehand.

[docs-packer-init]: https://developer.hashicorp.com/packer/docs/commands/init
[docs-packer-plugin-install]: https://developer.hashicorp.com/packer/docs/plugins/install-plugins
[docs-vsphere]: https://docs.vmware.com/en/VMware-vSphere/
[docs-vsphere-clone]: https://developer.hashicorp.com/packer/plugins/builders/vsphere/vsphere-clone
[docs-vsphere-iso]: https://developer.hashicorp.com/packer/plugins/builders/vsphere/vsphere-iso
[docs-vsphere-supervisor]: https://developer.hashicorp.com/packer/plugins/builders/vsphere/vsphere-supervisor
[docs-vsphere-plugin]: https://developer.hashicorp.com/packer/plugins/builders/vsphere
[golang-install]: https://golang.org/doc/install
[packer]: https://www.packer.io
[releases-vsphere-plugin]: https://github.com/hashicorp/packer-plugin-vsphere/releases
[product-lifecycle]: https://support.broadcom.com/group/ecx/productlifecycle

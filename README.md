# Packer Plugin for VMware vSphere

The Packer Plugin for VMware vSphere is a multi-component plugin can be used with [HashiCorp Packer][packer] to create virtual machine images for [VMware vSphere][docs-vsphere].

The plugin includes two builders which are able to create images, depending on your desired strategy:

* `vsphere-iso` - This builder starts from an ISO file and utilizes the vSphere API to build images on a vSphere cluster or ESXi host by connecting to a vCenter Server instance.

* `vsphere-clone` - This builder clones an existing virtual machine template, modifies the template, and then saves it as a new image. The builder uses the vSphere API to build images on a vSphere cluster or ESXi host by connecting to a vCenter Server instance.

* `vsphere-supervisor` - This builder creates a virtual machine on a vSphere Supervisor cluster by using the VM-Service API.

## Requirements

* [VMware vSphere][docs-vsphere] 6.5 or later.

    The provider supports versions in accordance with the VMware Product Lifecycle Matrix from General Availability to End of General Support.

    Learn more: [VMware Product Lifecycle Matrix][vmware-product-lifecycle-matrix]

* [Go 1.17][golang-install]

    Required if building the plugin.

## Installation

### Using Pre-built Releases

#### Automatic Installation

Packer v1.7.0 and later supports the `packer init` command which enables the automatic installation of Packer plugins. For more information, see the [Packer documentation][docs-packer-init].

To install this plugin, copy and paste this code (HCL2) into your Packer configuration and run `packer init`.

```hcl
packer {
  required_version = ">= 1.7.0"
  required_plugins {
    vsphere = {
      version = ">= 1.0.6"
      source  = "github.com/hashicorp/vsphere"
    }
  }
}
```

#### Manual Installation

You can download [pre-built binary releases][releases-vsphere-plugin] of the plugin on GitHub. Once you have downloaded the latest release archive for your target operating system and architecture, uncompress to retrieve the plugin binary file for your platform.

To install the downloaded plugin, please follow the Packer documentation on [installing a plugin][docs-packer-plugin-install].

### Using the Source

If you prefer to build the plugin from sources, clone the GitHub repository locally and run the command `go build` from the repository root directory. Upon successful compilation, a `packer-plugin-vsphere` plugin binary file can be found in the root directory.

To install the compiled plugin, please follow the Packer documentation on [installing a plugin][docs-packer-plugin-install].

### Configuration

For more information on how to configure the plugin, please see the plugin documentation

* `vsphere-iso` [builder documentation][docs-vsphere-iso]

* `vsphere-clone` [builder documentation][docs-vsphere-clone]

## Contributing

* If you think you've found a bug in the code or you have a question regarding the usage of this software, please reach out to us by opening an issue in this GitHub repository.

* Contributions to this project are welcome: if you want to add a feature or a fix a bug, please do so by opening a pull request in this GitHub repository. In case of feature contribution, we kindly ask you to open an issue to discuss it beforehand.

[docs-packer-init]: https://www.packer.io/docs/commands/init
[docs-packer-plugin-install]: https://www.packer.io/docs/extending/plugins/#installing-plugins
[docs-vsphere]: https://docs.vmware.com/en/VMware-vSphere/
[docs-vsphere-clone]: https://www.packer.io/docs/builders/vsphere/vsphere-clone
[docs-vsphere-iso]: https://www.packer.io/docs/builders/vsphere/vsphere-iso
[docs-vsphere-plugin]: https://www.packer.io/docs/builders/vsphere
[golang-install]: https://golang.org/doc/install
[packer]: https://www.packer.io
[releases-vsphere-plugin]: https://github.com/hashicorp/packer-plugin-vsphere/releases
[vmware-product-lifecycle-matrix]: https://lifecycle.vmware.com

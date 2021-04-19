## The Example Folder

This folder contains a fully working example of the plugin usage. The
example defines the `required_plugins` block. A pre-defined GitHub Action
will run `packer init`, `packer validate`, and `packer build` to test the
plugin with the latest version available of Packer.

The folder contains multiple HCL2 compatible files. The action will execute
Packer at this folder level running `packer init -upgrade .` and `packer build
.`.

If the plugin requires authentication, the configuration should be provided via
GitHub Secrets and set as environment variables in the
[test-plugin-example.yml](/.github/workflows/test-plugin-example.yml) file.
Example:

```yml
  - name: Build
    working-directory: ${{ github.event.inputs.folder }}
    run: PACKER_LOG=${{ github.event.inputs.logs }} packer build .
    env:
      AUTH_KEY: ${{ secrets.AUTH_KEY }}
      AUTH_PASSWORD: ${{ secrets.AUTH_PASSWORD }}
```

### Required variables

To run this example the following variables are required

```hcl
# file: vars.auto.pkrvars.hcl
bastion_host = "XXX.XXX.XXX.XXX"
bastion_user = "root"
datacenter_name = "datacenter"
esxi_host = "XXX.XXX.XXX.XXX"
esxi_password = "password"
esxi_user = "root"
vcenter_endpoint = "XXX.XXX.XXX.XXX"
vcenter_password = "password"
vcenter_user = "Administrator@vsphere.local"
vm_ip = "XXX.XXX.XXX.XXX"
gateway_ip = "XXX.XXX.XXX.XXX"
```
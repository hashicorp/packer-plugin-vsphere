Type: `vsphere-virtualmachine`
Artifact BuilderId: `vsphere.virtualmachine`

This data source retrieves information about existing virtual machines from vSphere
and return name of one virtual machine that matches all specified filters. This virtual
machine can be used in the vSphere Clone builder to select a template.

## Configuration Reference

### Filters Configuration

**Optional:**

<!-- Code generated from the comments of the Config struct in datasource/virtualmachine/data.go; DO NOT EDIT MANUALLY -->

- `name` (string) - Basic filter with glob support (e.g. `ubuntu_basic*`). Defaults to `*`.
  Using strict globs will not reduce execution time because vSphere API
  returns the full inventory. But can be used for better readability over
  regular expressions.

- `name_regex` (string) - Extended name filter with regular expressions support
  (e.g. `ubuntu[-_]basic[0-9]*`). Default is empty. The match of the
  regular expression is checked by substring. Use `^` and `$` to define a
  full string. For example, the `^[^_]+$` filter will search names
  without any underscores. The expression must use
  [Go Regex Syntax](https://pkg.go.dev/regexp/syntax).

- `template` (bool) - Filter to return only objects that are virtual machine templates.
  Defaults to `false` and returns all virtual machines.

- `host` (string) - Filter to search virtual machines only on the specified ESX host.

- `tag` ([]Tag) - Filter to return only that virtual machines that have attached all
  specified tags. Specify one or more `tag` blocks to define list of tags
   for the filter.

  HCL Example:

  ```hcl
  	tag {
  	  category = "team"
  	  name = "operations"
  	}
  	tag {
  	  category = "sla"
  	  name = "gold"
  	}
  ```

- `latest` (bool) - This filter determines how to handle multiple machines that were
  matched with all previous filters. Machine creation time is being used
  to find latest. By default, multiple matching machines results in an
  error.

<!-- End of code generated from the comments of the Config struct in datasource/virtualmachine/data.go; -->


### Tags Filter Configuration

**Required:**

<!-- Code generated from the comments of the Tag struct in datasource/virtualmachine/data.go; DO NOT EDIT MANUALLY -->

- `name` (string) - Name of the tag added to virtual machine which must pass the `tag`
  filter.

- `category` (string) - Name of the tag category that contains the tag.

  -> **Note:** Both `name` and `category` must be specified in the `tag`
  filter.

<!-- End of code generated from the comments of the Tag struct in datasource/virtualmachine/data.go; -->


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


## Output

<!-- Code generated from the comments of the DatasourceOutput struct in datasource/virtualmachine/data.go; DO NOT EDIT MANUALLY -->

- `vm_name` (string) - Name of the found virtual machine.

<!-- End of code generated from the comments of the DatasourceOutput struct in datasource/virtualmachine/data.go; -->


## Example Usage

This example demonstrates how to connect to vSphere cluster and search for the latest virtual machine
that matches the filters. The name of the machine is then output to the console as an output variable.
```hcl
data "vsphere-virtualmachine" "default" {
    vcenter_server = "vcenter.example.com"
    insecure_connection = true
    username = "administrator@vsphere.local"
    password = "VMware1!"
    datacenter = "dc-01"
    latest = true
    tag {
	  category = "team"
	  name = "operations"
	}
	tag {
	  category = "sla"
	  name = "gold"
	}

}

locals {
  vm_name = data.vsphere-virtualmachine.default.vm_name
}

source "null" "example" {
    communicator = "none"
}

build {
  sources = [
    "source.null.example"
  ]

  provisioner "shell-local" {
    inline = [
      "echo vm_name: ${local.vm_name}",
    ]
  }
}
```

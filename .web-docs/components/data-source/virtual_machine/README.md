Type: `vsphere-virtual_machine`
Artifact BuilderId: `vsphere.virtual_machine`

This datasource is able to get information about existing virtual machines from vSphere
and return name of one virtual machine that matches all specified filters. This virtual
machine can later be used in the vSphere Clone builder to select template.

## Configuration Reference

### Filters Configuration

**Optional:**

<!-- Code generated from the comments of the Config struct in datasource/virtual_machine/data.go; DO NOT EDIT MANUALLY -->

- `name` (string) - Basic filter with glob support (e.g. `nginx_basic*`). Defaults to `*`.
  Using strict globs will not reduce execution time because vSphere API returns the full inventory.
  But can be used for better readability over regular expressions.

- `name_regex` (string) - Extended name filter with regular expressions support (e.g. `nginx[-_]basic[0-9]*`). Default is empty.
  The match of the regular expression is checked by substring. Use `^` and `$` to define a full string.
  E.g. the `^[^_]+$` filter will search names without any underscores.
  The expression must use [Go Regex Syntax](https://pkg.go.dev/regexp/syntax).

- `template` (bool) - Filter to return only objects that are virtual machine templates.
  Defaults to `false` and returns all VMs.

- `node` (string) - Filter to search virtual machines only on the specified node.

- `vm_tags` ([]Tag) - Filter to return only that virtual machines that have attached all specifies tags.
  Specify one or more `vm_tags` blocks to define list of tags that will make up the filter.
  Should work since vCenter 6.7. To avoid incompatibility, REST client is being
  initialized only when at least one tag has been defined in the config.

- `latest` (bool) - This filter determines how to handle multiple machines that were matched with all
  previous filters. Machine creation time is being used to find latest.
  By default, multiple matching machines results in an error.

<!-- End of code generated from the comments of the Config struct in datasource/virtual_machine/data.go; -->


### Tags Filter Configuration

<!-- Code generated from the comments of the Tag struct in datasource/virtual_machine/data.go; DO NOT EDIT MANUALLY -->

Example of multiple vm_tags blocks in HCL format:
```

	vm_tags {
	  category = "team"
	  name = "operations"
	}
	vm_tags {
	  category = "SLA"
	  name = "gold"
	}

```

<!-- End of code generated from the comments of the Tag struct in datasource/virtual_machine/data.go; -->


**Required:**

<!-- Code generated from the comments of the Tag struct in datasource/virtual_machine/data.go; DO NOT EDIT MANUALLY -->

- `name` (string) - Tag with this name must be attached to virtual machine which should pass the Tags Filter.

- `category` (string) - Name of the category that contains this tag. Both tag and category must be specified.

<!-- End of code generated from the comments of the Tag struct in datasource/virtual_machine/data.go; -->


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

<!-- Code generated from the comments of the DatasourceOutput struct in datasource/virtual_machine/data.go; DO NOT EDIT MANUALLY -->

- `vm_name` (string) - Name of the found virtual machine.

<!-- End of code generated from the comments of the DatasourceOutput struct in datasource/virtual_machine/data.go; -->


## Example Usage

This is a very basic example that connects to vSphere cluster and tries to search
the latest virtual machine that matches all filters. The machine name is then printed
to console as output variable.
```hcl
data "vsphere-virtual_machine" "default" {
    vcenter_server = "vcenter.example.org"
    insecure_connection = true
    username = "administrator@example.org"
    password = "St4ongPa$$w0rd"
    datacenter = "AZ1"
    latest = true
    vm_tags {
	  category = "team"
	  name = "operations"
	}
	vm_tags {
	  category = "SLA"
	  name = "gold"
	}

}

locals {
  vm_name = data.vsphere-virtual_machine.default.vm_name
}

source "null" "basic-example" {
    communicator = "none"
}

build {
  sources = [
    "source.null.basic-example"
  ]

  provisioner "shell-local" {
    inline = [
      "echo vm_name: ${local.vm_name}",
    ]
  }
}


```

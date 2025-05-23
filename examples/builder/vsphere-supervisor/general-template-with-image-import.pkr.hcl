# A Packer template to deploy and publish a VM-Service VM using the vsphere-supervisor builder.

# image import configs.
variable "import_source_url" {
  type = string
}
variable "import_source_ssl_certificate" {
    type = string
}
variable "import_target_location_name" {
    type = string
}
variable "import_target_image_type" {
    type = string
    default = null
}
variable "import_target_image_name" {
    type = string
    default = null
}
variable "clean_imported_image" {
    type = bool
    default = false
}
variable "keep_import_request" {
    type = bool
    default = false
}

# VM-Service source VM configs.
variable "image_name" {
  type = string
  default = null
}
variable "class_name" {
  type = string
}
variable "storage_class" {
  type = string
}
variable "source_name" {
  type = string
  default = null
}
variable "bootstrap_provider" {
  type = string
  default = "CloudInit"
}
variable "bootstrap_data_file" {
  type = string
  default = null
}

# Supervisor cluster configs.
variable "kubeconfig_path" {
  type = string
  default = null
}
variable "supervisor_namespace" {
  type = string
  default = null
}

# SSH connection configs.
variable "communicator" {
  type = string
  default = "ssh"
}
variable "ssh_username" {
  type = string
  default = "packer"
}
variable "ssh_password" {
  type = string
  default = "packer"
  sensitive = true
}
variable "ssh_bastion_host" {
  type = string
  default = null
}
variable "ssh_bastion_username" {
  type = string
  default = null
}
variable "ssh_bastion_password" {
  type = string
  default = null
  sensitive = true
}

# Whether to keep the created source VM after the build.
variable "keep_input_artifact" {
  type = bool
  default = false
}

# VM publishing configs.
variable "publish_location_name" {
  type = string
  default = null
}
variable "publish_image_name" {
  type = string
  default = null
}

# Watch timeout related configs.
variable "watch_import_timeout_sec" {
  type = number
  default = 600
}
variable "watch_source_timeout_sec" {
  type = number
  default = 1800
}
variable "watch_publish_timeout_sec" {
  type = number
  default = 600
}

source "vsphere-supervisor" "vm" {
  kubeconfig_path = "${var.kubeconfig_path}"
  supervisor_namespace = "${var.supervisor_namespace}"
  class_name = "${var.class_name}"
  import_source_url = "${var.import_source_url}"
  import_source_ssl_certificate = "${var.import_source_ssl_certificate}"
  import_target_location_name = "${var.import_target_location_name}"
  import_target_image_type = "${var.import_target_image_type}"
  import_target_image_name = "${var.import_target_image_name}"
  clean_imported_image = "${var.clean_imported_image}"
  keep_import_request = "${var.keep_import_request}"
  source_name = "${var.source_name}"
  storage_class = "${var.storage_class}"
  bootstrap_provider = "${var.bootstrap_provider}"
  bootstrap_data_file = "${var.bootstrap_data_file}"
  communicator = "${var.communicator}"
  ssh_username = "${var.ssh_username}"
  ssh_password = "${var.ssh_password}"
  ssh_bastion_host = "${var.ssh_bastion_host}"
  ssh_bastion_username = "${var.ssh_bastion_username}"
  ssh_bastion_password = "${var.ssh_bastion_password}"
  keep_input_artifact = "${var.keep_input_artifact}"
  publish_location_name = "${var.publish_location_name}"
  publish_image_name = "${var.publish_image_name}"
  watch_import_timeout_sec = "${var.watch_import_timeout_sec}"
  watch_source_timeout_sec = "${var.watch_source_timeout_sec}"
  watch_publish_timeout_sec = "${var.watch_publish_timeout_sec}"
}

build {
  sources = ["source.vsphere-supervisor.vm"]
  provisioner "shell" {
    inline = [
      "echo 'Hello from Packer!' > ./hello-packer.txt",
    ]
  }
}

# A Packer template to deploy a VM-Service Windows VM using the vsphere-supervisor builder.

# Supervisor cluster configs.
variable "kubeconfig_path" {}
variable "supervisor_namespace" {}

# VM-Service source VM configs.
variable "class_name" {}
variable "image_name" {}
variable "source_name" {}
variable "storage_class" {}
variable "bootstrap_provider" {}
variable "bootstrap_data_file" {}

# SSH connection configs.
variable "ssh_username" {}
variable "ssh_password" {}
variable "ssh_bastion_host" {}
variable "ssh_bastion_username" {}
variable "ssh_bastion_password" {}

# Whether to keep the created source VM after the build.
variable "keep_input_artifact" {}

source "vsphere-supervisor" "vm" {
  kubeconfig_path = "${var.kubeconfig_path}"
  supervisor_namespace = "${var.supervisor_namespace}"
  class_name = "${var.class_name}"
  image_name = "${var.image_name}"
  source_name = "${var.source_name}"
  storage_class = "${var.storage_class}"
  bootstrap_provider = "${var.bootstrap_provider}"
  bootstrap_data_file = "${var.bootstrap_data_file}"
  ssh_username = "${var.ssh_username}"
  ssh_password = "${var.ssh_password}"
  ssh_bastion_host = "${var.ssh_bastion_host}"
  ssh_bastion_username = "${var.ssh_bastion_username}"
  ssh_bastion_password = "${var.ssh_bastion_password}"
  keep_input_artifact = "${var.keep_input_artifact}"
}

build {
  sources = ["source.vsphere-supervisor.vm"]
  provisioner "powershell" {
    inline = [
      "Set-Location $env:TEMP",
      "echo 'Hello from Packer!' | Out-File -FilePath ./hello-packer.txt",
    ]
  }
}

# Note: this is an example template file to show you how to use the vsphere-supervisor builder
# to deploy a VM with Nginx installed. You can use this file as a starting point for your own.

source "vsphere-supervisor" "vm" {
  image_name = "<Image name of the source VM, e.g. 'ubuntu-impish-21.10-cloudimg>'"
  class_name = "<VM class that describes the virtual hardware settings, e.g. 'best-effort-large'>"
  storage_class = "<Storage class that provides the backing storage for volume, e.g. 'wcplocal-storage-profile'>"
  kubeconfig_path = "<Path to kubeconfig file of the Supervisor cluster, e.g. '/home/user/.kube/config'>"
  supervisor_namespace = "<Namespace of the source VM in Supervisor cluster>"
  source_name = "<Name of the source VM and its required objects to deploy>"
  network_type = "<Network type of the source VM, e.g. 'nsx-t'>"
  ssh_username = "<SSH username to set in the source VM>"
  ssh_password = "<SSH password to set in the source VM>"
  ssh_bastion_host = "<IP address of the bastion host for Packer to access the source VM>"
  watch_source_timeout_sec = "<Timeout in seconds to wait for the source VM to be ready>"
  keep_input_artifact = "<Whether to keep the created source VM and its other objects>"
}

build {
  sources = ["source.vsphere-supervisor.example-vm"]
  provisioner "shell" {
    inline = [
      "yum install -qy nginx",
      "systemctl restart nginx",
      "systemctl status nginx",
      "echo 'Testing Nginx connectivity...'",
      "curl -sI http://localhost:80",
    ]
  }
  provisioner "ansible" {
    playbook_file = "<Path to the Ansible playbook file, e.g. 'cleanup-playbook.yml' provided in this folder>"
  }
}

# A Packer template to deploy a VM-Service VM using the vsphere-supervisor builder.
# It installs Jenkins and runs a sample hello-world job in the deployed VM.

# VM-Service source VM configs.
variable "image_name" {
  type = string
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

source "vsphere-supervisor" "vm" {
  kubeconfig_path = "${var.kubeconfig_path}"
  supervisor_namespace = "${var.supervisor_namespace}"
  class_name = "${var.class_name}"
  image_name = "${var.image_name}"
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
}

build {
  sources = ["source.vsphere-supervisor.vm"]

  # Jenkins job configuration file.
  provisioner "file" {
    destination = "/tmp/sample-job.xml"
    content = <<EOF
<?xml version='1.1' encoding='UTF-8'?>
<project>
  <description>A sample job</description>
  <builders>
    <hudson.tasks.Shell>
      <command>echo "Hello VM-Service from Jenkins"</command>
    </hudson.tasks.Shell>
  </builders>
</project>
EOF
  }

  provisioner "shell" {
    inline = [
      # Download Jenkins repository key and add it to the trusted keyrings.
      "curl -fsSL https://pkg.jenkins.io/debian/jenkins.io-2023.key | sudo gpg --dearmor -o /usr/share/keyrings/jenkins-keyring.gpg",
      "echo deb [signed-by=/usr/share/keyrings/jenkins-keyring.gpg] https://pkg.jenkins.io/debian binary/ | sudo tee /etc/apt/sources.list.d/jenkins.list",

      # Download the new Kubernetes community-owned repository key and add it to the trusted keyrings (to get apt-get update working).
      "curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.28/deb/Release.key | sudo gpg --dearmor -o /usr/share/keyrings/kubernetes-apt-keyring.gpg",
      "echo deb [signed-by=/usr/share/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.28/deb/ / | sudo tee /etc/apt/sources.list.d/kubernetes.list",

      # Sometimes apt-get uses IPv6 and causes failure, force to use IPv4 address.
      "sudo apt-get -qq -o Acquire::ForceIPv4=true update",
      "sudo apt-get -qq -o Acquire::ForceIPv4=true install -f -y ca-certificates openjdk-11-jre-headless",
      "sudo apt-get -qq -o Acquire::ForceIPv4=true install -f -y jenkins",
      # Restart Jenkins service, in case it didn't initialize successfully.
      "sudo systemctl restart jenkins",

      "export JENKINS_URL=http://localhost:8080/",
      "export USER=admin",
      "export PASSWORD=$(sudo cat /var/lib/jenkins/secrets/initialAdminPassword)",
      # Download Jenkins CLI to create and check job status.
      "wget -q -O /tmp/jenkins-cli.jar $JENKINS_URL/jnlpJars/jenkins-cli.jar",      
      # Create a new job from the above sample-job.xml file.
      "java -jar /tmp/jenkins-cli.jar -s $JENKINS_URL -auth $USER:$PASSWORD create-job sample-job < /tmp/sample-job.xml",
      # Build and wait for a successful completion of the job.
      "java -jar /tmp/jenkins-cli.jar -s $JENKINS_URL -auth $USER:$PASSWORD build sample-job -s -v",
    ]
  }
}

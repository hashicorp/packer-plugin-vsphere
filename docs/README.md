# VMware vSphere Components

The vSphere plugin is able to create vSphere virtual machines for use with any VMware product. 
To achieve this, the plugin comes with three builders, and two post-processors
to build the VM depending on the strategy you want to use.

### Builders:
- [vsphere-iso](/packer/plugins/builders/vsphere-iso.mdx) - This builder starts from an
  ISO file and utilizes the vSphere API to build on a remote esx instance.
  This allows you to build vms even if you do not have SSH access to your vSphere cluster.

- [vsphere-clone](/packer/plugins/builders/vsphere-clone.mdx) - This builder clones a
  vm from an existing template, then modifies it and saves it as a new
  template. It uses the vSphere API to build on a remote esx instance.
  This allows you to build vms even if you do not have SSH access to your vSphere cluster.

- [vsphere-supervisor](/packer/plugins/builders/vsphere-supervisor.mdx) - This builder deploys a
  vm to a vSphere Supervisor cluster, using the VM-Service API. This allows you to build
  vms without spec yaml files and configure them after using the Packer provisioners.

### Post-Processors
- [vsphere](/packer/plugins/post-processors/vsphere.mdx) - The Packer vSphere post-processor takes an artifact 
  and uploads it to a vSphere endpoint.

- [vsphere-template](/packer/plugins/post-processors/vsphere-template.mdx) - The Packer vSphere Template post-processor takes an 
  artifact from the vmware-iso builder, built on an ESXi host (i.e. remote) or an artifact from the 
  [vSphere](/packer/plugins/post-processors/vsphere) post-processor, marks the VM as a template, and leaves it in the path of 
  your choice.

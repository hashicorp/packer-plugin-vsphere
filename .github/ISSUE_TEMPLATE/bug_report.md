---
name: Bug Report
about: You're experiencing an issue with this Packer plugin that is different than the documented behavior.
labels: bug
---

When filing a bug please include the following headings, if possible. 

<!--
Any example text in this template can be deleted.
-->

#### Overview of the Issue

Please provide a clear and concise description of the issue you are experiencing.

#### Reproduction Steps

Please provide the steps to reproduce the issue.

### Packer Version

Please provide the Packer version.

<!--
    From `packer version`

    Example:
     - 1.7.8
-->

### Plugin Version and Builders

Please provide the plugin version.

<!--
    Example:    
     - 1.0.2
-->

Please select the builder.

<!--
    Please check the builder(s) that applies to this issue using "[x]".
-->

- [] `vsphere-iso`
- [] `vsphere-clone`

### VMware vSphere Version

Please provide the VMware vSphere version.

<!--
    Examples:
     - 7.0 Update 2
     - 7.0.2
-->

### Guest Operating System

<!--
    The guest operating system(s) being built.

    Examples:
     - Ubuntu 20.04 LTS x64
-->

### Simplified Packer Buildfile

If the file is longer than a few dozen lines, please include the URL to the [gist](https://gist.github.com/) of the log or use the [GitHub detailed format](https://gist.github.com/ericclemmons/b146fe5da72ca1f706b2ef72a20ac39d) instead of posting it directly in the issue.

### Operating System and Environment Details

Please add any information you can provide about the environment.

<!--
    Example:
     - Operating System: macOS Big Sur (Intel)
-->

### Log Fragments and `crash.log` Files

Include appropriate log fragments. If the log is longer than a few dozen lines, please include the URL to the [gist](https://gist.github.com/) of the log or use the [Github detailed format](https://gist.github.com/ericclemmons/b146fe5da72ca1f706b2ef72a20ac39d) instead of posting it directly in the issue.

Set the env var `PACKER_LOG=1` for maximum log detail.
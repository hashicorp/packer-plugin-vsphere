# Latest Release

Please refer to [releases](https://github.com/hashicorp/packer-plugin-vsphere/releases) for the latest CHANGELOG information.

---
## 1.0.2 (October 18, 2021)

### NOTES:
Support for the HCP Packer registry is currently in beta and requires
Packer v1.7.7 [GH-115] [GH-120]

### FEATURES:
* Add HCP Packer registry image metadata to all artifacts. [GH-115]
    [GH-120]

## IMPROVEMENTS:
* Add `floppy_content` parameter, which is similar to `http_content` and
    `cd_content`, but for floppy. [GH-117]
* Add skip_import parameter in content_library_destination configuration.
    [GH-121]
* Update packer-plugin-sdk to latest version v0.2.7 [GH-118]

## 1.0.1 (July 22, 2021)

* Pass `cd_content` config to vSphere ISO and Clone builders [GH-87]
* Added `snapshot_name` option for vSphere ISO and Clone builders [GH-85]
* Take the host configuration into account while cloning a VM [GH-79]
* Fix removing templates with snapshots [GH-78]

## 1.0.0 (June 14, 2021)

* Bump github.com/hashicorp/packer-plugin-sdk to v0.2.3 [GH-69]
* Fixed vSphere resource ID/value load order when importing VM template to Content Library [GH-58]
* Fix firmware reverting to BIOS when pushing to template in content library [GH-68]

## 0.0.1 (April 15, 2021)

* vSphere Plugin break out from Packer core. Changes prior to break out can be found in [Packer's CHANGELOG](https://github.com/hashicorp/packer/blob/master/CHANGELOG.md).

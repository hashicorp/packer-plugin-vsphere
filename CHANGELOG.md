## 1.0.1 (July 22, 2021)

* Pass `cd_content` config to vSphere ISO and Clone builders [GH-87]
* Added `snapshot_name` option for vSphere ISO and Clone builders [GH-85]
* Take the host configuration into account while cloning a VM [GH-79]
* Fix removing templates with snapshots [GH-78]

## 1.0.0 (June 14, 2021)

* Bump github.com/hashicorp/packer-plugin-sdk to v0.2.3 [GH-69]
* Fixed VMWare Resource ID/Value load order when importing VM template to Content Library [GH-58]
* Fix firmware reverting to BIOSwhen pushing to template in content library [GH-68]

## 0.0.1 (April 15, 2021)

* vSphere Plugin break out from Packer core. Changes prior to break out can be found in [Packer's CHANGELOG](https://github.com/hashicorp/packer/blob/master/CHANGELOG.md).

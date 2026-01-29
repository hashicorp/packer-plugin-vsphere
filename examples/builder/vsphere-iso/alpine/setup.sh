#!/bin/sh
# © Broadcom. All Rights Reserved.
# The term "Broadcom" refers to Broadcom Inc. and/or its subsidiaries.
# SPDX-License-Identifier: MPL-2.0

set -ex

sed -i 's/#http/http/' /etc/apk/repositories
apk update

apk add openssl
apk add open-vm-tools
apk add open-vm-tools-plugins-all
rc-update add open-vm-tools
/etc/init.d/open-vm-tools start

cat >/usr/local/bin/shutdown <<EOF
#!/bin/sh
poweroff
EOF
chmod +x /usr/local/bin/shutdown

sed -i "/#PermitRootLogin/c\PermitRootLogin yes" /etc/ssh/sshd_config
/etc/init.d/sshd restart

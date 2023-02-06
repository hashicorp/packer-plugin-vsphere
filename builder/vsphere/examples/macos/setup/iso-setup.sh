#!/bin/sh
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

set -eux

mkdir -p out/pkgroot
rm -rf /out/pkgroot/*

mkdir -p out/scripts
rm -rf /out/scripts/*
cp postinstall out/scripts/

pkgbuild \
  	--identifier io.packer.install \
    --root out/pkgroot \
  	--scripts out/scripts \
  	out/postinstall.pkg

mkdir -p out/iso
rm -rf out/iso/*
cp setup.sh out/iso/
chmod +x out/iso/setup.sh

productbuild --package out/postinstall.pkg out/iso/postinstall.pkg

rm -f out/setup.iso
hdiutil makehybrid -iso -joliet -default-volume-name setup -o out/setup.iso out/iso
cd out
shasum -a 256 setup.iso >sha256sums

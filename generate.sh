#!/bin/bash

cd ipxe/src
make bin/ipxe.usb bin/ipxe.iso EMBED=../../../build/embed.ipxe
mv bin/ipxe.usb ../../../build/ipxe.usb.img
mv bin/ipxe.iso ../../../build/ipxe.iso

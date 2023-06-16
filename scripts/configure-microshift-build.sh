#!/usr/bin/env bash

. $(dirname $0)/../demo.conf

[[ $EUID -ne 0 ]] && exit_on_error "Must run as root"

##
## Add the necessary repos for Microshift
##

subscription-manager repos \
    --enable rhocp-4.13-for-rhel-9-$(uname -i)-rpms \
    --enable fast-datapath-for-rhel-9-$(uname -i)-rpms


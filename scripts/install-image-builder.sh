#!/usr/bin/env bash

. $(dirname $0)/../demo.conf

[[ $EUID -ne 0 ]] && exit_on_error "Must run as root"

##
## Install the packages
##

dnf -y install osbuild-composer composer-cli cockpit-composer jq \
    bash-completion golang

dnf -y install container-tools

##
## Start the socket listeners
##

systemctl enable --now osbuild-composer.socket cockpit.socket

##
## Add user to weldr group
##

[[ ! -z "$SUDO_USER" ]] && usermod -aG weldr $SUDO_USER

##
## Open up port for edge device installations
##

firewall-cmd --permanent --add-port=${IB_PORT}/tcp
firewall-cmd --reload


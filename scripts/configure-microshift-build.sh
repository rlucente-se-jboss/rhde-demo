#!/usr/bin/env bash

. $(dirname $0)/../demo.conf

[[ $EUID -ne 0 ]] && exit_on_error "Must run as root"

##
## Add the necessary repos for Microshift
##

subscription-manager repos \
    --enable rhocp-4.12-for-rhel-8-$(uname -i)-rpms \
    --enable fast-datapath-for-rhel-8-$(uname -i)-rpms

##
## Install various yum utilities
##

dnf -y install yum-utils createrepo_c

##
## Sync Microshift packages to the build host
##

REPO_PATH=/var/repos
mkdir -p ${REPO_PATH}

reposync --arch=$(uname -i) --arch=noarch --gpgcheck \
    --download-path ${REPO_PATH}/microshift-local \
    --repo=rhocp-4.12-for-rhel-8-$(uname -i)-rpms \
    --repo=fast-datapath-for-rhel-8-$(uname -i)-rpms

##
## Remove coreos packages to avoid conflicts
##

find ${REPO_PATH}/microshift-local -name \*coreos\* -exec rm -f {} \;

##
## Create a local RPM repository
##

createrepo ${REPO_PATH}/microshift-local

##
## Create an image build source for this new repo
##

tee ${REPO_PATH}/microshift-local/microshift.toml > /dev/null <<EOF
id = "microshift-local"
name = "MicroShift local repo"
type = "yum-baseurl"
url = "file://${REPO_PATH}/microshift-local/"
check_gpg = false
check_ssl = false
system = false
EOF

##
## Add the source file to image builder
##

composer-cli sources add ${REPO_PATH}/microshift-local/microshift.toml


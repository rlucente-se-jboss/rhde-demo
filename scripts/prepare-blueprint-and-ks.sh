#!/usr/bin/env bash

. $(dirname $0)/../demo.conf

[[ $EUID -eq 0 ]] && exit_on_error "Do not run as root"

##
## Generate ssh keys for edge user
##

ssh-keygen -f $HOME/.ssh/id_$EDGE_USER -t rsa -P "" \
           -C $EDGE_USER@localhost.localdomain
cp $HOME/.ssh/id_$EDGE_USER.pub .

##
## Create the blueprint file for running microshift
##

cat > microshift-blueprint.toml <<EOF
name = "Microshift"
description = "Microshift edge blueprint"
version = "0.0.1"

[[packages]]
name = "microshift"
version = "*"

[[packages]]
name = "microshift-greenboot"
version = "*"

[[containers]]
source = "quay.io/redhatgov/ads-b-service:v0.0.5"

[[containers]]
source = "quay.io/redhatgov/ads-b-map:v0.1.6"

[customizations.services]
enabled = ["microshift"]

[[customizations.user]]
name = "$EDGE_USER"
description = "default edge user"
password = "$(openssl passwd -6 $EDGE_PASS)"
key = "$(cat id_$EDGE_USER.pub)"
home = "/home/$EDGE_USER/"
shell = "/usr/bin/bash"
groups = [ "wheel" ]

[[customizations.sshkey]]
user = "$EDGE_USER"
key = "$(cat id_$EDGE_USER.pub)"

[customizations.firewall]
ports = ["6443:tcp", "80:tcp", "443:tcp", "30000:tcp"]

EOF

##
## Create the kickstart file for running Microshift
##

cat > microshift.ks <<EOF
# Partition disk such that it contains an LVM volume group called 'rhel' with a
# 10GB+ system root but leaving free space for the LVMS CSI driver for storing data.
#
# For example, a 20GB disk would be partitioned in the following way:
#
# NAME          MAJ:MIN RM SIZE RO TYPE MOUNTPOINT
# sda             8:0    0  20G  0 disk
# ├─sda1          8:1    0 200M  0 part /boot/efi
# ├─sda1          8:1    0 800M  0 part /boot
# └─sda2          8:2    0  19G  0 part
#  └─rhel-root  253:0    0  10G  0 lvm  /sysroot
#
zerombr
clearpart --all --initlabel
part /boot/efi --fstype=efi --size=200
part /boot --fstype=xfs --asprimary --size=800
part pv.01 --grow
volgroup rhel pv.01
logvol / --vgname=rhel --fstype=xfs --size=10000 --name=root

rootpw --lock
user --name=${EDGE_USER} --groups=wheel --password="$(openssl passwd -6 ${EDGE_PASS})" --iscrypted
 
text

ostreesetup --nogpg --url=http://${IB_SERVER}:${IB_PORT}/repo/ --osname=rhel --remote=edge --ref=rhel/9/x86_64/edge

reboot --eject

%post --log=/var/log/anaconda/post-install.log --erroronfail

# Add the pull secret to CRI-O and set root user-only read/write permissions
cat > /etc/crio/openshift-pull-secret <<EOF1
$(cat pull-secret.txt)
EOF1

chmod 600 /etc/crio/openshift-pull-secret

# Configure the firewall with the mandatory rules for MicroShift
firewall-offline-cmd --zone=trusted --add-source=10.42.0.0/16
firewall-offline-cmd --zone=trusted --add-source=169.254.169.0/24

%end
EOF


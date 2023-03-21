#!/usr/bin/env bash

##
## This script builds an extremely small web service container image
## that also includes an ADS-B dataset.
##

. $(dirname $0)/../demo.conf

[[ $USER = "root" ]] && exit_on_error "Must NOT run as root"
[[ $EUID -ne 0 ]] && exit_on_error "Must run in a 'buildah unshare' environment"

pushd $(dirname $0) &> /dev/null

# build the golang app and statically link all dependencies
CGO_ENABLED=0 GOOS=linux go build -a \
    -ldflags '-extldflags "-static"' ../src/$WS_APP_NAME.go

# build the container from the scratch (empty) container
newcontainer=$(buildah from scratch)
scratchmnt=$(buildah mount $newcontainer)

mkdir -p $scratchmnt/data
cp $WS_APP_NAME $scratchmnt
cp ../data/ads-b-data.json $scratchmnt/data

buildah config --entrypoint "[\"/$WS_APP_NAME\"]" --port ${WS_PORT} \
    --user 1000 $newcontainer
buildah config --author "$WS_AUTHOR" --label name="$WS_APP_NAME" $newcontainer

buildah commit $newcontainer $WS_APP_NAME
buildah unmount $newcontainer

buildah rm $newcontainer

rm $WS_APP_NAME

popd

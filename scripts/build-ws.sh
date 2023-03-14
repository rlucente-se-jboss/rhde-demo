#!/usr/bin/env bash

# TODO THIS NEEDS A LOT OF WORK

# This script builds the smallest web service container image
# possible that also includes the sample dataset.

APP_NAME="ads-b-service"
AUTHOR="Rich Lucente"

#
# build the golang app and statically link all dependencies
#

CGO_ENABLED=0 GOOS=linux go build -a \
    -ldflags '-extldflags "-static"' src/$APP_NAME.go

#
# build the container from scratch
#

newcontainer=$(buildah from scratch)

export $newcontainer
buildah unshare
scratchmnt=$(buildah mount $newcontainer)

mkdir -p $scratchmnt/data
cp $APP_NAME $scratchmnt
cp data/sample-ads-b-data.json $scratchmnt/data/ads-b-data.json

buildah config --entrypoint '["/$APP_NAME"]' --port 8080 \
    --user 1000 $newcontainer
buildah config --author "$AUTHOR" --label name="$APP_NAME" $newcontainer

buildah unmount $newcontainer
buildah commit $newcontainer $APP_NAME

buildah rm $newcontainer


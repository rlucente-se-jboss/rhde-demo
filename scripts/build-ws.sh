#!/usr/bin/env bash

pushd $(dirname $0) &> /dev/null

# This script builds the smallest web service container image
# possible that also includes the sample dataset.

APP_NAME="ads-b-service"
AUTHOR="Rich Lucente"

#
# build the golang app and statically link all dependencies
#

CGO_ENABLED=0 GOOS=linux go build -a \
    -ldflags '-extldflags "-static"' ../src/$APP_NAME.go

#
# build the container from scratch
#

newcontainer=$(buildah from scratch)
scratchmnt=$(buildah mount $newcontainer)

mkdir -p $scratchmnt/data
cp $APP_NAME $scratchmnt
cp ../data/ads-b-data.json $scratchmnt/data

buildah config --entrypoint "[\"/$APP_NAME\"]" --port 8080 \
    --user 1000 $newcontainer
buildah config --author "$AUTHOR" --label name="$APP_NAME" $newcontainer

buildah commit $newcontainer $APP_NAME
buildah unmount $newcontainer

buildah rm $newcontainer

rm $APP_NAME

popd

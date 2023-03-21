#!/usr/bin/env bash

. $(dirname $0)/../demo.conf

[[ $EUID -eq 0 ]] && exit_on_error "Must NOT run as root"

pushd $(dirname $0) &> /dev/null

# place each sample in its own json file

RAWDATA=../data/raw_capture_data
mkdir -p $RAWDATA

for i in $(seq -f %03g 1 $SAMPLE_PTS)
do
    echo -n "$i "
    curl -u "${OPENSKY_USER}:${OPENSKY_PASS}" -s https://opensky-network.org/api/states/all?lamin=${LATMIN}\&lomin=${LONMIN}\&lamax=${LATMAX}\&lomax=${LONMAX} > $RAWDATA/data-$i.json &
    sleep $SAMPLE_DELAY
done
echo

popd &> /dev/null


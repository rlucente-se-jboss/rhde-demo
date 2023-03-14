#!/usr/bin/env bash

USER=__YOUR_OPENSKY_USERNAME__
PASS=__YOUR_OPENSKY_PASSWORD__

# matches background map for Dulles airport in Virginia

LATMIN=38.8686
LATMAX=39.0352
LONMIN=-77.5707
LONMAX=-77.1965

# gather 450 samples with two second inter-sample delay so roughly
# fifteen minutes of data

NUMPTS=450
DELAY=2

# place each sample in its own json file

RAWDATA=$WORKDIR/../data/raw_capture_data
mkdir -p $RAWDATA

for i in $(seq -f %03g 1 $NUMPTS)
do
    echo -n "$i "
    curl -u "${USER}:${PASS}" -s https://opensky-network.org/api/states/all?lamin=${LATMIN}\&lomin=${LONMIN}\&lamax=${LATMAX}\&lomax=${LONMAX} > $RAWDATA/data-$i.json
    sleep $DELAY
done
echo


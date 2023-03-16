# WIP RHDE Demo 
This is very much a work-in-progress. May or may not ever be finished.

## Aircraft sample data
Automatic Dependent Surveillance-Broadcast (ADS-B) data was captured
from the [OpenSky Network](https://opensky-network.org/) for a small
area around Dulles airport in northern Virginia.  The
`states` array data is described in the [OpenSky Network
API](https://openskynetwork.github.io/opensky-api/rest.html#id4).

The sample data covers aircraft events within the following ranges:

| field     | minimum    | maximum    |
| -----     | -------    | -------    |
| time      | 1678744752 | 1678745843 |
| latitude  |  38.8686   |  39.0352   |
| longitude | -77.5707   | -77.1965   |

The `time` field is the number of seconds since epoch (Unix time).

The captured data file includes ADS-B position reports collected
every two seconds for thirty-five aircraft over a period of eighteen
minutes. Each position report has many fields including the callsign,
time, latitude, longitude, and true track (clockwise decimal degrees
with north=0&deg;).

The simple REST web service included in this project will rebase
the earliest position report time to the web service start time so
that the aircraft tracks always appear to be current.

## Images for map and icons
The map was downloaded from [OpenStreetMap](https://www.openstreetmap.org/)
and covers an area with the same lat/lon ranges as the sample data.

The three aircraft icons (64x64) show an aircraft with a transparent
background facing north, northeast, and east. Each image can be
flipped vertically and/or horizontally to get an icon for all eight
compass points.

## ADS-B web service
The simple REST web service takes no parameters and returns a json
file containing an array of current aircraft position reports. The
REST endpoint is accessible at:

    http://localhost:8080/ads-b-states

This service is packaged as a lightweight container with the full 
dataset. Volume mounts can be used for an lternative dataset.

After signing in to [quay.io](https://quay.io), you can pull a
pre-built container [here](https://quay.io/rlucente-se-jboss/ads-b-service).

### Build the web service
To build the web service, simply execute the following command:

    buildah unshare ./scripts/build-ws.sh

### Run the web service
To run the web service with the full data set, use the following
command:

    podman run --rm -d -p 8080:8080 localhost/ads-b-service:latest

The container includes the full dataset but you can use volume
mounts to override that to a different dataset. This would look
like:

    podman run --rm -d -p 8080:8080 \
        -v data:/data localhost/ads-b-service:latest

### Test the service
To test the running service, use the command:

    ./scripts/test-ws.sh


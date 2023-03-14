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
every two seconds for thirty-eight aircraft over a period of fifteen
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
file containing an array of aircraft position reports matching
current time (within +/- 1 second).

This service will be packaged as a lightweight container that has
a small sample dataset. Volume mounts can be used to replace the
data with the more expansive dataset.


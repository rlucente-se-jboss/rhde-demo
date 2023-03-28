# RHDE Demo 
This demo runs two containers on Microshift within a RHEL for Edge
deployment. One container simulates receiving Automatic Dependent
Surveillance-Broadcast (ADS-B) reports and making those reports
available as a simple REST web service. The other container hosts
a web front end to display the reports on a graphical map.

## Install a base system
Install a minimal RHEL 8.7 system. Next, edit `demo.conf` to include
your [Red Hat Customer Portal](https://access.redhat.com) credentials.
Then run the following script to register the system and apply all
updates.

    sudo ./scripts/setup-rhel.sh
    sudo reboot

Next, install the image builder tooling and the web console. This will
enable rpm-ostree builds both graphically and from the command line.

    sudo ./scripts/install-image-builder.sh

The default blueprint for the rpm-ostree image uses the pre-built
container images for the applications. If you're not gathering your
own ADS-B data or rebuilding the applications, you can skip to
[Setup the build environment](#setup-the-build-environment).

## Aircraft sample data
ADS-B data was captured from the [OpenSky Network](https://opensky-network.org/)
for the Washington, DC metro area including the three major airports,
IAD, BWI, and DCA. The `states` array data is described in the
[OpenSky Network API](https://openskynetwork.github.io/opensky-api/rest.html#id4).

The sample data covers aircraft events in the following ranges:

| field     | minimum    | maximum    |
| -----     | -------    | -------    |
| time      | 1679163823 | 1679164732 |
| latitude  |  38.25469  |  39.51589  |
| longitude | -77.96168  | -77.17943  |

The `time` field is the number of seconds since epoch (Unix time).

The captured data file (`data/ads-b-data.json`) includes ADS-B
position reports collected every second for 112 aircraft over a
period of fifteen minutes.  Each position report has many fields
including the callsign, time, latitude, longitude, and true track
(clockwise decimal degrees with north=0&deg;).

### Capturing your own ADS-B data
You can capture your own data by first editing the `demo.conf` file
to modify the geographic area, number of sample points, and delay
between samples (`LATMIN`, `LATMAX`, `LONMIN`, `LONMAX`, `SAMPLE_PTS`,
and `SAMPLE_DELAY`, respectively).

Sign up for an account with [OpenSky Network](https://opensky-network.org)
to be able to access their REST API. Modify `demo.conf` to include
your OpenSky Network credentials and then run the following command
to pull the data.

    ./scripts/capture-opensky-rest-data.sh

Once the capture of raw data is complete, you can convert it to the
expected format for the web service by running the following command.

    ./scripts/convert-raw-data.sh

The `data` directory will be populated with two files: `ads-b-data.json`
which includes the entire dataset and `sample-ads-b-data.json` which
includes only the first one hundred entries from the full dataset.

## ADS-B web service
This REST web service simulates an ADS-B receiver that provides a
simplified report to clients. The web service will rebase the
earliest position report time to the web service start time so that
the aircraft tracks always appear to be current. Additionally,
positions are linearly interpolated from the given dataset so that
reports occur every second for each aircraft.

The service takes no parameters and returns a JSON file containing
an array of current aircraft position reports. The REST endpoint
is accessible at:

    http://localhost:8888/ads-b-states

This service is packaged as a lightweight container with the full 
dataset. Volume mounts can be used for an alternative dataset.

You can pull a pre-built container
[here](https://quay.io/rlucente-se-jboss/ads-b-service).

### Build the ADS-B web service
Edit the `demo.conf` file to modify the application name, author,
and listening port (`WS_APP_NAME`, `WS_AUTHOR`, and `WS_PORT`,
respectively). To build the containerized web service, simply execute
the following command:

    buildah unshare ./scripts/build-ws.sh

Buildah is used to create the smallest possible container image
that only includes the statically linked golang executable and the
ADS-B dataset. The resulting container image is under 10 MiB in
size.

### Run the web service as a container
For testing, the web service can be run directly via podman using
the following command:

    podman run --rm -d -p 8888:8888 localhost/ads-b-service:v0.0.3

The container includes the full dataset but you can use volume
mounts to override with a different dataset. This would look like:

    podman run --rm -d -p 8888:8888 \
        -v data:/data localhost/ads-b-service:v0.0.3

### Run the web service as an executable
You can build and test the web service on any platform with golang
installed.

    cd ./src
    go build ads-b-service.go

Then run the web service with the desired dataset.

    ./ads-b-service -f ../data/ads-b-data.json

Use CTRL-C to stop the service.

### Test the service
To test the running service, use the command:

    ./scripts/test-ws.sh

## Setup the build environment
The following script adheres to the [Microshift product documentation](https://access.redhat.com/documentation/en-us/red_hat_build_of_microshift/4.12/html-single/installing/index)
so please refer to that as you go through the process to enable
including microshift in rpm-ostree images.

    sudo ./scripts/configure-microshift-build.sh

You can confirm that the Microshift source was added correctly using
the commands:

    composer-cli sources list
    composer-cli sources info microshift-local

## Create the edge blueprint file
You'll need to download your [OpenShift pull secret](https://console.redhat.com/openshift/create/local).
Make sure the `pull-secret.txt` file is in the same directory as
`demo.conf`.

Edit `demo.conf` to set the edge user credentials (`EDGE_USER` and
`EDGE_PASS`) and the IP address and port of the server running
image-builder for the rpm-ostree content (`IB_SERVER` and `IB_PORT`).

Run the following script to generate the blueprint and kickstart files.

    ./scripts/prepare-blueprint-and-ks.sh

The local server is now set up to build an rpm-ostree image that
enables microshift and serves the rpm-ostree content and kickstart
to an edge device.

## Build the rpm-ostree image
Push the blueprint file to the image-builder service using the
following commands:

    composer-cli blueprints push microshift-blueprint.toml
    composer-cli blueprints list

Launch an image builder compose using the blueprint file. For this
example, an edge-commit image will be created so that later the
rpm-ostree image contents and a kickstart file can be available
over the network.

    composer-cli compose start-ostree Microshift edge-commit

Wait for the build to complete by monitoring the status of the
compose.

    watch composer-cli compose status

When the compose has a status of FINISHED, hit CTRL-C to stop the
status command.

## Create simple web server for edge device installation
List the composes and select the one you wish to install on the
edge device. Note the UUID of the compose.

    composer-cli compose list

Download the compose from the image builder by substituting the
desired UUID in the following command.

    composer-cli compose image <UUID>

Create a temporary directory and populate it with the rpm-ostree
content and the kickstart file. This directory is where we'll run
a simple web server for the edge device installation.

    mkdir -p tmp
    cd tmp
    tar xvf ../<UUID>-commit.tar
    ln -s ../microshift.ks .

Launch the simple web server using the following commands.

    . ../demo.conf
    python3 -m http.server ${IB_PORT}

## Install the edge device
Download the RHEL 8.7+ [boot ISO](https://access.redhat.com/downloads/content/479/ver=/rhel---8/8.7/x86_64/product-software).
Use the ISO to boot either a virtual machine or physical server
with access to the simple web server above. Append the following
text to the kernel boot line:

    inst.ks=http://IB_SERVER:IB_PORT/microshift.ks

where both `IB_SERVER` and `IB_PORT` match the values in `demo.conf`.

On the host running the small web server, please note the IP address
of the edge device as it pulls both the kickstart file and the
rpm-ostree contents.

The edge device will install automatically and reboot when installation
is complete. Make sure to remove the install media from the physical
edge device so it doesn't attempt to install again.

## Get kubeconfig for MicroShift instance
Copy the file `~/.ssh/id_core` from the image-builder host to your
local workstation or laptop. This should be in the directory `~/.ssh`
with permissions 0600.

Login to the edge device and copy the `kubeconfig` file to the local
user's home directory.

    ssh -i ~/.ssh/id_core core@EDGE_DEVICE_IP
    sudo cp /var/lib/microshift/resources/kubeadmin/kubeconfig .
    sudo chown core: kubeconfig
    exit

Make sure the parameter EDGE_DEVICE_IP above matches the IP address
of the edge device.

Copy the kubeconfig file from the edge device to your local desktop
or workstation with the following commands:

    mkdir -p ~/.kube
    scp -i ~/.ssh/id_core EDGE_USER@EDGE_DEVICE_IP:kubeconfig ~/.kube/config
    chmod og-r ~/.kube/config

The user name `EDGE_USER` should match the value you set earlier
in the `demo.conf` file.

Edit the file `~/.kube/config` and change the IP address 127.0.0.1
to the IP address of the edge device.

## Connect to the Microshift cluster
You can verify that the local `~/.kube/config` is correct by issuing
the following command.

    oc get all -A

A listing of all the various resources running on the MicroShift
server should appear.

## Deploy the demo application
Run the following command to deploy all of the resources for the
demo application.

    oc apply -k app

## Browse to the demo application
You'll need to edit your `/etc/hosts` file first on the computer
where you'll be browsing to the web application. Make sure to add
the following entries to `/etc/hosts` that map to the actual IP
address of the edge device.

    192.168.8.60    ads-b-map.local
    192.168.8.60    ads-b-service.local

Then simply browse to the URL http://ads-b-map.local to see the
demo.


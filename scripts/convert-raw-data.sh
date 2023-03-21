#!/usr/bin/env bash

. $(dirname $0)/../demo.conf

[[ $EUID -eq 0 ]] && exit_on_error "Must NOT run as root"

pushd $(dirname $0) &> /dev/null

RAWDATA=$(dirname $0)/../data/raw_capture_data
OUTPUT=$(dirname $0)/../data/ads-b-data.json
SAMPLE=$(dirname $0)/../data/sample-ads-b-data.json

# Create the file header
echo '{"states":[' > $OUTPUT

# Extract all the ads-b position reports from the raw data by doing
# the following:
#
#   Extract each aircraft state as a single line
#   Add a comma to the end of each line
#   Remove any trailing whitespace in the comma separated fields
#   Sort all of the lines to get the unique ones
#   Put the aircraft states in time order
#   Remove the trailing comma from the last line

jq -c .states[] $RAWDATA/*.json | \
    sed 's/\]/],/g' | \
    sed 's/ *\"/\"/g' | \
    sort -u | \
    sort -n -t, -k4 | \
    sed '$ s/.$//' >> $OUTPUT

# Append the correct footer
echo "]}" >> $OUTPUT

# Compress the file
jq -c . $OUTPUT > tmp.out
mv tmp.out $OUTPUT

# Create the smaller sample dataset
echo '{"states":' > $SAMPLE
jq .states[0:100] $OUTPUT >> $SAMPLE
echo "}" >> $SAMPLE
jq -c . $SAMPLE > tmp.out
mv tmp.out $SAMPLE

popd &> /dev/null


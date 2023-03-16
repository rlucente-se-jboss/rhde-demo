#!/usr/bin/env bash

RAWDATA=$(dirname $0)/../data/raw_capture_data
OUTDATA=$(dirname $0)/../data/ads-b-data.json
SAMPLE=$(dirname $0)/../data/sample-ads-b-data.json

# Create the file header
echo '{"states":[' > $OUTDATA

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
    sed '$ s/.$//' >> $OUTDATA

# Append the correct footer
echo "]}" >> $OUTDATA

# Compress the file
jq -c . $OUTDATA > tmp.out
mv tmp.out $OUTDATA

# Create the smaller sample dataset
echo '{"states":' > $SAMPLE
jq .states[0:20] $OUTDATA >> $SAMPLE
echo "}" >> $SAMPLE
jq -c . $SAMPLE > tmp.out
mv tmp.out $SAMPLE


#!/usr/bin/env bash

RAWDATA=$(dirname $0)/../data/raw_capture_data
OUTDATA=$(dirname $0)/../data/ads-b-data.json
SAMPLE=$(dirname $0)/../data/sample-ads-b-data.json

# create the file header
echo '{"states":[' > $OUTDATA

# extract all the ads-b position reports from the raw data
jq .states[] $RAWDATA/*.json | \
    sed 's/\]/],/g' >> $OUTDATA

# remove the last unnecessary comma from the output
sed -i.bak '$ s/.$//' $OUTDATA
rm -f $OUTDATA.bak

# append the correct footer
echo "]}" >> $OUTDATA

# compress the file
jq -c . $OUTDATA > tmp.out
mv tmp.out $OUTDATA

# create the smaller sample dataset
echo '{"states":' > $SAMPLE
jq .states[0:20] $OUTDATA >> $SAMPLE
echo "}" >> $SAMPLE
jq -c . $SAMPLE > tmp.out
mv tmp.out $SAMPLE


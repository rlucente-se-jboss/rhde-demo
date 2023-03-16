#!/usr/bin/env bash

while true
do
    curl -s http://localhost:8080/ads-b-states | jq .
    echo
    sleep 1
done


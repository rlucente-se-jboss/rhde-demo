#!/usr/bin/env bash

. $(dirname $0)/../demo.conf

[[ $EUID -eq 0 ]] && exit_on_error "Must NOT run as root"

while true
do
    curl -vs http://localhost:${WS_PORT}/ads-b-states | jq .
    echo
    sleep 1
done


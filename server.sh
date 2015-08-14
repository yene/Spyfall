#!/bin/bash

set -e # stop on error
reset
killall spyfall
go build
./spyfall &
echo "server is ready"

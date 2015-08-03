#!/bin/bash
reset
killall spyfall
go build
./spyfall &
echo "server is ready"

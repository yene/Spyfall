#!/bin/bash
reset
echo "building and restarting"
killall spyfall
go build
./spyfall &

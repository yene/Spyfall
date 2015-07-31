#!/bin/bash
echo "building and restarting"
killall spyfall
go build
./spyfall &

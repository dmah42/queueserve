#!/bin/bash

export GOPATH=`pwd`

echo "building and testing qserver"
go build qserver
go test qserver

echo "running qserver"
./qserver &
PID=$!

echo "installing and testing qclient"
go install qclient
go test qclient

echo "killing qserver"
kill $PID

echo "done"

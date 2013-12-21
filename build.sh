#!/bin/bash

export GOPATH=`pwd`

echo "building and testing qserver"
go build qserver
go test qserver

echo "running qserver"
./qserver --port=4242 &
PID=$!

echo "installing and testing qclient"
go install qclient
go test qclient --port=4242 --host=localhost

echo "killing qserver"
kill $PID

echo "done"

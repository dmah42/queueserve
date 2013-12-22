#!/bin/bash

export GOPATH=`pwd`

./build.sh

echo "testing qserver"
go test qserver

echo "running qserver"
./qserver --port=4242 &
PID=$!

echo "testing qclient"
go test qclient --port=4242 --host=localhost

echo "killing qserver"
kill $PID

echo "done"

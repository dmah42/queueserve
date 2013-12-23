#!/bin/bash

export GOPATH=`pwd`

./build.sh

echo "testing qserver"
go test qserver -v

echo "running qserver"
./qserver --port=4242 &
PID=$!

echo "testing qclient"
go test qclient --port=4242 --host=localhost -v

echo "killing qserver"
kill $PID

echo "done"

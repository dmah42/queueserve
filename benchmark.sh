#!/bin/bash

export GOPATH=`pwd`

./build.sh

echo "qserver"
go test qserver --bench=. 

./qserver --port=4242 &
PID=$!

echo "client"
go test qclient --port=4242 --host=localhost --bench=. 

echo "killing qserver"
kill $PID

echo "done"

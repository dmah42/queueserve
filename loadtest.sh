#!/bin/bash

export GOPATH=`pwd`

go install qcommon

echo "qserver"
go build qserver
go test qserver --bench=. 

./qserver --port=4242 &
PID=$!

# echo "one client"
# go install qclient
# go test qclient --port=4242 --host=localhost --bench=. 
# 
# echo "multiple clients"
# for i in {0..$1}; do
#   go test qclient --port=4242 --host=localhost --bench=.
# done

echo "killing qserver"
kill $PID

echo "done"

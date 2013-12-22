#!/bin/bash

export GOPATH=`pwd`

./build.sh

echo "qserver"
# BenchmarkRead runs too long if --benchtime > ~0.2
go test qserver --bench=. --benchtime=0.2

./qserver --port=4242 &
PID=$!

echo "qclient"
go test qclient --port=4242 --host=localhost --bench=. --benchtime=0.5

echo "killing qserver"
kill $PID

echo "done"

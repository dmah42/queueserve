#!/bin/bash

export GOPATH=`pwd`

./build.sh

HOST=$1
PORT=$2
CLIENT_COUNT=$3
OP_COUNT=$4

# defaults
if [ -z "$HOST" ]; then
  echo "need to specify host" && exit
fi

if [ -z "$PORT" ]; then
  PORT=4242
fi

if [ -z "$CLIENT_COUNT" ]; then
  CLIENT_COUNT=10
fi

if [ -z "$OP_COUNT" ]; then
  OP_COUNT=100
fi

echo "creating queue"
STATUS=$(curl --data "name=q" -o /dev/null -s -w '%{http_code}' http://$HOST:$PORT/create)
if [ "200" != "$STATUS" ]; then
  echo "failed to create queue: $STATUS" && exit
fi

echo "starting $CLIENT_COUNT test clients"
for i in {0..$CLIENT_COUNT}; do
  ./testqclient --host=$HOST --port=$PORT --count=$OP_COUNT &
done

echo "waiting for clients to finish"
wait $(jobs -p)

echo "deleting queue"
STATUS=$(curl --data "id=q" -o /dev/null -s -w '%{http_code}' http://$HOST:$PORT/delete)
if [ "200" != "$STATUS" ]; then
  echo "failed to delete queue: $STATUS" && exit
fi

echo "done"


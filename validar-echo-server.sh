#!/bin/bash

MESSAGE="Hello World!"
SERVER_PORT=12345

docker network inspect validate_network >/dev/null 2>&1 || docker network create validate_network

if ! docker ps -q -f name=server >/dev/null; then
    docker run -d --rm --name server --network "$NETWORK_NAME" server:latest
fi

sleep 2

RESPONSE=$(echo "$MESSAGE" | docker run --rm --network validate_network busybox nc server "$SERVER_PORT")

if [ "$RESPONSE" = "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi

#!/bin/bash

SERVER_CONTAINER="server"
MESSAGE="Hello World!"
SERVER_PORT=12345

docker network inspect validate_network >/dev/null 2>&1 || docker network create validate_network

# docker run -d --rm --name server --network validate_network server:latest

sleep 2

RESPONSE=$(echo "$MESSAGE" | docker run --rm --network validate_network busybox nc server "$SERVER_PORT")

if [ "$RESPONSE" = "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi

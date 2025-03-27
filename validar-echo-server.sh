#!/bin/bash

MESSAGE="Mensaje de prueba"
SERVER_PORT=12345
SERVER_NAME="server"
NETWORKNAME=$(docker network ls --format '{{.Name}}' | grep "testing_net" | head -n 1)

RESPONSE=$(docker run --rm --network "$NETWORKNAME" busybox sh -c "echo '$MESSAGE' | nc $SERVER_NAME $SERVER_PORT")

if [ "$RESPONSE" = "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
else
    echo "action: test_echo_server | result: fail"
fi

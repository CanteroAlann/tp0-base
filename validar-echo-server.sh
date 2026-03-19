#!/bin/bash

SERVER_TARGET="server" 
SERVER_PORT="12345"
NETWORK_NAME="testing-net"
MESSAGE="testing_echo_service"

RESPONSE=$(echo "$MESSAGE" | docker run --rm --network "$NETWORK_NAME" busybox nc -w 2 "$SERVER_TARGET" "$SERVER_PORT" 2>/dev/null)

if [ "$RESPONSE" = "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
    exit 0
else
    echo "action: test_echo_server | result: fail"
    exit 1
fi
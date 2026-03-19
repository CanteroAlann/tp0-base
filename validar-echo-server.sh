#!/bin/bash

SERVER_IP="172.25.125.2"
SERVER_PORT="12345"
MESSAGE="testing_echo_service"


RESPONSE=$(echo "$MESSAGE" | nc -w 2 "$SERVER_IP" "$SERVER_PORT" 2>/dev/null)

if [ "$RESPONSE" = "$MESSAGE" ]; then
    echo "action: test_echo_server | result: success"
    exit 0
else
    echo "action: test_echo_server | result: fail"
    exit 1
fi
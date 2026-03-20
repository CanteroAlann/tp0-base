#!/bin/bash

SERVER_TARGET="server" 
SERVER_PORT="12345"
NETWORK_NAME=tp0_testing-net
MESSAGE="testing_echo_service"
MAX_RETRIES=5
RETRY_COUNT=0

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
   
    RESPONSE=$(echo "$MESSAGE" | docker run -i --rm --network "$NETWORK_NAME" busybox nc -w 2 "$SERVER_TARGET" "$SERVER_PORT" 2>/dev/null)

    if [ "$RESPONSE" = "$MESSAGE" ]; then
        echo "action: test_echo_server | result: success"
        exit 0
    fi

    RETRY_COUNT=$((RETRY_COUNT + 1))
    sleep 1
done

echo "action: test_echo_server | result: fail"
exit 1
#!/bin/sh
FILE_TO_CHECK="/tmp/telegraf.conf"

if [ -f "$FILE_TO_CHECK" ]; then
    if nc -zv localhost 9273; then
        echo "Connection successful"
    else
        exit 1
    fi
else
    echo "No active redis instances"
fi
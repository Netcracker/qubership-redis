#!/bin/sh
set -e

SECRETS_DIR=/var/run/secrets/redis

if [ -f "${SECRETS_DIR}/username" ]; then
    export REDIS_DBAAS_USER=$(cat "${SECRETS_DIR}/username")
fi
if [ -f "${SECRETS_DIR}/password" ]; then
    export REDIS_DBAAS_PASSWORD=$(cat "${SECRETS_DIR}/password")
fi
if [ -f "${SECRETS_DIR}/redis-password" ]; then
    export REDIS_PASSWORD=$(cat "${SECRETS_DIR}/redis-password")
elif [ -z "${REDIS_PASSWORD}" ]; then
    export REDIS_PASSWORD=redis
fi

exec /docker-entrypoint.sh "$@"

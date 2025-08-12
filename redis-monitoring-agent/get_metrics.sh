#!/bin/bash
DB=$1
LATENCY_CHECK_SECONDS=$2
POD_LABEL=$3
POD_PORT=$4
POD_PASS=$5
TLS=$6
#TS=$(date +%s)

REDIS_COMMAND="redis-cli -h $POD_LABEL -p $POD_PORT -a ${POD_PASS}"

if [[ "$TLS" == "true" ]]; then
    REDIS_COMMAND="$REDIS_COMMAND --tls --cacert $TLS_ROOTCERT"
fi

AVG_LATENCY="$(timeout -s 9 -k $LATENCY_CHECK_SECONDS $LATENCY_CHECK_SECONDS $REDIS_COMMAND --latency | tail -n 1 | awk '{print $3}')"

SET_TEST="$($REDIS_COMMAND set monitoring_test 1 | tail -n 1)"
if [[ "$SET_TEST" == "OK" ]]
then
    SET_TEST=1
else
    SET_TEST=0
fi

GET_TEST="$($REDIS_COMMAND get monitoring_test | tail -n 1)"
if [[ $GET_TEST != 1 ]]
then
    GET_TEST=0
fi

DEL_TEST="$($REDIS_COMMAND del monitoring_test | tail -n 1 | awk '{print $1}')"
if [[ $DEL_TEST != 1 ]]
then
    DEL_TEST=0
fi

echo "redis,host=$DB,port=$POD_PORT,server=$POD_LABEL,replication_role=master avg_latency=${AVG_LATENCY:=-1},set_op=${SET_TEST:=0},get_op=${GET_TEST:=0},del_op=${DEL_TEST:=0}"

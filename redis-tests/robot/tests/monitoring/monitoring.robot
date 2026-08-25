*** Settings ***
Resource  ../shared/keywords.robot
Suite Setup  Prepare Heartbeat Connection


*** Keywords ***
Prepare Heartbeat Connection
    ${REDIS_HEARTBEAT_DATA}=  Generate Random Data
    Set Suite Variable  ${REDIS_HEARTBEAT_DATA}
    ${REDIS_HOST}=  Set Variable  redis.${REDIS_NAMESPACE}
    Set Suite Variable  ${REDIS_HOST}


*** Test Cases ***
Test Connect To RedisDB For Heartbeat
    [Tags]  redis  monitoring-heartbeat
    Skip If  'monitoring-heartbeat' not in '${TAGS}'  This test only runs when explicitly enabled via the monitoring-heartbeat tag.
    ${REDIS_CONN}=  Connect To RedisDB  ${REDIS_HOST}  ${REDIS_PORT}  0  ${REDIS_PASSWORD}  ${REDIS_TLS_ENABLED}  ${REDIS_TLS_ROOTCERT}
    Set Suite Variable  ${REDIS_CONN}

Test Create Dashboard Heartbeat Key
    [Tags]  redis  monitoring-heartbeat
    Skip If  'monitoring-heartbeat' not in '${TAGS}'  This test only runs when explicitly enabled via the monitoring-heartbeat tag.
    Set Data To Redis  ${REDIS_CONN}  dashboard-check:heartbeat  ${REDIS_HEARTBEAT_DATA}

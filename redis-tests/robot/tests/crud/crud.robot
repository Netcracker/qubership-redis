*** Variables ***

*** Settings ***
Library  String
Library    OperatingSystem
#Library  RedisLibrary
Resource  ../shared/keywords.robot
Suite Setup  Preparation


*** Keywords ***
Create Dbaassession With TLS
    [Arguments]  ${host}  ${port}  ${auth}  ${verify}
    Create Session    dbaassession    https://${host}:${port}  auth=${auth}   verify=${verify}

Create Dbaassession Without TLS
    [Arguments]  ${host}  ${port}  ${auth}
    Create Session    dbaassession    http://${host}:${port}  auth=${auth}

Preparation
    ${REDIS_KEY}=  Generate Random Data
    Set Suite Variable  ${REDIS_KEY}
    ${REDIS_TEST_DATA}=  Generate Random Data
    Set Suite Variable  ${REDIS_TEST_DATA}
    ${headers}  Create Dictionary  Content-Type=application/json  Accept=application/json
    Set Suite Variable   ${headers}
    ${auth}=  Create List  ${REDIS_DBAAS_USER}  ${REDIS_DBAAS_PASSWORD}
    ${verify}=    Get Environment Variable    name=TLS_ROOTCERT    default=False
    ${dbass_tls}=    Get Environment Variable    name=TLS_ENABLED    default=False
    ${REDIS_DBAAS_ADAPTER_PORT}=    Get Environment Variable    name=REDIS_DBAAS_ADAPTER_PORT    default=8080
    ${https_aggregator_enabled}=  Evaluate  "https" in "${DBAAS_AGGREGATOR_REGISTRATION_ADDRESS}"
    Run Keyword If  '${dbass_tls}' == 'true' and '${https_aggregator_enabled}' == 'True'
    ...  Create Dbaassession With TLS  ${REDIS_DBAAS_ADAPTER_HOST}  ${REDIS_DBAAS_ADAPTER_PORT}  auth=${auth}  verify=${verify}
    ...  ELSE  Create Dbaassession Without TLS  ${REDIS_DBAAS_ADAPTER_HOST}  ${REDIS_DBAAS_ADAPTER_PORT}  auth=${auth}
    Run Keyword If  '${DBAAS_ENABLED}' == 'false'
    ...  Set Suite Variable  ${REDIS_HOST}  redis.${REDIS_NAMESPACE}

*** Test Cases ***
Test Creating DB Via Dbaas Adapter
    [Tags]  redis  smoke  dbaas
    Skip If  '${DBAAS_ENABLED}' == 'false'  Redis Dbaas is not enabled. Skip test.
    ${REDIS_HOST}=  Create DB Via Dbaas Adapter
    Set Suite Variable  ${REDIS_HOST}
    Sleep  20s

Test Connect To RedisDB
    [Tags]  redis  smoke  crud
    ${REDIS_CONN}=  Connect To RedisDB  ${REDIS_HOST}  ${REDIS_PORT}  0  ${REDIS_PASSWORD}  ${REDIS_TLS_ENABLED}  ${REDIS_TLS_ROOTCERT} 
    Set Suite Variable  ${REDIS_CONN}

Test Add Data
    [Tags]  redis  smoke  crud
    Add Data To Redis  ${REDIS_CONN}  ${REDIS_KEY}  ${REDIS_TEST_DATA}

Test Read Data
    [Tags]  redis  smoke  crud
    Get Data From Keyspace  ${REDIS_CONN}  ${REDIS_KEY}

Test Update Data
    [Tags]  redis  smoke  crud
    Update Data In Redis  ${REDIS_CONN}  ${REDIS_KEY}

Test Delete Data Drom Redis
    [Tags]  redis  smoke  crud
    Delete Data From Keyspace  ${REDIS_CONN}  ${REDIS_KEY}

Test Get DB Via Dbaas Adapter
    [Tags]  redis  smoke  dbaas
    Skip If  '${DBAAS_ENABLED}' == 'false'  Redis Dbaas is not enabled. Skip test.
    Get DB Via Dbaas Adapter  ${REDIS_HOST}

Test Delete DB Via Dbaas Adapter
    [Tags]  redis  smoke  dbaas
    Skip If  '${DBAAS_ENABLED}' == 'false'  Redis Dbaas is not enabled. Skip test.
    Delete DB Via Dbaas Adapter  ${REDIS_HOST}
    ${all_redis_deployments}=  Get Deployment Entity Names For Service  ${REDIS_NAMESPACE}  ${REDIS_HOST}  label_name=name
    Sleep  10s
    Should Not Contain	${all_redis_deployments}  ${REDIS_HOST}

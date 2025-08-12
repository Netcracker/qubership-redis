*** Settings ***
Library     String
Library     Collections
Library     RequestsLibrary
Library     ../lib/RedisLibraryKeywords.py
Library     PlatformLibrary    managed_by_operator=true


*** Variables ***
${REDIS_NAMESPACE}              %{OPENSHIFT_WORKSPACE_WA}
${REDIS_PORT}                   %{REDIS_PORT}
${REDIS_PASSWORD}               %{REDIS_PASSWORD}
${REDIS_DBAAS_ADAPTER_HOST}     %{REDIS_DBAAS_ADAPTER_HOST}
${REDIS_DBAAS_ADAPTER_PORT}     %{REDIS_DBAAS_ADAPTER_PORT}
${REDIS_DBAAS_USER}             %{REDIS_DBAAS_USER=dbaas-aggregator}
${REDIS_DBAAS_PASSWORD}         %{REDIS_DBAAS_PASSWORD=dbaas-aggregator}
${DBAAS_AGGREGATOR_REGISTRATION_ADDRESS}    %{DBAAS_AGGREGATOR_REGISTRATION_ADDRESS}
${DBAAS_ENABLED}                %{DBAAS_ENABLED=true}

${REDIS_TLS_ENABLED}            %{TLS_ENABLED=false}
${REDIS_TLS_ROOTCERT}           %{TLS_ROOTCERT=/usr/ssl/ca.crt}


*** Keywords ***
Generate Random Data
    ${random_string}=    Generate Random String    10    [LOWER]
    RETURN    ${random_string}

Get Dbaas Aggregator version 
    ${apiVersion}=    Set Variable    v2
    RETURN    ${apiVersion}

Create DB Via Dbaas Adapter
    ${dbaas_api_version}=    Get Dbaas Aggregator version
    Set Suite Variable  ${dbaas_api_version}
    ${redis_db_name}=    Generate Random Data
    ${data}=    Set Variable
    ...    { "metadata": { "classifier": { "microserviceName": "Service-test", "isServiceDb": true } }, "settings": { "redisDbSettings": { "maxmemory":"150mb" }, "redisDbResources": {"requests": {"memory": "150Mi"}, "limits": {"memory": "200Mi"}}, "redisDbWaitStartServiceSecond": 111}, "dbName":"${redis_db_name}", "password":"${REDIS_PASSWORD}", "namePrefix":"test" }
    ${resp}=    POST On Session
    ...    dbaassession
    ...    url=/api/${dbaas_api_version}/dbaas/adapter/redis/databases
    ...    data=${data}
    ...    headers=${headers}
    Sleep    10s
    Should Be Equal As Strings    ${resp.status_code}    201
    ${deployment_name}=    Get Deployment Entity Names For Service
    ...    ${REDIS_NAMESPACE}
    ...    test-${redis_db_name}
    ...    label=name
    List Should Contain Value    ${deployment_name}    test-${redis_db_name}
    ${redis_db_host}=    Catenate    SEPARATOR=    test-${redis_db_name}
    RETURN    ${redis_db_host}

Connect To RedisDB
    [Arguments]    ${REDIS_HOST}    ${REDIS_PORT}    ${KEYSPACE_NUMBER}    ${REDIS_PASSWORD}    ${REDIS_TLS_ENABLED}    ${REDIS_TLS_ROOTCERT}
    ${redis_conn}=    Connect To Redis
    ...    ${REDIS_HOST}
    ...    ${REDIS_PORT}
    ...    ${KEYSPACE_NUMBER}
    ...    ${REDIS_PASSWORD}
    ...    ${REDIS_TLS_ENABLED}
    ...    ${REDIS_TLS_ROOTCERT}
    Set Suite Variable    ${redis_conn}
    RETURN    ${redis_conn}

Add Data To Redis
    [Arguments]    ${redis_conn}    ${REDIS_KEY}    ${REDIS_TEST_DATA}
    ${code}=    Append To Redis    ${redis_conn}    ${REDIS_KEY}    ${REDIS_TEST_DATA}
    Should Be Equal As Strings    ${code}    10

Get Data From Keyspace
    [Arguments]    ${redis_conn}    ${REDIS_KEY}
    ${code}=    Get From Redis    ${redis_conn}    ${REDIS_KEY}
    Should Be Equal As Strings    ${code}    ${REDIS_TEST_DATA}

Update Data In Redis
    [Arguments]    ${redis_conn}    ${REDIS_KEY}
    ${updated_data}=    Generate Random Data
    ${data}=    Set To Redis    ${redis_conn}    ${REDIS_KEY}    ${updated_data}
    Should Be Equal As Strings    ${data}    True

Delete Data From Keyspace
    [Arguments]    ${redis_conn}    ${REDIS_KEY}
    ${code}=    Delete From Redis    ${redis_conn}    ${REDIS_KEY}
    Should Be Equal As Strings    ${code}    1

Get DB Via Dbaas Adapter
    [Arguments]    ${redis_host}
    ${resp}=    GET On Session    dbaassession    url=/api/${dbaas_api_version}/dbaas/adapter/redis/databases
    Should Be Equal As Strings    ${resp.status_code}    200
    Should Contain    str(${resp.content})    ${redis_host}

Delete DB Via Dbaas Adapter
    [Arguments]    ${redis_host}
    ${data}=    Set Variable
    ...    [{"kind":"Deployment","name":"${redis_host}"},{"kind":"Service","name":"${redis_host}"},{"kind":"ConfigMap","name":"${redis_host}"},{"kind":"Secret","name":"${redis_host}-credentials"}]
    ${resp}=    POST On Session
    ...    dbaassession
    ...    url=/api/${dbaas_api_version}/dbaas/adapter/redis/resources/bulk-drop
    ...    data=${data}
    ...    headers=${headers}
    Should Be Equal As Strings    ${resp.status_code}    200

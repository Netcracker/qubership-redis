This section describes using robotframework integration tests

- Run Tests in Openshift.
- Run Tests Locally.


### Running Tests Locally

For running robotframework integration tests locally you should start one of the robot's files.

Prerequisite:

- installed python v.3.6, robotframework, redis python driver

- exported necessary variables to env. for tests that you will run.

for running : robot ./integration-tests/robot that command runs robotframework tests.

| Variable name             | Value example                                                                                                         | Description                                         |
|---------------------------|-----------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------|
| CLOUD_URL                 | cloud.openshift.com                                                                           | Cloud URL without port and protocol                 |
| OPENSHIFT_WORKSPACE       | redis                                                                                                                 | Project name where redis is installed               |
| REDIS_DBAAS_ADAPTER_HOST  | dbaas-redis-adapter.redis                                                                                             | Host to dbaas-redis-adapter                         |
| REDIS_DBAAS_ADAPTER_PORT  | 8080                                                                                                                  | Port for dbaas-redis-adapter                        |
| REDIS_DBAAS_USER          | dbaas-aggregator                                                                                                      | Username for dbaas-redis-adapter                    |
| REDIS_DBAAS_PASSWORD      | dbaas-aggregator                                                                                                      | Password for dbaas-redis-adapter                    |
| REDIS_PORT                | 6379                                                                                                                  | Port for redis pod                                  |
| REDIS_PASSWORD            | redis                                                                                                                 | Password for redis                                  |
| OS_TOKEN                  |                                                                                                                       | Token instead of user/password to login to OpenShift|
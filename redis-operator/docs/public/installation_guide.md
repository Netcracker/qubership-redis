This section provides information about the deployment of Redis service.

<!-- #GFCFilterMarkerStart# -->
[[_TOC_]]
<!-- #GFCFilterMarkerEnd# -->

# Prerequisites

The prerequisites described in the following sections should be met before your start deploying Redis service using Helm.

## Common

For information about the hardware prerequisites, refer to [HWE](#hwe).

The prerequisites to deploy DBaaS Redis Operator are as follows:

* The deployer user (SA) must have the following role bound:

  ```
  apiVersion: rbac.authorization.k8s.io/v1
  kind: Role
  metadata:
    name: nc-role
  rules:
    - apiGroups:
        - netcracker.com
      resources:
        - '*'
      verbs:
        - create
        - get
        - list
        - patch
        - update
        - watch
        - delete
  ```

* The project or namespace should be created.
* The cloud administrator should create the Custom Resource Definition (CRD).
* If deployed to OpenShift with restricted SCC, the project supplemental group annotation must have the same UID as in the `podSecurityContext.runAsUser` and `podSecurityContext.fsGroup` parameters.
* If the Pod Security Policy is enabled on the Kubernetes (K8s) cluster, it is mandatory to set the `podSecurityContext.fsGroup` and `podSecurityContext.runAsUser` parameters. For more information, refer to [https://kubernetes.io/docs/concepts/policy/pod-security-policy/](https://kubernetes.io/docs/concepts/policy/pod-security-policy/).
* In case of Prometheus Monitoring stack deployment, you need to have the rights to create the `integreatly.org/v1alpha1` and `monitoring.coreos.com/v1` objects.

The following is an example of such a role:

```
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: generic-monitoring-role
rules:
  - apiGroups:
      - "monitoring.coreos.com"
    resources:
      - servicemonitors
      - prometheusrules
    verbs:
      - get
      - list
      - create
      - update
      - delete
      - watch
      - patch
  - apiGroups:
      - "integreatly.org"
    resources:
      - grafanadashboards
    verbs:
      - get
      - list
      - create
      - update
      - delete
      - watch
      - patch

```

### Apply New Custom Resource Definition Version

#### Automated CRD Upgrade

The upgrade of CRD happens automatically through pre-deploy scripts.

**Note**: Automation CRD upgrade requires the following cluster rights for the deploy user:

```yaml
- apiGroups: ["apiextensions.k8s.io"]
  resources: ["customresourcedefinitions"]
  verbs: ["get", "create", "patch"]
```

To disable this feature, add the `DISABLE_CRD: true` parameter.

#### Manual CRD Upgrade

You can find multiple CRDs in the `charts/helm/redis-operator/` directory:

* `legacy_crd/crd.yaml` - CRD for Kubernetes version below 1.22 and OpenShift below 4.9.
* `crds/k8s_1.22_crd.yaml` - CRD for Kubernetes version 1.22+ and OpenShift 4.9+.

Apply the new version of CRD using the following command:

`kubectl replace -f charts/helm/redis-operator/crds/<crd_name>.yaml`

Specify `--skip-crds` in the `ADDITIONAL_OPTIONS` parameter of the DP Deployer Job.

Specify `DISABLE_CRD=true;` in the `CUSTOM_PARAMS` parameter of the App Deployer Job.


## HWE

**Note**: The provided HWE is for one Redis instance in a namespace.

**Note**: The CPU limit for monitoring must be calculated per Redis instance

Redis services resources can be selected during deployment using parameter `global.profile` that takes values: `small`, `medium`, `large`.

### Small

Recommended for development purposes, PoC, and demos.


| Module                 | CPU Requests, cores | RAM Requests, Mb | CPU Limits, cores | RAM Limits, Mb |
| ---------------------- | ------------------- | ---------------- | ----------------- | -------------- |
| Dbaas Redis Operator   | 0.05                | 64               | 0.1               | 128            |
| Redis Monitoring Agent | 0.05                | 64               | 0.1               | 128            |
| Robot Tests            | 0.2                 | 128              | 0.2               | 256            |
| Redis                  | 0.05                | 64               | 0.25              | 256            |
| **Total**              | **0.35**            | **320**          | **0.65**          | **768**        |

### Medium

Recommended for deployments with average load.

| Module                 | CPU Requests, cores | RAM Requests, Mb | CPU Limits, cores | RAM Limits, Mb |
| ---------------------- | ------------------- | ---------------- | ----------------- | -------------- |
| Dbaas Redis Operator   | 0.05                | 64               | 0.1               | 128            |
| Redis Monitoring Agent | 0.05                | 64               | 0.1               | 128            |
| Robot Tests            | 0.2                 | 128              | 0.2               | 256            |
| Redis                  | 1                   | 2048             | 2                 | 4096           |
| **Total**              | **1.3**             | **2304**         | **2.4**           | **4608**       |

### Large

Recommended for deployments with high workload and large amount of data.

| Module                 | CPU Requests, cores | RAM Requests, Mb | CPU Limits, cores | RAM Limits, Mb |
| ---------------------- | ------------------- | ---------------- | ----------------- | -------------- |
| Dbaas Redis Operator   | 0.05                | 64               | 0.1               | 128            |
| Redis Monitoring Agent | 0.05                | 64               | 0.1               | 128            |
| Robot Tests            | 0.2                 | 128              | 0.2               | 256            |
| Redis                  | 2                   | 4096             | 4                 | 8192           |
| **Total**              | **2.3**             | **4352**         | **4.4**           | **8704**       |

## Parameters

The [values.yaml](charts/helm/redis-operator/values.yaml) file contains the description and example of each parameter and its default value. Some parameters are self-explanatory.

The following sections provide the list of parameters.

There are no mandatory parameters. By default, Redis with DBaaS adapter is installed.

### Operator Parameters

The operator parameters are specified below.

| Parameter                                      | Mandatory | Type       | Default | Description                                                                                                           |
| ---------------------------------------------- | --------- | ---------- | ------- | --------------------------------------------------------------------------------------------------------------------- |
| `operator.resources.requests.cpu`              | false     | string/int | 50m     | The CPU requests for the operator.                                                                                    |
| `operator.resources.requests.memory`           | false     | string/int | 64mi    | The RAM requests for the operator.                                                                                    |
| `operator.resources.limits.cpu`                | false     | string/int | 100m    | The CPU limits for the operator.                                                                                      |
| `operator.resources.limits.memory`             | false     | string/int | 128mi   | The RAM limits for the operator.                                                                                      |
| `securityContext.fsGroup`                      | false     | int        | 1001    | The fsGroup of all containers.                                                                                        |
| `securityContext.runAsUser`                    | false     | string     | 1001    | The user to run all container under.                                                                                  |
| `securityContext.supplementalGroups`           | false     | array      | ""      | The supplementalGroups of all containers.                                                                             |
| `policies.tolerations[$idx].key`               | false     | string     | ""      | The taint key the toleration applies to.                                                                              |
| `policies.tolerations[$idx].operator`          | false     | string     | ""      | The key relationship to the value.                                                                                    |
| `policies.tolerations[$idx].value`             | false     | int        | ""      | The taint value the toleration matches to.                                                                            |
| `policies.tolerations[$idx].effect`            | false     | string     | ""      | The taint effect to the match.                                                                                        |
| `policies.tolerations[$idx].tolerationSeconds` | false     | int        | ""      | The period the toleration (which must be of effect `NoExecute`, otherwise this field is ignored) tolerates the taint. |

### DBaaS Redis Adapter Parameters

The list of DBaaS Redis Adapter parameters is as follows:

| Parameter                                             | Mandatory | Type   | Default                            | Description                                                                              |
| ----------------------------------------------------- | --------- | ------ | ---------------------------------- | ---------------------------------------------------------------------------------------- |
| `dbaas.install`                                       | false     | bool   | true                               | If the DBaaS adapter needs to be installed.                                              |
| `dbaas.aggregator.physicalDatabaseIdentifier`         | false     | string | redis                              | The database identifier in the DBaaS aggregator.                                         |
| `dbaas.aggregator.physicalDatabaseLabels`             | false     | string | ""                                 | The database labels in the DBaaS aggregator.                                             |
| `dbaas.aggregator.dbaasAggregatorRegistrationAddress` | false     | string | http://dbaas-aggregator.dbaas:8080 | The address of the DBaaS aggregator.                                                     |
| `dbaas.aggregator.address`                            | false     | string | http://dbaas-aggregator.dbaas:8080 | The address of the aggregator where the adapter registers its physical database cluster. |
| `dbaas.aggregator.username`                           | false     | string | cluster-dba                        | The username for DBaaS adapter registration in DBaaS.                                    |
| `dbaas.aggregator.password`                           | false     | string | Bnmq5567_PO                        | The password for DBaaS adapter registration in DBaaS.                                    |
| `dbaas.aggregator.secretName`                         | false     | string | dbaas-aggregator-credentials       | The name of the secret that holds Aggregator credentials.                                |
| `dbaas.adapter.username`                              | false     | string | dbaas-aggregator                   | The username for the database adapter.                                                   |
| `dbaas.adapter.password`                              | false     | string | dbaas-aggregator                   | The password for the database adapter.                                                   |
| `dbaas.adapter.secretName`                            | false     | string | dbaas-adapter-credentials          | The secret name of the adapter credentials.                                              |

### Redis Parameters

The list of Redis parameters is specified below.

| Parameter                                   | Mandatory | Type              | Default | Description                                                                                          |
| ------------------------------------------- | --------- | ----------------- | ------- | ---------------------------------------------------------------------------------------------------- |
| `redis.conf`                                | false     | map[string]string |         | The map of the Redis configuration.                                                                  |
| `redis.flavor`                              | false     | string            | small   | The flavor of redis deployment resources. Possible values are  `small`, `medium`, `large`.           |
| `redis.tls.enabled`                         | false     | bool              | false   | If TLS needs to be enabled.                                                                          |
| `redis.tls.tlsPort`                         | false     | int               | 6379    | The port for TLS connections.                                                                        |
| `redis.tls.generateCerts.enabled`           | false     | bool              | false   | If an integration with Cert Manager needs to be enabled.                                             |
| `redis.tls.generateCerts.clusterIssuerName` | false     | string            | ""      | The name of ClusterIssuer to integrate with Cert Manager.                                            |
| `redis.tls.generateCerts.duration`          | false     | string            | 365     | The certificate validity period.                                                                     |
| `redis.tls.rootCAFileName`                  | false     | string            | ca.crt  | The key in the Kubernetes secret `tls.rootCASecretName` that holds the CA certificate.               |
| `redis.tls.privateKeyFileName`              | false     | string            | tls.key | The key in the Kubernetes secret `tls.rootCASecretName` that holds the private key.                  |
| `redis.tls.signedCRTFileName`               | false     | string            | tls.crt | The key in the Kubernetes secret `tls.rootCASecretName` that holds the Signed Redis certificate. |
| `redis.tls.certificateSecretName`           | false     | string            | root-ca | The name of the secret that holds a certificate.                                                     |
| `redis.maxmem`                              | false     | string            |         | The memory limit of Redis.                                                                           |
| `redis.password`                            | false     | string            | redis   | The password of Redis.                                                                               |
| `redis.dockerImage`                         | false     | string            | ""      | The Docker image of Redis.                                                                           |
| `redis.nodeLabels`                          | false     | string            | ""      | The additional node labels for the Redis replica.                                                    |
| `redis.resources.requests.cpu`              | false     | int               | 150m    | The CPU request of the Redis replica. Ignored if `redis.flavor`` specified.                          |
| `redis.resources.requests.memory`           | false     | int               | 120Mi   | The memory request of the Redis replica. Ignored if `redis.flavor`` specified.                       |
| `redis.resources.limits.cpu`                | false     | int               | 250m    | The CPU limit of the Redis replica. Ignored if `redis.flavor`` specified.                            |
| `redis.resources.limits.memory`             | false     | int               | 250Mi   | The memory limit of the Redis replica. Ignored if `redis.flavor`` specified.                         |


To override the default Redis parameters, use the following command:

```
redis:
  conf:
    maxclients: 30000
```

### Monitoring Agent Parameters

The list of Monitoring Agent parameters is specified below.

| Parameter                                                | Mandatory | Type   | Default | Description                                                                                                            |
| -------------------------------------------------------- | --------- | ------ | ------- | ---------------------------------------------------------------------------------------------------------------------- |
| `monitoringAgent.install`                                | false     | bool   | false   | If the Monitoring agent needs to be installed.                                                                         |
| `monitoringAgent.dockerImage`                            | false     | string | ""      | The Docker image of the Monitoring agent.                                                                              |
| `monitoringAgent.nodeLabels`                             | false     | string | ""      | The labels on the node.                                                                                                |
| `monitoringAgent.resources.limits.memory`                | false     | int    | 64Mi    | The maximum amount of memory for the Monitoring agent replica.                                                         |
| `monitoringAgent.resources.limits.cpu`                   | false     | int    | 50m     | The maximum number of CPUs for the Monitoring agent replica.                                                           |
| `monitoringAgent.resources.requests.memory`              | false     | int    | 128Mi   | The minimum amount of memory for the Monitoring agent replica.                                                         |
| `monitoringAgent.resources.requests.cpu`                 | false     | int    | 100m    | The minimum number of CPUs for the Monitoring agent replica.                                                           |
| `monitoringAgent.monitoringInterval`                     | false     | int    | 202     | The monitoring interval in seconds.                                                                                    |
| `monitoringAgent.prometheus.alerts.cpuThreshold`         | false     | int    | 95      | The threshold for the CPU usage in percentage at which an alert will be raised.                                            |
| `monitoringAgent.prometheus.alerts.memThreshold`         | false     | int    | 95      | The threshold for the RAM usage in percentage at which an alert will be raised.                                            |
| `monitoringAgent.prometheus.alerts.latencyThresholdMs`   | false     | int    | 20      | The threshold latency in ms at which an alert will be raised.                                                          |
| `monitoringAgent.prometheus.alerts.connectionsThreshold` | false     | int    | 90      | The threshold for the ratio of the current connections to redis_maxclients in percentage at which an alert will be raised. |


### Robot Tests Parameters

The list of Robot Tests parameters is specified below.

| Parameter                              | Mandatory | Type   | Default | Description                                                                          |
| -------------------------------------- | --------- | ------ | ------- | ------------------------------------------------------------------------------------ |
| `robotTests.install`                   | false     | bool   | false   | If Robot tests need to be installed.                                                |
| `robotTests.tags`                      | false     | string | ""      | The tags of Robot tests. The possible values are "redis", "smoke", and "dbaas crud". |
| `robotTests.nodeLabels`                | false     | string | ""      | The additional node labels for the Robot tests' replica.                             |
| `robotTests.dockerImage`               | false     | string | ""      | The Docker image of Robot tests.                                                     |
| `robotTests.resources.requests.cpu`    | false     | int    | 200m    | The CPU request of the Robot tests' replica.                                         |
| `robotTests.resources.requests.memory` | false     | int    | 128Mi   | The RAM request of the Robot tests' replica.                                         |
| `robotTests.resources.limits.cpu`      | false     | int    | 200m    | The CPU limit of the Robot tests' replica.                                           |
| `robotTests.resources.limits.memory`   | false     | int    | 256Mi   | The RAM limit of the Robot tests' replica.                                           |

### Parameters Examples

The following sections provide parameter examples for different scenarios of deployment.

##### Example With all Parameters

The following is an example with all parameters:

```
operator:
  dockerImage: ghcr.io/netcracker/qubership-redis-operator:main
  resources:
    requests:
      cpu: 150m
      memory: 120Mi
    limits:
      cpu: 250m
      memory: 250Mi

dbaas:
  install: true

  adapter:
    username: dbaas-aggregator
    password: dbaas-aggregator
  aggregator:
    username: cluster-dba
    password: Bnmq5567_PO
    address: "http://dbaas-aggregator.dbaas:8080"
    physicalDatabaseIdentifier: "redis"
    dbaasAggregatorRegistrationAddress: "http://dbaas-aggregator.dbaas:8080"

redis:
  maxmem: 200mb
  password: redis
  dockerImage: "redis:7.4.2-alpine"
  nodeLabels:
    region: databases
  parameters:
    label: redis-dbaas-adapter
  resources:
    requests:
      cpu: 150m
      memory: 120Mi
    limits:
      cpu: 250m
      memory: 250Mi

monitoringAgent:
  install: true
  dockerImage: ghcr.io/netcracker/qubership-redis-monitoring-agent:feature_initial
  nodeLabels:
    region: databases
  resources:
    requests:
      cpu: 200m
      memory: 128Mi
    limits:
      cpu: 200m
      memory: 256Mi
  monitoringInterval: 20s

robotTests:
  install: false
  dockerImage: ghcr.io/netcracker/qubership-redis-tests:feature_initial
  tags:
  nodeLabels:
    'kubernetes.io/hostname': ci-master-node-1-1
  resources:
    requests:
      cpu: 200m
      memory: 128Mi
    limits:
      cpu: 200m
      memory: 256Mi
```


# Installation

The installation process is described in the below sub-sections.

## Before You Begin

The below sub-sections provide information about the preparations to be done before begining the installation process.

### Helm

Before starting the manual deployment of the Redis service using Helm, ensure you have Helm 3 release.

Alternatively, you can install the operator using the Helm chart from the `charts/helm/redis-operator` folder. Install the Helm CLI on your machine. For more information about Helm v3.0.0, see [https://github.com/helm/helm/releases/tag/v3.0.0](https://github.com/helm/helm/releases/tag/v3.0.0).

To use the CRD-based configuration, specify the **values.yaml** file as shown in the following example:

Also, you have to specify the proper microservice images in the **values.yaml** file. <!-- #GFCFilterMarkerStart# --> The list of microservices can be found in the Microservice versions section of each release on the Releases page at https://github.com/Netcracker/qubership-redis/redis-operator/-/releases.<!-- #GFCFilterMarkerEnd# --> Follow each microservice tag link and find the Artifacts section with the Docker image name.

Also, in the **values.yaml** file in the **dockerImage** parameters you need to change the links to docker images from the current release that needs to be deployed.

1. Clone the project to a local machine using the following command:

   ```git clone git@github.com/Netcracker/qubership-redis/redis-operator.git```

1. Navigate to the Redis operator directory using the following command:

   ```cd redis-operator/charts/helm/redis-operator```

1. Authenticate to the Kubernetes cloud.

1. Deploy the operator using `helm install redis-operator` helm.

## On-Prem

By default the schema with DBaas integration is deployed.

### DBaaS Adapter Scheme

No required parameters for DBaaS Adapter Scheme.

This schema will deploy only `dbaas-redis-operator`, `redis-monitoring-agent` and `robot-tests` deployments.
There won't be any `Redis` deployment. All `Redis` instances are to be created on demand via `DBaas aggreator`.

### Single Redis Scheme

This schema will deploy `dbaas-redis-operator`, `redis-monitoring-agent` and single `redis` deployments.

It will not allow to create `Redis` instances on demand.

An example of single scheme is as follows:

```
dbaas:
  install: false
```

### Deployment with TLS

SSL/TLS is supported by Redis starting with version 6 (2.6.0 redis-operator version).

To enable TLS, set the `redis.tls.enabled` parameter to "true".

TLS port can be set in the `redis.tls.tlsPort` parameter. By default, it is set to "6379".

To enable automatic certificate generation with Cert Manager, set the `redis.tls.generateCerts.enabled` parameter to "true" and specify `ClusterIssuer` name in `redis.tls.generateCerts.clusterIssuerName`.
Note that dbaas-redis-operator requires the following role during the deployment:

```
- apiGroups:
  - cert-manager.io
  resources:
  - '*'
  verbs:
  - create
  - get
  - list
  - update
```

This role must be bound to the deployer service account.

#### Set Certificates Manually

Note: this option is only for deployement without DbaaS integration.
Creating of certificates for each instance of Redis would require CA private key, which cannot be provided for security reasons.

To pass pre-generated certificates as deploy parameters use the following parameters:

`tls.redis.certificates.ca_crt` - a base 64 encoded CA certificate

`tls.redis.certificates.tls_crt` - a base 64 encoded certificate for Redis

`tls.redis.certificates.tls_key` - a base 64 encoded key for Redis

`tls.dbaas.certificates.tls_crt` - a base 64 encoded certificate for Dbaas Adapter, DNS name is `dbaas-redis-adapter.<namespace>.svc`

`tls.dbaas.certificates.tls_key` - a base 64 encoded key for Dbaas Adapter


Example:

```
dbaas:
  install: false
redis:
  tls:
    install: true
    generateCerts:
      enabled: false
    certificates:
      tls_key: "LS0tLS1CRUd...LS0tLQo="
      tls_crt: "LS0tLS1...tLS0tLQo="
      ca_crt: "LS0tLS1CRUdJ...FLS0tLS0K"
```

# Upgrade

This section provides the information about the upgrade procedure from one operator version to another.
The upgrade procedure is identical to a clean installation. The only difference is that DEPLOY_MODE needs to be set to Rolling Update. If needed, change the deployment parameters.

### Prerequisites

Ensure you use the same type of deployer for the previous and current installation.
For example, if App Deployer was used for the previous installation, it should be used for the current installation as well.

#### CRD Upgrade

For information relating to CRD upgrade, refer to the [Apply New Custom Resource Definition Version](#apply-new-custom-resource-definition-version) section.


#### Recommend parameters Value.

`monitoringAgent.resources.limits.cpu` should be set 100m per redis instance. For instance if we need 5 redis instances set it to 500m.
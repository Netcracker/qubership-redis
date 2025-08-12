This chapter describes the architectural features of Redis Operator. 

# Overview

Redis is an open source in-memory data structure store that is used as a database, cache, and message broker. It is trusted by thousands of companies for its high performance, scalability, and versatility.

Redis was designed to meet the needs of modern applications that require fast, real-time data processing. Its design objectives include:

* High performance in-memory data storage and retrieval
* Horizontal scalability through clustering and sharding
* Support for multiple data structures such as strings, hashes, lists, sets, and sorted sets
* Persistence options for data durability
* Built-in support for pub/sub messaging and other advanced features
* Flexible configuration options for tuning performance and behavior

Redis is often used in combination with other databases and caching layers to provide a complete data processing solution for modern applications. Its ease of use, flexibility, and performance makes it a popular choice for developers and operations' teams alike.

## Netcracker Redis Delivery and Features

The Netcracker platform provides Redis deployment to Kubernetes/OpenShift using its own helm chart with an operator and additional features.

The deployment procedure and additional features include the following:

* Support of Netcracker deployment jobs for scheme with DBaaS support. For more information, refer to the [DbaaS Redis Adapter Parameters](/docs/public/installation_guide.md) section. 
* Monitoring integration with Grafana Dashboard and Prometheus Alerts. For more information, refer to the [Redis Operator Monitoring](/docs/public/dashboard_overview.md) chapter in _Cloud Platform Monitoring Guide_.

# Redis Components

The architecture diagram for Redis components is shown below.

![Redis Components](/docs/public/images/architecture.png)

## DBaaS Redis Operator

The DBaaS Redis Operator is a mandatory microservice written with Operator-SDK and designed specifically for Kubernetes environments.
With DBaaS integration, it allows to create a Redis instance by an application request. The operator also takes care of managing supplementary services, ensuring seamless integration and efficient resource utilization. In addition to that, DBaaS Redis Operator also performs an upgrade scenario.

## Redis Instances

Redis is an official docker image distributed by Redis Ltd. Redis pods are created by a request to DBaaS. For more information, refer to [Deployment Scheme with DBaaS](#deployment-scheme-with-dbaas).

## Redis Monitoring Agent

The Redis monitoring agent is a microservice that collects metrics from all Redis instances in a namespace and exports them in the Prometheus format. 
This metrics are later used in Grafana Dashboard and Alert Manager.

## Robot Tests

Robot Tests is a microservice that performs integration testing after all components of Redis deployment are installed.

# Supported Deployment Schemes

The supported deployment schemes are specified below.

## On-Prem

The deployment schemes for On-Prem are specified in the below sub-sections.

### Non-HA Deployment Scheme

This deploys a single Redis instance.

### Deployment Scheme with DBaaS 

This deploys only `dbaas-redis-operator`, `redis-monitoring-agent` and `robot-tests` deployments.    
There won't be any `Redis` deployment. All `Redis` instances are to be created on demand via `Dbaas aggreator`.

![Redis with DbaaS](/docs/public/images/redis_sequence.png)

### DR Deployment Scheme

For Redis instances replication, **Cluster Replicator** must be installed on the environment as prerequisite.

In Redis deployment parameters, section`disasterRecovery` must be configured:
```yaml
disasterRecovery:
  enabled: true
  replicatorNamespace: replicator-dc2
```

Where `replicatorNamespace` - Cluster Replicator namespace. 


### Google Cloud

Not Applicable; the default Deployment Scheme with DBaaS is used for the deployment to Google Cloud.

### AWS

Not Applicable; the default Deployment Scheme with DBaaS is used for the deployment to AWS.

### Azure

Not Applicable; the default Deployment Scheme with DBaaS is used for the deployment to Azure.

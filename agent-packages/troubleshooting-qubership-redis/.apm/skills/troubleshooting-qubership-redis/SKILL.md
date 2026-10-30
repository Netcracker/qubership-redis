---
name: troubleshooting-qubership-redis
description: Diagnose Qubership Redis failures by matching an exact alert name, error string, or symptom to a section of references/troubleshooting.md. Covers the alerts defined in prometheus_alerts.yaml of the redis-operator chart (Redis node is down, Redis CPU Usage, Redis Memory Usage, latency, connections count) and failed robot integration tests (references/robot-tests.md).
---

## Finding the section

Run from the skill directory:

1. `grep -n "^## " references/troubleshooting.md` — lists every issue header with its line number. Issue titles are
   `##` lines.
2. Look the alert name, error string, or symptom up in the tables below. Tables give the exact header text; use it to
   pick the line number from step 1.
3. Every section has the same `###` parts (Description, Alerts, Stack trace(s), How to solve, Recommendations). They
   belong to the `##` section above them; read the whole `##` section, from its line to the next `##` line.
4. To search by literal string instead, run `grep -n -F "<exact string>" references/troubleshooting.md` (for example
   `grep -n -F "max number of clients" references/troubleshooting.md`).
5. Read only that line range with offset/limit.

## Alert name → section

Alert names are the `alert:` values in `redis-operator/charts/helm/redis-operator/templates/prometheus_alerts.yaml`
(`PrometheusRule` `prometheus-redis-rules`). The number in an alert name is a chart value, so match on the wording. The
guide's section titles are not the alert names, so use this table.

| Alert name | Fires when | Where the signal shows | Header in references/troubleshooting.md |
|---|---|---|---|
| `Redis node is down` (severity high) | `redis_avg_latency == -1`: the monitoring agent cannot read metrics from the node | `redis-monitoring-agent` pod logs; status and events of the Redis pod | Redis Node Down Incident Report |
| `Redis CPU Usage` | container CPU above `cpuThreshold` % of its CPU limit (default 95) | CPU metric of the pod named in the alert's `pod` label | High Redis CPU Usage |
| `Redis Memory Usage` | container memory above `memThreshold` % of its memory limit (default 95) | memory metric of the pod named in the alert's `pod` label | High Redis Memory Usage |
| `Latency on a Redis node is more than <N> ms` | `redis_avg_latency` above `latencyThresholdMs` (default 20) | metric `redis_avg_latency` | Latency on a Redis Node is More than N ms |
| `Connections count to a Redis node is more than <N>% of the limit` | `redis_clients` / `redis_maxclients` above `connectionsThreshold` % (default 90) | metrics `redis_clients` and `redis_maxclients` | Connections Count to a Redis Node is More Than N% of the Limit |

The `Redis CPU Usage` and `Redis Memory Usage` expressions match every pod in the release namespace, not only Redis
pods; the `pod` label of the firing alert identifies which pod crossed the threshold.

## Error string / state → section

| Exact string or state | Header in references/troubleshooting.md |
|---|---|
| `ERR max number of clients reached` returned to clients | Connections Count to a Redis Node is More Than N% of the Limit |
| `OOM command not allowed when used memory > 'maxmemory'.` returned to clients | High Redis Memory Usage |
| Redis pod last state reason `OOMKilled` | High Redis Memory Usage |
| `robot-tests` pod in `Error` / `CrashLoopBackOff` / `OOMKilled`, or `RobotTests failed` in `dbaas-redis-operator` logs | `references/robot-tests.md` (not `troubleshooting.md`) |

## Symptom (no exact string) → section

| Symptom | Header in references/troubleshooting.md |
|---|---|
| Redis pod not Running; metrics for a Redis node missing in monitoring | Redis Node Down Incident Report |
| Slow commands, `SLOWLOG` entries, high CPU on Redis pods | High Redis CPU Usage |
| Responses from Redis slower than the threshold | Latency on a Redis Node is More than N ms |
| Clients cannot connect, connection count close to `maxclients` | Connections Count to a Redis Node is More Than N% of the Limit |
| Redis memory growing close to its limit, evictions, memory-related pod restarts | High Redis Memory Usage |


## Config values

| Parameter | Lives on | Notes |
|---|---|---|
| `redis.conf.maxclients` (default `50000`) | chart values `redis.conf` → ConfigMap `redis-default-conf`, key `config` | reaches the running Redis pods through a rolling upgrade job |
| `redis.conf.maxmemory` (default `200mb`), `redis.conf.maxmemory-policy` (default `allkeys-lru`) | chart values `redis.conf` → ConfigMap `redis-default-conf`, key `config` | same rolling upgrade |
| `redis.resources.requests.cpu` / `.memory`, `redis.resources.limits.cpu` / `.memory` (chart defaults 50m / 64Mi and 250m / 256Mi) | chart values `redis.resources` → `DbaasRedisAdapter` CR `spec.redis.resources` | used only when `global.profile` is empty |
| `global.profile` (`small` / `medium` / `large`; env `REDIS_PROFILE` takes precedence) | chart values `global.profile` → `DbaasRedisAdapter` CR `spec.redis.resources` | fixed limits: small 250m CPU / 256Mi, medium 2 CPU / 4096Mi, large 4 CPU / 8192Mi; any other non-empty value fails the deploy |
| `monitoringAgent.prometheus.alerts.cpuThreshold`, `.memThreshold`, `.latencyThresholdMs`, `.connectionsThreshold` (defaults 95 / 95 / 20 / 90) | chart values `monitoringAgent.prometheus.alerts` → `PrometheusRule` `prometheus-redis-rules` | changes when the alert fires, not Redis behavior |
| `name` (default `dbaas-redis-service`) | chart value `name` → `DbaasRedisAdapter` CR `metadata.name` | Redis pods carry the labels `app` and `name` with this value |

A value changed inside a running Redis instance (the guide shows `CONFIG SET`) affects only that instance; the
persistent setting is the `redis.conf.*` value above.

## Config changes that lose or reject data

| Change | Exact parameter | Effect | Must be true first |
|---|---|---|---|
| Lower the memory cap | `redis.conf.maxmemory` | once used memory exceeds the new value, Redis evicts keys according to `maxmemory-policy`; under `noeviction` it rejects writes | current used memory is below the new value, and the application tolerates evicted keys |
| Change the eviction policy | `redis.conf.maxmemory-policy` | `noeviction` rejects writes with `OOM command not allowed when used memory > 'maxmemory'.` at the limit; any other policy deletes keys to make room | the application tolerates the chosen behavior |

## Guardrails

Parameters are listed in "Config values" with the resource each one lives on.

**Do not run cluster or component CLI commands.** Even though `references/troubleshooting.md` may mention `kubectl`, `redis-cli`, `SLOWLOG`, `CONFIG SET`, `INFO MEMORY`, or `MEMORY USAGE`. If a required artifact — pod logs, pod/CR status, metrics snapshot, Redis `INFO`/`SLOWLOG` output, etc. — is not provided, request it.

| Do not | Exact parameters / scope | Why |
|---|---|---|
| Run `kubectl`, `redis-cli`, or any other cluster or component CLI command | any situation where the reference file mentions checking pod status/logs, Redis `INFO`, `SLOWLOG`, `CONFIG SET`, or `MEMORY USAGE` | the agent has no live cluster or Redis access; request the artifact instead |
| Raise only the memory limit to give Redis more data capacity | `redis.resources.limits.memory` (or the `global.profile` memory limit) and `redis.conf.maxmemory` | the data cap is `maxmemory` (default `200mb`), not the container limit (default `256Mi`); raise both |
| Set `maxmemory` at or above the memory limit | `redis.conf.maxmemory`, `redis.resources.limits.memory` | the pod can be `OOMKilled` before Redis starts evicting keys |
| Lower `maxclients` for `Connections count to a Redis node is more than <N>% of the limit` | `redis.conf.maxclients` (default `50000`) | the guide's fix for that alert is to raise it; lowering is the guide's option only for `Redis CPU Usage` when too many clients are connected |
| Recommend alert thresholds for `Redis node is down` | `monitoringAgent.prometheus.alerts.*` | that alert is `redis_avg_latency == -1` and has no threshold parameter; thresholds exist only for the other four alerts |

## Robot integration tests failure

If the failing component is the `robot-tests` pod/deployment, load `references/robot-tests.md` instead of
`references/troubleshooting.md`.

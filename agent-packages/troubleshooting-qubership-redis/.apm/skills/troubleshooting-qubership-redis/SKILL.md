---
name: troubleshooting-qubership-redis
description: Diagnose and resolve Redis node down alerts, high CPU/memory/latency/connections alerts, ArgoCD rollback stuck on DbaasRedisAdapter, DBaaS Redis provisioning failures, Disaster Recovery switchover issues, redis-operator reconciliation failures, and failed robot integration tests. Match the reported symptom to the documented troubleshooting section; fall back to a general diagnostic checklist when no specific match exists.
---

## How to use the reference file

1. Grep issue headers:
   `grep -n "^## " references/troubleshooting.md`
2. Read only the matching section: offset at its line number, limit through to the next header's line number. Never load the whole file for one lookup.

## Symptom → reference section

| Symptom | Section in references/troubleshooting.md |
|---|---|
| ArgoCD rollback stuck / `DbaasRedisAdapter` not reconciling | Race Condition in ArgoCD Rolling Updates |
| Alert: "Redis Node Down" / monitoring cannot collect metrics | Redis Node Down Incident Report |
| Alert: "Connections Count to a Redis Node is More Than N% of the Limit" | Connections Count to a Redis Node |
| Alert: "Latency on a Redis Node is More than N ms" | Latency on a Redis Node |
| Alert: "High Redis CPU Usage" / slow commands / SLOWLOG | High Redis CPU Usage |
| Alert: "High Redis Memory Usage" / OOMKilled Redis pod | High Redis Memory Usage |
| `robot-tests` pod/deployment failed / operator reports `RobotTests failed` | Troubleshooting Robot Integration Tests |

## Robot integration tests failure

If the failing component is the `robot-tests` pod/deployment, load
`references/robot-tests.md` instead of `references/troubleshooting.md`.

## Alert name → service component

Redis alert names identify the resource category directly (e.g. `Redis Node Down`,
`High Redis CPU Usage`). Use the alert name to target the right component for logs
and metrics rather than inspecting all pods:

- **Node / pod availability** → Redis pod status and restart/OOM events (request artifact if not provided)
- **CPU / memory / latency** → Redis pod resource utilization metrics (request artifact if not provided)
- **Metrics collection failure** → `redis-monitoring-agent` pod logs (request artifact if not provided)
- **Operator / reconciliation** → `dbaas-redis-operator` pod logs (request artifact if not provided)

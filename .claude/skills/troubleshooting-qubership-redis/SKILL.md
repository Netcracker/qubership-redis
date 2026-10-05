---
name: troubleshooting-qubership-redis
description: "Diagnose and resolve Redis issues in qubership-redis: Node Down alerts, high CPU/memory/latency/connections alerts, ArgoCD rollback stuck on DbaasRedisAdapter, DBaaS Redis provisioning failures, Disaster Recovery switchover issues, and redis-operator reconciliation failures. Use this skill whenever the user reports any Redis problem, alert, or unexpected behavior — even without a named alert. Always start with logs and events, then match to reference sections."
---

## Diagnostic flow (always follow this order)

1. **Analyze provided logs and events first** — before looking anything up in references. Read all attached files. Look for:
   - Redis Operator logs — errors, reconciliation failures, panics
   - Redis instance events — status conditions, last transitions
   - Redis pod logs — crashes, OOM, connection errors
   - Monitoring agent logs — scrape failures, config errors

2. **Identify symptom** — match what you see in logs/events to the table below

3. **Read the reference section** — load only the matching section from `references/troubleshooting.md`

## How to use the reference file

Load only the relevant section — do NOT read the full file.

1. Grep section headers: `grep -n "^## " references/troubleshooting.md`
2. Match symptom to section
3. Read only that section using `offset` and `limit`

## Symptom → reference section

| Symptom | Section in references/troubleshooting.md |
|---|---|
| Alert: "Redis Node Down" / monitoring cannot collect metrics | Redis Node Down Incident Report |
| Alert: "Connections Count to a Redis Node is More Than N% of the Limit" | Connections Count to a Redis Node |
| Alert: "Latency on a Redis Node is More than N ms" | Latency on a Redis Node |
| Alert: "High Redis CPU Usage" / slow commands / SLOWLOG | High Redis CPU Usage |
| Alert: "High Redis Memory Usage" / OOMKilled Redis pod | High Redis Memory Usage |
| ArgoCD rollback stuck / `DbaasRedisAdapter` not reconciling | Race Condition in ArgoCD Rolling Updates |

## Alert name → service component

Use alert name to target the right component for logs and metrics:

- **Node / pod availability** → Redis pod status and restart/OOM events
- **CPU / memory / latency** → Redis pod resource utilization metrics
- **Metrics collection failure** → `redis-monitoring-agent` pod logs
- **Operator / reconciliation** → `dbaas-redis-operator` pod logs

## General diagnostic checklist (no alert / no reference match)

When no alert name is given and no section matches, work through this checklist:

1. **Operator health** — `kubectl get pods -n <namespace> | grep redis-operator` — is it running and not crash-looping?
2. **CRD status** — `kubectl get dbaasredisadapter -n <namespace>` — check `STATUS` column and `.status.conditions`
3. **Recent events** — `kubectl get events -n <namespace> --sort-by=.lastTimestamp | grep -i redis | tail -30`
4. **Redis pod health** — are pods Running? any restarts? `kubectl get pods -n <namespace> -l app=redis`
5. **Resource pressure** — OOMKilled or CPU throttling? `kubectl describe pod <redis-pod> -n <namespace>`
6. **Reconciliation errors** — grep operator logs for `ERROR` or `reconcile`
7. **DBaaS aggregator** — if provisioning path: check aggregator connectivity and credentials secret
8. **DR state** — if Disaster Recovery mode: check replication lag and `ClusterReplicator` status

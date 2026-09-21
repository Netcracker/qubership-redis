# Troubleshooting Robot Integration Tests

`redis-tests/robot/` contains the Robot Framework integration-test suite for Qubership Redis. It is packaged as a Docker image and run as the `robot-tests` Kubernetes Deployment when `robotTests.install` is true in the CR.

The general failure signatures are:

- The `robot-tests` pod/deployment is in `Error`, `CrashLoopBackOff`, or `OOMKilled` state.
- The operator reconciliation fails with `RobotTests failed`.
- Specific test groups fail or are unexpectedly skipped.

## Robot tests chart parameters

| Parameter | Effect on test execution |
|---|---|
| `robotTests.install` | Creates the `robot-tests` Deployment. If false, no test pod is deployed. |
| `robotTests.dockerImage` | Test runner image; default is `ghcr.io/netcracker/qubership-redis-tests:main`. |
| `robotTests.tags` | Sets the `TAGS` env var, used by Robot as an `--include` tag filter. Leave empty to run all tests. |
| `robotTests.priorityClassName` | Kubernetes PriorityClass for the test pod. |
| `robotTests.nodeLabels` | Node selector for the test pod. |
| `robotTests.resources` | CPU/memory requests and limits for the test pod. |
| `spec.waitTimeout` | How long (seconds) the operator waits for `robot-tests` Deployment to finish before treating it as failed. |

## How tests are organized

- `redis-tests/robot/tests/crud/crud.robot` — CRUD operations against Redis via DBaaS adapter and direct connection. Key tags:
  - `smoke` — all CRUD cases; run in every pipeline.
  - `dbaas` — cases that call the DBaaS adapter API to create/get/delete a Redis DB. Skipped when `DBAAS_ENABLED=false`.
  - `crud` — direct Redis read/write/update/delete operations.
- `redis-tests/robot/tests/monitoring/monitoring.robot` — Heartbeat write test. Key tags:
  - `monitoring-heartbeat` — runs only when this tag is explicitly included via `TAGS`; skipped otherwise.
- `redis-tests/robot/tests/image_tests/image_tests.robot` — Container image and hardening verification. Key tags:
  - `check_redis_images` — compares running container image tags against the deploy descriptor ConfigMap `tests-config`. Skipped if the ConfigMap is absent.
  - `redis_container_hardening` — runs container hardening checks (`CH12`, `CH4` are excluded by default).
  - `smoke` — included in every pipeline.
- `redis-tests/robot/tests/shared/keywords.robot` — shared library keywords (Redis connect, CRUD helpers, DBaaS adapter calls).
- `redis-tests/robot/tests/lib/RedisLibraryKeywords.py` — Python Redis client wrapper used by keywords.
- `redis-tests/robot/tests/lib/KubernetesClient.py` — Kubernetes API helper (deployments, custom resources).
- `redis-tests/Dockerfile` — builds the `redis-tests` image on top of `ghcr.io/netcracker/qubership-docker-integration-tests`.
- `redis-tests/entrypoint.sh` — loads credentials from `/var/run/secrets/redis/` into env vars before handing off to the base image entrypoint.

## Environment variables injected by the operator

| Variable | Source | Purpose |
|---|---|---|
| `OPENSHIFT_WORKSPACE_WA` | `cr.Namespace` | Kubernetes namespace for test pod |
| `REDIS_PORT` | Hardcoded `6379` | Redis port |
| `REDIS_DBAAS_ADAPTER_HOST` | `dbaas-redis-adapter.<namespace>.svc` | DBaaS adapter service host |
| `REDIS_DBAAS_ADAPTER_PORT` | TLS-aware (8080 or 8443) | DBaaS adapter port |
| `DBAAS_AGGREGATOR_REGISTRATION_ADDRESS` | `cr.Spec.Dbaas.Aggregator` | DBaaS aggregator URL |
| `DBAAS_ENABLED` | `cr.Spec.Dbaas.Install` | Whether DBaaS-mode tests run |
| `TAGS` | `cr.Spec.RobotTests.Tags` | Robot `--include` tag filter |
| `WAIT_TIMEOUT` | `cr.Spec.WaitTimeout` | Overall test wait timeout |
| `TLS_ENABLED` | Derived from aggregator URL + Redis TLS config | Enables TLS for Redis and adapter connections |
| `REDIS_PASSWORD` | Mounted from secret at `/var/run/secrets/redis/redis-password` | Redis auth password |
| `REDIS_DBAAS_USER` / `REDIS_DBAAS_PASSWORD` | Mounted from DBaaS adapter secret | DBaaS adapter credentials |

## Common failure modes and troubleshooting steps

### 1. `robot-tests` pod is stuck or crashing

- Inspect the `robot-tests` pod status and events (request artifact if not provided).
- Look for `ImagePullBackOff`, `OOMKilled`, `RunContainerError`, or volume mount errors in the pod events.
- Read the `robot-tests` container logs to find the Robot Framework output (request artifact if not provided).

### 2. DBaaS tests fail or are skipped unexpectedly

- When `DBAAS_ENABLED=false`, all `dbaas`-tagged tests are skipped via `Skip If`. This is expected when Redis is deployed in non-DBaaS mode.
- If `DBAAS_ENABLED=true` but the adapter is unreachable, `Test Creating DB Via Dbaas Adapter` will fail with a connection error. Check `dbaas-redis-adapter` pod status and logs (request artifact if not provided) to verify the adapter is running and the `REDIS_DBAAS_ADAPTER_HOST` / port are correct.

### 3. CRUD tests fail with Redis connection errors

- The `Test Connect To RedisDB` case connects directly to `redis.<namespace>` on port 6379.
- Check Redis pod status and events to confirm the pod is running (request artifact if not provided).
- Verify `REDIS_PASSWORD` is correctly populated — it is read from the secret mounted at `/var/run/secrets/redis/redis-password`. Check the `robot-tests` pod env vars or the mounted secret (request artifact if not provided).
- If TLS is enabled, verify `TLS_ENABLED=true` and the CA cert is present at `TLS_ROOTCERT` path in the pod.

### 4. Image check tests fail

- `check_redis_images` reads expected image tags from the `tests-config` ConfigMap. If the ConfigMap is absent, the test is skipped.
- If the test fails, a running container's image tag does not match the tag in the deploy descriptor. This usually indicates a partial upgrade. Check the `tests-config` ConfigMap content and compare with the actual pod images (request artifact if not provided).

### 5. Operator reports `RobotTests failed`

- The operator calls `helperImpl.WaitForTestsReady` on the `robot-tests` Deployment and fails if the Deployment does not reach a successful state within `spec.waitTimeout`.
- Read `dbaas-redis-operator` logs for the `RobotTests failed` message and surrounding context (request artifact if not provided).
- Read the full `robot-tests` pod logs to identify the failing test case (request artifact if not provided).
- Increase `spec.waitTimeout` if tests time out due to slow cluster conditions.

### 6. `monitoring-heartbeat` tests are always skipped

- This is by design: `monitoring.robot` skips all cases unless `monitoring-heartbeat` is explicitly in the `TAGS` value. Set `robotTests.tags: monitoring-heartbeat` in the CR or Helm values to enable them.

## Relevant source files

- Test image: `redis-tests/Dockerfile`, `redis-tests/entrypoint.sh`
- Shared keywords and env setup: `redis-tests/robot/tests/shared/keywords.robot`
- Python Redis client library: `redis-tests/robot/tests/lib/RedisLibraryKeywords.py`
- Operator test pod template: `redis-operator/api/v2/impl/robotTests/templates.go` → `RobotDeployment`
- Operator orchestration: `redis-operator/api/v2/impl/robotTests/templates.go` → `RobotBuilder.Build`
- Chart parameters: `redis-operator/charts/helm/redis-operator/values.yaml` (`robotTests`)

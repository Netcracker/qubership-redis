# AGENTS.md

Guidance for agents when working in this repository.

## Repository Layout

Multi-module Go workspace with three components:

- `redis-operator/` — Kubernetes operator (Operator SDK, controller-runtime) that reconciles the
  `DbaasRedisAdapter` CRD. See [redis-operator/README.md](redis-operator/README.md) and
  [redis-operator/docs/public/architecture.md](redis-operator/docs/public/architecture.md) for the
  supported deployment schemes (DBaaS, Non-HA, Disaster Recovery).
- `redis-monitoring-agent/` — Telegraf-based sidecar; auto-generates `telegraf.conf`, exposes
  Prometheus metrics.
- `redis-tests/` — RobotFramework integration/availability test suite.

### redis-operator layout

- `api/v1/` — legacy API version, kept only so the conversion webhook can convert to/from v2.
- `api/v2/` — storage version (`//+kubebuilder:storageversion`); all new fields go here.
- `api/v2/impl/` — builders per concern: `adapter/` (DBaaS provisioning), `redis/` (non-DBaaS
  deployment), `monitoring/`, `robotTests/`.
- `controllers/` — `DbaasRedisAdapterReconciler`; delegates to the external
  `nosqldb-operator-core` shared reconciler.
- `dbaas/pkg/` — `core/` constants, `entity/` DTOs, `redis/` client abstraction + mocks,
  `services/` DB lifecycle administration, `templates/` manifest templates.
- `charts/helm/redis-operator/` — Helm chart; CRD manifests under `crds/` and `legacy_crd/` (see
  "CRD code generation" below). Resource profiles: `small` (dev/PoC), `medium`, `large`.

Run a single test:
```bash
cd redis-operator && go test ./path/to/package/... -run TestName -v
```

## CRD Compatibility Rules

- `api/v2/` is the storage version; `api/v1/` exists only for conversion to/from v2. Do not add
  fields to `api/v1/` alone.
- New fields in `api/v2/` MUST be additive and optional. Never remove or rename an existing field,
  change its type, or turn an optional field into a required one.
- Every field change must keep the v1↔v2 conversion working — update the conversion code and its
  tests in the same change, not in a follow-up.
- If a change cannot be made without breaking an existing field or the conversion, stop and ask a
  human for explicit approval before making it.

## CRD Code Generation

After changing types in `api/v1/` or `api/v2/`, run from `redis-operator/`:
```bash
make manifests   # regenerates charts/helm/redis-operator/crds/k8s_1.22_crd.yaml + RBAC
make generate    # regenerates */zz_generated.deepcopy.go
```
Pass condition: both commands exit 0, and `git diff` touches only `k8s_1.22_crd.yaml`, RBAC
manifests, and `zz_generated.deepcopy.go` files.

Never hand-edit a `zz_generated.deepcopy.go` file — it is only ever produced by `make generate`.

`make manifests` regenerates only `charts/helm/redis-operator/crds/k8s_1.22_crd.yaml`; the
controller-gen rule that would regenerate `legacy_crd/crd.yaml` is commented out in the
`Makefile`. After every `make manifests`, manually port the same schema change into
`charts/helm/redis-operator/legacy_crd/crd.yaml` (`apiextensions.k8s.io/v1beta1` syntax) so both
chart CRDs describe the same resource.

## Key Dependencies

- `controller-runtime v0.22.4` — operator framework
- `k8s.io v0.34.1` — Kubernetes client
- `qubership-nosqldb-operator-core` — shared reconciler library (external)
- `qubership-dbaas-adapter-core` — DBaaS integration (external)
- `go-redis` — Redis client
- `fiber` — HTTP framework (adapter endpoints)
- `zap` — structured logging

## Ports

| Port    | Purpose             |
|---------|----------------------|
| 8070    | Webhook server       |
| 6060    | pprof profiling      |
| metrics | Disabled by default  |

## Build & Test Commands

All commands below run from `redis-operator/`:

```bash
# Build
make build       # Build manager binary

# Code quality
make fmt         # go fmt ./...
make vet         # go vet ./...

# Tests (also runs manifests, generate, fmt, vet first; requires envtest)
make test
```

After any Go change under `redis-operator/`, run:
```bash
cd redis-operator && make fmt vet test
```
Pass condition: `go fmt` reports no files, `go vet` prints nothing, and `go test` exits 0. Do not
consider the change done until all three pass.

# Qubership Redis
A set of components to run **Redis** in Kubernetes:

- **redis-operator** — CRD & controller that provisions Redis instances via the `DbaasRedisAdapter` custom resource.
- **redis-monitoring-agent** — Telegraf-based side service that auto-generates `telegraf.conf` and exposes Prometheus metrics.
- **redis-tests** — RobotFramework tests for quick CRUD / availability checks.

## Repository layout

```
redis-operator/          # Operator code, Helm chart, CRDs, docs
redis-monitoring-agent/  # Monitoring agent (Dockerfile + sources)
redis-tests/             # Robot tests (RobotFramework)
```

Useful operator docs:
- `redis-operator/docs/public/architecture.md`
- `redis-operator/docs/public/configuration.md`
- `redis-operator/docs/public/installation_guide.md`
- `redis-operator/docs/public/operations_guide.md`
- `redis-operator/docs/public/troubleshooting.md`


## License & Contributions

- License: see `LICENCE`
- Contributions: see `CONTRIBUTING.md` and `CODE-OF-CONDUCT.md`
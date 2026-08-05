# Vulnerability Fix Report — 2026-08-05

Source: PSUPCLCAP-5182 (TELUS SCA scan), XRAY report
`dbaas-redis_redis8-4.0.6-20260208.031114-2-RELEASE.xlsx`, branch `fix/vuln_redis_telus`.

Scope: High/Critical findings only, `qubership-redis-tests` excluded (test microservice),
deduplicated by CVE + Component. 134 unique findings after filtering.

## Fixes Applied

| Component | Change | Resolves |
|---|---|---|
| `redis-monitoring-agent/Dockerfile` | `telegraf:1.39.1-alpine` → `1.39.2-alpine` | CVE-2026-46595, -39831, -39834, -42508, -39830, -39833, -39832, -39829, -46597 (x/crypto); CVE-2026-55969, -48586 (thrift); CVE-2026-46600, -39821, -33814 (x/net, telegraf side); CVE-2026-56852 (x/text, telegraf side); CVE-2025-68615 (net-snmp), CVE-2026-4878 (libcap) via base-image `apk upgrade` |
| `redis-monitoring-agent/source/go.mod` | `golang.org/x/net` 0.55.0 → 0.57.0, `golang.org/x/text` 0.37.0 → 0.40.0 | CVE-2026-46600, CVE-2026-56852 (monitoring-watcher binary side) |
| `redis-operator/go.mod` | `github.com/go-openapi/spec` 0.21.0 → 0.22.9 | Advisory (no CVE id), "PowerShell Smart-Quote..."-class swag/spec issue — fix version v0.22.9 |
| `redis-operator/Dockerfile` | final stage `alpine:3.23.5` → `alpine:3.24.1` | CVE-2026-9080, -11352, -9547, -12064, -8286, -10536, -8932, -11564, -8924, -9079, -8927, -9545, -8926, -8925, -11856, -11586, -9546, -5773, -6276, -3805 (curl/libcurl, all — verified `curl-8.21.0-r0`/`libcurl-8.21.0-r0` shipped on Alpine 3.24, not yet backported to 3.23); CVE-2026-27135 (nghttp2-libs) |

Already fixed on this branch before this pass (verified, no action needed):
- `docker-redis/8/Dockerfile` → `redis:8.8.0-alpine` — pulled and inspected the image: `libcrypto3-3.5.7-r0`, `libssl3-3.5.7-r0`, `musl-1.2.5-r23`, `musl-utils-1.2.5-r23`, all at/above required fix versions (32 CVEs).
- `redis-operator/go.mod` → `gofiber/fiber/v2 v2.52.12` (≥ required 2.52.12), `go.opentelemetry.io/otel v1.41.0` (CVE-2025-66630, CVE-2026-25882, CVE-2026-29181).
- `redis-monitoring-agent/source/go.mod` → `go-jose/go-jose/v4 v4.1.4` already at fix version.
- `redis-operator/Dockerfile` builder / `redis-monitoring-agent/Dockerfile` builder → `golang:1.26.5-alpine` already satisfies all 22 Go-stdlib CVEs found in the monitoring-watcher binary (highest requirement was 1.26.5).

## 1. Unfixed / Needs Review

| CVE | Severity | Components | Reason | Impact |
|---|---|---|---|---|
| CVE-2026-54572 | High | `github.com/rclone/rclone` (embedded in telegraf binary) | No telegraf release ships rclone ≥ 1.74.4 yet. Verified telegraf v1.39.2 (latest tag, 2026-07-20) go.mod pins `rclone v1.74.1-0.20260628215305-6bbc28cf02dc`, a pre-fix snapshot from 2026-06-28; the rclone fix (v1.74.4) landed 2026-07-08, after that snapshot. rclone is statically compiled into the telegraf binary, so it can't be patched independently of a telegraf release. | Requires `-l/--links` against an attacker-controlled remote. This deployment's `telegraf.conf` (statsd/tcp/udp inputs, prometheus output — see `redis-operator/api/v2/impl/monitoring/templates.go`) does not configure any rclone-backed sync plugin, so the vulnerable code path isn't reachable in normal operation, but the binary still carries it. |
| CVE-2026-49980 | Critical | `github.com/rclone/rclone` (telegraf) | Same upstream gap; fix needs rclone ≥ 1.74.3, telegraf 1.39.2 still below that. | Vulnerable feature is `rclone rcd --rc-serve`; not started or exposed by this deployment. |
| (no CVE id — GHSA) "PowerShell Smart-Quote Filename Injection Enables SFTP Server-Side Command Execution" | High | `github.com/rclone/rclone` (telegraf) | Same upstream gap; fix needs rclone ≥ v1.75.0. | rclone SFTP server component not used by this deployment. |
| CVE-2026-65819 | High | `github.com/gopacket/gopacket` (telegraf) | telegraf 1.39.2 ships gopacket v1.7.0; fix is v1.7.1 — one patch release behind, not available in any current telegraf tag. | gopacket backs telegraf's packet-capture (pcap) input plugins, which are not enabled in this deployment's telegraf config. |
| CVE-2026-58208 | High | `github.com/nats-io/nats-server/v2` (telegraf) | XRAY report lists no Fix Version. | Likely already resolved: reported vulnerable range is `2.13.0 ≤ version < 2.14.3`; telegraf 1.39.2 ships nats-server v2.14.3, exactly at/above the upper bound. Flagging for scanner re-verification once XRAY publishes an explicit fix version. |

## 2. Transitive Issues

None — no findings had `ncdiag`/`diagtools` in the physical path.

## 3. End-of-Use Candidates

None — all unfixed findings are 2026 CVEs with active upstream components; no pre-2023 CVEs without a fix path.

## Follow-up

Recommend commenting on PSUPCLCAP-5182 that CVE-2026-46595 and the majority of findings are resolved in this branch, but CVE-2026-54572 / CVE-2026-49980 (rclone) and CVE-2026-65819 (gopacket) remain open pending the next telegraf release — track and re-scan once InfluxData cuts a new tag.

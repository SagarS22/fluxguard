# Distributed Rate Limiter (Library + Service) — Agile Delivery Plan

## 1) Project Overview
Build a **battle-tested distributed rate limiter** implemented as:
1. **Reusable libraries** in **Python** and **Go**.
2. A **sidecar microservice** exposing a stable API for non-native clients.

Core algorithms:
- **Token Bucket**
- **Sliding Window**
- **Leaky Bucket**

Distribution and consistency layer:
- **Redis** with **Lua scripts** for atomic state transitions.

Validation goals:
- Throughput benchmarks at **10k RPS** and **100k RPS** (k6 + containerized setup).
- Correctness argument/proof for race conditions and concurrent clients.

---

## 2) Scope and Non-Goals

### In Scope
- Algorithm implementations with deterministic, atomic Redis Lua scripts.
- Common policy model (limit, burst, refill rate/window, key dimensions).
- Python and Go SDKs with matching behavior and response schema.
- Sidecar service (HTTP/JSON) using same Redis scripts and policy model.
- Dockerized local/dev/perf environment.
- Load test suite and benchmark report.
- Correctness documentation including invariants and race-condition analysis.

### Out of Scope (initial release)
- Multi-region Redis replication semantics beyond documented best-effort.
- Full UI/dashboard.
- Proprietary auth providers (basic API key/mTLS support only).
- Dynamic policy management UI (file/env/Redis config only in v1).

---

## 3) Target Architecture

## Components
1. **Core Script Layer**
   - Lua scripts per algorithm with atomic update + decision logic.
   - Shared key naming and TTL strategy.

2. **Go Library (`/go/ratelimiter`)**
   - Client abstraction + policy definition.
   - `Allow()`, `Reserve()`, optional `Reset()` APIs.
   - Pluggable Redis client and observability hooks.

3. **Python Library (`/python/ratelimiter`)**
   - Symmetric API with Go library.
   - Async + sync interfaces (if feasible in v1; at least sync).

4. **Sidecar Service (`/service`)**
   - HTTP endpoints:
     - `POST /v1/check`
     - `POST /v1/reserve`
     - `GET /healthz`
     - `GET /metrics`
   - Stateless service sharing the same script engine.

5. **Benchmark Harness (`/bench`)**
   - k6 scripts for 10k and 100k RPS scenarios.
   - Result capture + summary generation.

6. **Ops/Packaging (`/deploy`)**
   - Dockerfiles, docker-compose, Make targets.
   - CI pipeline with tests, linting, perf smoke checks.

---

## 4) Agile Execution Model (Solo Builder)
- Cadence: **2-week sprints**.
- Rituals: sprint planning, daily personal check-in, end-of-sprint demo note, retrospective.
- Tracking: Epics → deliverables → tasks, each with acceptance criteria.
- Definition of Done:
  - Code + tests + docs + metrics + CI green.
  - Backward-compatible API changes unless versioned.
  - Self-review checklist completed for design, tests, and docs.

---

## 5) Work Breakdown Structure (Epics, Deliverables, Acceptance Criteria)

## Epic A — Product & Technical Foundations
**Deliverables**
- A1. Architecture Decision Records (ADRs) for algorithm semantics, API model, Redis key strategy.
- A2. Monorepo structure + coding conventions + lint/test tooling.
- A3. CI skeleton.

**Acceptance Criteria**
- ADRs approved and versioned.
- `make test`, `make lint`, `make bench-smoke` targets exist.
- CI runs on push/PR with pass/fail gates.

---

## Epic B — Redis Lua Algorithm Engine
**Deliverables**
- B1. Lua script: Token Bucket (atomic refill + consume).
- B2. Lua script: Sliding Window (time-bucketed or sorted-set model with bounded memory).
- B3. Lua script: Leaky Bucket (queue/drain-time model).
- B4. Script loading + SHA caching + fallback reload.
- B5. Shared protocol for result payload (allowed, remaining, retry_after, reset_at).

**Acceptance Criteria**
- Unit tests cover edge cases: clock skew assumptions, zero burst, over-limit, TTL expiry.
- Property tests for monotonicity/invariants under randomized workloads.
- Atomicity guaranteed via single-script execution for each decision.

---

## Epic C — Go Library
**Deliverables**
- C1. Public API for policy registration and checks.
- C2. Context-aware calls, timeout handling, Redis/network error strategy.
- C3. Instrumentation (latency histogram, decision counters, error counters).
- C4. Examples and README usage snippets.

**Acceptance Criteria**
- API docs generated and reviewed.
- Integration tests with real Redis in Docker.
- p95 latency budgets defined and measured in local bench.

---

## Epic D — Python Library
**Deliverables**
- D1. Matching policy and decision API with Go feature parity.
- D2. Packaging (`pyproject.toml`), versioning, publish-ready artifacts.
- D3. Optional async client interface (if scope allows; otherwise tracked for v1.1).
- D4. Examples and notebook/simple script demos.

**Acceptance Criteria**
- Contract tests prove parity with Go for same inputs.
- Wheels/sdist build in CI.
- Typed API stubs or inline type hints included.

---

## Epic E — Sidecar Microservice
**Deliverables**
- E1. OpenAPI spec for `/v1/check` and `/v1/reserve`.
- E2. Request validation, policy resolution, algorithm dispatch.
- E3. Health/readiness probes + Prometheus metrics.
- E4. Basic auth mode (API key or mTLS toggle).

**Acceptance Criteria**
- Stateless horizontal scale works behind load balancer.
- Contract tests pass against OpenAPI schema.
- Graceful degradation behavior documented for Redis outages.

---

## Epic F — Correctness & Race Condition Proof
**Deliverables**
- F1. Formalized invariants for each algorithm.
- F2. Concurrency model and assumptions (Redis single-threaded script execution, time source strategy).
- F3. Proof sketch/document (safety + bounded liveness) under high concurrency.
- F4. Jepsen-lite style stress suite / randomized concurrent test runner.

**Acceptance Criteria**
- Documented proof completed with a self-review checklist (invariants, assumptions, counterexamples).
- Stress tests demonstrate no invariant violations in repeated runs.
- Known limitations and assumptions explicitly listed.

---

## Epic G — Performance Benchmarking (10k/100k RPS)
**Deliverables**
- G1. k6 scenarios for:
  - Single key hotspot
  - High-cardinality keys
  - Mixed algorithm workloads
- G2. Benchmark environment profile (CPU/memory/network settings).
- G3. Results report with throughput, p50/p95/p99 latency, Redis CPU/mem, error rate.
- G4. Optimization pass (script micro-optimizations, pipelining, pooling).

**Acceptance Criteria**
- Reproducible benchmark scripts checked into repo.
- Achieves stable 10k RPS and target exploration at 100k RPS with documented hardware.
- Bottlenecks and recommended production sizing published.

---

## Epic H — Packaging, Deployment, and Release
**Deliverables**
- H1. Docker images for service + test harness.
- H2. docker-compose for local stack (service + Redis + optional Grafana/Prometheus).
- H3. Release workflow (semantic versioning, changelog, tags).
- H4. Production runbook (SLOs, alert thresholds, scaling guidance).

**Acceptance Criteria**
- One-command local bring-up (`docker compose up`).
- Versioned releases for Python package, Go module, and service image.
- Runbook validated in game-day scenario (Redis restart/network hiccup).

---

## 6) Sprint-by-Sprint Plan (8 Sprints / 16 Weeks)

## Sprint 1 — Foundations
- Finalize requirements and ADRs.
- Set up repo structure, CI, lint/test framework.
- Implement baseline Redis connectivity and script loader.

**Sprint Deliverable:** working scaffold with CI and script execution harness.

## Sprint 2 — Token Bucket End-to-End
- Implement Lua token bucket + tests.
- Expose in Go + Python minimal APIs.
- Initial sidecar `/v1/check` support.

**Sprint Deliverable:** token bucket functional across library + service.

## Sprint 3 — Sliding Window End-to-End
- Implement Lua sliding window with memory bounds.
- Add integration tests and parity tests.
- Extend service routing + policy support.

**Sprint Deliverable:** sliding window fully integrated.

## Sprint 4 — Leaky Bucket End-to-End
- Implement Lua leaky bucket.
- Validate edge cases and failure behavior.
- Harden response schema consistency across all algorithms.

**Sprint Deliverable:** all three algorithms available.

## Sprint 5 — Hardening & Correctness
- Build invariant/property tests.
- Write correctness proof draft.
- Concurrency stress tooling and race-focused test scenarios.

**Sprint Deliverable:** correctness package v1 (tests + document).

## Sprint 6 — Performance Engineering
- Build k6 workload suite.
- Run 10k RPS benchmarks, identify hotspots.
- Optimize scripts/client pooling/serialization.

**Sprint Deliverable:** 10k RPS benchmark report + optimization changelog.

## Sprint 7 — 100k RPS Campaign + Operability
- Execute 100k RPS tests in tuned environment.
- Improve observability, metrics dashboards, alert guidelines.
- Validate autoscaling and failure recovery behavior.

**Sprint Deliverable:** 100k RPS report + production sizing guide.

## Sprint 8 — Release Readiness
- Final docs, examples, runbooks.
- API stabilization and versioning.
- Release candidates, bug bash, GA release.

**Sprint Deliverable:** v1.0 libraries + sidecar service release.

---

## 7) Testing Strategy
- **Unit tests:** algorithm math, boundary conditions.
- **Integration tests:** Redis-backed script behavior.
- **Contract tests:** parity between Go/Python/service responses.
- **Concurrency tests:** randomized parallel callers.
- **Fault injection:** Redis timeout, connection drops, script cache misses.
- **Performance tests:** sustained and spike k6 workloads.

Quality gates:
- >=90% coverage on critical algorithm paths.
- Zero known invariant violations in stress suite.
- Benchmark variance within agreed tolerance (e.g., ±10%).

---

## 8) Risks and Mitigations
1. **Redis bottleneck at high RPS**
   - Mitigation: key sharding, connection pooling, script optimization, Redis cluster guidance.
2. **Clock-related anomalies**
   - Mitigation: server-side timestamp usage; document assumptions.
3. **Cross-language behavior drift**
   - Mitigation: shared test vectors + contract suite.
4. **Memory growth in sliding window**
   - Mitigation: bounded bucketization and TTL cleanup.
5. **Operational complexity**
   - Mitigation: sidecar defaults, sane configs, runbook and dashboards.

---

## 9) Deliverables Checklist (Final)
- [ ] Redis Lua scripts for token/sliding/leaky algorithms.
- [ ] Go library (versioned) with docs/examples.
- [ ] Python library (versioned) with docs/examples.
- [ ] Sidecar service with OpenAPI and Docker image.
- [ ] k6 benchmark suite + 10k/100k RPS reports.
- [ ] Correctness proof document + stress test evidence.
- [ ] CI/CD release pipeline and operational runbook.

---

## 10) Suggested Repository Layout
```text
/plan.md
/go/ratelimiter
/python/ratelimiter
/service
/scripts/lua
/bench/k6
/deploy/docker
/docs
  /adr
  /proof
  /runbook
```

---

## 11) Success Metrics
- Functional: all three algorithms pass conformance and parity tests.
- Performance: documented throughput at 10k and 100k RPS with p95/p99 latency.
- Reliability: no invariant violations under stress runs.
- Usability: SDK + sidecar quickstart <15 minutes to first successful limit.
- Operability: actionable metrics, alerts, and runbook tested.

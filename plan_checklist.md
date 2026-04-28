# Distributed Rate Limiter — Implementation Checklist

## Foundations
- [ ] Finalize scope and v1 requirements
- [ ] Define repository structure
- [ ] Add coding standards and tooling
- [ ] Set up CI for linting and tests
- [ ] Add local dev commands (`make` or equivalent)

## Architecture and Design
- [ ] Write ADR for overall architecture
- [ ] Define shared policy model
- [ ] Define shared response schema
- [ ] Define Redis key naming strategy
- [ ] Define TTL and cleanup strategy
- [ ] Define error-handling behavior

## Redis Lua Core
- [ ] Set up Redis script loader
- [ ] Implement token bucket Lua script
- [ ] Implement sliding window Lua script
- [ ] Implement leaky bucket Lua script
- [ ] Add script SHA caching and reload logic
- [ ] Add unit tests for algorithm edge cases

## Go Library
- [ ] Create Go package skeleton
- [ ] Implement Redis client integration
- [ ] Implement policy configuration
- [ ] Implement `Allow()` API
- [ ] Implement `Reserve()` API
- [ ] Add tests
- [ ] Add examples and docs

## Python Library
- [ ] Create Python package skeleton
- [ ] Implement Redis client integration
- [ ] Implement policy configuration
- [ ] Implement check/reserve APIs
- [ ] Add tests
- [ ] Add packaging config
- [ ] Add examples and docs

## Sidecar Service
- [ ] Choose service implementation language/runtime
- [ ] Define HTTP API
- [ ] Write OpenAPI spec
- [ ] Implement `/v1/check`
- [ ] Implement `/v1/reserve`
- [ ] Implement `/healthz`
- [ ] Implement `/metrics`
- [ ] Add request validation
- [ ] Add tests

## Correctness and Concurrency
- [ ] Define invariants for each algorithm
- [ ] Document race-condition assumptions
- [ ] Write correctness proof/proof sketch
- [ ] Add concurrency stress tests
- [ ] Add randomized/property-based tests
- [ ] Document limitations and tradeoffs

## Performance and Benchmarking
- [ ] Set up benchmark environment
- [ ] Create k6 scripts for 10k RPS
- [ ] Create k6 scripts for 100k RPS
- [ ] Benchmark hotspot key scenarios
- [ ] Benchmark high-cardinality scenarios
- [ ] Benchmark mixed algorithm scenarios
- [ ] Capture latency and throughput results
- [ ] Optimize bottlenecks
- [ ] Write benchmark summary

## Observability and Operations
- [ ] Add Prometheus metrics
- [ ] Add structured logging
- [ ] Define alerting suggestions
- [ ] Add Dockerfiles
- [ ] Add docker-compose setup
- [ ] Test Redis restart/failure handling
- [ ] Write production runbook

## Parity and Quality
- [ ] Add cross-language parity tests
- [ ] Validate response consistency across SDKs and service
- [ ] Add regression tests
- [ ] Review API ergonomics
- [ ] Review configuration ergonomics

## Release Readiness
- [ ] Finalize documentation
- [ ] Finalize versioning strategy
- [ ] Prepare release workflow
- [ ] Verify local setup from scratch
- [ ] Verify benchmark reproducibility
- [ ] Cut v1.0 release

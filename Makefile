.PHONY: help test test-go test-python lint lint-go lint-python bench-smoke bench docker-up docker-down clean install-go install-python

help:
	@echo "FluxGuard - Distributed Rate Limiter"
	@echo ""
	@echo "Available targets:"
	@echo "  make install-go     - Install Go dependencies"
	@echo "  make install-python - Install Python dependencies"
	@echo "  make test          - Run all tests"
	@echo "  make test-go       - Run Go tests"
	@echo "  make test-python   - Run Python tests"
	@echo "  make lint         - Run linters"
	@echo "  make lint-go      - Run Go linter (golangci-lint)"
	@echo "  make lint-python - Run Python linter (ruff)"
	@echo "  make bench-smoke - Run benchmark smoke tests"
	@echo "  make bench        - Run full benchmarks"
	@echo "  make docker-up    - Start Docker services"
	@echo "  make docker-down - Stop Docker services"
	@echo "  make clean       - Clean build artifacts"

install-go:
	cd go/ratelimiter && go mod download

install-python:
	cd python && pip install -e .

test: test-go test-python

test-go:
	cd go/ratelimiter && go test -v -race ./...

test-python:
	cd python && python -m pytest -v

lint: lint-go lint-python

lint-go:
	cd go/ratelimiter && go vet ./... && golangci-lint run

lint-python:
	cd python && ruff check .

bench-smoke:
	cd bench && k6 run --vus=10 --duration=10s smoke.js

bench:
	cd bench && k6 run 10k-rps.js

docker-up:
	docker compose -f deploy/docker-compose.yaml up -d

docker-down:
	docker compose -f deploy/docker-compose.yaml down

clean:
	cd go/ratelimiter && go clean
	find . -type d -name __pycache__ -exec rm -rf {} + 2>/dev/null || true
	find . -type f -name "*.pyc" -delete
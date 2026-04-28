# Distributed Rate Limiter

Go and Python SDK with sidecar service.

## Project Structure

```
/go/ratelimiter      - Go SDK (in progress)
/python/ratelimiter - Python SDK (in progress)
/service           - Sidecar microservice
/scripts/lua       - Redis Lua scripts
/bench             - k6 benchmark harness
/deploy            - Dockerfiles & docker-compose
/docs/adr          - Architecture Decision Records
```

## Quickstart (coming soon)

```bash
# Start local stack
make docker-up

# Run tests
make test

# Run linters
make lint
```

## Development

See `plan.md` for full roadmap.

## License

MIT
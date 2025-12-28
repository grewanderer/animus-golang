# Contributing

Thanks for contributing to Animus DataPilot's open integration layer. This repository targets production-grade enterprise pilots; changes must be verifiable and audit-ready.

## Development setup

Prerequisites:

- Go 1.25+
- Python 3.10+

Bootstrap:

```bash
make bootstrap
```

## Coding standards

- Go: `gofmt` must pass and `go vet` must be clean (`make lint`).
- Python: `python -m compileall` must pass (`make lint`).
- Avoid TODOs and stub implementations.
- Prefer small, testable changes with clear commit messages.

## Tests

```bash
make test
```

## Pull request checklist

- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] OpenAPI specs updated if endpoints change
- [ ] SDK docs updated if usage changes
- [ ] Security implications reviewed

## Commit style

There is no enforced commit format. Use a short, descriptive subject and include scope where helpful (e.g., `experiments: add evidence bundle verification`).

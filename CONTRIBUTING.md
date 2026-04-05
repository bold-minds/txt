# Contributing to `each`

Thanks for your interest in contributing. This guide covers the operational process. For the **why** — the design principles every contribution is tested against — see **[bold-minds/oss/PRINCIPLES.md](https://github.com/bold-minds/oss/blob/main/PRINCIPLES.md)**.

## 🎯 Before You Start

Every contribution is measured against the four Bold Minds principles: **outcome naming**, **one way to do each thing**, **get out of the way**, and **non-goals explicit**. If your proposed change doesn't honor these, it will not be merged.

**Read [PRINCIPLES.md](https://github.com/bold-minds/oss/blob/main/PRINCIPLES.md) first.**

## 🔧 Development Setup

**Requirements:** Go 1.21 or later, Git, Bash.

```bash
git clone https://github.com/bold-minds/each.git
cd each
go test ./...              # unit tests
go test -race ./...        # race detection
go test -bench=. ./...     # benchmarks
./scripts/validate.sh      # full validation pipeline
./scripts/validate.sh ci   # strict CI mode
```

Your contribution must pass `./scripts/validate.sh ci` before submitting.

## 📁 Project Structure

```
each/
├── each.go                # Implementation (single file)
├── each_test.go           # Unit tests with adversarial coverage
├── bench_test.go          # Benchmarks
├── examples/              # Runnable examples
├── scripts/
│   └── validate.sh        # Validation pipeline
├── README.md
├── CONTRIBUTING.md        # This file
├── CHANGELOG.md
├── CODE_OF_CONDUCT.md
├── SECURITY.md
├── LICENSE
└── go.mod
```

Flat layout. No `internal/` directory.

## 🎨 Code Style

### Naming
- Outcome naming per PRINCIPLES.md. Function names describe the predicate-based operation performed.

### Error Handling
- Functions **must not panic** on valid input.
- No error returns — operations either succeed or return sensible defaults (zero values, non-nil empty slices/maps).
- Every function is nil-safe.

### Documentation
- Every exported function has a doc comment describing behavior, edge cases (nil/empty input), and ordering guarantees.
- Panic risks from caller-side misuse (non-comparable key types from `GroupBy`/`KeyBy`) are documented in the package doc.

### Dependencies
- **Zero external dependencies.** `each` is pure stdlib.

## 🧪 Testing

**Coverage target: 100% of exported functions.**

**Every PR must include adversarial tests.** The lesson from the wider Bold Minds library family: passing validation is necessary but insufficient. Tests must explicitly verify:

1. **Non-nil return guarantees** — no function returns `nil` for empty results when the docs promise "non-nil empty"
2. **Immutability** — input slices are byte-identical before and after the call
3. **Result non-aliasing** — mutating the returned slice does not affect the input
4. **Custom comparable key types** — `GroupBy`/`KeyBy` must work with named types, structs, and any other comparable Go type
5. **Short-circuit evaluation** — `Every` must stop on the first false; write a counter-based test that would catch a regression
6. **Stateful predicates** — closures with captured state must work (common real-world pattern)

```bash
go test -v ./...
go test -race ./...
go test -cover ./...
go test -bench=. -benchmem ./...
```

## 📝 Pull Request Process

### PR Checklist

- [ ] **Outcome naming** — does the function name describe what the caller gets?
- [ ] **One way** — does any existing function (this library or stdlib) already do this?
- [ ] **Get out of the way** — can a Go dev use this from the signature alone?
- [ ] **Non-goals** — does this violate any of the library's stated non-goals?
- [ ] Tests cover 100% of new code
- [ ] Adversarial tests included (nil returns, immutability, aliasing)
- [ ] Benchmarks added for new exported functions
- [ ] README updated (if adding or changing exported functions)
- [ ] CHANGELOG.md updated
- [ ] `./scripts/validate.sh ci` passes locally

## 🆕 Adding a New Function

`each` is deliberately scoped to seven functions. New additions must clear a high bar:

1. Read the library's non-goals in [README.md](README.md#-whats-deliberately-not-here) and [PRINCIPLES.md](https://github.com/bold-minds/oss/blob/main/PRINCIPLES.md)
2. Prove the stdlib gap. Current Go's `slices` package is more capable than many realize.
3. Confirm the function operates on a single slice with a predicate or key function (operations on multiple slices belong in [`bold-minds/list`](https://github.com/bold-minds/list)).
4. Show real-world evidence of the pain.
5. Reject anything that is a thin wrapper around `slices.IndexFunc`, `slices.ContainsFunc`, or other stdlib helpers.

## 🏷️ Versioning and Releases

- Semantic versioning
- v0.x: API may change between minor versions
- v1.0+: breaking changes require major version bump

## 🙏 Code of Conduct

See [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md).

## 📄 License

By contributing, you agree your contributions are licensed under the MIT License.

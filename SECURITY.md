# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 1. **Do Not** Create a Public Issue

Please do not report security vulnerabilities through public GitHub issues, discussions, or pull requests.

### 2. Report Privately

Send an email to **security@boldminds.tech** with:

- **Subject**: Security Vulnerability in bold-minds/each
- **Description**: Detailed description of the vulnerability
- **Steps to Reproduce**: Clear steps to reproduce the issue
- **Impact**: Potential impact and severity assessment
- **Suggested Fix**: If you have ideas for a fix (optional)

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution**: Varies based on complexity, typically within 30 days

## Security Considerations

`each` is a pure-computation library with a very small attack surface:

- **No network I/O.** `each` does not make network calls.
- **No file I/O.** `each` does not read or write files.
- **No reflection.** All operations use Go's generics and concrete type constraints.
- **No external dependencies.** Pure Go stdlib.
- **Immutable.** `each` never modifies input slices.
- **Nil-safe.** All functions handle nil slices without panicking.

### Known runtime-panic sources from caller misuse

`each` does not panic on any documented input. However, `GroupBy` and
`KeyBy` use the caller-provided key function's return value as a Go map
key. If the key function returns a non-comparable dynamic type (e.g.,
a slice or map stored inside an `any`), Go's map implementation will
panic at runtime with a "hash of unhashable type" error. `each` does
not recover from these panics. Callers must ensure their key functions
return comparable values.

### Predicates with side effects

All `each` functions accept a predicate or key function supplied by the
caller. These functions are evaluated one or more times per element.
`each` does not sandbox predicate execution — if a predicate panics,
the panic propagates to the caller. If a predicate has side effects
(I/O, global state mutation, etc.), those side effects occur during
the call. For `Every`, the short-circuit behavior means that elements
after the first false are not visited, so their predicate side effects
will not fire.

## Security Updates

Security updates will be released as patch versions (e.g., 0.1.1),
documented in CHANGELOG.md, and announced through GitHub releases.

## Acknowledgments

We appreciate responsible disclosure and will acknowledge security
researchers who help improve the security of this project.

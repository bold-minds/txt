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

- **Subject**: Security Vulnerability in bold-minds/txt
- **Description**: Detailed description of the vulnerability
- **Steps to Reproduce**: Clear steps to reproduce the issue
- **Impact**: Potential impact and severity assessment
- **Suggested Fix**: If you have ideas for a fix (optional)

### 3. Response Timeline

- **Initial Response**: Within 48 hours
- **Status Update**: Within 7 days
- **Resolution**: Varies based on complexity, typically within 30 days

## Security Considerations

`txt` is a small, self-contained string library with a narrow attack surface:

- **No network I/O.** `txt` does not make network calls.
- **No file I/O.** `txt` does not read or write files (apart from `Print`, which writes a formatted string to stdout via `fmt.Println`).
- **No external dependencies.** Pure Go stdlib.
- **Minimal reflection.** Reflection is used in exactly one place — `formatValue` calls `reflect.TypeOf(v).Kind() == reflect.Chan` to render channel arguments for `Format`. All other type handling goes through a concrete type switch.
- **Cryptographically-random string generation.** `Random` draws from `crypto/rand` (not `math/rand`) via a rejection-sampling helper that avoids modulo bias.

### Known behaviors callers must be aware of

#### `Truncate` can produce invalid UTF-8

`Truncate` operates on bytes, not runes. Calling it on a string that contains multibyte UTF-8 characters can cut mid-sequence and yield a byte slice that is not valid UTF-8 (e.g. `txt.Truncate("héllo", 2, "")` → `"h\xc3"`). This is a deliberate trade-off for predictable byte-length bounds. For UTF-8 safety, bound the input with `Substring` first, which operates on runes.

This is not a panic or memory-safety issue, but it can surface in downstream systems that assume well-formed UTF-8 (database columns, JSON encoders, Protobuf fields). Treat `Truncate` as a byte operation and encode accordingly.

#### `Random` and `randInt` panic on `crypto/rand` failure

`Random` and the internal `randInt` helper panic with `"txt: crypto/rand.Read failed: <err>"` if `crypto/rand.Read` returns a non-nil error. On a healthy system this never happens — `crypto/rand` only fails on extraordinary system-level faults (e.g. a sealed getrandom syscall, an exhausted file descriptor table). The library does not attempt to recover because there is no meaningful fallback for a broken entropy source, and silently returning predictable output would be worse than a panic.

#### `Random` entropy use case

`Random` is suitable for non-key secrets such as invite codes, correlation IDs, password resets, or test fixtures. It is **not** a substitute for key derivation or high-entropy cryptographic material — use `crypto/rand`, HKDF, or Argon2id directly for those.

### No panics on caller input

Apart from the `crypto/rand` failure path above, `txt` never panics on any documented caller input. `Substring` clamps out-of-range indices, `Truncate` tolerates negative and oversized `maxLen`, and `Format` leaves unfilled placeholders in place when arguments are missing so the bug is visible to whoever reads the log line.

## Security Updates

Security updates will be released as patch versions (e.g., 0.1.1),
documented in CHANGELOG.md, and announced through GitHub releases.

## Acknowledgments

We appreciate responsible disclosure and will acknowledge security
researchers who help improve the security of this project.

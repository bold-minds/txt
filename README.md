# txt

[![Go Reference](https://pkg.go.dev/badge/github.com/bold-minds/txt.svg)](https://pkg.go.dev/github.com/bold-minds/txt)
[![Build](https://img.shields.io/github/actions/workflow/status/bold-minds/txt/test.yaml?branch=main&label=tests)](https://github.com/bold-minds/txt/actions/workflows/test.yaml)
[![Go Version](https://img.shields.io/github/go-mod/go-version/bold-minds/txt)](go.mod)
[![Coverage](https://img.shields.io/endpoint?url=https://gist.githubusercontent.com/clairevnext/eb5318a268371b987ef7b15fef8f9bee/raw/coverage.json)](https://github.com/bold-minds/txt/actions/workflows/test.yaml)

**Outcome-named string formatting and manipulation — the stdlib helpers Go doesn't ship.**

Go's `fmt` package covers every case but speaks in terse verbs. Go's `strings` package covers everything but makes you chain idioms. `txt` fills the gap with outcome-named formatters (`Hex`, `Binary`, `Float`, ...) and direct one-line helpers for the string operations that otherwise take three lines: extract between delimiters, collapse whitespace, safe substring with negative indexing, byte-bounded truncation, and cryptographically-random strings with configurable charsets.

```go
// Before — fmt verb vocabulary + manual bounds + multi-line squish
hex := fmt.Sprintf("%x", 255)
msg := fmt.Sprintf("user %d logged in from %s", userID, host)
fields := strings.Fields("  hello   world  ")
cleaned := strings.Join(fields, " ")

// After
hex := txt.FormatAs(txt.Hex, 255)
msg := txt.Format("user {} logged in from {}", userID, host)
cleaned := txt.Squish("  hello   world  ")
```

## ✨ Why txt?

- 🎯 **Outcome-named verbs** — `txt.FormatAs(txt.Hex, 255)` reads as "format as hex," not "format with `%x`"
- 💬 **`{}` placeholders** — `txt.Format("user {} not found", id)` skips the verb-to-type mental pattern match for error messages
- 🔍 **`Between`** — extract substrings between delimiters in one call, no manual index math
- ✂️ **`Squish`** — collapse runs of whitespace and trim in one outcome-named call (no `Fields`+`Join` idiom)
- 🔢 **Unicode-safe `Substring`** — rune-counting, negative indices, out-of-range clamping, never panics
- 🪓 **Byte-bounded `Truncate`** — shorten with a suffix, guaranteed `len(out) <= maxLen`
- 🎲 **Cryptographically-random `Random`** — `crypto/rand`-backed, 12 charset options, safe by default for invite codes and correlation IDs
- 🪶 **~20 functions, one file, zero dependencies** — pure stdlib
- 🔗 **Pairs with [`bold-minds/each`](https://github.com/bold-minds/each) and [`bold-minds/list`](https://github.com/bold-minds/list)** — same outcome-naming convention across the family

## 📦 Installation

```bash
go get github.com/bold-minds/txt
```

Requires Go 1.21 or later.

## 🚀 Quick Start

```go
package main

import (
    "fmt"

    "github.com/bold-minds/txt"
)

func main() {
    // {} placeholder formatting — skips fmt's verb vocabulary
    msg := txt.Format("user {} connected from {}:{}", 42, "192.168.0.1", 8080)
    fmt.Println(msg) // "user 42 connected from 192.168.0.1:8080"

    // Outcome-named verbs instead of %x, %b, %f, ...
    fmt.Println(txt.FormatAs(txt.Hex, 255))                  // "ff"
    fmt.Println(txt.FormatAs(txt.Binary, 42))                // "101010"
    fmt.Println(txt.FormatAs(txt.Float.Precision(2), 3.14159)) // "3.14"

    // Extract between delimiters
    path := txt.Between("GET /api/users/42/profile HTTP/1.1", "/users/", "/")
    fmt.Println(path) // "42"

    // Collapse whitespace
    fmt.Println(txt.Squish("  the   quick\tbrown\n\nfox"))   // "the quick brown fox"

    // Rune-safe substring with negative indexing
    fmt.Println(txt.Substring("héllo world", 1, 4))          // "éllo"
    fmt.Println(txt.Substring("hello", -3, 3))               // "llo"

    // Byte-bounded truncation
    fmt.Println(txt.Truncate("The quick brown fox", 12, "...")) // "The quick..."

    // Cryptographically-random strings
    code := txt.Random(8, txt.AlphaNum(), txt.AlphaStart())
    fmt.Println(code) // e.g. "K3x9aB2p"

    token := txt.Random(16, txt.Letters(), txt.Exclude('l', 'I', '0', 'O'))
    fmt.Println(token) // e.g. "XzmpHvDfKtRjBqWn" — no confusable chars
}
```

## 🔧 Core Features

### `Format` — `{}` placeholder building

Replaces each `{}` in the template with the matching argument, using type-aware formatting (decimals for integers, `%g` for floats, `"Error: <msg>"` for errors). Extras are ignored; missing ones leave placeholders in place so bugs are visible.

```go
txt.Format("user {} not found", userID)
txt.Format("failed to connect to {}:{}", host, port)
errors.New(txt.Format("invalid value: {}", val))
```

For control over precision, base, or quoting, use `FormatAs` with the exported `FmtType` constants.

### `FormatAs` — outcome-named verbs

Twelve exported format constants wrap fmt verbs in outcome-named handles:

| Constant | Verb | Purpose |
|---|---|---|
| `Binary` | `%b` | base 2 |
| `Octal` | `%o` | base 8 |
| `Hex` | `%x` | lowercase hex |
| `HexUpper` | `%X` | uppercase hex |
| `Char` | `%c` | rune literal |
| `Float` | `%f` | decimal, no exponent |
| `Scientific` | `%e` | scientific notation |
| `ScientificUpper` | `%E` | scientific (uppercase) |
| `Quoted` | `%q` | Go-quoted string / rune |
| `Unicode` | `%U` | `U+XXXX` |
| `Type` | `%T` | Go type name |
| `Pointer` | `%p` | pointer address |

```go
txt.FormatAs(txt.Hex, 255)                    // "ff"
txt.FormatAs(txt.Float.Precision(2), 3.14159) // "3.14"
txt.FormatAs(txt.Binary, 42)                  // "101010"
txt.FormatAs(txt.Hex, 1, 2, 3)                // "1 2 3" — space-joined
```

`Precision` returns a fresh `FmtType` — calling it never mutates the exported constants.

### `Print` — `Format` to stdout

Thin convenience wrapper for one-liner CLI output. Use `Format` directly when you need the string for logging or errors.

```go
txt.Print("ready")
txt.Print("user {} logged in", userID)
```

### `Between` — extract between delimiters

Returns the substring between the first `start` and the next `end`. Returns `""` if either delimiter is missing. Empty `start` anchors at the beginning; empty `end` anchors at the end.

```go
txt.Between("foo [bar] baz", "[", "]")        // "bar"
txt.Between("a=1&b=2", "a=", "&")             // "1"
txt.Between("BEGIN hello END", "BEGIN ", " END") // "hello"
txt.Between("prefix:value", "", ":")          // "prefix"
txt.Between("prefix:value", ":", "")          // "value"
txt.Between("no markers", "[", "]")           // ""
```

### `Squish` — collapse and trim whitespace

Collapses every run of whitespace in `s` to a single space and trims leading/trailing whitespace. Equivalent to `strings.Join(strings.Fields(s), " ")`, but outcome-named.

```go
txt.Squish("  hello   world  ")  // "hello world"
txt.Squish("\tfoo\n\nbar")       // "foo bar"
txt.Squish("   ")                // ""
```

### `Substring` — rune-safe extraction with negative indices

Returns `length` **runes** (not bytes) starting at `start`. Negative `start` counts from the end. Out-of-range indices clamp to the string boundaries rather than panicking.

```go
txt.Substring("hello", 0, 3)    // "hel"
txt.Substring("hello", -2, 2)   // "lo"
txt.Substring("héllo", 1, 3)    // "éll" — counts runes, not bytes
txt.Substring("日本語", 1, 2)    // "本語"
txt.Substring("hi", 5, 10)      // "" — start past end, no panic
txt.Substring("anything", 0, 0) // "" — zero length
```

### `Truncate` — byte-bounded with suffix

Shortens `s` to at most `maxLen` **bytes**, appending `suffix` if shortened. The output's byte length is guaranteed to be `<= maxLen`. If `maxLen <= len(suffix)`, a prefix of `suffix` of that length is returned.

```go
txt.Truncate("Hello world", 8, "...")  // "Hello..."
txt.Truncate("short", 20, "...")       // "short"
txt.Truncate("abcdef", 2, "...")       // ".."
txt.Truncate("anything", -1, "...")    // ""
```

`Truncate` operates on bytes. For UTF-8 safety on multibyte strings, bound the rune count with `Substring` first.

### `Random` — cryptographically-random strings

Backed by `crypto/rand` with rejection-sampling for uniform distribution. Safe by default for non-key secrets like invite codes, correlation IDs, and test fixtures.

```go
txt.Random(16)                                         // 16 chars, full printable ASCII
txt.Random(8, txt.AlphaNum())                          // 8 alphanumeric
txt.Random(12, txt.Letters(), txt.AlphaStart())        // 12 letters, first is alpha
txt.Random(20, txt.Lowercase(), txt.Exclude('l'))      // no confusable 'l'
txt.Random(6, txt.Numbers())                           // 6-digit code
txt.Random(32, txt.AlphaNum(), txt.Exclude('0','O','I','l')) // URL-safe, no confusables
```

Charset options:

| Option | Charset |
|---|---|
| `All()` | all printable ASCII (default) |
| `AlphaNum()` | letters + digits |
| `Letters()` | upper + lower letters |
| `Lowercase()` | a–z |
| `Uppercase()` | A–Z |
| `Numbers()` | 0–9 |
| `Symbols()` | `!@#$%^&*()_+-=[]{}|;:,.<>?` |
| `Chars(...)` | exactly these characters |

Modifiers:

| Option | Effect |
|---|---|
| `Include(...)` | add characters to the active charset |
| `Exclude(...)` | remove characters from the active charset (wins over `Include`) |
| `AlphaStart()` | force first character to be alphabetic |
| `RandomLength()` | return length in `[1, maxLen]` instead of exactly `maxLen` |

> ⚠️ **Not a key-derivation primitive.** `Random` is designed for user-facing random strings (codes, IDs, tokens). For cryptographic keys or long-lived secrets, use `crypto/rand` or an HKDF directly.

## 🛡️ Safety guarantees

- **Never panics on valid input.** Nil is accepted (rendered as `<nil>`), out-of-range indices clamp rather than panic, empty charsets return empty strings, unknown types fall through to `fmt.Sprintf("%v", ...)`.
- **Immutable.** `txt` never modifies input strings or values.
- **`Precision` never mutates exported constants.** `Float.Precision(2)` returns a fresh `FmtType`.
- **Unicode-safe `Substring`.** Operates on runes, returns valid UTF-8 for every valid call.
- **Cryptographic `Random`.** Backed by `crypto/rand` with rejection sampling — no modulo bias.
- **Zero dependencies.** Pure stdlib.
- **`Random` panics only on `crypto/rand.Read` failure**, which is a system-level fault, not caller input.

## 🏎️ Performance

Measured on Go 1.26 (Intel Ultra 9 275HX; library targets Go 1.21+). Format and the mutation helpers are sub-100-nanosecond. `Random` is dominated by `crypto/rand` syscalls and costs ~500ns per 16-character output.

```
BenchmarkFormat_NoArgs-24                1000000000      0.69 ns/op       0 B/op    0 allocs/op
BenchmarkFormat_SingleArg-24               16878091     76.47 ns/op      26 B/op    2 allocs/op
BenchmarkFormat_MultipleArgs-24             6149194    210.0  ns/op      80 B/op    6 allocs/op
BenchmarkFormatAs_Hex-24                   21151851     47.86 ns/op       8 B/op    1 allocs/op
BenchmarkFormatAs_FloatPrecision-24         8647272    143.9  ns/op      24 B/op    3 allocs/op
BenchmarkBetween-24                       182318179      6.28 ns/op       0 B/op    0 allocs/op
BenchmarkSquish-24                         11898037    109.8  ns/op     128 B/op    2 allocs/op
BenchmarkSubstring_ASCII-24                11810703     98.72 ns/op     176 B/op    1 allocs/op
BenchmarkSubstring_Unicode-24              18801601     68.23 ns/op       0 B/op    0 allocs/op
BenchmarkTruncate_Under-24               1000000000      0.12 ns/op       0 B/op    0 allocs/op
BenchmarkTruncate_Over-24                 182279376      6.49 ns/op       0 B/op    0 allocs/op
BenchmarkRandom_Default16-24                2487652    531.1  ns/op      80 B/op    2 allocs/op
BenchmarkRandom_AlphaNum32-24               1222848   1007    ns/op      96 B/op    2 allocs/op
BenchmarkRandom_LettersAlphaStart8-24       2965905    409.6  ns/op     192 B/op    6 allocs/op
```

`Between`, `Truncate`, and the no-arg `Format` path are zero-allocation. `Substring_Unicode` is zero-allocation when the result fits in stack-escape-analyzed bounds.

## 🧪 Testing

```bash
go test ./...                      # unit tests
go test -race ./...                # race detection
go test -bench=. -benchmem ./...   # benchmarks
```

Current coverage: 99.4%. The uncovered branches are the `crypto/rand.Read` panic path and the rejection-sampling retry loop in `randInt`, both unreachable without fault injection.

## 📚 API Reference

```go
// Format replaces each "{}" placeholder with the matching argument.
func Format(template string, args ...any) string

// FormatAs formats values using an outcome-named FmtType.
func FormatAs(f FmtType, values ...any) string

// Print writes Format's result to stdout followed by a newline.
func Print(template string, args ...any)

// FmtType is an opaque format-verb handle. Use the exported constants below.
type FmtType struct { /* opaque */ }

// Precision returns a copy of f with the given precision.
func (f FmtType) Precision(p int) FmtType

// Format verb constants.
var Binary, Octal, Hex, HexUpper, Char FmtType
var Float, Scientific, ScientificUpper FmtType
var Quoted, Unicode, Type, Pointer FmtType

// Between returns the substring of s between start and end.
func Between(s, start, end string) string

// Squish collapses whitespace runs and trims.
func Squish(s string) string

// Substring returns length runes of s starting at start (negative start
// counts from the end; out-of-range clamps).
func Substring(s string, start, length int) string

// Truncate shortens s to at most maxLen bytes, appending suffix if shortened.
func Truncate(s string, maxLen int, suffix string) string

// Random returns a cryptographically-random string up to maxLen bytes.
func Random(maxLen int, opts ...RandomOption) string

// RandomOption is the functional-option interface for Random.
type RandomOption interface { /* sealed */ }

// Charset options.
func All() RandomOption
func AlphaNum() RandomOption
func Letters() RandomOption
func Lowercase() RandomOption
func Uppercase() RandomOption
func Numbers() RandomOption
func Symbols() RandomOption
func Chars(chars ...byte) RandomOption

// Modifiers.
func Include(chars ...byte) RandomOption
func Exclude(chars ...byte) RandomOption
func AlphaStart() RandomOption
func RandomLength() RandomOption
```

## 🤝 Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md). Bold Minds Go libraries follow a shared set of design principles; read [PRINCIPLES.md](https://github.com/bold-minds/oss/blob/main/PRINCIPLES.md) before opening a PR.

## 📄 License

MIT. See [LICENSE](LICENSE).

## 🔗 Related Projects

- [`bold-minds/each`](https://github.com/bold-minds/each) — slice operations (Find, Filter, GroupBy, KeyBy, Partition, Count, Every). Same outcome-naming convention.
- [`bold-minds/list`](https://github.com/bold-minds/list) — set operations on slices (Unique, Union, Intersect, Minus, Without).
- [`bold-minds/to`](https://github.com/bold-minds/to) — safe value conversion. Pair with `txt.Format` when building messages from untyped config: `txt.Format("port {}", to.IntOr(cfg["port"], 8080))`.
- [`bold-minds/dig`](https://github.com/bold-minds/dig) — nested data navigation. Common pattern: `dig` out a leaf, `to` convert it, `txt.Format` build a message.
- Go standard library [`fmt`](https://pkg.go.dev/fmt) — the full-fidelity formatting package. `txt.FormatAs` is a thin outcome-named façade over `fmt.Sprintf`; use `fmt` directly when you need the full verb vocabulary.
- Go standard library [`strings`](https://pkg.go.dev/strings) — the mechanical foundation. `txt` layers outcome-naming on top of `strings` for the operations that otherwise take multiple lines.
- Go standard library [`crypto/rand`](https://pkg.go.dev/crypto/rand) — the entropy source used by `txt.Random`. Use `crypto/rand` directly when you need raw random bytes or high-entropy cryptographic material.

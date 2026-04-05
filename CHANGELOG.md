# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-04-05

Initial release.

### Added
- `Format(template string, args ...any) string` — `{}` placeholder string building with type-aware formatting
- `FormatAs(f FmtType, values ...any) string` — outcome-named format verbs
- `FmtType` with `Precision(p int) FmtType` method and twelve exported verb constants (`Binary`, `Octal`, `Hex`, `HexUpper`, `Char`, `Float`, `Scientific`, `ScientificUpper`, `Quoted`, `Unicode`, `Type`, `Pointer`)
- `Print(template string, args ...any)` — convenience wrapper writing `Format`'s result to stdout
- `Between(s, start, end string) string` — extract substring between delimiters, empty-anchor semantics for unset delimiters
- `Squish(s string) string` — collapse whitespace runs and trim
- `Substring(s string, start, length int) string` — rune-safe extraction with negative indexing and out-of-range clamping
- `Truncate(s string, maxLen int, suffix string) string` — byte-bounded shortening with suffix
- `Random(maxLen int, opts ...RandomOption) string` — `crypto/rand`-backed string generation with rejection sampling
- Twelve `RandomOption` constructors: `All`, `AlphaNum`, `Letters`, `Lowercase`, `Uppercase`, `Numbers`, `Symbols`, `Chars`, `Include`, `Exclude`, `AlphaStart`, `RandomLength`

### Design
- Outcome-naming convention shared with `bold-minds/each` and `bold-minds/list`
- Zero dependencies (pure stdlib)
- Go 1.21+
- Never panics on caller input; `Random` panics only on `crypto/rand.Read` system failure
- 99.4% test coverage, including adversarial tests for Unicode boundaries, empty/negative inputs, charset overrides, `Include`/`Exclude` ordering, and `AlphaStart` fall-through

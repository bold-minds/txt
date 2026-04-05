# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed — BREAKING

- **`Truncate` signature changed** from `func(s, maxLen, suffix) string` to `func(s, maxLen, suffix) (kept, removed string)`. The kept value has the same semantics as the 0.1.x return; the second return holds the portion of `s` that was dropped to make room for `suffix`. Callers that only need the truncated output can discard it with `kept, _ := txt.Truncate(...)`. The two-value return is load-bearing for callers that need to report what was cut — error messages, log lines, "show more" UI affordances. Benchmarks and examples updated accordingly.

  Migration:
  ```go
  // 0.1.x
  s := txt.Truncate(input, 80, "...")

  // 0.2.0
  s, _ := txt.Truncate(input, 80, "...")
  // or, if you need the dropped portion:
  s, dropped := txt.Truncate(input, 80, "...")
  ```

### Added

- **`Mutate(s string, opts ...MutateOption) string`** — composable string transformation pipeline. Each option is `func(string) string`, so `txt.Squish` can be passed directly (its signature matches) and user-defined closures compose naturally. Example:
  ```go
  result := txt.Mutate(input, txt.Squish, txt.TruncateOp(80, "..."))
  ```
- **`MutateOption`** type (`func(string) string`) exposed for callers who want to build their own pipeline steps.
- **`SubstringOp(start, length int) MutateOption`** — parameter-carrying wrapper around `Substring` for Mutate pipelines.
- **`TruncateOp(maxLen int, suffix string) MutateOption`** — parameter-carrying wrapper around `Truncate` that keeps the kept portion and discards `removed`. Use `Truncate` directly if you need both halves inside a pipeline.
- **`BetweenOp(start, end string) MutateOption`** — parameter-carrying wrapper around `Between` for Mutate pipelines.
- **`Print` now supports three modes**, dispatched deterministically by argument shape:
  1. **Map mode** — exactly two args, a string template and a `map[string]any`: the template's `{key}` placeholders are substituted from the map.
  2. **Format mode** — first arg is a string containing `{}` placeholders; remaining args fill them positionally (the existing v0.1.x behavior).
  3. **Multi-line mode** — anything else: each arg is printed on its own line.

  The signature changed from `Print(template string, args ...any)` to `Print(args ...any)` to accommodate multi-line and map inputs, but every v0.1.x call site continues to work because a single-string arg or a template-plus-args still dispatches to the same Format-mode path.

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

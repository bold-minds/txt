// Package txt provides outcome-named string manipulation and formatting.
//
// txt is the string-side counterpart to bold-minds/each and bold-minds/list:
// instead of the terse verbs of fmt.Sprintf it exposes ergonomic outcome-named
// formatters (Hex, Binary, Float, ...) plus direct helpers for the operations
// Go's strings package keeps behind multi-line idioms — Between, Squish,
// Substring, Truncate — and a cryptographically-random string generator with
// configurable charsets.
//
// The library is pure stdlib, zero dependencies, and never panics on valid
// input. See the README at https://github.com/bold-minds/txt for the full
// surface and examples.
package txt

import (
	"crypto/rand"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// =============================================================================
// Format — {}-placeholder string building
// =============================================================================

// Format replaces each "{}" placeholder in template with the matching argument,
// using type-aware formatting (integers as decimals, floats as %g, errors as
// "Error: <msg>", everything else via fmt's %v). Arguments are applied in
// order; extras are ignored, missing ones leave their placeholders in place.
//
// Format is a lightweight fast path for user-visible messages where the rigid
// "%d / %s / %v" vocabulary of fmt.Sprintf is noise:
//
//	txt.Format("user {} not found", userID)
//	txt.Format("failed to connect to {}:{}", host, port)
//	errors.New(txt.Format("invalid value: {}", val))
//
// For control over precision, base, or quoting, use FormatAs with the
// exported FmtType constants.
func Format(template string, args ...any) string {
	if len(args) == 0 {
		return template
	}
	// Single left-to-right scan so substituted text is never re-scanned:
	// if a previously-substituted arg contains "{}", it must not be
	// mistaken for the next placeholder.
	var b strings.Builder
	b.Grow(len(template))
	i, ai := 0, 0
	for i < len(template) {
		if ai < len(args) && i+1 < len(template) && template[i] == '{' && template[i+1] == '}' {
			b.WriteString(formatValue(args[ai]))
			ai++
			i += 2
			continue
		}
		b.WriteByte(template[i])
		i++
	}
	return b.String()
}

// formatValue renders a single argument using the type switches Format needs
// for its fast path. Kept in one place so the 16 branch cases are obvious in
// coverage output.
func formatValue(v any) string {
	switch x := v.(type) {
	case nil:
		return "<nil>"
	case string:
		return x
	case bool:
		if x {
			return "true"
		}
		return "false"
	case int:
		return strconv.FormatInt(int64(x), 10)
	case int8:
		return strconv.FormatInt(int64(x), 10)
	case int16:
		return strconv.FormatInt(int64(x), 10)
	case int32:
		return strconv.FormatInt(int64(x), 10)
	case int64:
		return strconv.FormatInt(x, 10)
	case uint:
		return strconv.FormatUint(uint64(x), 10)
	case uint8:
		return strconv.FormatUint(uint64(x), 10)
	case uint16:
		return strconv.FormatUint(uint64(x), 10)
	case uint32:
		return strconv.FormatUint(uint64(x), 10)
	case uint64:
		return strconv.FormatUint(x, 10)
	case float32:
		return strconv.FormatFloat(float64(x), 'g', -1, 32)
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64)
	case error:
		return "Error: " + x.Error()
	default:
		if rt := reflect.TypeOf(v); rt != nil && rt.Kind() == reflect.Chan {
			return fmt.Sprintf("chan %s", rt.Elem())
		}
		return fmt.Sprintf("%v", v)
	}
}

// =============================================================================
// FormatAs — outcome-named verbs
// =============================================================================

// FmtType is an opaque format-verb handle used by FormatAs. Use the exported
// constants below (Hex, Binary, Float, ...) and optionally add precision with
// the Precision method.
type FmtType struct {
	verb         string
	precision    int
	hasPrecision bool
}

// Precision returns a copy of f with the given decimal/character precision.
// The semantics match fmt: %f/%e/%E/%g/%G get that many decimal places;
// %q/%s/%x/%X take that many input characters. Negative values are clamped
// to 0 so a malformed precision never produces fmt error output like
// "%!-(float64=3)1f".
//
//	txt.FormatAs(txt.Float.Precision(2), 3.14159)   // "3.14"
//	txt.FormatAs(txt.Scientific.Precision(3), 1e10) // "1.000e+10"
func (f FmtType) Precision(p int) FmtType {
	if p < 0 {
		p = 0
	}
	return FmtType{verb: f.verb, precision: p, hasPrecision: true}
}

// Exported format verb constants. Each wraps a single fmt verb in an
// outcome-named handle so call sites read as "format AS hex" rather than
// "format with %x". These are vars (not consts) so FmtType can stay opaque.
var (
	Binary          = FmtType{verb: "b"} // %b — base 2
	Octal           = FmtType{verb: "o"} // %o — base 8
	Hex             = FmtType{verb: "x"} // %x — lowercase base 16
	HexUpper        = FmtType{verb: "X"} // %X — uppercase base 16
	Char            = FmtType{verb: "c"} // %c — rune literal
	Float           = FmtType{verb: "f"} // %f — decimal, no exponent
	Scientific      = FmtType{verb: "e"} // %e — scientific notation
	ScientificUpper = FmtType{verb: "E"} // %E — scientific (uppercase)
	Quoted          = FmtType{verb: "q"} // %q — Go-quoted string / rune
	Unicode         = FmtType{verb: "U"} // %U — U+XXXX
	Type            = FmtType{verb: "T"} // %T — Go type name
	Pointer         = FmtType{verb: "p"} // %p — pointer address
)

// FormatAs formats one or more values using f. Single values return the
// formatted string directly; multiple values are joined by a single space.
//
//	txt.FormatAs(txt.Hex, 255)                    // "ff"
//	txt.FormatAs(txt.Float.Precision(2), 3.14159) // "3.14"
//	txt.FormatAs(txt.Binary, 42)                  // "101010"
//	txt.FormatAs(txt.Hex, 1, 2, 3)                // "1 2 3"
func FormatAs(f FmtType, values ...any) string {
	if len(values) == 0 {
		return ""
	}
	var format string
	if f.hasPrecision {
		format = fmt.Sprintf("%%.%d%s", f.precision, f.verb)
	} else {
		format = "%" + f.verb
	}
	if len(values) == 1 {
		return fmt.Sprintf(format, values[0])
	}
	var b strings.Builder
	for i, v := range values {
		if i > 0 {
			b.WriteByte(' ')
		}
		fmt.Fprintf(&b, format, v)
	}
	return b.String()
}

// =============================================================================
// Print — fmt.Println with {} placeholders
// =============================================================================

// Print expands a Format template and writes it to stdout followed by a
// newline. It is a convenience on top of Format + fmt.Println — use Format
// directly when you need the string (for logs, errors, etc.) rather than
// stdout output.
//
//	txt.Print("user {} logged in", userID)
//	txt.Print("ready")
func Print(template string, args ...any) {
	fmt.Println(Format(template, args...))
}

// =============================================================================
// Between, Squish, Substring, Truncate
// =============================================================================

// Between returns the substring of s that lies between the first occurrence
// of start and the next occurrence of end after it. Returns "" if either
// delimiter is missing.
//
//	txt.Between("foo [bar] baz", "[", "]")  // "bar"
//	txt.Between("a=1&b=2", "a=", "&")       // "1"
//	txt.Between("no markers", "[", "]")     // ""
//
// Empty start anchors at the beginning of s; empty end anchors at the end:
//
//	txt.Between("prefix:value", "", ":")    // "prefix"
//	txt.Between("prefix:value", ":", "")    // "value"
func Between(s, start, end string) string {
	i := 0
	if start != "" {
		idx := strings.Index(s, start)
		if idx == -1 {
			return ""
		}
		i = idx + len(start)
	}
	if end == "" {
		return s[i:]
	}
	j := strings.Index(s[i:], end)
	if j == -1 {
		return ""
	}
	return s[i : i+j]
}

// Squish collapses every run of whitespace in s to a single space and trims
// leading and trailing whitespace. Equivalent to strings.Join(strings.Fields(s), " ")
// but outcome-named.
//
//	txt.Squish("  hello   world  ")  // "hello world"
//	txt.Squish("\tfoo\n\nbar")       // "foo bar"
//	txt.Squish("   ")                // ""
func Squish(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

// Substring returns length runes of s starting at index start. Negative
// start counts from the end. Out-of-range indices clamp to the string
// boundaries rather than panicking, and length <= 0 returns "".
//
//	txt.Substring("hello", 0, 3)    // "hel"
//	txt.Substring("hello", -2, 2)   // "lo"
//	txt.Substring("héllo", 1, 3)    // "éll" — counts runes, not bytes
//	txt.Substring("hi", 5, 10)      // "" — start past end, no panic
//	txt.Substring("anything", 0, 0) // "" — zero length
func Substring(s string, start, length int) string {
	if length <= 0 {
		return ""
	}
	runes := []rune(s)
	n := len(runes)
	if n == 0 {
		return ""
	}
	if start < 0 {
		start += n
	}
	if start < 0 {
		start = 0
	}
	if start >= n {
		return ""
	}
	end := start + length
	if end > n || end < start { // end < start guards int overflow on huge length
		end = n
	}
	return string(runes[start:end])
}

// Truncate shortens s to at most maxLen bytes, appending suffix if s was
// actually shortened. The returned string's byte length is guaranteed to
// be <= maxLen. If maxLen <= len(suffix), a prefix of suffix of that length
// is returned (so Truncate("hello", 2, "...") → "..").
//
//	txt.Truncate("Hello world", 8, "...")  // "Hello..."
//	txt.Truncate("short", 20, "...")       // "short"
//	txt.Truncate("abcdef", 2, "...")       // ".."
//	txt.Truncate("anything", -1, "...")    // ""
//
// WARNING: Truncate operates on bytes, not runes. Calling it on a string
// containing multibyte UTF-8 characters can cut mid-sequence and produce
// an invalid-UTF-8 result (e.g. Truncate("héllo", 2, "") → "h\xc3"). For
// UTF-8 safety, bound the input with Substring first.
func Truncate(s string, maxLen int, suffix string) string {
	if maxLen < 0 {
		return ""
	}
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= len(suffix) {
		return suffix[:maxLen]
	}
	return s[:maxLen-len(suffix)] + suffix
}

// =============================================================================
// Random — cryptographically-random strings with configurable charsets
// =============================================================================

// Predefined charsets used by the option constructors below.
const (
	charsLower   = "abcdefghijklmnopqrstuvwxyz"
	charsUpper   = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	charsDigit   = "0123456789"
	charsSymbol  = "!@#$%^&*()_+-=[]{}|;:,.<>?"
	charsDefault = charsLower + charsUpper + charsDigit + charsSymbol
)

// randomConfig is the internal configuration built up by applying
// RandomOption values. Finalization (Include/Exclude pass) happens in Random.
type randomConfig struct {
	charset      string
	randomLength bool
	alphaStart   bool
	excluded     map[byte]struct{}
	included     []byte
}

// RandomOption is the functional-option interface accepted by Random.
type RandomOption interface {
	apply(c *randomConfig)
}

// Random returns a cryptographically-random string of up to maxLen bytes,
// drawn from a charset configured by the options. With no options it uses
// all printable ASCII characters (letters + digits + symbols).
//
//	txt.Random(16)                                       // 16 chars, full printable ASCII
//	txt.Random(8, txt.AlphaNum())                        // 8 alphanumeric
//	txt.Random(12, txt.Letters(), txt.AlphaStart())      // 12 letters, first is alpha
//	txt.Random(20, txt.Lowercase(), txt.Exclude('l'))    // 20 lowercase minus 'l'
//
// Random uses crypto/rand for entropy, making it safe for non-key secrets
// such as invite codes, correlation IDs, or test fixtures. It is NOT a
// replacement for key derivation or high-entropy cryptographic material —
// use crypto/rand or hkdf directly for that.
//
// Random panics only if crypto/rand.Read itself fails (an extraordinary
// system-level error), never on caller input.
func Random(maxLen int, opts ...RandomOption) string {
	if maxLen <= 0 {
		return ""
	}
	c := &randomConfig{charset: charsDefault}
	for _, o := range opts {
		o.apply(c)
	}
	c.finalizeCharset()
	if len(c.charset) == 0 {
		return ""
	}

	length := maxLen
	if c.randomLength {
		length = randInt(maxLen) + 1 // 1..maxLen inclusive
	}

	out := make([]byte, length)
	if c.alphaStart {
		alpha := filterAlpha(c.charset)
		if len(alpha) > 0 {
			out[0] = alpha[randInt(len(alpha))]
			for i := 1; i < length; i++ {
				out[i] = c.charset[randInt(len(c.charset))]
			}
			return string(out)
		}
		// No alpha characters in the active charset — silently fall through
		// to uniform fill rather than returning "" or panicking.
	}
	for i := range out {
		out[i] = c.charset[randInt(len(c.charset))]
	}
	return string(out)
}

// finalizeCharset applies Include then Exclude in a second pass so that the
// two options work regardless of the order they appear in the option list
// relative to the base charset. Include runs before Exclude so a later
// Exclude can remove characters that Include added.
func (c *randomConfig) finalizeCharset() {
	if len(c.included) > 0 {
		seen := make(map[byte]struct{}, len(c.charset))
		for i := 0; i < len(c.charset); i++ {
			seen[c.charset[i]] = struct{}{}
		}
		var b strings.Builder
		b.WriteString(c.charset)
		for _, ch := range c.included {
			if _, ok := seen[ch]; !ok {
				b.WriteByte(ch)
				seen[ch] = struct{}{}
			}
		}
		c.charset = b.String()
	}
	if len(c.excluded) > 0 {
		var b strings.Builder
		for i := 0; i < len(c.charset); i++ {
			if _, bad := c.excluded[c.charset[i]]; !bad {
				b.WriteByte(c.charset[i])
			}
		}
		c.charset = b.String()
	}
}

// randInt returns a uniformly-random integer in [0, n) using crypto/rand,
// with rejection sampling to avoid modulo bias. Panics on crypto/rand.Read
// failure (see Random for rationale).
func randInt(n int) int {
	if n <= 0 {
		return 0
	}
	const maxUint64 = ^uint64(0)
	limit := maxUint64 - (maxUint64 % uint64(n))
	var buf [8]byte
	for {
		if _, err := rand.Read(buf[:]); err != nil {
			panic("txt: crypto/rand.Read failed: " + err.Error())
		}
		v := uint64(buf[0])<<56 | uint64(buf[1])<<48 | uint64(buf[2])<<40 | uint64(buf[3])<<32 |
			uint64(buf[4])<<24 | uint64(buf[5])<<16 | uint64(buf[6])<<8 | uint64(buf[7])
		if v < limit {
			// v%uint64(n) is strictly < uint64(n), and n was passed in as a
			// positive int — so the result fits in int by construction.
			return int(v % uint64(n)) //nolint:gosec // G115: bounded by n which is already int
		}
	}
}

// filterAlpha returns the subset of s that consists of ASCII letters.
func filterAlpha(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
			b.WriteByte(c)
		}
	}
	return b.String()
}

// -----------------------------------------------------------------------------
// Random option constructors
// -----------------------------------------------------------------------------

type charsetOpt struct{ charset string }

func (o charsetOpt) apply(c *randomConfig) { c.charset = o.charset }

type alphaStartOpt struct{}

func (alphaStartOpt) apply(c *randomConfig) { c.alphaStart = true }

type randomLengthOpt struct{}

func (randomLengthOpt) apply(c *randomConfig) { c.randomLength = true }

type excludeOpt struct{ chars []byte }

func (o excludeOpt) apply(c *randomConfig) {
	if c.excluded == nil {
		c.excluded = make(map[byte]struct{}, len(o.chars))
	}
	for _, b := range o.chars {
		c.excluded[b] = struct{}{}
	}
}

type includeOpt struct{ chars []byte }

func (o includeOpt) apply(c *randomConfig) {
	c.included = append(c.included, o.chars...)
}

// All uses the full printable-ASCII charset (letters + digits + symbols).
// This is the default when no charset option is provided.
func All() RandomOption { return charsetOpt{charset: charsDefault} }

// AlphaNum uses upper + lower case letters and digits.
func AlphaNum() RandomOption { return charsetOpt{charset: charsLower + charsUpper + charsDigit} }

// Letters uses upper + lower case letters.
func Letters() RandomOption { return charsetOpt{charset: charsLower + charsUpper} }

// Lowercase uses lowercase letters only.
func Lowercase() RandomOption { return charsetOpt{charset: charsLower} }

// Uppercase uses uppercase letters only.
func Uppercase() RandomOption { return charsetOpt{charset: charsUpper} }

// Numbers uses decimal digits only.
func Numbers() RandomOption { return charsetOpt{charset: charsDigit} }

// Symbols uses the standard printable ASCII symbol set.
func Symbols() RandomOption { return charsetOpt{charset: charsSymbol} }

// Chars uses exactly the provided characters as the charset. Duplicate
// bytes are removed so each character has equal selection probability —
// Chars('a', 'a', 'b') is identical to Chars('a', 'b').
func Chars(chars ...byte) RandomOption {
	seen := make(map[byte]struct{}, len(chars))
	var b strings.Builder
	b.Grow(len(chars))
	for _, ch := range chars {
		if _, ok := seen[ch]; ok {
			continue
		}
		seen[ch] = struct{}{}
		b.WriteByte(ch)
	}
	return charsetOpt{charset: b.String()}
}

// Include adds the given characters to the active charset. Applied after
// any charset option so additions are not overwritten.
func Include(chars ...byte) RandomOption { return includeOpt{chars: chars} }

// Exclude removes the given characters from the active charset. Applied
// after both the charset option and Include, so removals always stick.
func Exclude(chars ...byte) RandomOption { return excludeOpt{chars: chars} }

// AlphaStart forces the first character of the output to be alphabetic,
// independent of the rest of the charset. If the charset has no alphabetic
// characters this option is silently ignored.
func AlphaStart() RandomOption { return alphaStartOpt{} }

// RandomLength makes Random return a string of random length in [1, maxLen]
// rather than exactly maxLen characters.
func RandomLength() RandomOption { return randomLengthOpt{} }

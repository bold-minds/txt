package txt

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"unicode/utf8"
)

// =============================================================================
// Format
// =============================================================================

func TestFormat_NoArgs(t *testing.T) {
	if got := Format("no args here"); got != "no args here" {
		t.Errorf("Format: got %q, want unchanged template", got)
	}
}

func TestFormat_SingleArg(t *testing.T) {
	if got := Format("user {} logged in", 42); got != "user 42 logged in" {
		t.Errorf("Format: got %q", got)
	}
}

func TestFormat_MultipleArgs(t *testing.T) {
	got := Format("{}:{} from {}", "host", 8080, "admin")
	if got != "host:8080 from admin" {
		t.Errorf("Format: got %q", got)
	}
}

func TestFormat_ExtraArgsIgnored(t *testing.T) {
	// Two placeholders, three args — the third is silently dropped.
	got := Format("{} and {}", "a", "b", "c")
	if got != "a and b" {
		t.Errorf("Format: got %q, extras should be ignored", got)
	}
}

func TestFormat_MissingArgsLeavePlaceholders(t *testing.T) {
	// Three placeholders, one arg — remaining placeholders are left as-is
	// so the bug is visible to whoever reads the log line.
	got := Format("{} {} {}", "only")
	if got != "only {} {}" {
		t.Errorf("Format: got %q", got)
	}
}

func TestFormat_AllTypedArgs(t *testing.T) {
	cases := []struct {
		name string
		arg  any
		want string
	}{
		{"string", "hi", "hi"},
		{"bool-true", true, "true"},
		{"bool-false", false, "false"},
		{"int", int(-5), "-5"},
		{"int8", int8(-8), "-8"},
		{"int16", int16(-16), "-16"},
		{"int32", int32(-32), "-32"},
		{"int64", int64(-64), "-64"},
		{"uint", uint(5), "5"},
		{"uint8", uint8(8), "8"},
		{"uint16", uint16(16), "16"},
		{"uint32", uint32(32), "32"},
		{"uint64", uint64(64), "64"},
		{"float32", float32(1.5), "1.5"},
		{"float64", float64(2.5), "2.5"},
		{"error", errors.New("boom"), "Error: boom"},
		{"nil", nil, "<nil>"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Format("{}", tc.arg); got != tc.want {
				t.Errorf("Format(%v): got %q, want %q", tc.arg, got, tc.want)
			}
		})
	}
}

func TestFormat_ChannelType(t *testing.T) {
	ch := make(chan int)
	got := Format("{}", ch)
	if !strings.HasPrefix(got, "chan int") {
		t.Errorf("Format(chan int): got %q, want prefix \"chan int\"", got)
	}
}

// TestFormat_SubstitutedArgIsNotRescanned pins the single-pass scan: once an
// argument has been substituted into the output, any "{}" sequence inside
// that rendered value must NOT be treated as another placeholder. A previous
// strings.Replace-based implementation failed this: the next arg would land
// inside the first arg's text.
func TestFormat_SubstitutedArgIsNotRescanned(t *testing.T) {
	cases := []struct {
		name string
		tmpl string
		args []any
		want string
	}{
		{"literal-{}-in-arg", "a={} b={}", []any{"{}", "X"}, "a={} b=X"},
		{"embedded-{}-in-arg", "{} + {}", []any{"a{}b", "c"}, "a{}b + c"},
		{"nested-braces", "x={} y={}", []any{"{{}}", "Z"}, "x={{}} y=Z"},
		{"arg-is-bare-placeholder", "{}-{}", []any{"{}", "end"}, "{}-end"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Format(tc.tmpl, tc.args...); got != tc.want {
				t.Errorf("Format(%q, %v): got %q, want %q", tc.tmpl, tc.args, got, tc.want)
			}
		})
	}
}

func TestFormat_StructFallback(t *testing.T) {
	// Unknown types route through fmt.Sprintf("%v", ...).
	type point struct{ X, Y int }
	got := Format("{}", point{1, 2})
	if got != "{1 2}" {
		t.Errorf("Format(struct): got %q", got)
	}
}

// =============================================================================
// FormatAs + FmtType
// =============================================================================

func TestFormatAs_ZeroValues(t *testing.T) {
	if got := FormatAs(Hex); got != "" {
		t.Errorf("FormatAs with no values: got %q, want \"\"", got)
	}
}

func TestFormatAs_Single(t *testing.T) {
	cases := []struct {
		name string
		f    FmtType
		v    any
		want string
	}{
		{"binary", Binary, 42, "101010"},
		{"octal", Octal, 8, "10"},
		{"hex", Hex, 255, "ff"},
		{"hex-upper", HexUpper, 255, "FF"},
		{"char", Char, 'A', "A"},
		{"float", Float, 3.14, "3.140000"},
		{"sci", Scientific, 1000.0, "1.000000e+03"},
		{"sci-upper", ScientificUpper, 1000.0, "1.000000E+03"},
		{"quoted", Quoted, "hi", `"hi"`},
		{"unicode", Unicode, 'A', "U+0041"},
		{"type", Type, 42, "int"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatAs(tc.f, tc.v); got != tc.want {
				t.Errorf("FormatAs(%s): got %q, want %q", tc.name, got, tc.want)
			}
		})
	}
}

func TestFormatAs_Pointer(t *testing.T) {
	// Pointer addresses are machine-dependent; just verify it produced a
	// 0x-prefixed hex string rather than asserting a specific value.
	x := 1
	got := FormatAs(Pointer, &x)
	if !strings.HasPrefix(got, "0x") {
		t.Errorf("FormatAs(Pointer): got %q, want 0x-prefixed address", got)
	}
}

func TestFormatAs_Precision(t *testing.T) {
	if got := FormatAs(Float.Precision(2), 3.14159); got != "3.14" {
		t.Errorf("FormatAs(Float.Precision(2)): got %q", got)
	}
	if got := FormatAs(Scientific.Precision(3), 1e10); got != "1.000e+10" {
		t.Errorf("FormatAs(Scientific.Precision(3)): got %q", got)
	}
}

// TestFormatAs_PrecisionNegativeClamped pins the clamping behavior for
// negative precision. A previous version let negative values flow into the
// format string, producing fmt error output like "%!-(float64=3)1f" that
// a caller would then log or show to a user.
func TestFormatAs_PrecisionNegativeClamped(t *testing.T) {
	if got := FormatAs(Float.Precision(-1), 3.14); got != "3" {
		t.Errorf("FormatAs(Float.Precision(-1), 3.14): got %q, want %q", got, "3")
	}
	if got := FormatAs(Float.Precision(-100), 2.718); got != "3" {
		t.Errorf("FormatAs(Float.Precision(-100), 2.718): got %q, want %q", got, "3")
	}
}

func TestFormatAs_PrecisionDoesNotMutateOriginal(t *testing.T) {
	// Precision must return a fresh FmtType, not mutate the package-level
	// Float constant. If it did, FormatAs(Float, x) elsewhere would start
	// carrying precision.
	_ = Float.Precision(5)
	if Float.hasPrecision {
		t.Errorf("Float.Precision mutated the original Float constant")
	}
}

func TestFormatAs_Multiple(t *testing.T) {
	if got := FormatAs(Hex, 1, 2, 3); got != "1 2 3" {
		t.Errorf("FormatAs(Hex, 1, 2, 3): got %q", got)
	}
	if got := FormatAs(Float.Precision(1), 1.23, 4.56); got != "1.2 4.6" {
		t.Errorf("FormatAs(Float.Precision(1), ...): got %q", got)
	}
}

// =============================================================================
// Print
// =============================================================================

func TestPrint_WritesToStdout(t *testing.T) {
	// Swap os.Stdout so we can capture Print's output deterministically.
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w

	Print("value {}", 42)

	if cerr := w.Close(); cerr != nil {
		t.Fatalf("w.Close: %v", cerr)
	}
	os.Stdout = orig

	var buf bytes.Buffer
	if _, cerr := io.Copy(&buf, r); cerr != nil {
		t.Fatalf("io.Copy: %v", cerr)
	}
	if got := buf.String(); got != "value 42\n" {
		t.Errorf("Print: got %q, want \"value 42\\n\"", got)
	}
}

// =============================================================================
// Between
// =============================================================================

func TestBetween(t *testing.T) {
	cases := []struct {
		name       string
		s, a, b    string
		want       string
	}{
		{"basic-brackets", "foo [bar] baz", "[", "]", "bar"},
		{"query-param", "a=1&b=2", "a=", "&", "1"},
		{"missing-start", "no markers", "[", "]", ""},
		{"missing-end", "has [start only", "[", "]", ""},
		{"empty-start-anchor", "prefix:value", "", ":", "prefix"},
		{"empty-end-anchor", "prefix:value", ":", "", "value"},
		{"both-empty", "whole string", "", "", "whole string"},
		{"adjacent-delimiters", "[]", "[", "]", ""},
		{"multichar-delimiters", "BEGIN content END", "BEGIN ", " END", "content"},
		{"start-appears-in-content", "xxfooxxbarxx", "xx", "xx", "foo"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Between(tc.s, tc.a, tc.b); got != tc.want {
				t.Errorf("Between(%q, %q, %q): got %q, want %q", tc.s, tc.a, tc.b, got, tc.want)
			}
		})
	}
}

// =============================================================================
// Squish
// =============================================================================

func TestSquish(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"leading-trailing", "  hello   world  ", "hello world"},
		{"tabs-newlines", "\tfoo\n\nbar", "foo bar"},
		{"all-whitespace", "   \t\n  ", ""},
		{"empty", "", ""},
		{"single-word", "hello", "hello"},
		{"already-normalized", "a b c", "a b c"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Squish(tc.in); got != tc.want {
				t.Errorf("Squish(%q): got %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

// =============================================================================
// Substring
// =============================================================================

func TestSubstring(t *testing.T) {
	cases := []struct {
		name           string
		s              string
		start, length  int
		want           string
	}{
		{"basic", "hello", 0, 3, "hel"},
		{"middle", "hello", 1, 3, "ell"},
		{"end", "hello", 2, 10, "llo"},
		{"negative-start", "hello", -2, 2, "lo"},
		{"negative-start-clamped", "hello", -100, 3, "hel"},
		{"start-past-end", "hi", 5, 10, ""},
		{"zero-length", "anything", 0, 0, ""},
		{"negative-length", "hello", 0, -1, ""},
		{"empty-string", "", 0, 5, ""},
		{"unicode-runes", "héllo", 1, 3, "éll"},
		{"unicode-all", "αβγδ", 0, 4, "αβγδ"},
		{"length-overflow", "hello", 1, int(^uint(0) >> 1), "ello"}, // max int
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Substring(tc.s, tc.start, tc.length); got != tc.want {
				t.Errorf("Substring(%q, %d, %d): got %q, want %q",
					tc.s, tc.start, tc.length, got, tc.want)
			}
		})
	}
}

func TestSubstring_UnicodeSafeBoundaries(t *testing.T) {
	// Every returned substring must still be valid UTF-8 — regression guard
	// against a future byte-slice optimization that slices mid-rune.
	for _, s := range []string{"héllo", "αβγ", "日本語", "🚀rocket"} {
		for start := -5; start <= 5; start++ {
			for length := 0; length <= 6; length++ {
				got := Substring(s, start, length)
				if !utf8.ValidString(got) {
					t.Errorf("Substring(%q, %d, %d) returned invalid UTF-8: %q",
						s, start, length, got)
				}
			}
		}
	}
}

// =============================================================================
// Truncate
// =============================================================================

func TestTruncate(t *testing.T) {
	cases := []struct {
		name        string
		s           string
		maxLen      int
		suffix      string
		wantKept    string
		wantRemoved string
	}{
		{"under-limit", "short", 20, "...", "short", ""},
		{"exact-limit", "hello", 5, "...", "hello", ""},
		{"truncates-with-suffix", "Hello world", 8, "...", "Hello...", " world"},
		{"suffix-fills-maxlen", "abcdef", 2, "...", "..", "abcdef"},
		{"suffix-exact-maxlen", "abcdef", 3, "...", "...", "abcdef"},
		{"empty-suffix", "Hello world", 5, "", "Hello", " world"},
		{"negative-maxlen", "anything", -1, "...", "", "anything"},
		{"zero-maxlen", "anything", 0, "...", "", "anything"},
		{"empty-string", "", 10, "...", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			kept, removed := Truncate(tc.s, tc.maxLen, tc.suffix)
			if kept != tc.wantKept {
				t.Errorf("Truncate(%q, %d, %q): kept = %q, want %q",
					tc.s, tc.maxLen, tc.suffix, kept, tc.wantKept)
			}
			if removed != tc.wantRemoved {
				t.Errorf("Truncate(%q, %d, %q): removed = %q, want %q",
					tc.s, tc.maxLen, tc.suffix, removed, tc.wantRemoved)
			}
			// Byte-length guarantee: kept <= maxLen when maxLen >= 0.
			if tc.maxLen >= 0 && len(kept) > tc.maxLen {
				t.Errorf("Truncate(%q, %d, %q): kept %q len %d > maxLen",
					tc.s, tc.maxLen, tc.suffix, kept, len(kept))
			}
		})
	}
}

// =============================================================================
// Mutate
// =============================================================================

func TestMutate_Empty(t *testing.T) {
	// No options → input returned unchanged.
	if got := Mutate("hello"); got != "hello" {
		t.Errorf("Mutate(\"hello\") with no options = %q, want %q", got, "hello")
	}
}

func TestMutate_SingleOption(t *testing.T) {
	// Squish matches the func(string) string signature directly —
	// it can be passed as a MutateOption without a closure.
	if got := Mutate("  hello   world  ", Squish); got != "hello world" {
		t.Errorf("Mutate(Squish) = %q, want %q", got, "hello world")
	}
}

func TestMutate_Pipeline(t *testing.T) {
	// Squish of this input produces "Hello world" (11 bytes). A
	// TruncateOp budget of 11 triggers the under-limit path and
	// returns the string unchanged — verifies that options compose
	// without an off-by-one at the equality boundary.
	short := "  Hello  world  "
	if got := Mutate(short, Squish, TruncateOp(11, "...")); got != "Hello world" {
		t.Errorf("Mutate at exact boundary = %q, want %q", got, "Hello world")
	}

	// Longer input: squish to a canonical form, then truncate with
	// ellipsis. Exercises the full pipeline and the cut-with-suffix
	// path of Truncate via TruncateOp.
	long := "   The   quick   brown   fox   "
	got := Mutate(long, Squish, TruncateOp(11, "..."))
	// Squish → "The quick brown fox" (19 chars). TruncateOp(11, "...")
	// → cut = 11 - 3 = 8, kept = "The quic" + "..." = "The quic..."
	if got != "The quic..." {
		t.Errorf("Mutate pipeline = %q, want %q", got, "The quic...")
	}
}

func TestMutate_SubstringOp(t *testing.T) {
	got := Mutate("Hello World", SubstringOp(0, 5))
	if got != "Hello" {
		t.Errorf("Mutate(SubstringOp(0,5)) = %q, want %q", got, "Hello")
	}
}

func TestMutate_BetweenOp(t *testing.T) {
	got := Mutate("prefix [target] suffix", BetweenOp("[", "]"))
	if got != "target" {
		t.Errorf("Mutate(BetweenOp) = %q, want %q", got, "target")
	}
}

func TestMutate_OrderMatters(t *testing.T) {
	// Substring before Truncate vs Truncate before Substring — the
	// intermediate result changes the final output. ASCII "..." is
	// used throughout to avoid the multi-byte-ellipsis byte-budget
	// interaction with Truncate.
	in := "Hello, World!"

	// Substring(in, 0, 5) = "Hello" (5 bytes).
	// TruncateOp(4, "...") on "Hello": len=5 > 4, maxLen(4) > len(suffix)(3),
	// cut = 4-3 = 1, kept = "H" + "..." = "H..." (4 bytes).
	a := Mutate(in, SubstringOp(0, 5), TruncateOp(4, "..."))
	if a != "H..." {
		t.Errorf("Mutate(Substring, Truncate) = %q, want %q", a, "H...")
	}

	// TruncateOp(8, "...") on in ("Hello, World!", 13 bytes):
	// cut = 5, kept = "Hello" + "..." = "Hello..." (8 bytes).
	// Then SubstringOp(0, 5) takes first 5 runes = "Hello".
	b := Mutate(in, TruncateOp(8, "..."), SubstringOp(0, 5))
	if b != "Hello" {
		t.Errorf("Mutate(Truncate, Substring) = %q, want %q", b, "Hello")
	}
}

// =============================================================================
// Random — charsets, options, adversarial
// =============================================================================

func TestRandom_ZeroLength(t *testing.T) {
	if got := Random(0); got != "" {
		t.Errorf("Random(0): got %q, want \"\"", got)
	}
	if got := Random(-5); got != "" {
		t.Errorf("Random(-5): got %q, want \"\"", got)
	}
}

func TestRandom_DefaultCharset(t *testing.T) {
	out := Random(100)
	if len(out) != 100 {
		t.Fatalf("Random(100): len %d", len(out))
	}
	for i, b := range []byte(out) {
		if b < 33 || b > 126 { // printable ASCII, excluding space
			t.Errorf("Random default: byte %d at index %d is not printable ASCII", b, i)
		}
	}
}

func TestRandom_CharsetOptions(t *testing.T) {
	cases := []struct {
		name  string
		opt   RandomOption
		check func(byte) bool
	}{
		{"lowercase", Lowercase(), func(b byte) bool { return b >= 'a' && b <= 'z' }},
		{"uppercase", Uppercase(), func(b byte) bool { return b >= 'A' && b <= 'Z' }},
		{"numbers", Numbers(), func(b byte) bool { return b >= '0' && b <= '9' }},
		{"letters", Letters(), func(b byte) bool {
			return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
		}},
		{"alphanum", AlphaNum(), func(b byte) bool {
			return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9')
		}},
		{"symbols", Symbols(), func(b byte) bool {
			return strings.ContainsRune(charsSymbol, rune(b))
		}},
		{"all", All(), func(b byte) bool { return b >= 33 && b <= 126 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			out := Random(200, tc.opt)
			for i, b := range []byte(out) {
				if !tc.check(b) {
					t.Errorf("Random(%s) byte %d at index %d fails charset check (%q)",
						tc.name, b, i, out)
					return
				}
			}
		})
	}
}

func TestRandom_Chars(t *testing.T) {
	out := Random(50, Chars('a', 'b', 'c'))
	for _, b := range []byte(out) {
		if b != 'a' && b != 'b' && b != 'c' {
			t.Errorf("Random(Chars): unexpected byte %q in %q", b, out)
			return
		}
	}
}

// TestChars_Dedup pins that duplicate bytes passed to Chars do not skew the
// distribution. We draw a large sample from Chars('a','a','b') and expect
// roughly equal counts of 'a' and 'b'; the pre-dedup implementation would
// have given 'a' about twice the probability of 'b'.
func TestChars_Dedup(t *testing.T) {
	const n = 4000
	out := Random(n, Chars('a', 'a', 'b'))
	var aCount, bCount int
	for _, c := range []byte(out) {
		switch c {
		case 'a':
			aCount++
		case 'b':
			bCount++
		default:
			t.Fatalf("Chars dedup: unexpected byte %q", c)
		}
	}
	// Equal probability → expected ratio 1.0. Allow a wide 0.85–1.15 band
	// so crypto/rand jitter never flakes this, but the pre-dedup ratio of
	// ~2.0 is well outside it.
	ratio := float64(aCount) / float64(bCount)
	if ratio < 0.85 || ratio > 1.15 {
		t.Errorf("Chars dedup: a/b ratio = %.2f (a=%d, b=%d), want near 1.0",
			ratio, aCount, bCount)
	}
}

func TestRandom_Exclude(t *testing.T) {
	out := Random(200, Lowercase(), Exclude('a', 'e', 'i', 'o', 'u'))
	for _, b := range []byte(out) {
		if strings.ContainsRune("aeiou", rune(b)) {
			t.Errorf("Random(Exclude vowels): output %q contained %q", out, b)
			return
		}
	}
}

func TestRandom_Include(t *testing.T) {
	// Include adds to the existing charset — with Numbers() + Include('a'),
	// 'a' should appear in a long-enough output.
	seen := false
	for i := 0; i < 5 && !seen; i++ {
		out := Random(500, Numbers(), Include('a'))
		if strings.ContainsRune(out, 'a') {
			seen = true
		}
	}
	if !seen {
		t.Errorf("Random(Numbers + Include('a')): 'a' never appeared across 5 runs of len 500")
	}
}

func TestRandom_IncludeExcludeOrdering(t *testing.T) {
	// Exclude must beat Include, because the internal pass runs Include
	// first then Exclude. If we Include('a') and Exclude('a') the output
	// must not contain 'a'.
	out := Random(500, Numbers(), Include('a'), Exclude('a'))
	if strings.ContainsRune(out, 'a') {
		t.Errorf("Random: Exclude should win over Include, got %q", out)
	}
}

func TestRandom_IncludeDoesNotDuplicate(t *testing.T) {
	// Including 'a' into a charset that already contains 'a' must not
	// change the character distribution — there's no test for distribution
	// directly, but we can verify that the charset construction path runs
	// without error and still produces only lowercase letters.
	out := Random(100, Lowercase(), Include('a'))
	for _, b := range []byte(out) {
		if b < 'a' || b > 'z' {
			t.Errorf("Random(Lowercase + Include('a')): unexpected byte %q", b)
			return
		}
	}
}

func TestRandom_AlphaStart(t *testing.T) {
	for i := 0; i < 20; i++ {
		out := Random(8, AlphaNum(), AlphaStart())
		if len(out) == 0 {
			t.Fatalf("Random empty")
		}
		first := out[0]
		isAlpha := (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')
		if !isAlpha {
			t.Errorf("AlphaStart: first char %q is not alphabetic (iter %d, output %q)",
				first, i, out)
			return
		}
	}
}

func TestRandom_AlphaStartNoAlphaFallsThrough(t *testing.T) {
	// AlphaStart with a digits-only charset has no alpha to choose from;
	// it must NOT panic and must NOT return "" — silently fall through to
	// uniform fill.
	out := Random(10, Numbers(), AlphaStart())
	if len(out) != 10 {
		t.Errorf("AlphaStart-no-alpha: got len %d, want 10", len(out))
	}
	for _, b := range []byte(out) {
		if b < '0' || b > '9' {
			t.Errorf("AlphaStart-no-alpha: unexpected byte %q", b)
			return
		}
	}
}

func TestRandom_EmptyCharsetReturnsEmpty(t *testing.T) {
	// Exclude every character of a single-char charset → empty charset →
	// empty return (no panic).
	out := Random(10, Chars('a'), Exclude('a'))
	if out != "" {
		t.Errorf("Random with empty charset: got %q, want \"\"", out)
	}
}

func TestRandom_RandomLength(t *testing.T) {
	// With RandomLength the returned length must be in [1, maxLen].
	seenLengths := make(map[int]bool)
	for i := 0; i < 200; i++ {
		out := Random(16, RandomLength())
		if len(out) < 1 || len(out) > 16 {
			t.Fatalf("RandomLength: got len %d, want [1,16]", len(out))
		}
		seenLengths[len(out)] = true
	}
	// Should have seen a reasonable spread of lengths in 200 iterations.
	if len(seenLengths) < 5 {
		t.Errorf("RandomLength: only saw %d distinct lengths across 200 iters (suspicious)",
			len(seenLengths))
	}
}

func TestRandom_DistinctOutputs(t *testing.T) {
	// Two large-enough Random() calls should never collide in practice.
	// This is a smoke test against future bugs that would e.g. hardcode a
	// seed or cache the first output.
	a := Random(32)
	b := Random(32)
	if a == b {
		t.Errorf("Random: two 32-char outputs collided: %q", a)
	}
}

func TestRandom_ChainedOptions(t *testing.T) {
	// Later charset option should win — Lowercase() then Uppercase() must
	// produce only uppercase.
	out := Random(100, Lowercase(), Uppercase())
	for _, b := range []byte(out) {
		if b < 'A' || b > 'Z' {
			t.Errorf("Lowercase then Uppercase: got %q in output", b)
			return
		}
	}
}

// =============================================================================
// randInt — internal uniformity smoke test
// =============================================================================

func TestRandInt_Uniform(t *testing.T) {
	// Chi-squared would be more rigorous, but a simple "every bucket sees
	// some hits" check catches obvious bias bugs. 10 buckets, 5000 samples.
	const buckets = 10
	const samples = 5000
	counts := make([]int, buckets)
	for i := 0; i < samples; i++ {
		counts[randInt(buckets)]++
	}
	for i, c := range counts {
		if c == 0 {
			t.Errorf("randInt: bucket %d was never hit across %d samples", i, samples)
		}
		// Each bucket should see ~500 hits; tolerate ±60% to avoid flakiness.
		if c < 200 || c > 800 {
			t.Errorf("randInt: bucket %d had %d hits (expected ~500)", i, c)
		}
	}
}

func TestRandInt_ZeroAndNegative(t *testing.T) {
	if randInt(0) != 0 {
		t.Errorf("randInt(0): must return 0")
	}
	if randInt(-5) != 0 {
		t.Errorf("randInt(-5): must return 0")
	}
}

func TestRandInt_Bounded(t *testing.T) {
	for i := 0; i < 1000; i++ {
		v := randInt(7)
		if v < 0 || v >= 7 {
			t.Errorf("randInt(7) = %d out of range", v)
		}
	}
}

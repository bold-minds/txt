package txt

import "testing"

// Benchmarks cover the hot paths: string formatting, the stdlib-style
// mutation helpers (where a concise helper beats a multi-line idiom), and
// Random at the sizes typical for invite codes and correlation IDs.

// -----------------------------------------------------------------------------
// Format
// -----------------------------------------------------------------------------

func BenchmarkFormat_NoArgs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Format("no placeholders here")
	}
}

func BenchmarkFormat_SingleArg(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Format("user {} logged in", 42)
	}
}

func BenchmarkFormat_MultipleArgs(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Format("{}:{}/{}?{}", "host", 8080, "api", "v1")
	}
}

// -----------------------------------------------------------------------------
// FormatAs
// -----------------------------------------------------------------------------

func BenchmarkFormatAs_Hex(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = FormatAs(Hex, 0xdeadbeef)
	}
}

func BenchmarkFormatAs_FloatPrecision(b *testing.B) {
	b.ReportAllocs()
	f := Float.Precision(2)
	for i := 0; i < b.N; i++ {
		_ = FormatAs(f, 3.14159)
	}
}

// -----------------------------------------------------------------------------
// Between / Squish / Substring / Truncate
// -----------------------------------------------------------------------------

func BenchmarkBetween(b *testing.B) {
	b.ReportAllocs()
	s := "GET /api/users/42/profile?format=json HTTP/1.1"
	for i := 0; i < b.N; i++ {
		_ = Between(s, "/users/", "/")
	}
}

func BenchmarkSquish(b *testing.B) {
	b.ReportAllocs()
	s := "  the   quick\tbrown\n\nfox  jumps   over  "
	for i := 0; i < b.N; i++ {
		_ = Squish(s)
	}
}

func BenchmarkSubstring_ASCII(b *testing.B) {
	b.ReportAllocs()
	s := "The quick brown fox jumps over the lazy dog"
	for i := 0; i < b.N; i++ {
		_ = Substring(s, 4, 11)
	}
}

func BenchmarkSubstring_Unicode(b *testing.B) {
	b.ReportAllocs()
	s := "日本語のテキスト操作ライブラリ"
	for i := 0; i < b.N; i++ {
		_ = Substring(s, 2, 5)
	}
}

func BenchmarkTruncate_Under(b *testing.B) {
	b.ReportAllocs()
	s := "short"
	for i := 0; i < b.N; i++ {
		_, _ = Truncate(s, 100, "...")
	}
}

func BenchmarkTruncate_Over(b *testing.B) {
	b.ReportAllocs()
	s := "The quick brown fox jumps over the lazy dog"
	for i := 0; i < b.N; i++ {
		_, _ = Truncate(s, 20, "...")
	}
}

// Benchmark_Mutate exercises the composable pipeline path that chains
// Squish into TruncateOp. This is the primary shape v0.2.0 shipped and
// regressions in the pipeline overhead (e.g. per-option allocation) show
// up here.
func Benchmark_Mutate_SquishTruncate(b *testing.B) {
	b.ReportAllocs()
	s := "   The   quick   brown   fox   jumps   over   the   lazy   dog   "
	for i := 0; i < b.N; i++ {
		_ = Mutate(s, Squish, TruncateOp(20, "..."))
	}
}

// -----------------------------------------------------------------------------
// Random
// -----------------------------------------------------------------------------

func BenchmarkRandom_Default16(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Random(16)
	}
}

func BenchmarkRandom_AlphaNum32(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Random(32, AlphaNum())
	}
}

func BenchmarkRandom_LettersAlphaStart8(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = Random(8, Letters(), AlphaStart())
	}
}

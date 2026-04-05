package txt_test

import (
	"fmt"

	"github.com/bold-minds/txt"
)

func ExampleFormat() {
	fmt.Println(txt.Format("user {} logged in from {}", 42, "10.0.0.1"))
	// Output: user 42 logged in from 10.0.0.1
}

func ExampleFormat_missingArgs() {
	// Missing arguments leave their placeholders in place so the bug is
	// visible in the rendered string rather than silently swallowed.
	fmt.Println(txt.Format("a={} b={} c={}", 1, 2))
	// Output: a=1 b=2 c={}
}

func ExampleFormatAs() {
	fmt.Println(txt.FormatAs(txt.Hex, 255))
	fmt.Println(txt.FormatAs(txt.Binary, 42))
	fmt.Println(txt.FormatAs(txt.Float.Precision(2), 3.14159))
	// Output:
	// ff
	// 101010
	// 3.14
}

func ExampleBetween() {
	fmt.Println(txt.Between("foo [bar] baz", "[", "]"))
	fmt.Println(txt.Between("a=1&b=2", "a=", "&"))
	// Output:
	// bar
	// 1
}

func ExampleSquish() {
	fmt.Println(txt.Squish("  hello   world  "))
	// Output: hello world
}

func ExampleSubstring() {
	fmt.Println(txt.Substring("hello", 0, 3))
	fmt.Println(txt.Substring("hello", -2, 2))
	fmt.Println(txt.Substring("héllo", 1, 3)) // rune-counted
	// Output:
	// hel
	// lo
	// éll
}

func ExampleTruncate() {
	kept, removed := txt.Truncate("Hello world", 8, "...")
	fmt.Printf("kept=%q removed=%q\n", kept, removed)
	kept, removed = txt.Truncate("short", 20, "...")
	fmt.Printf("kept=%q removed=%q\n", kept, removed)
	// Output:
	// kept="Hello..." removed=" world"
	// kept="short" removed=""
}

func ExampleMutate() {
	// Pipeline: squish whitespace, then truncate with ellipsis.
	fmt.Println(txt.Mutate(
		"   The   quick   brown   fox   ",
		txt.Squish,
		txt.TruncateOp(15, "..."),
	))
	// Output: The quick br...
}

func ExamplePrint_multiLine() {
	// Multi-line mode: each arg prints on its own line.
	txt.Print("line 1", "line 2", "line 3")
	// Output:
	// line 1
	// line 2
	// line 3
}

func ExamplePrint_mapTemplate() {
	// Map mode: {key} placeholders substituted from the map.
	txt.Print("Hello {name}, age {age}", map[string]any{
		"name": "Alice",
		"age":  30,
	})
	// Output: Hello Alice, age 30
}

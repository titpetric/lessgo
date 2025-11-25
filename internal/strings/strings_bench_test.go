package strings

import (
	stdstrings "strings"
	"testing"
)

// Prevent compiler optimizations
var (
	benchSink  string
	benchSinkB bool
)

// Benchmark our zero-alloc TrimSpace vs standard library
func BenchmarkTrimSpace(b *testing.B) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"no_trim", "hello"},
		{"left_trim", "  hello"},
		{"right_trim", "hello  "},
		{"both_trim", "  hello  "},
		{"heavy_trim", "          hello world          "},
		{"mixed_whitespace", "  \t\r\nhello\r\n\t  "},
	}

	for _, tt := range tests {
		b.Run("custom/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				benchSink = TrimSpace(tt.input)
			}
		})

		b.Run("stdlib/"+tt.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				benchSink = stdstrings.TrimSpace(tt.input)
			}
		})
	}
}

// Benchmark individual operations
func BenchmarkTrimSpaceVsStdlib(b *testing.B) {
	const input = "  hello world  "

	b.Run("custom", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchSink = TrimSpace(input)
		}
	})

	b.Run("stdlib", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchSink = stdstrings.TrimSpace(input)
		}
	})
}

// Benchmark with realistic LESS values
func BenchmarkTrimSpaceLESS(b *testing.B) {
	lessValues := []string{
		"@primary-color: #3498db;",
		"  font-family: Arial, sans-serif;  ",
		"  color: rgb(255, 0, 0);  ",
		"\t\tmargin: 10px;\t\t",
		"transform: scale(1.5);",
	}

	for _, val := range lessValues {
		b.Run("custom", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchSink = TrimSpace(val)
			}
		})

		b.Run("stdlib", func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				benchSink = stdstrings.TrimSpace(val)
			}
		})
	}
}

// Benchmark split functions
func BenchmarkSplit(b *testing.B) {
	const input = "selector1, selector2, selector3, selector4, selector5"

	// Benchmark custom comma split
	b.Run("custom_comma", func(b *testing.B) {
		b.ReportAllocs()
		buf := make([]string, 0, 8)
		for i := 0; i < b.N; i++ {
			SplitCommaNoAlloc(input, &buf)
			benchSink = buf[0] // Use result to prevent optimization
		}
	})

	// Benchmark stdlib split
	b.Run("stdlib_split", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			parts := stdstrings.Split(input, ",")
			benchSink = parts[0]
		}
	})

	// Benchmark byte split
	const declInput = "color: red; font-size: 12px; margin: 10px; padding: 5px"

	b.Run("custom_byte", func(b *testing.B) {
		b.ReportAllocs()
		buf := make([]string, 0, 16)
		for i := 0; i < b.N; i++ {
			SplitByteNoAlloc(declInput, ';', &buf)
			benchSink = buf[0]
		}
	})

	b.Run("stdlib_split_byte", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			parts := stdstrings.Split(declInput, ";")
			benchSink = parts[0]
		}
	})
}

// Benchmark function aliases overhead
func BenchmarkAliases(b *testing.B) {
	const (
		haystack = "hello world hello"
		needle   = "hello"
	)

	b.Run("alias_HasPrefix", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchSinkB = HasPrefix(haystack, needle)
		}
	})

	b.Run("stdlib_HasPrefix", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchSinkB = stdstrings.HasPrefix(haystack, needle)
		}
	})

	b.Run("alias_Contains", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchSinkB = Contains(haystack, needle)
		}
	})

	b.Run("stdlib_Contains", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			benchSinkB = stdstrings.Contains(haystack, needle)
		}
	})

	b.Run("alias_Index", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = Index(haystack, needle)
		}
	})

	b.Run("stdlib_Index", func(b *testing.B) {
		b.ReportAllocs()
		for i := 0; i < b.N; i++ {
			_ = stdstrings.Index(haystack, needle)
		}
	})
}

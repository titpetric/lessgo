package dst

import (
	"bufio"
	"testing"

	"github.com/titpetric/lessgo/internal/strings"
)

// Sample LESS file for benchmarking
const sampleLESS = `
// Variables
@primary-color: #3498db;
@secondary-color: #2ecc71;
@base-font-size: 16px;
@spacing: 10px;

/* Main styles */
body {
    font-size: @base-font-size;
    color: @primary-color;
    margin: 0;
    padding: 0;
}

/* Button styles */
.button {
    padding: @spacing 20px;
    background: @primary-color;
    color: white;
    border: none;
    cursor: pointer;
    
    &:hover {
        background: darken(@primary-color, 10%);
    }
    
    &.large {
        padding: 20px 40px;
        font-size: 18px;
    }
    
    &.small {
        padding: 5px 10px;
        font-size: 12px;
    }
}

/* Card component */
.card {
    background: white;
    border: 1px solid #ddd;
    padding: @spacing;
    margin: @spacing;
    
    .header {
        font-weight: bold;
        padding-bottom: @spacing;
        border-bottom: 1px solid #eee;
    }
    
    .body {
        padding: @spacing 0;
    }
    
    .footer {
        text-align: right;
        padding-top: @spacing;
    }
}

/* Utility classes */
.text-center { text-align: center; }
.text-right { text-align: right; }
.text-left { text-align: left; }

/* Responsive */
@media (max-width: 768px) {
    .button { padding: 10px; }
    .card { padding: 5px; }
}
`

// BenchmarkParser benchmarks the original parser
func BenchmarkParser(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		parser := NewParser(strings.NewReader(sampleLESS))
		_, _ = parser.Parse()
	}
}

// BenchmarkParserNoAlloc benchmarks the no-allocation parser
func BenchmarkParserNoAlloc(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		parser := NewParserNoAlloc(strings.NewReader(sampleLESS))
		_, _ = parser.Parse()
	}
}

// BenchmarkParserReuse benchmarks no-allocation parser with buffer reuse
func BenchmarkParserNoAllocReuse(b *testing.B) {
	parser := NewParserNoAlloc(strings.NewReader(""))
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		parser.scanner = bufio.NewScanner(strings.NewReader(sampleLESS))
		parser.eof = false
		_, _ = parser.Parse()
	}
}

// TestParserComparison verifies both parsers produce identical output
func TestParserComparison(t *testing.T) {
	// Parse with original parser
	original, err := NewParser(strings.NewReader(sampleLESS)).Parse()
	if err != nil {
		t.Fatalf("Original parser failed: %v", err)
	}

	// Parse with no-allocation parser
	noalloc, err := NewParserNoAlloc(strings.NewReader(sampleLESS)).Parse()
	if err != nil {
		t.Fatalf("NoAlloc parser failed: %v", err)
	}

	// Compare results
	if len(original.Nodes) != len(noalloc.Nodes) {
		t.Errorf("Node count mismatch: original=%d, noalloc=%d",
			len(original.Nodes), len(noalloc.Nodes))
	}

	for i, orig := range original.Nodes {
		if i >= len(noalloc.Nodes) {
			break
		}
		noa := noalloc.Nodes[i]

		// Basic type check
		if getNodeType(orig) != getNodeType(noa) {
			t.Errorf("Node %d type mismatch: %T vs %T", i, orig, noa)
		}
	}
}

// getNodeType returns string representation of node type
func getNodeType(n Node) string {
	switch n.(type) {
	case *Comment:
		return "Comment"
	case *Decl:
		return "Decl"
	case *Block:
		return "Block"
	case *MixinCall:
		return "MixinCall"
	case *BlockVariable:
		return "BlockVariable"
	case *Each:
		return "Each"
	default:
		return "Unknown"
	}
}

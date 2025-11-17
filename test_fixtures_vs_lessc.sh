#!/bin/bash

FIXTURES_DIR="testdata/fixtures"
LESSGO_BIN="./bin/lessgo"
LESSC_BIN="/usr/bin/lessc"

# Build lessgo if needed
if [ ! -f "$LESSGO_BIN" ]; then
    go build -o "$LESSGO_BIN" ./cmd/lessgo
fi

echo "Comparing fixtures against lessc..."
echo ""

# Find all .less files (excluding helper files starting with _)
for less_file in $(find "$FIXTURES_DIR" -name "*.less" ! -name "_*" | sort); do
    base_name=$(basename "$less_file" .less)
    css_file="$FIXTURES_DIR/${base_name}.css"
    
    if [ ! -f "$css_file" ]; then
        echo "⚠️  SKIP: $base_name (no .css file)"
        continue
    fi
    
    # Compile with lessgo
    lessgo_output=$("$LESSGO_BIN" compile "$less_file" 2>&1)
    lessgo_exit=$?
    
    # Compile with lessc
    lessc_output=$(cd "$FIXTURES_DIR" && "$LESSC_BIN" "$(basename "$less_file")" 2>&1)
    lessc_exit=$?
    
    # Normalize whitespace for comparison
    normalize() {
        echo "$1" | sed 's/^[[:space:]]*//g; s/[[:space:]]*$//g' | grep -v '^$' | sort
    }
    
    lessgo_norm=$(normalize "$lessgo_output")
    lessc_norm=$(normalize "$lessc_output")
    
    if [ "$lessgo_norm" = "$lessc_norm" ]; then
        echo "✓ $base_name"
    else
        echo "✗ $base_name - OUTPUT DIFFERS"
        echo "  lessgo:"
        echo "$lessgo_output" | sed 's/^/    /'
        echo "  lessc:"
        echo "$lessc_output" | sed 's/^/    /'
        echo ""
    fi
done

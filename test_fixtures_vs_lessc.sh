#!/bin/bash

# Regenerate fixture .css files from lessc, then test lessgo against them

FIXTURES_DIR="testdata/fixtures"
LESSC_BIN="/usr/bin/lessc"
LESSGO_BIN="./bin/lessgo"
REGENERATED=0
PASSED=0
FAILED=0
ERRORS=0

# Build lessgo if needed
if [ ! -f "$LESSGO_BIN" ]; then
    go build -o "$LESSGO_BIN" ./cmd/lessgo
fi

echo "Regenerating fixture .css files from lessc..."
echo ""

# Find all .less files and regenerate from lessc
for less_file in $(find "$FIXTURES_DIR" -maxdepth 1 -name "*.less" | sort); do
    base_name=$(basename "$less_file" .less)
    css_file="$FIXTURES_DIR/${base_name}.css"
    
    # Compile with official lessc (run from fixtures dir)
    lessc_output=$(cd "$FIXTURES_DIR" && "$LESSC_BIN" "$base_name.less" 2>&1)
    lessc_exit=$?
    
    if [ $lessc_exit -ne 0 ]; then
        echo "✗ $base_name - lessc error:"
        echo "$lessc_output" | sed 's/^/    /'
        ((ERRORS++))
        continue
    fi
    
    # Write to .css file
    echo "$lessc_output" > "$css_file"
    echo "✓ $base_name: Regenerated from lessc"
    ((REGENERATED++))
done

echo ""
echo "================================="
echo "Testing lessgo against regenerated fixtures..."
echo ""

# Now test lessgo against the fixture .css files
for less_file in $(find "$FIXTURES_DIR" -maxdepth 1 -name "*.less" | sort); do
    base_name=$(basename "$less_file" .less)
    css_file="$FIXTURES_DIR/${base_name}.css"
    
    if [ ! -f "$css_file" ]; then
        continue
    fi
    
    # Compile with lessgo (1 second timeout)
    lessgo_output=$(timeout 1s "$LESSGO_BIN" compile "$less_file" 2>&1)
    lessgo_exit=$?
    
    if [ $lessgo_exit -ne 0 ]; then
        echo "✗ $base_name - lessgo error"
        ((FAILED++))
        continue
    fi
    
    # Normalize whitespace for comparison
    normalize() {
        echo "$1" | sed 's/^[[:space:]]*//g; s/[[:space:]]*$//g' | grep -v '^$'
    }
    
    lessgo_norm=$(normalize "$lessgo_output")
    fixture_norm=$(normalize "$(cat "$css_file")")
    
    if [ "$lessgo_norm" = "$fixture_norm" ]; then
        echo "✓ $base_name"
        ((PASSED++))
    else
        echo "✗ $base_name - OUTPUT DIFFERS"
        ((FAILED++))
    fi
done

echo ""
echo "================================="
echo "Regenerated: $REGENERATED | lessgo Results: $PASSED passed, $FAILED failed | Errors: $ERRORS"

exit $((FAILED + ERRORS))

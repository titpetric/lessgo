#!/bin/bash

# Test lessgo against lessc-generated fixtures
#
# Usage:
#   ./test_fixtures_vs_lessc.sh              # Test all fixtures
#   ./test_fixtures_vs_lessc.sh 999          # Test only fixtures matching prefix "999"
#   ./test_fixtures_vs_lessc.sh 001-         # Test only fixtures with "001-" prefix
#
# This script:
# 1. Regenerates fixture .css files from the official lessc compiler
# 2. Tests lessgo output against the official lessc output
# 3. Reports pass/fail for each fixture
#
# Preferred testing flow:
#   - Run with prefix to focus on specific failing tests: ./test_fixtures_vs_lessc.sh 999
#   - Fix bugs in lessgo
#   - Re-run with same prefix to verify
#   - When all pass, run full test without prefix
#
# For local debugging:
#   lessc testdata/fixtures/999-sinog-index.less           # See lessc output
#   ./bin/lessgo compile testdata/fixtures/999-sinog-index.less  # See lessgo output
#   diff -u <(lessc ...) <(lessgo compile ...)             # Compare

FIXTURES_DIR="testdata/fixtures"
LESSC_BIN="/usr/bin/lessc"
LESSGO_BIN="./bin/lessgo"
REGENERATED=0
PASSED=0
FAILED=0
ERRORS=0
PREFIX="${1:-}"  # Optional prefix filter (e.g., "999" or "001-")

# Build lessgo if needed
if [ ! -f "$LESSGO_BIN" ]; then
    go build -o "$LESSGO_BIN" ./cmd/lessgo
fi

echo "Regenerating fixture .css files from lessc..."
if [ -n "$PREFIX" ]; then
    echo "Filtering by prefix: $PREFIX"
fi
echo ""

# Find all .less files and regenerate from lessc
for less_file in $(find "$FIXTURES_DIR" -maxdepth 1 -name "${PREFIX}*.less" | sort); do
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
for less_file in $(find "$FIXTURES_DIR" -maxdepth 1 -name "${PREFIX}*.less" | sort); do
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

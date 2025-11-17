#!/bin/bash

# Verify lessgo output against fixture .css files
# Uses same whitespace normalization as verify_fixtures.sh

FIXTURES_DIR="testdata/fixtures"
FAILURES=0
SUCCESSES=0
TIMEOUTS=0
BUILD_FAILURES=0

normalize_whitespace() {
    # Remove trailing whitespace and blank lines
    sed 's/[[:space:]]*$//' | grep -v '^$'
}

echo "Building lessgo..."
if ! timeout 10s go build -o bin/lessgo ./cmd/lessgo; then
    echo "✗ Build failed"
    exit 1
fi
echo "✓ Build succeeded"
echo ""

echo "Verifying lessgo output..."
echo "================================="

# Find all .less files (excluding imports subdirectory)
for less_file in $(find "$FIXTURES_DIR" -maxdepth 1 -name "*.less" -type f | sort); do
    base_name=$(basename "$less_file" .less)
    css_file="$FIXTURES_DIR/${base_name}.css"
    
    # Skip import files that start with underscore
    if [[ "$base_name" =~ ^_ ]]; then
        continue
    fi
    
    if [ ! -f "$css_file" ]; then
        echo "⚠️  $base_name: No .css file found"
        continue
    fi
    
    # Compile with lessgo (2 second timeout)
    lessgo_output=$(timeout 2s ./bin/lessgo compile "$less_file" 2>&1) || {
        exit_code=$?
        if [ $exit_code -eq 124 ]; then
            echo "⏱ $base_name: lessgo timeout (2s)"
            ((TIMEOUTS++))
        else
            echo "✗ $base_name: lessgo error (exit code $exit_code)"
            echo "  $lessgo_output"
            ((FAILURES++))
        fi
        continue
    }
    
    # Read the expected .css file and normalize both
    expected_css=$(cat "$css_file" | normalize_whitespace)
    lessgo_normalized=$(echo "$lessgo_output" | normalize_whitespace)
    
    # Compare
    if [ "$lessgo_normalized" = "$expected_css" ]; then
        echo "✓ $base_name"
        ((SUCCESSES++))
    else
        echo "✗ $base_name: Output differs"
        # Show first difference
        diff_output=$(diff -u <(echo "$expected_css") <(echo "$lessgo_normalized") | head -20)
        echo "$diff_output" | while IFS= read -r line; do
            echo "  $line"
        done
        ((FAILURES++))
    fi
done

echo ""
echo "================================="
echo "Results: $SUCCESSES passed, $FAILURES failed, $TIMEOUTS timeouts"

exit $FAILURES

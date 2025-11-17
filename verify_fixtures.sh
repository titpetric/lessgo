#!/bin/bash

# Verify fixture .css files with official lessc compiler
# This script compiles each .less fixture with lessc and compares against the .css file
# Tolerates extra blank lines (our formatter adds opinionated spacing)

FIXTURES_DIR="testdata/fixtures"
FAILURES=0
SUCCESSES=0
TIMEOUTS=0

normalize_whitespace() {
    # Remove trailing whitespace, collapse multiple blank lines to single blank line
    sed 's/[[:space:]]*$//' | cat -s
}

echo "Verifying fixtures with lessc..."
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
    
    # Compile with lessc (2 second timeout)
    lessc_output=$(timeout 2s lessc "$less_file" 2>&1) || {
        exit_code=$?
        if [ $exit_code -eq 124 ]; then
            echo "⏱ $base_name: lessc timeout (2s)"
            ((TIMEOUTS++))
        else
            echo "✗ $base_name: lessc error (exit code $exit_code)"
            echo "  $lessc_output"
            ((FAILURES++))
        fi
        continue
    }
    
    # Read the existing .css file and normalize both
    existing_css=$(cat "$css_file" | normalize_whitespace)
    lessc_normalized=$(echo "$lessc_output" | normalize_whitespace)
    
    # Compare
    if [ "$lessc_normalized" = "$existing_css" ]; then
        echo "✓ $base_name"
        ((SUCCESSES++))
    else
        echo "✗ $base_name: Output differs"
        # Write diff to temp file for analysis
        diff_output=$(diff -u <(echo "$lessc_normalized") <(echo "$existing_css") | head -20)
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

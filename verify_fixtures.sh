#!/bin/bash
# Verify all fixtures by compiling with lessgo and comparing against expected output

FIXTURES_DIR="testdata/fixtures"
LESSGO="timeout 1s ./lessgo-verify"
PASSED=0
FAILED=0
FAILED_TESTS=""

# Get all .less fixture files (excluding imports)
for less_file in "$FIXTURES_DIR"/*.less; do
    # Skip import files (starting with _)
    basename=$(basename "$less_file")
    if [[ "$basename" == _* ]]; then
        continue
    fi
    
    css_file="${less_file%.less}.css"
    test_name=$(basename "$less_file" .less)
    
    # Skip if expected CSS doesn't exist
    if [ ! -f "$css_file" ]; then
        continue
    fi
    
    # Compile with lessgo
    output=$($LESSGO compile "$less_file" 2>&1)
    exit_code=$?
    
    if [ $exit_code -eq 0 ] || [ $exit_code -eq 124 ]; then
        # 124 is timeout exit code, treat as compilation error
        if [ $exit_code -eq 124 ]; then
            echo "✗ $test_name (timeout)"
            ((FAILED++))
            FAILED_TESTS="$FAILED_TESTS\n  $test_name"
        else
            # Compare output with expected
            expected=$(cat "$css_file")
            if [ "$output" = "$expected" ]; then
                echo "✓ $test_name"
                ((PASSED++))
            else
                echo "✗ $test_name (output mismatch)"
                ((FAILED++))
                FAILED_TESTS="$FAILED_TESTS\n  $test_name"
            fi
        fi
    else
        echo "✗ $test_name (compilation error)"
        ((FAILED++))
        FAILED_TESTS="$FAILED_TESTS\n  $test_name"
    fi
done

echo ""
echo "Results: $PASSED passed, $FAILED failed"

if [ $FAILED -gt 0 ]; then
    echo -e "Failed tests:$FAILED_TESTS"
    exit 1
fi

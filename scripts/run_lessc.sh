#!/bin/bash

# Regenerate fixture .css files using official lessc compiler
# This ensures fixtures match the canonical LESS output

FIXTURES_DIR="testdata/fixtures"
REGENERATED=0
SKIPPED=0
ERRORS=0

echo "Regenerating fixture .css files from lessc..."
echo "================================="

# Find all .less files (excluding imports subdirectory)
for less_file in $(find "$FIXTURES_DIR" -maxdepth 1 -name "*.less" -type f | sort); do
    base_name=$(basename "$less_file" .less)
    css_file="$FIXTURES_DIR/${base_name}.css"
    
    # Skip import files that start with underscore
    if [[ "$base_name" =~ ^_ ]]; then
        echo "⊘ $base_name: Skipped (import file)"
        ((SKIPPED++))
        continue
    fi
    
    # Compile with official lessc
    /usr/bin/lessc "$less_file" > /dev/null
    echo "✓ $base_name: Regenerated from lessc"
    ((REGENERATED++))
done

echo ""
echo "================================="
echo "Results: $REGENERATED regenerated, $SKIPPED skipped, $ERRORS errors"

exit $ERRORS

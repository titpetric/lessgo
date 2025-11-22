#!/bin/bash

# Benchmark lessgo against lessc
#
# Usage:
#   ./benchmark_with_lessc.sh              # Benchmark all fixtures
#   ./benchmark_with_lessc.sh 999          # Benchmark only fixtures matching prefix "999"
#   ./benchmark_with_lessc.sh 001-         # Benchmark fixtures with "001-" prefix
#
# This script measures compilation time for each fixture with both lessc and lessgo,
# tracks performance metrics, and provides a summary comparison.

FIXTURES_DIR="testdata/fixtures"
LESSC_BIN="/usr/bin/lessc"
LESSGO_BIN="./bin/lessgo"
PREFIX="${1:-}"
RUNS=10  # Number of runs per fixture for averaging

# Build lessgo if needed
set -e
go build -o "$LESSGO_BIN" ./cmd/lessgo
set +e

# Initialize counters
declare -A lessc_times
declare -A lessgo_times
declare -a fixture_names

total_lessc_time=0
total_lessgo_time=0
lessc_failures=0
lessgo_failures=0
fixture_count=0

if [ -n "$PREFIX" ]; then
    echo "Filtering by prefix: $PREFIX"
fi

echo ""
echo "================================="
echo "Benchmarking lessgo vs lessc"
echo "Running $RUNS iterations per fixture..."
echo ""

# Find all .less fixture files
fixtures=$(find "$FIXTURES_DIR" -maxdepth 1 -name "${PREFIX}*.less" -type f | sort)

for less_file in $fixtures; do
    base_name=$(basename "$less_file" .less)
    
    # Skip import files that start with underscore
    if [[ "$base_name" =~ ^_ ]]; then
        continue
    fi
    
    fixture_names+=("$base_name")
    
    # Benchmark lessc
    lessc_avg=0
    lessc_success=0
    for ((i=1; i<=RUNS; i++)); do
        start=$(date +%s%N)
        $LESSC_BIN "$less_file" > /dev/null
        lessc_exit=$?
        end=$(date +%s%N)
        
        if [ $lessc_exit -eq 0 ]; then
            elapsed=$((($end - $start) / 1000000))  # Convert nanoseconds to milliseconds
            lessc_avg=$((lessc_avg + elapsed))
            ((lessc_success++))
        else
            ((lessc_failures++))
        fi
    done
    
    if [ $lessc_success -gt 0 ]; then
        lessc_avg=$((lessc_avg / lessc_success))
    else
        lessc_avg=-1
    fi
    
    # Benchmark lessgo
    lessgo_avg=0
    lessgo_success=0
    for ((i=1; i<=RUNS; i++)); do
        start=$(date +%s%N)
        $LESSGO_BIN generate "$less_file" > /dev/null
        lessgo_exit=$?
        end=$(date +%s%N)
        
        if [ $lessgo_exit -eq 0 ]; then
            elapsed=$((($end - $start) / 1000000))  # Convert nanoseconds to milliseconds
            lessgo_avg=$((lessgo_avg + elapsed))
            ((lessgo_success++))
        else
            ((lessgo_failures++))
        fi
    done
    
    if [ $lessgo_success -gt 0 ]; then
        lessgo_avg=$((lessgo_avg / lessgo_success))
    else
        lessgo_avg=-1
    fi
    
    lessc_times["$base_name"]=$lessc_avg
    lessgo_times["$base_name"]=$lessgo_avg
    
    # Determine status and calculate speedup
    if [ $lessc_avg -gt 0 ] && [ $lessgo_avg -gt 0 ]; then
        if [ $lessgo_avg -lt $lessc_avg ]; then
            speedup=$(echo "scale=2; $lessc_avg / $lessgo_avg" | bc)
            printf "%-50s | lessc: %5dms | lessgo: %5dms | speedup: %.2fx\n" "$base_name" "$lessc_avg" "$lessgo_avg" "$speedup"
        else
            slowdown=$(echo "scale=2; $lessgo_avg / $lessc_avg" | bc)
            printf "%-50s | lessc: %5dms | lessgo: %5dms | slowdown: %.2fx\n" "$base_name" "$lessc_avg" "$lessgo_avg" "$slowdown"
        fi
        total_lessc_time=$((total_lessc_time + lessc_avg))
        total_lessgo_time=$((total_lessgo_time + lessgo_avg))
        ((fixture_count++))
    elif [ $lessc_avg -lt 0 ]; then
        printf "%-50s | LESSC FAILED\n" "$base_name"
    else
        printf "%-50s | LESSGO FAILED\n" "$base_name"
    fi
done

echo ""
echo "================================="
echo "BENCHMARK SUMMARY"
echo "================================="
echo "Total fixtures tested: $fixture_count"
echo "lessc failures: $lessc_failures"
echo "lessgo failures: $lessgo_failures"
echo ""

if [ $fixture_count -gt 0 ]; then
    avg_lessc=$((total_lessc_time / fixture_count))
    avg_lessgo=$((total_lessgo_time / fixture_count))
    
    echo "Total compilation time (all fixtures, averaged across $RUNS runs):"
    echo "  lessc:  ${total_lessc_time}ms (avg: ${avg_lessc}ms per fixture)"
    echo "  lessgo: ${total_lessgo_time}ms (avg: ${avg_lessgo}ms per fixture)"
    echo ""
    
    if [ $avg_lessgo -lt $avg_lessc ]; then
        speedup=$(echo "scale=2; $avg_lessc / $avg_lessgo" | bc)
        echo "lessgo is ${speedup}x FASTER than lessc on average"
    else
        slowdown=$(echo "scale=2; $avg_lessgo / $avg_lessc" | bc)
        echo "lessgo is ${slowdown}x SLOWER than lessc on average"
    fi
fi

echo ""

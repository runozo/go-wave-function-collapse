#!/usr/bin/env bash
#
# Benchmark helper for the WFC package.
#
# Usage:
#   ./scripts/bench.sh                 # run benchmarks on the current tree
#   ./scripts/bench.sh <git-rev>       # compare the current tree against <git-rev>
#
# Environment overrides:
#   BENCH_PATTERN   benchmark regexp (default: BenchmarkElaborateCellSweep)
#   BENCH_TIME      -benchtime value (default: 2s)
#   BENCH_COUNT     -count value     (default: 5)
#
# Examples:
#   ./scripts/bench.sh                               # current sweep benchmark
#   BENCH_PATTERN=. ./scripts/bench.sh               # all benchmarks
#   ./scripts/bench.sh a92eb05                       # A/B sweep benchmark
set -euo pipefail

BENCH_PATTERN="${BENCH_PATTERN:-BenchmarkElaborateCellSweep}"
BENCH_TIME="${BENCH_TIME:-2s}"
BENCH_COUNT="${BENCH_COUNT:-5}"

ROOT="$(git rev-parse --show-toplevel)"

run_benchmarks() {
	local dir="$1" label="$2"
	echo "== ${label} =="
	(
		cd "$dir"
		go test ./wfc/... \
			-run '^$' \
			-bench "$BENCH_PATTERN" \
			-benchtime "$BENCH_TIME" \
			-count "$BENCH_COUNT"
	)
	echo
}

if [ "$#" -eq 0 ]; then
	run_benchmarks "$ROOT" "current tree"
	exit 0
fi

REV="$1"
TMP="$(mktemp -d "${TMPDIR:-/tmp}/wfc-bench.XXXXXX")"
cleanup() { rm -rf "$TMP"; }
trap cleanup EXIT

git -C "$ROOT" archive "$REV" | tar -x -C "$TMP"

run_benchmarks "$TMP" "$REV"
run_benchmarks "$ROOT" "current tree"

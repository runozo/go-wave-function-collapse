# go-wave-function-collapse

[![CI](https://github.com/runozo/go-wave-function-collapse/actions/workflows/main.yml/badge.svg)](https://github.com/runozo/go-wave-function-collapse/actions/workflows/main.yml)
[![Release](https://github.com/runozo/go-wave-function-collapse/actions/workflows/release.yml/badge.svg)](https://github.com/runozo/go-wave-function-collapse/actions/workflows/release.yml)

Wave Function Collapse algorithm, implemented in [Go](https://golang.org), with [ebitengine](https://github.com/hajimehoshi/ebiten) game library.

![GIF animation of WFC algorithm](gifs/gowfc.gif)

## Live demo

The WASM demo is rebuilt and deployed to GitHub Pages on every push to `main`:

<https://runozo.github.io/go-wave-function-collapse/>

## Features

- Fast: tile options are stored as bitmasks over integer tile IDs, so
  constraint propagation is a few machine-word operations per cell
- Weighted random tile selection
- Contradiction handling with automatic restart
- Animated, frame-by-frame generation (works on WebAssembly too)

<!--
## Quickstart

[Download](https://github.com/runozo/go-wave-function-collapse/releases/latest) a release suitable for your platform and run it.
-->

## Run with Go

```go run main.go```

## Usage

Press ```Spacebar``` to generate a new map. 

Press ```Esc``` to exit.

## Performance

The generator was rewritten around two changes that keep the exact same Wave
Function Collapse semantics:

- **Bitmask options.** Every tile gets an integer ID and a cell's remaining
  options are a single `uint64` (the current tile set has 40 ground tiles).
  Filtering becomes an `AND`, union an `OR` and entropy a population count,
  replacing quadratic string-slice intersections and millions of allocations.
- **Worklist propagation (arc-consistency).** After a collapse, only the cells
  actually affected are revisited, instead of sweeping the whole grid on every
  collapse.

Together with precomputed compatibility masks, reused buffers and a
non-allocating render loop, a full 31x18 map generation went from **~1.43 s**
and **~5.4 GB allocated** to **~1.97 ms** and **~62 KB**, with identical output.

| Metric | Before | After | Improvement |
|---|---|---|---|
| Time per generation | ~1.43 s | ~1.97 ms | ~728x |
| Allocated per generation | ~5.4 GB | ~62 KB | ~86,000x |
| Allocations per generation | ~3.54 M | 49 | ~72,000x |

Measured with `BenchmarkGeneration` (Ryzen 5 5600, `GOMAXPROCS=12`, Go 1.27,
31x18 grid, real tile set). The remaining time is dominated by the
`neighborAllowed` union and the entropy scan.

## Benchmarks

The main benchmark is `BenchmarkGeneration`: a full map generation on the 31x18
grid used by the application, with the real tile set.

```sh
# All benchmarks on the current tree
make benchmark

# A single benchmark on the current tree
BENCH_PATTERN=BenchmarkGeneration ./scripts/bench.sh

# Compare the current tree against a git revision that contains the benchmark
# (extracts the revision to a temp dir, so the working tree is untouched)
./scripts/bench.sh <git-rev>
make benchmark-compare REV=<git-rev>
```

Tunable via environment variables:

| Variable | Default | Meaning |
|---|---|---|
| `BENCH_PATTERN` | `BenchmarkGeneration` | Benchmark regexp |
| `BENCH_TIME` | `2s` | `go test -benchtime` |
| `BENCH_COUNT` | `5` | `go test -count` |

## Releases

Pushing a tag (e.g. `v1.0.0`) triggers the `Release` workflow, which builds on
native runners (Ebitengine needs CGO on Linux/macOS) and publishes prebuilt
archives plus `checksums.txt` on the GitHub Releases page:

| OS | Arch |
|---|---|
| Linux | amd64 |
| Windows | amd64 |
| macOS | amd64, arm64 |

```sh
git tag v1.0.0
git push origin v1.0.0
```

## Credits

- Inspired by [The Coding Train](https://thecodingtrain.com/challenges/171-wave-function-collapse)
- Spritesheet by [Kenney](https://kenney.nl)

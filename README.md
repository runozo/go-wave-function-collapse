# go-wave-function-collapse

[![goreleaser](https://github.com/runozo/go-wave-function-collapse/actions/workflows/go.yml/badge.svg)](https://github.com/runozo/go-wave-function-collapse/actions/workflows/go.yml)

Wave Function Collapse algorithm, implemented in [Go](https://golang.org), with [ebitengine](https://github.com/hajimehoshi/ebiten) game library.

![GIF animation of WFC algorithm](gifs/gowfc.gif)

## Features

- Fast
- Concurrent
- Weighted random tile selection

<!--
## Quickstart

[Download](https://github.com/runozo/go-wave-function-collapse/releases/latest) a release suitable for your platform and run it.
-->

## Run with Go

```go run main.go```

## Usage

Press ```Spacebar``` to generate a new map. 

Press ```Esc``` to exit.

## Benchmarks

The WFC package ships benchmarks, the most important being `BenchmarkElaborateCellSweep`,
which measures the concurrent full-grid constraint-propagation sweep that runs after every
collapse in `Iterate`.

```sh
# All benchmarks on the current tree
make benchmark

# A single benchmark on the current tree
BENCH_PATTERN=BenchmarkElaborateCellSweep ./scripts/bench.sh

# Compare the current tree against a git revision (extracts it to a temp dir)
./scripts/bench.sh a92eb05
make benchmark-compare REV=a92eb05
```

Tunable via environment variables:

| Variable | Default | Meaning |
|---|---|---|
| `BENCH_PATTERN` | `BenchmarkElaborateCellSweep` | Benchmark regexp |
| `BENCH_TIME` | `2s` | `go test -benchtime` |
| `BENCH_COUNT` | `5` | `go test -count` |

Reference A/B (Ryzen 5 5600, `GOMAXPROCS=12`, `-benchtime=2s -count=5`): the
snapshot-based sweep (`ElaborateGrid`, one barrier for the whole grid) is about
**1.76x faster** than the previous per-row sweep (~196 µs vs ~345 µs per sweep,
no overlap across samples).

## Credits

- Inspired by [The Coding Train](https://thecodingtrain.com/challenges/171-wave-function-collapse)
- Spritesheet by [Kenney](https://kenney.nl)

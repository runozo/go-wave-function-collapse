# go-wave-function-collapse

[![CI](https://github.com/runozo/go-wave-function-collapse/actions/workflows/main.yml/badge.svg)](https://github.com/runozo/go-wave-function-collapse/actions/workflows/main.yml)
[![Release](https://github.com/runozo/go-wave-function-collapse/actions/workflows/release.yml/badge.svg)](https://github.com/runozo/go-wave-function-collapse/actions/workflows/release.yml)

Wave Function Collapse algorithm, implemented in [Go](https://golang.org), with [ebitengine](https://github.com/hajimehoshi/ebiten) game library.

![GIF animation of WFC algorithm](gifs/gowfc.gif)

## Live demo

The WASM demo is rebuilt and deployed to GitHub Pages on every push to `main`:

<https://runozo.github.io/go-wave-function-collapse/>

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

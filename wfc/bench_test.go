package wfc

import "testing"

// Grid used by the benchmarks, matching the application (31x18).
const (
	benchGridW = 31
	benchGridH = 18
)

// benchCenter is an interior cell with four neighbors.
const benchCenter = (benchGridH/2)*benchGridW + benchGridW/2

// BenchmarkNewWfc measures construction: buildIndex (compatibility masks) plus
// the initial Reset, on the real tile set.
func BenchmarkNewWfc(b *testing.B) {
	entries := loadRealTileEntries(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewWfc(benchGridW, benchGridH, entries)
	}
}

// BenchmarkReset measures resetting the whole grid to the initial
// superposition.
func BenchmarkReset(b *testing.B) {
	entries := loadRealTileEntries(b)
	w := NewWfc(benchGridW, benchGridH, entries)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Reset()
	}
}

// BenchmarkLeastEntropyCellIndexes measures the entropy scan over a fresh grid
// (worst case: every cell has the same entropy, so all are collected).
func BenchmarkLeastEntropyCellIndexes(b *testing.B) {
	entries := loadRealTileEntries(b)
	w := NewWfc(benchGridW, benchGridH, entries)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.LeastEntropyCellIndexes()
	}
}

// BenchmarkRandomOptionWithWeight measures weighted tile selection over a full
// option mask.
func BenchmarkRandomOptionWithWeight(b *testing.B) {
	entries := loadRealTileEntries(b)
	w := NewWfc(1, 1, entries)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.RandomOptionWithWeight(0)
	}
}

// BenchmarkCollapseCell measures collapsing a single cell.
func BenchmarkCollapseCell(b *testing.B) {
	entries := loadRealTileEntries(b)
	w := NewWfc(1, 1, entries)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w.Tiles[0] = Tile{Mask: w.initialMask}
		w.CollapseCell(0)
	}
}

// BenchmarkNeighborAllowed measures the union over a neighbor's remaining
// options (worst case: a full mask).
func BenchmarkNeighborAllowed(b *testing.B) {
	entries := loadRealTileEntries(b)
	w := NewWfc(benchGridW, benchGridH, entries)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.neighborAllowed(benchCenter, dirRight)
	}
}

// BenchmarkAllowedMask measures intersecting an interior cell with its four
// neighbors.
func BenchmarkAllowedMask(b *testing.B) {
	entries := loadRealTileEntries(b)
	w := NewWfc(benchGridW, benchGridH, entries)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = w.allowedMask(benchCenter)
	}
}

// BenchmarkPropagate measures arc-consistency propagation from a collapsed
// cell on a fresh grid: this is the largest cascade (a single collapse can
// narrow the whole map). The grid setup is excluded from the timer.
func BenchmarkPropagate(b *testing.B) {
	entries := loadRealTileEntries(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		w := NewWfc(benchGridW, benchGridH, entries)
		w.CollapseCell(benchCenter)
		b.StartTimer()

		_ = w.propagate(benchCenter)
	}
}

// BenchmarkIterate measures a single collapse + propagation step on a fresh
// grid, i.e. the first step, whose cascade is the largest. The average step
// over a whole generation is BenchmarkGeneration divided by TotalTiles. The
// grid setup is excluded from the timer.
func BenchmarkIterate(b *testing.B) {
	entries := loadRealTileEntries(b)
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		w := NewWfc(benchGridW, benchGridH, entries)
		w.BeginRender()
		b.StartTimer()

		w.Iterate()
	}
}

// BenchmarkGeneration measures a full map generation (all cells collapsed) on
// the 31x18 grid used by the application, with the real tile set.
func BenchmarkGeneration(b *testing.B) {
	entries := loadRealTileEntries(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWfc(benchGridW, benchGridH, entries)
		w.StartRender()
	}
}

// BenchmarkGenerationSizes shows how a full generation scales with the grid
// size.
func BenchmarkGenerationSizes(b *testing.B) {
	entries := loadRealTileEntries(b)
	sizes := []struct {
		name string
		w, h int
	}{
		{"31x18", 31, 18},
		{"62x36", 62, 36},
		{"124x72", 124, 72},
	}
	for _, size := range sizes {
		b.Run(size.name, func(b *testing.B) {
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				w := NewWfc(size.w, size.h, entries)
				w.StartRender()
			}
		})
	}
}

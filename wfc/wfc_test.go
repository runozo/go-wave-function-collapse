package wfc

import (
	"testing"

	"github.com/runozo/go-wave-function-collapse/assets"
)

func TestFilterOptions(t *testing.T) {
	tests := []struct {
		name     string
		orig     []string
		options  []string
		expected []string
	}{
		{"empty orig", []string{}, []string{"a", "b"}, []string{}},
		{"empty options", []string{"a", "b"}, []string{}, []string{}},
		{"no matches", []string{"a", "b"}, []string{"c", "d"}, []string{}},
		{"some matches", []string{"a", "b", "c"}, []string{"a", "c"}, []string{"a", "c"}},
		{"all matches", []string{"a", "b"}, []string{"a", "b"}, []string{"a", "b"}},
		{"options with duplicates", []string{"a", "b"}, []string{"a", "a", "b"}, []string{"a", "b"}},
		{"orig with duplicates", []string{"a", "a", "b"}, []string{"a", "b"}, []string{"a", "a", "b"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wfc := &Wfc{}
			actual := wfc.FilterOptions(tt.orig, tt.options)
			if !sliceEqual(actual, tt.expected) {
				t.Errorf("FilterOptions(%v, %v) = %v, want %v", tt.orig, tt.options, actual, tt.expected)
			}
		})
	}
}

func BenchmarkFilterOptions(b *testing.B) {
	wfc := &Wfc{}
	orig := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	options := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	for i := 0; i < b.N; i++ {
		wfc.FilterOptions(orig, options)
	}
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// optionsMap builds a direction -> allowed tiles map as expected by TileEntry.Options.
func optionsMap(up, right, down, left []string) map[string][]string {
	return map[string][]string{
		"up":    up,
		"right": right,
		"down":  down,
		"left":  left,
	}
}

// groundEntry builds a ground TileEntry usable by NewWfc (needs >= 4 option keys).
func groundEntry(name string, weight int, options map[string][]string) assets.TileEntry {
	return assets.TileEntry{
		Name:    name,
		Type:    "ground",
		Weight:  weight,
		Options: options,
	}
}

// TestRandomOptionWithWeight verifies that a single-option cell is deterministic
// and that, with two weighted options, the heavier one is chosen more often.
func TestRandomOptionWithWeight(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A"}, []string{"A"}, []string{"A"}, []string{"A"})),
		"B": groundEntry("B", 100, optionsMap([]string{"B"}, []string{"B"}, []string{"B"}, []string{"B"})),
	}
	wfc := &Wfc{
		TileEntries: entries,
		Tiles: []Tile{
			{Options: []string{"A", "B"}},
			{Options: []string{"B"}},
		},
	}

	// A cell with a single option must always return that option.
	for i := 0; i < 10; i++ {
		if got := wfc.RandomOptionWithWeight(1); got != "B" {
			t.Fatalf("single-option cell returned %q, want %q", got, "B")
		}
	}

	// B has weight 100 vs A weight 1, so it must clearly dominate.
	const runs = 1000
	counts := map[string]int{}
	for i := 0; i < runs; i++ {
		got := wfc.RandomOptionWithWeight(0)
		if got != "A" && got != "B" {
			t.Fatalf("returned option %q not in available options", got)
		}
		counts[got]++
	}
	if counts["B"] <= counts["A"] {
		t.Errorf("weighted selection: got A=%d B=%d, want B > A", counts["A"], counts["B"])
	}
}

// TestCollapseCell verifies that collapsing pins the cell to a single option.
func TestCollapseCell(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A"}, []string{"A"}, []string{"A"}, []string{"A"})),
	}
	wfc := &Wfc{
		TileEntries: entries,
		Tiles:       []Tile{{Options: []string{"A"}}},
	}

	wfc.CollapseCell(0)

	tile := wfc.Tiles[0]
	if !tile.Collapsed {
		t.Errorf("Collapsed = false, want true")
	}
	if tile.Name != "A" {
		t.Errorf("Name = %q, want %q", tile.Name, "A")
	}
	if len(tile.Options) != 1 || tile.Options[0] != "A" {
		t.Errorf("Options = %v, want [A]", tile.Options)
	}
}

// TestGetAvailableOptions verifies that the allowed tiles are the union of the
// neighbor's options lists for the requested direction.
func TestGetAvailableOptions(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap(nil, []string{"A", "B"}, nil, nil)),
		"B": groundEntry("B", 1, optionsMap(nil, []string{"C"}, nil, nil)),
	}
	wfc := &Wfc{
		TileEntries: entries,
		Tiles:       []Tile{{Options: []string{"A", "B"}}},
	}

	got := wfc.GetAvailableOptions(0, "right")
	want := []string{"A", "B", "C"}
	if !sliceEqual(got, want) {
		t.Errorf("GetAvailableOptions(0, right) = %v, want %v", got, want)
	}
}

// TestElaborateCell verifies that a cell is narrowed by its neighbors and that
// boundary cells (no neighbor in a direction) are handled without panicking.
func TestElaborateCell(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A", "B"}, []string{"A", "B"}, []string{"A", "B"}, []string{"A"})),
		"B": groundEntry("B", 1, optionsMap([]string{"A", "B"}, []string{"B"}, []string{"A", "B"}, []string{"B"})),
	}

	// 3x1 grid: right neighbor only allows "A", left neighbor allows everything.
	wfc := &Wfc{
		TileEntries: entries,
		numOfTilesX: 3,
		numOfTilesY: 1,
		Tiles: []Tile{
			{Options: []string{"A", "B"}},
			{Options: []string{"A", "B"}},
			{Options: []string{"A"}},
		},
	}

	wfc.ElaborateCell(1, 0)

	want := []string{"A"}
	if !sliceEqual(wfc.Tiles[1].Options, want) {
		t.Errorf("center Options = %v, want %v", wfc.Tiles[1].Options, want)
	}

	// Boundary case: a 1x1 grid has no neighbors, options must be untouched.
	single := &Wfc{
		TileEntries: entries,
		numOfTilesX: 1,
		numOfTilesY: 1,
		Tiles:       []Tile{{Options: []string{"A", "B"}}},
	}
	single.ElaborateCell(0, 0)
	if !sliceEqual(single.Tiles[0].Options, []string{"A", "B"}) {
		t.Errorf("1x1 Options = %v, want [A B]", single.Tiles[0].Options)
	}
}

// TestElaborateCellContradiction verifies that incompatible neighbors reduce the
// options of a cell to the empty set, and that such a cell is reported by
// LeastEntropyCellIndexes (entropy 0).
func TestElaborateCellContradiction(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A"}, []string{"A"}, []string{"A"}, []string{"A"})),
		"B": groundEntry("B", 1, optionsMap([]string{"B"}, []string{"B"}, []string{"B"}, []string{"B"})),
	}

	// 3x1 grid: left cell is "A" and right cell is "B", both incompatible with
	// each other across the center cell.
	wfc := &Wfc{
		TileEntries: entries,
		numOfTilesX: 3,
		numOfTilesY: 1,
		Tiles: []Tile{
			{Options: []string{"A"}},
			{Options: []string{"A", "B"}},
			{Options: []string{"B"}},
		},
	}

	wfc.ElaborateCell(1, 0)

	if len(wfc.Tiles[1].Options) != 0 {
		t.Fatalf("center Options = %v, want empty (contradiction)", wfc.Tiles[1].Options)
	}

	got := wfc.LeastEntropyCellIndexes()
	if !wfc.IntInSlice(1, got) {
		t.Errorf("LeastEntropyCellIndexes() = %v, want it to contain the contradictory cell 1", got)
	}
}

// TestLeastEntropyCellIndexes verifies that all cells sharing the minimum number
// of options are returned and that collapsed cells are ignored.
func TestLeastEntropyCellIndexes(t *testing.T) {
	wfc := &Wfc{
		TileEntries: map[string]assets.TileEntry{
			"A": {Name: "A"},
			"B": {Name: "B"},
		},
		Tiles: []Tile{
			{Options: []string{"A", "B"}, Collapsed: true},
			{Options: []string{"A", "B"}},
			{Options: []string{"A"}},
			{Options: []string{"B"}},
			{Options: []string{"A", "B"}},
		},
	}

	got := wfc.LeastEntropyCellIndexes()
	if len(got) != 2 || !wfc.IntInSlice(2, got) || !wfc.IntInSlice(3, got) {
		t.Errorf("LeastEntropyCellIndexes() = %v, want [2 3]", got)
	}
}

// TestIterateRendered verifies that a fully collapsed grid reports Rendered and
// flips the running/rendered flags accordingly.
func TestIterateRendered(t *testing.T) {
	wfc := &Wfc{
		TileEntries: map[string]assets.TileEntry{"A": {Name: "A"}},
		numOfTilesX: 1,
		numOfTilesY: 1,
		Tiles:       []Tile{{Options: []string{"A"}, Collapsed: true, Name: "A"}},
		IsRunning:   true,
	}

	if got := wfc.Iterate(1, 1); got != Rendered {
		t.Fatalf("Iterate() = %v, want Rendered", got)
	}
	if wfc.IsRunning {
		t.Errorf("IsRunning = true, want false")
	}
	if !wfc.IsRendered {
		t.Errorf("IsRendered = false, want true")
	}
}

// TestIterateContradiction verifies that a cell with zero options is detected
// instead of being collapsed (which would panic).
func TestIterateContradiction(t *testing.T) {
	wfc := &Wfc{
		TileEntries: map[string]assets.TileEntry{
			"A": {Name: "A"},
			"B": {Name: "B"},
		},
		numOfTilesX: 1,
		numOfTilesY: 1,
		Tiles:       []Tile{{Options: []string{}}},
		IsRunning:   true,
	}

	if got := wfc.Iterate(1, 1); got != Contradiction {
		t.Fatalf("Iterate() = %v, want Contradiction", got)
	}
	if !wfc.IsRunning {
		t.Errorf("IsRunning = false, want true (contradiction does not stop the run)")
	}
}

// TestIterateProgress verifies that a normal step collapses exactly one cell and
// reports Iterating.
func TestIterateProgress(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A"}, []string{"A"}, []string{"A"}, []string{"A"})),
		"B": groundEntry("B", 1, optionsMap([]string{"B"}, []string{"B"}, []string{"B"}, []string{"B"})),
	}
	wfc := &Wfc{
		TileEntries: entries,
		numOfTilesX: 1,
		numOfTilesY: 1,
		Tiles:       []Tile{{Options: []string{"A", "B"}}},
		IsRunning:   true,
	}

	if got := wfc.Iterate(1, 1); got != Iterating {
		t.Fatalf("Iterate() = %v, want Iterating", got)
	}
	if !wfc.Tiles[0].Collapsed {
		t.Errorf("cell not collapsed after Iterate")
	}
	if wfc.ProcessedTiles != 1 {
		t.Errorf("ProcessedTiles = %d, want 1", wfc.ProcessedTiles)
	}
}

// TestRandomOptionWithWeightEmpty verifies that a cell with no options does not
// panic and returns an empty string.
func TestRandomOptionWithWeightEmpty(t *testing.T) {
	wfc := &Wfc{
		TileEntries: map[string]assets.TileEntry{},
		Tiles:       []Tile{{Options: []string{}}},
	}

	if got := wfc.RandomOptionWithWeight(0); got != "" {
		t.Errorf("RandomOptionWithWeight() = %q, want empty string", got)
	}
}

// TestRandomOptionWithWeightZeroWeights verifies that when every weight is zero
// the function falls back to a uniform choice instead of panicking.
func TestRandomOptionWithWeightZeroWeights(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": {Name: "A", Weight: 0},
		"B": {Name: "B", Weight: 0},
	}
	wfc := &Wfc{
		TileEntries: entries,
		Tiles:       []Tile{{Options: []string{"A", "B"}}},
	}

	for i := 0; i < 100; i++ {
		got := wfc.RandomOptionWithWeight(0)
		if got != "A" && got != "B" {
			t.Fatalf("returned option %q not in available options", got)
		}
	}
}

// TestCollapseCellNoOptions verifies that collapsing an empty cell is a no-op.
func TestCollapseCellNoOptions(t *testing.T) {
	wfc := &Wfc{
		TileEntries: map[string]assets.TileEntry{},
		Tiles:       []Tile{{Options: []string{}}},
	}

	wfc.CollapseCell(0)

	if wfc.Tiles[0].Collapsed {
		t.Errorf("empty cell was collapsed, want it left untouched")
	}
}

// TestStartRenderRendersValidGrid verifies that StartRender completes a
// contradiction-free rule set and leaves the expected final state.
func TestStartRenderRendersValidGrid(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A", "B"}, []string{"A", "B"}, []string{"A", "B"}, []string{"A", "B"})),
		"B": groundEntry("B", 1, optionsMap([]string{"A", "B"}, []string{"A", "B"}, []string{"A", "B"}, []string{"A", "B"})),
	}
	wfc := NewWfc(5, 5, entries)
	if wfc.MaxRestarts != DefaultMaxRestarts {
		t.Errorf("MaxRestarts = %d, want %d", wfc.MaxRestarts, DefaultMaxRestarts)
	}

	wfc.StartRender()

	if wfc.IsRunning {
		t.Errorf("IsRunning = true, want false")
	}
	if !wfc.IsRendered {
		t.Errorf("IsRendered = false, want true")
	}
	if wfc.ProcessedTiles != wfc.TotalTiles {
		t.Errorf("ProcessedTiles = %d, want %d", wfc.ProcessedTiles, wfc.TotalTiles)
	}
}

// TestStartRenderRestartsAndGivesUp verifies that a rule set which always leads
// to a contradiction restarts up to MaxRestarts and then gives up cleanly.
func TestStartRenderRestartsAndGivesUp(t *testing.T) {
	// Both tiles have no compatible neighbor on the left/right, so whichever
	// cell collapses first empties its neighbor: contradiction on every attempt.
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{}, []string{}, []string{}, []string{})),
		"B": groundEntry("B", 1, optionsMap([]string{}, []string{}, []string{}, []string{})),
	}
	wfc := NewWfc(2, 1, entries)
	wfc.MaxRestarts = 3

	wfc.StartRender()

	if wfc.IsRunning {
		t.Errorf("IsRunning = true, want false")
	}
	if !wfc.IsRendered {
		t.Errorf("IsRendered = false, want true")
	}
	// Reset clears ProcessedTiles at every attempt, and each attempt collapses
	// exactly one cell before the contradiction is detected.
	if wfc.ProcessedTiles != 1 {
		t.Errorf("ProcessedTiles = %d, want 1", wfc.ProcessedTiles)
	}
}

// TestElaborateGridMatchesSerialFixpoint verifies that the concurrent,
// snapshot-based sweep reaches the same arc-consistency fixpoint as the serial
// ElaborateCell loop.
func TestElaborateGridMatchesSerialFixpoint(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A"}, []string{"A"}, []string{"A"}, []string{"A"})),
		"B": groundEntry("B", 1, optionsMap([]string{"B"}, []string{"B"}, []string{"B"}, []string{"B"})),
	}
	const width, height = 4, 3

	newGrid := func() *Wfc {
		wfc := &Wfc{
			TileEntries: entries,
			numOfTilesX: width,
			numOfTilesY: height,
			Tiles:       make([]Tile, width*height),
		}
		for i := range wfc.Tiles {
			wfc.Tiles[i] = Tile{Options: []string{"A", "B"}}
		}
		// Collapse the top-left corner to "A": the constraint must propagate.
		wfc.Tiles[0] = Tile{Options: []string{"A"}, Name: "A", Collapsed: true}
		return wfc
	}

	concurrent := newGrid()
	serial := newGrid()

	for i := 0; i < width*height; i++ {
		concurrent.ElaborateGrid(width, height)
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				serial.ElaborateCell(x, y)
			}
		}
	}

	for i := range concurrent.Tiles {
		if !sliceEqual(concurrent.Tiles[i].Options, serial.Tiles[i].Options) {
			t.Fatalf("cell %d: concurrent %v != serial %v",
				i, concurrent.Tiles[i].Options, serial.Tiles[i].Options)
		}
	}
}

// BenchmarkElaborateCellSweep measures the concurrent full-grid propagation
// sweep performed after every collapse in Iterate.
func BenchmarkElaborateCellSweep(b *testing.B) {
	const (
		width  = 31
		height = 18
	)

	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A", "B"}, []string{"A", "B"}, []string{"A", "B"}, []string{"A", "B"})),
		"B": groundEntry("B", 1, optionsMap([]string{"A", "B"}, []string{"A", "B"}, []string{"A", "B"}, []string{"A", "B"})),
	}
	wfc := NewWfc(width, height, entries)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		wfc.ElaborateGrid(width, height)
	}
}

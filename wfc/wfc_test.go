package wfc

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/runozo/go-wave-function-collapse/assets"
)

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

// twoCompatibleTiles returns two tiles that accept each other on every side.
func twoCompatibleTiles() map[string]assets.TileEntry {
	all := []string{"A", "B"}
	return map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap(all, all, all, all)),
		"B": groundEntry("B", 1, optionsMap(all, all, all, all)),
	}
}

// mutuallyExclusiveTiles returns two tiles with no compatible side at all.
func mutuallyExclusiveTiles() map[string]assets.TileEntry {
	empty := []string{}
	return map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap(empty, empty, empty, empty)),
		"B": groundEntry("B", 1, optionsMap(empty, empty, empty, empty)),
	}
}

func containsInt(a int, slice []int) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}

// TestNewWfcInitialState verifies the initial superposition and grid size.
func TestNewWfcInitialState(t *testing.T) {
	wfc := NewWfc(2, 3, twoCompatibleTiles())

	if wfc.TotalTiles != 6 {
		t.Errorf("TotalTiles = %d, want 6", wfc.TotalTiles)
	}
	if len(wfc.Tiles) != 6 {
		t.Errorf("len(Tiles) = %d, want 6", len(wfc.Tiles))
	}
	if wfc.initialMask != 0b11 {
		t.Errorf("initialMask = %b, want 0b11", wfc.initialMask)
	}
	for i := range wfc.Tiles {
		if wfc.Tiles[i].Collapsed {
			t.Errorf("cell %d already collapsed", i)
		}
		if wfc.Tiles[i].Mask != wfc.initialMask {
			t.Errorf("cell %d mask = %b, want %b", i, wfc.Tiles[i].Mask, wfc.initialMask)
		}
	}
	if wfc.MaxRestarts != DefaultMaxRestarts {
		t.Errorf("MaxRestarts = %d, want %d", wfc.MaxRestarts, DefaultMaxRestarts)
	}
}

// TestCollapseCell verifies that collapsing pins the cell to a single option.
func TestCollapseCell(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A"}, []string{"A"}, []string{"A"}, []string{"A"})),
	}
	wfc := NewWfc(1, 1, entries)

	wfc.CollapseCell(0)

	tile := wfc.Tiles[0]
	if !tile.Collapsed {
		t.Errorf("Collapsed = false, want true")
	}
	if tile.Name != "A" {
		t.Errorf("Name = %q, want %q", tile.Name, "A")
	}
	if tile.Mask != 1<<uint(wfc.nameToID["A"]) {
		t.Errorf("Mask = %b, want single bit for A", tile.Mask)
	}
}

// TestCollapseCellNoOptions verifies that collapsing an empty cell is a no-op.
func TestCollapseCellNoOptions(t *testing.T) {
	wfc := NewWfc(1, 1, twoCompatibleTiles())
	wfc.Tiles[0] = Tile{Mask: 0}

	wfc.CollapseCell(0)

	if wfc.Tiles[0].Collapsed {
		t.Errorf("empty cell was collapsed, want it left untouched")
	}
}

// TestRandomOptionWithWeight verifies deterministic single-option selection and
// that the heavier option dominates.
func TestRandomOptionWithWeight(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A"}, []string{"A"}, []string{"A"}, []string{"A"})),
		"B": groundEntry("B", 100, optionsMap([]string{"B"}, []string{"B"}, []string{"B"}, []string{"B"})),
	}
	wfc := NewWfc(1, 1, entries)

	// A cell with a single option must always return that option.
	wfc.Tiles[0] = Tile{Mask: 1 << uint(wfc.nameToID["A"])}
	for i := 0; i < 10; i++ {
		if got := wfc.RandomOptionWithWeight(0); got != "A" {
			t.Fatalf("single-option cell returned %q, want %q", got, "A")
		}
	}

	// B has weight 100 vs A weight 1, so it must clearly dominate.
	wfc.Tiles[0] = Tile{Mask: wfc.initialMask}
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

// TestRandomOptionWithWeightEmpty verifies that a cell with no options does not
// panic and returns an empty string.
func TestRandomOptionWithWeightEmpty(t *testing.T) {
	wfc := NewWfc(1, 1, twoCompatibleTiles())
	wfc.Tiles[0] = Tile{Mask: 0}

	if got := wfc.RandomOptionWithWeight(0); got != "" {
		t.Errorf("RandomOptionWithWeight() = %q, want empty string", got)
	}
}

// TestRandomOptionWithWeightZeroWeights verifies the uniform fallback when every
// weight is zero.
func TestRandomOptionWithWeightZeroWeights(t *testing.T) {
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 0, optionsMap([]string{"A"}, []string{"A"}, []string{"A"}, []string{"A"})),
		"B": groundEntry("B", 0, optionsMap([]string{"B"}, []string{"B"}, []string{"B"}, []string{"B"})),
	}
	wfc := NewWfc(1, 1, entries)
	wfc.Tiles[0] = Tile{Mask: wfc.initialMask}

	for i := 0; i < 100; i++ {
		got := wfc.RandomOptionWithWeight(0)
		if got != "A" && got != "B" {
			t.Fatalf("returned option %q not in available options", got)
		}
	}
}

// TestPropagateNarrows verifies that a collapsed cell restricts its neighbor.
func TestPropagateNarrows(t *testing.T) {
	// A only accepts A on its right; B only accepts B on its right.
	entries := map[string]assets.TileEntry{
		"A": groundEntry("A", 1, optionsMap([]string{"A", "B"}, []string{"A"}, []string{"A", "B"}, []string{"A", "B"})),
		"B": groundEntry("B", 1, optionsMap([]string{"A", "B"}, []string{"B"}, []string{"A", "B"}, []string{"A", "B"})),
	}
	wfc := NewWfc(2, 1, entries)
	idA := wfc.nameToID["A"]
	wfc.Tiles[0] = Tile{Collapsed: true, Name: "A", Mask: 1 << uint(idA)}

	if !wfc.propagate(0) {
		t.Fatalf("propagate returned a contradiction, want success")
	}
	if wfc.Tiles[1].Mask != 1<<uint(idA) {
		t.Errorf("neighbor mask = %b, want only A", wfc.Tiles[1].Mask)
	}
}

// TestPropagateContradiction verifies that incompatible neighbors empty a cell
// and that propagate reports the contradiction.
func TestPropagateContradiction(t *testing.T) {
	wfc := NewWfc(2, 1, mutuallyExclusiveTiles())
	idA := wfc.nameToID["A"]
	wfc.Tiles[0] = Tile{Collapsed: true, Name: "A", Mask: 1 << uint(idA)}

	if wfc.propagate(0) {
		t.Fatalf("propagate returned success, want contradiction")
	}
	if wfc.Tiles[1].Mask != 0 {
		t.Errorf("neighbor mask = %b, want 0", wfc.Tiles[1].Mask)
	}
}

// TestLeastEntropyCellIndexes verifies that all cells sharing the minimum number
// of options are returned and that collapsed cells are ignored.
func TestLeastEntropyCellIndexes(t *testing.T) {
	wfc := NewWfc(4, 1, twoCompatibleTiles())
	idA := wfc.nameToID["A"]
	idB := wfc.nameToID["B"]

	wfc.Tiles[0] = Tile{Collapsed: true, Name: "A", Mask: 1 << uint(idA)}
	wfc.Tiles[1] = Tile{Mask: wfc.initialMask}
	wfc.Tiles[2] = Tile{Mask: 1 << uint(idA)}
	wfc.Tiles[3] = Tile{Mask: 1 << uint(idB)}

	got := wfc.LeastEntropyCellIndexes()
	if len(got) != 2 || !containsInt(2, got) || !containsInt(3, got) {
		t.Errorf("LeastEntropyCellIndexes() = %v, want [2 3]", got)
	}
}

// TestIterateRendered verifies that a fully collapsed grid reports Rendered and
// flips the running/rendered flags accordingly.
func TestIterateRendered(t *testing.T) {
	wfc := NewWfc(1, 1, twoCompatibleTiles())
	idA := wfc.nameToID["A"]
	wfc.Tiles[0] = Tile{Collapsed: true, Name: "A", Mask: 1 << uint(idA)}
	wfc.IsRunning = true

	if got := wfc.Iterate(); got != Rendered {
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
	wfc := NewWfc(1, 1, twoCompatibleTiles())
	wfc.Tiles[0] = Tile{Mask: 0}
	wfc.IsRunning = true

	if got := wfc.Iterate(); got != Contradiction {
		t.Fatalf("Iterate() = %v, want Contradiction", got)
	}
	if !wfc.IsRunning {
		t.Errorf("IsRunning = false, want true (contradiction does not stop the run)")
	}
}

// TestIterateProgress verifies that a normal step collapses exactly one cell and
// reports Iterating.
func TestIterateProgress(t *testing.T) {
	wfc := NewWfc(1, 1, twoCompatibleTiles())
	wfc.BeginRender()

	if got := wfc.Iterate(); got != Iterating {
		t.Fatalf("Iterate() = %v, want Iterating", got)
	}
	if !wfc.Tiles[0].Collapsed {
		t.Errorf("cell not collapsed after Iterate")
	}
	if wfc.ProcessedTiles != 1 {
		t.Errorf("ProcessedTiles = %d, want 1", wfc.ProcessedTiles)
	}
}

// TestBeginRender verifies that BeginRender prepares a fresh generation without
// running it.
func TestBeginRender(t *testing.T) {
	wfc := NewWfc(2, 2, twoCompatibleTiles())
	wfc.ProcessedTiles = 7
	wfc.IsRendered = true

	wfc.BeginRender()

	if !wfc.IsRunning {
		t.Errorf("IsRunning = false, want true")
	}
	if wfc.IsRendered {
		t.Errorf("IsRendered = true, want false")
	}
	if wfc.ProcessedTiles != 0 {
		t.Errorf("ProcessedTiles = %d, want 0", wfc.ProcessedTiles)
	}
	if wfc.attempts != 0 {
		t.Errorf("attempts = %d, want 0", wfc.attempts)
	}
}

// TestStepProgress verifies that a single Step collapses one cell and reports
// Iterating while the generation is still in progress.
func TestStepProgress(t *testing.T) {
	wfc := NewWfc(1, 1, twoCompatibleTiles())
	wfc.BeginRender()

	if got := wfc.Step(); got != Iterating {
		t.Fatalf("Step() = %v, want Iterating", got)
	}
	if !wfc.Tiles[0].Collapsed {
		t.Errorf("cell not collapsed after Step")
	}
	if wfc.ProcessedTiles != 1 {
		t.Errorf("ProcessedTiles = %d, want 1", wfc.ProcessedTiles)
	}
}

// TestStepRendered verifies that Step reports Rendered once the grid is fully
// collapsed and clears the running flag.
func TestStepRendered(t *testing.T) {
	wfc := NewWfc(1, 1, twoCompatibleTiles())
	idA := wfc.nameToID["A"]
	wfc.Tiles[0] = Tile{Collapsed: true, Name: "A", Mask: 1 << uint(idA)}
	wfc.IsRunning = true

	if got := wfc.Step(); got != Rendered {
		t.Fatalf("Step() = %v, want Rendered", got)
	}
	if wfc.IsRunning {
		t.Errorf("IsRunning = true, want false")
	}
	if !wfc.IsRendered {
		t.Errorf("IsRendered = false, want true")
	}
}

// TestStepRestartsAndGivesUp verifies that Step transparently restarts the
// generation on contradictions and eventually gives up after MaxRestarts,
// without ever panicking.
func TestStepRestartsAndGivesUp(t *testing.T) {
	wfc := NewWfc(2, 1, mutuallyExclusiveTiles())
	wfc.MaxRestarts = 2
	wfc.BeginRender()

	result := Iterating
	for steps := 0; result == Iterating && steps < 50; steps++ {
		result = wfc.Step()
	}

	if result != Rendered {
		t.Fatalf("Step() loop ended with %v, want Rendered", result)
	}
	if wfc.IsRunning {
		t.Errorf("IsRunning = true, want false")
	}
	if !wfc.IsRendered {
		t.Errorf("IsRendered = false, want true")
	}
	if wfc.attempts != wfc.MaxRestarts {
		t.Errorf("attempts = %d, want %d", wfc.attempts, wfc.MaxRestarts)
	}
}

// TestStartRenderRendersValidGrid verifies that StartRender completes a
// contradiction-free rule set and leaves the expected final state.
func TestStartRenderRendersValidGrid(t *testing.T) {
	wfc := NewWfc(5, 5, twoCompatibleTiles())

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
	for i := range wfc.Tiles {
		if !wfc.Tiles[i].Collapsed || wfc.Tiles[i].Name == "" {
			t.Errorf("cell %d not properly collapsed: %+v", i, wfc.Tiles[i])
		}
	}
}

// TestStartRenderRestartsAndGivesUp verifies that a rule set which always leads
// to a contradiction restarts up to MaxRestarts and then gives up cleanly.
func TestStartRenderRestartsAndGivesUp(t *testing.T) {
	wfc := NewWfc(2, 1, mutuallyExclusiveTiles())
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

// TestGenerationWithRealTiles runs a full generation with the real tile set and
// checks that every cell ends up collapsed.
func TestGenerationWithRealTiles(t *testing.T) {
	wfc := NewWfc(31, 18, loadRealTileEntries(t))

	wfc.StartRender()

	if !wfc.IsRendered {
		t.Fatalf("IsRendered = false, want true")
	}
	if wfc.ProcessedTiles != wfc.TotalTiles {
		t.Fatalf("ProcessedTiles = %d, want %d", wfc.ProcessedTiles, wfc.TotalTiles)
	}
	for i := range wfc.Tiles {
		if !wfc.Tiles[i].Collapsed || wfc.Tiles[i].Name == "" {
			t.Fatalf("cell %d not properly collapsed: %+v", i, wfc.Tiles[i])
		}
	}
}

// loadRealTileEntries loads the real tile set used by the application.
func loadRealTileEntries(tb testing.TB) map[string]assets.TileEntry {
	tb.Helper()
	data, err := os.ReadFile(filepath.Join("..", "data", "mapped_tiles.json"))
	if err != nil {
		tb.Fatalf("read tiles: %v", err)
	}
	var list []assets.TileEntry
	if err := json.Unmarshal(data, &list); err != nil {
		tb.Fatalf("parse tiles: %v", err)
	}
	entries := make(map[string]assets.TileEntry, len(list))
	for _, e := range list {
		entries[e.Name] = e
	}
	return entries
}

// BenchmarkGeneration measures a full map generation (all cells collapsed) on
// the same 31x18 grid used by the application, with the real tile set.
func BenchmarkGeneration(b *testing.B) {
	entries := loadRealTileEntries(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := NewWfc(31, 18, entries)
		w.StartRender()
	}
}

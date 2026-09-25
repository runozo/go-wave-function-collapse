package wfc

import (
	"log"
	"runtime"
	"sync"

	"math/rand"

	"github.com/runozo/go-wave-function-collapse/assets"
)

// The tile to be drawn on screen
type Tile struct {
	Collapsed bool
	Name      string
	Options   []string
}

// IterateResult describes the outcome of a single Iterate call.
type IterateResult int

const (
	// Iterating means the generation is still in progress and more cells can be collapsed.
	Iterating IterateResult = iota
	// Rendered means every cell has been collapsed and the map is complete.
	Rendered
	// Contradiction means at least one uncollapsed cell has no options left.
	Contradiction
)

// DefaultMaxRestarts is the default number of times StartRender restarts a
// generation after hitting a contradiction before giving up.
const DefaultMaxRestarts = 50

type Wfc struct {
	Tiles          []Tile
	TileEntries    map[string]assets.TileEntry
	IsRunning      bool
	IsRendered     bool
	numOfTilesX    int
	numOfTilesY    int
	TotalTiles     int
	ProcessedTiles int
	MaxRestarts    int

	// sweepBuffer is reused by ElaborateGrid as the read-only snapshot of Tiles
	// for the duration of a concurrent sweep, avoiding per-sweep allocations.
	sweepBuffer []Tile
}

// NewWfc returns a new WFC with the given number of tiles in the X and Y directions, and the given tile entries.
//
// Parameters:
// - numOfTilesX: the number of tiles in the X direction.
// - numOfTilesY: the number of tiles in the Y direction.
// - tileEntries: a map of tile names to TileEntries.
//
// Returns:
// - *Wfc: a pointer to a new WFC.
func NewWfc(numOfTilesX, numOfTilesY int, tileEntries map[string]assets.TileEntry) *Wfc {
	// filter in only groud type tiles
	filteredTileEntries := make(map[string]assets.TileEntry)
	for k, v := range tileEntries {
		if v.Type == "ground" {
			filteredTileEntries[k] = v
		}
	}
	wfc := &Wfc{
		Tiles:          make([]Tile, numOfTilesX*numOfTilesY),
		TileEntries:    filteredTileEntries,
		numOfTilesX:    numOfTilesX,
		numOfTilesY:    numOfTilesY,
		TotalTiles:     numOfTilesX * numOfTilesY,
		IsRunning:      false,
		IsRendered:     false,
		ProcessedTiles: 0,
		MaxRestarts:    DefaultMaxRestarts,
	}
	wfc.Reset()
	return wfc
}

// IntInSlice checks if a given integer is present in a slice of integers.
//
// Parameters:
// - a: the integer to search for.
// - slice: the slice of integers to search in.
//
// Returns:
// - bool: true if the integer is found in the slice, false otherwise.
func (wfc *Wfc) IntInSlice(a int, slice []int) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}

// FilterOptions filters the original options based on the provided options slice.
//
// It takes in two parameters:
// - orig []string: the original options slice
// - options []string: the options to filter by
// Returns []string: the filtered options slice
func (wfc *Wfc) FilterOptions(orig, options []string) []string {
	filtered := make([]string, 0, len(orig))

	for _, o := range orig {
		for _, b := range options {
			if b == o {
				filtered = append(filtered, o)
				break
			}
		}
	}
	return filtered
}

// Reset resets all the tiles in the WFC to their initial state, with all options available.
// It does not reset the TileEntries, so the same tile entries will be used as previously.

func (wfc *Wfc) Reset() {
	// create a slice of all the options available
	initialOptions := []string{}

	for k, v := range wfc.TileEntries {
		if len(v.Options) >= 4 {
			// log.Println("appending", k)
			initialOptions = append(initialOptions, k)
		}
	}

	// setup tiles with all the options enabled
	for i := 0; i < len(wfc.Tiles); i++ {
		wfc.Tiles[i] = Tile{
			Collapsed: false,
			Options:   initialOptions,
		}
	}
	wfc.ProcessedTiles = 0
}

// LeastEntropyCellIndexes returns the indexes of the cells with the least entropy.
//
// This function iterates through all the tiles and identifies those that are not
// collapsed and have the fewest available options, which represents the least entropy.
// It returns a slice of indexes corresponding to these cells. If multiple cells have
// the same minimum entropy, all their indexes are included in the result.
//
// Returns:
// - []int: a slice of integers representing the indexes of the cells with the least entropy.

func (wfc *Wfc) LeastEntropyCellIndexes() []int {
	minEntropy := len(wfc.TileEntries)
	minEntropyIndexes := []int{}
	for index, tile := range wfc.Tiles {
		if !tile.Collapsed && len(tile.Options) < minEntropy {
			minEntropy = len(tile.Options)
			minEntropyIndexes = []int{index}
		} else if !tile.Collapsed && len(tile.Options) == minEntropy {
			minEntropyIndexes = append(minEntropyIndexes, index)
		}
	}
	// log.Println("minEntropyIndexes", len(minEntropyIndexes), "minEntropy", minEntropy)
	return minEntropyIndexes
}

// RandomOptionWithWeight selects a random option for a tile at the given index,
// taking into account the weights of each option.
//
// It calculates the total weight of all available options, then chooses a random
// weight within that total. The function iterates through the options, summing
// their weights until the random weight is less than the cumulative weight, at
// which point it returns the current option. This ensures that options with
// higher weights have a higher probability of being selected.
//
// Parameters:
// - index: the index of the tile for which to select an option.
//
// Returns:
// - string: the randomly selected option based on weights.

func (wfc *Wfc) RandomOptionWithWeight(index int) string {
	options := wfc.Tiles[index].Options
	if len(options) == 0 {
		// A cell with no options is a contradiction: return early instead of
		// panicking on rand.Intn(0). Iterate detects this before collapsing.
		return ""
	}

	var totalWeight int
	for _, option := range options {
		totalWeight += wfc.TileEntries[option].Weight
	}

	if totalWeight <= 0 {
		// All weights are zero (or negative): fall back to a uniform choice.
		return options[rand.Intn(len(options))]
	}

	randomWeight := rand.Intn(totalWeight)

	totalWeight = 0
	for _, option := range options {
		totalWeight += wfc.TileEntries[option].Weight
		if randomWeight < totalWeight {
			return option
		}
	}
	// should never reach here
	return options[rand.Intn(len(options))]
}

// CollapseCell collapses a cell.
//
// Parameters:
// - cellIndex: the index of the cell to collapse.
//
// Returns:
//   - nothing. It modifies the cell at the given index to have a collapsed state
//     with a randomly chosen option.
func (wfc *Wfc) CollapseCell(index int) {
	if len(wfc.Tiles[index].Options) == 0 {
		// Nothing to collapse: leave the cell as-is so the contradiction stays
		// detectable instead of turning it into a collapsed, empty-named tile.
		return
	}
	// collapse a cell with least entropy
	randomOption := wfc.RandomOptionWithWeight(index)
	wfc.Tiles[index] = Tile{
		Options:   []string{randomOption},
		Name:      randomOption,
		Collapsed: true,
	}
}

// GetAvailableOptions takes a cell index and a direction, and returns a slice of strings that represent all the available options for the given direction.
//
// It reads the current (live) state of the grid. The concurrent sweep uses
// ElaborateGrid instead, which reads from a stable snapshot.
//
// Parameters:
// - cellIndex: the index of the cell to retrieve the available options for.
// - direction: a string representing the direction to retrieve the available options for. Can be "up", "right", "down", or "left".
//
// Returns:
// - []string: a slice of strings representing the available options
func (wfc *Wfc) GetAvailableOptions(cellIndex int, direction string) []string {
	return wfc.getAvailableOptions(wfc.Tiles, cellIndex, direction)
}

// getAvailableOptions is the read-only core of GetAvailableOptions. Neighbor
// options are read from src, which during a concurrent sweep is a snapshot of
// the grid taken before the sweep started.
func (wfc *Wfc) getAvailableOptions(src []Tile, cellIndex int, direction string) []string {
	availableOptions := make([]string, 0, len(src[cellIndex].Options))
	for _, o := range src[cellIndex].Options {
		availableOptions = append(availableOptions, wfc.TileEntries[o].Options[direction]...)
	}
	return availableOptions
}

// ElaborateCell takes an x and y coordinate and elaborates a cell by filtering
// its available options based on the options of its adjacent cells.
//
// It reads and writes the live grid, so it is meant for serial use (tests and
// single-threaded callers). The concurrent sweep must use ElaborateGrid.
//
// Parameters:
// - x: the x coordinate of the cell to elaborate.
// - y: the y coordinate of the cell to elaborate.
//
// Returns:
//   - nothing. It modifies the options of the cell at the given x and y
//     coordinates.
func (wfc *Wfc) ElaborateCell(x, y int) {
	wfc.elaborateCell(wfc.Tiles, wfc.numOfTilesX, wfc.numOfTilesY, x, y)
}

// elaborateCell filters the options of the cell at (x, y) using the neighbor
// options read from src, and writes the result back into wfc.Tiles. Only the
// cell at (x, y) is written, so distinct cells can be elaborated concurrently
// as long as src is not being mutated.
func (wfc *Wfc) elaborateCell(src []Tile, numOfTilesX, numOfTilesY, x, y int) {
	index := y*numOfTilesX + x
	if wfc.Tiles[index].Collapsed {
		return
	}
	// Look UP
	if y > 0 {
		wfc.Tiles[index].Options = wfc.FilterOptions(
			wfc.Tiles[index].Options,
			wfc.getAvailableOptions(src, (y-1)*numOfTilesX+x, "down"),
		)
	}
	// Look RIGHT
	if x < numOfTilesX-1 {
		wfc.Tiles[index].Options = wfc.FilterOptions(
			wfc.Tiles[index].Options,
			wfc.getAvailableOptions(src, y*numOfTilesX+x+1, "left"),
		)
	}
	// Look DOWN
	if y < numOfTilesY-1 {
		wfc.Tiles[index].Options = wfc.FilterOptions(
			wfc.Tiles[index].Options,
			wfc.getAvailableOptions(src, (y+1)*numOfTilesX+x, "up"),
		)
	}
	// Look LEFT
	if x > 0 {
		wfc.Tiles[index].Options = wfc.FilterOptions(
			wfc.Tiles[index].Options,
			wfc.getAvailableOptions(src, y*numOfTilesX+x-1, "right"),
		)
	}
}

// ElaborateGrid performs a full-grid constraint propagation sweep in parallel.
//
// It is race-free: before launching the workers it takes a snapshot of Tiles
// (reusing sweepBuffer) and every worker reads neighbor options only from that
// snapshot, writing solely its own cell in wfc.Tiles. Because no options slice
// is ever mutated in place (FilterOptions and CollapseCell always allocate a new
// slice), a shallow snapshot is enough and stays stable for the whole sweep.
//
// Parameters:
// - numOfTilesX: the number of tiles in the X direction.
// - numOfTilesY: the number of tiles in the Y direction.
func (wfc *Wfc) ElaborateGrid(numOfTilesX, numOfTilesY int) {
	n := numOfTilesX * numOfTilesY
	if len(wfc.sweepBuffer) != n {
		wfc.sweepBuffer = make([]Tile, n)
	}
	copy(wfc.sweepBuffer, wfc.Tiles[:n])

	var wg sync.WaitGroup
	wg.Add(n)
	for y := 0; y < numOfTilesY; y++ {
		for x := 0; x < numOfTilesX; x++ {
			go func(x, y int) {
				defer wg.Done()
				wfc.elaborateCell(wfc.sweepBuffer, numOfTilesX, numOfTilesY, x, y)
			}(x, y)
		}
	}
	wg.Wait()
}

// Iterate iteratively collapses cells with the least entropy until no more collapsable cells are available or the rendering
// process is stopped.
//
// It first checks if there are any more collapsable cells. If not, it marks the map as rendered and returns Rendered.
// If the least-entropy cells have no options left, the wave function has collapsed into a contradiction: it returns
// Contradiction without collapsing anything, so the caller can restart the generation.
// Otherwise it randomly selects one of the least-entropy cells and collapses it, then iteratively
// elaborates the adjacent cells by filtering their available options based on the options of the adjacent cells.
//
// Parameters:
// - numOfTilesX: the number of tiles in the X direction.
// - numOfTilesY: the number of tiles in the Y direction.
//
// Returns:
// - IterateResult: Rendered when complete, Contradiction when stuck, Iterating otherwise.
func (wfc *Wfc) Iterate(numOfTilesX, numOfTilesY int) IterateResult {
	leastEntropyIndexes := wfc.LeastEntropyCellIndexes()

	if len(leastEntropyIndexes) == 0 {
		// log.Println("Playfiled is rendered. No more collapsable cells.", "tiles involved", len(wfc.Tiles))
		wfc.IsRunning = false
		wfc.IsRendered = true
		return Rendered
	}

	if len(wfc.Tiles[leastEntropyIndexes[0]].Options) == 0 {
		// At least one cell has no options left: the constraint propagation
		// reached an impossible state.
		return Contradiction
	}

	collapseIndex := leastEntropyIndexes[rand.Intn(len(leastEntropyIndexes))]
	wfc.CollapseCell(collapseIndex)
	wfc.ProcessedTiles++
	wfc.ElaborateGrid(numOfTilesX, numOfTilesY)
	return Iterating
}

// StartRender initializes and starts the rendering process using the Wave Function Collapse algorithm.
//
// This method first checks if the rendering is already running. If so, it logs a message and returns.
// If not running, it sets the `IsRunning` flag to true and resets the state. Then, it iteratively
// collapses cells with the least entropy until no more collapsable cells are available or the rendering
// process is stopped.
//
// If the propagation reaches a contradiction (a cell with no options left), the generation is
// restarted from scratch, up to MaxRestarts times. When the limit is reached the map is left as-is
// and the method gives up without panicking.

func (wfc *Wfc) StartRender() {
	if wfc.IsRunning {
		log.Println("wfc is already running")
		return
	}
	wfc.IsRunning = true
	wfc.IsRendered = false

	for attempt := 0; attempt <= wfc.MaxRestarts; attempt++ {
		if attempt > 0 {
			log.Printf("wfc: contradiction detected, restarting (%d/%d)", attempt, wfc.MaxRestarts)
		}
		wfc.Reset()

		for {
			switch wfc.Iterate(wfc.numOfTilesX, wfc.numOfTilesY) {
			case Rendered:
				return
			case Contradiction:
				// Break the inner loop to restart the whole generation.
				goto restart
			case Iterating:
				if !wfc.IsRunning {
					return
				}
				runtime.Gosched()
			}
		}
	restart:
	}

	log.Println("wfc: max restarts reached, giving up")
	wfc.IsRendered = true
	wfc.IsRunning = false
}

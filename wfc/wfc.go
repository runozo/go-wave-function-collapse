package wfc

import (
	"log"
	"sync"

	"math/rand"

	"github.com/runozo/go-wave-function-collapse/assets"
)

type Tile struct {
	Collapsed bool
	Name      string
	Options   []string
}

type Wfc struct {
	Tiles       []Tile
	TileEntries map[string]assets.TileEntry
	IsRunning   bool
	numOfTilesX int
	numOfTilesY int
}

func NewWfc(numOfTilesX, numOfTilesY int, tileEntries map[string]assets.TileEntry) *Wfc {
	wfc := &Wfc{
		Tiles:       make([]Tile, numOfTilesX*numOfTilesY),
		TileEntries: tileEntries,
		numOfTilesX: numOfTilesX,
		numOfTilesY: numOfTilesY,
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

// CollapseCell collapses a cell with least entropy.
//
// Parameters:
// - cellIndex: the index of the cell to collapse.
//
// Returns:
//   - nothing. It modifies the cell at the given index to have a collapsed state
//     with a randomly chosen option.
func (wfc *Wfc) CollapseCell(cellIndex int) {
	// collapse a cell with least entropy
	randomOption := wfc.Tiles[cellIndex].Options[rand.Intn(len(wfc.Tiles[cellIndex].Options))]
	wfc.Tiles[cellIndex] = Tile{
		Options:   []string{randomOption},
		Name:      randomOption,
		Collapsed: true,
	}
}

func (wfc *Wfc) GetAvailableOptions(cellIndex int, direction string) []string {
	availableOptions := make([]string, 0, len(wfc.Tiles[cellIndex].Options))
	for _, o := range wfc.Tiles[cellIndex].Options {
		availableOptions = append(availableOptions, wfc.TileEntries[o].Options[direction]...)
	}
	return availableOptions
}

// ElaborateCell takes an x and y coordinate and elaborates a cell by filtering
// its available options based on the options of its adjacent cells.
//
// Parameters:
// - x: the x coordinate of the cell to elaborate.
// - y: the y coordinate of the cell to elaborate.
//
// Returns:
//   - nothing. It modifies the options of the cell at the given x and y
//     coordinates.
func (wfc *Wfc) ElaborateCell(x, y int) {
	numOfTilesX := wfc.numOfTilesX
	numOfTilesY := wfc.numOfTilesY
	index := y*numOfTilesX + x
	if !wfc.Tiles[index].Collapsed {
		// Look UP
		if y > 0 {
			wfc.Tiles[index].Options = wfc.FilterOptions(
				wfc.Tiles[index].Options,
				wfc.GetAvailableOptions((y-1)*numOfTilesX+x, "down"),
			)
		}
		// Look RIGHT
		if x < numOfTilesX-1 {
			wfc.Tiles[index].Options = wfc.FilterOptions(
				wfc.Tiles[index].Options,
				wfc.GetAvailableOptions(y*numOfTilesX+x+1, "left"),
			)
		}
		// Look DOWN
		if y < numOfTilesY-1 {
			wfc.Tiles[index].Options = wfc.FilterOptions(
				wfc.Tiles[index].Options,
				wfc.GetAvailableOptions((y+1)*numOfTilesX+x, "up"),
			)
		}
		// Look LEFT
		if x > 0 {
			wfc.Tiles[index].Options = wfc.FilterOptions(
				wfc.Tiles[index].Options,
				wfc.GetAvailableOptions(y*numOfTilesX+x-1, "right"),
			)
		}
	}
}

// Iterate collapses one cell with least entropy, then elaborates all cells.
// The collapsing and elaboration is done concurrently, but the elaboration
// of each row is done sequentially to avoid race conditions.
// If all cells are collapsed, it sets IsRunning to false and returns false.
// Otherwise, it returns true.
func (wfc *Wfc) Iterate(numOfTilesX, numOfTilesY int) bool {
	leastEntropyIndexes := wfc.LeastEntropyCellIndexes()

	if len(leastEntropyIndexes) == 0 {
		// log.Println("Playfiled is rendered. No more collapsable cells.", "tiles involved", len(wfc.Tiles))
		wfc.IsRunning = false
		return false
	} else {
		collapseIndex := leastEntropyIndexes[rand.Intn(len(leastEntropyIndexes))]
		wfc.CollapseCell(collapseIndex)
		var wg sync.WaitGroup
		for y := 0; y < numOfTilesY; y++ {
			wg.Add(numOfTilesX)
			for x := 0; x < numOfTilesX; x++ {
				go func(x, y int) {
					defer wg.Done()
					wfc.ElaborateCell(x, y)
				}(x, y)
			}
			wg.Wait()
		}
	}
	return true
}

// StartRender initializes and starts the rendering process using the Wave Function Collapse algorithm.
//
// This method first checks if the rendering is already running. If so, it logs a message and returns.
// If not running, it sets the `IsRunning` flag to true and resets the state. Then, it iteratively
// collapses cells with the least entropy until no more collapsable cells are available or the rendering
// process is stopped.

func (wfc *Wfc) StartRender() {
	if wfc.IsRunning {
		log.Println("wfc is already running")
		return
	}
	wfc.IsRunning = true
	wfc.Reset()
	for wfc.Iterate(wfc.numOfTilesX, wfc.numOfTilesY) {
		if !wfc.IsRunning {
			return
		}
	}
}

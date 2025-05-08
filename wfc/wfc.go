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

func (wfc *Wfc) Reset() {
	// create a slice of all the options available
	initialOptions := []string{}
	for k, v := range wfc.TileEntries {
		if len(v.Options) >= 4 {
			// log.Println("appending", k)
			initialOptions = append(initialOptions, k)
		}
	}

	// setup tiles with all the options enabled and reset the image name
	for i := 0; i < len(wfc.Tiles); i++ {
		wfc.Tiles[i] = Tile{
			Collapsed: false,
			Options:   initialOptions,
		}
	}
}

// GetLeastEntropyIndexes returns a slice of integers representing the indexes of the tiles with the least entropy.
//
// Parameters:
// - tiles: a pointer to a slice of Tile structs representing the tiles.
//
// Return:
// - []int: a slice of integers representing the indexes of the tiles with the least entropy.
func (wfc *Wfc) LeastEntropyCells() []int {
	minEntropy := len(wfc.TileEntries)
	minEntropyIndexes := make([]int, 0, 10)
	for index, tile := range wfc.Tiles {
		if !tile.Collapsed {
			cellEntropy := len(tile.Options)
			if cellEntropy < minEntropy {
				minEntropy = cellEntropy
				minEntropyIndexes = []int{index}
			} else if cellEntropy == minEntropy {
				minEntropyIndexes = append(minEntropyIndexes, index)
			}
		}
	}
	// log.Println("minEntropyIndexes", len(minEntropyIndexes), "minEntropy", minEntropy)
	return minEntropyIndexes
}

// CollapseCell collapses a cell with the least entropy.
//
// Parameters:
// - game: a pointer to a Game instance.
// - randomIndex: an integer representing the index of the cell to collapse.
func (wfc *Wfc) CollapseCell(cellIndex int) {
	// collapse a cell with least entropy
	randomOption := wfc.Tiles[cellIndex].Options[rand.Intn(len(wfc.Tiles[cellIndex].Options))]
	wfc.Tiles[cellIndex] = Tile{
		Options:   []string{randomOption},
		Name:      randomOption,
		Collapsed: true,
	}
}

func (wfc *Wfc) ElaborateCell(x, y int) {
	numOfTilesX := wfc.numOfTilesX
	numOfTilesY := wfc.numOfTilesY
	index := y*numOfTilesX + x
	if !wfc.Tiles[index].Collapsed {
		// Look UP
		if y > 0 {
			var availableOptions []string
			for _, o := range wfc.Tiles[(y-1)*numOfTilesX+x].Options {
				availableOptions = append(availableOptions, wfc.TileEntries[o].Options["down"]...)
			}
			wfc.Tiles[index].Options = wfc.FilterOptions(
				wfc.Tiles[index].Options,
				availableOptions,
			)
		}
		// Look RIGHT
		if x < numOfTilesX-1 {
			var availableOptions []string
			for _, o := range wfc.Tiles[y*numOfTilesX+x+1].Options {
				availableOptions = append(availableOptions, wfc.TileEntries[o].Options["left"]...)
			}
			wfc.Tiles[index].Options = wfc.FilterOptions(
				wfc.Tiles[index].Options,
				availableOptions,
			)

		}
		// Look DOWN
		if y < numOfTilesY-1 {
			var availableOptions []string
			for _, o := range wfc.Tiles[(y+1)*numOfTilesX+x].Options {
				availableOptions = append(availableOptions, wfc.TileEntries[o].Options["up"]...)
			}
			wfc.Tiles[index].Options = wfc.FilterOptions(
				wfc.Tiles[index].Options,
				availableOptions,
			)
		}
		// Look LEFT
		if x > 0 {
			var availableOptions []string
			for _, o := range wfc.Tiles[y*numOfTilesX+x-1].Options {
				availableOptions = append(availableOptions, wfc.TileEntries[o].Options["right"]...)
			}
			wfc.Tiles[index].Options = wfc.FilterOptions(
				wfc.Tiles[index].Options,
				availableOptions,
			)
		}
	}
}

func (wfc *Wfc) Iterate(numOfTilesX, numOfTilesY int) bool {
	// pick the minimum entropy indexes
	leastEntropyIndexes := wfc.LeastEntropyCells()

	if len(leastEntropyIndexes) == 0 {
		log.Println("Playfiled is rendered. No more collapsable cells.", "tiles involved", len(wfc.Tiles))
		return false
	} else {
		wfc.CollapseCell(leastEntropyIndexes[rand.Intn(len(leastEntropyIndexes))])
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

var isWfcRunning bool = false

func (wfc *Wfc) StartRender() {
	if isWfcRunning {
		log.Println("wfc is already running")
		return
	}
	isWfcRunning = true
	defer func() {
		isWfcRunning = false
	}()
	wfc.Reset()
	for wfc.Iterate(wfc.numOfTilesX, wfc.numOfTilesY) {
	}
}

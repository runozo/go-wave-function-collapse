package wfc

import (
	"log"

	"math/rand"
)

type Tile struct {
	Collapsed bool
	ImageName string
	Options   []string
}

type Wfc struct {
	Tiles       []Tile
	IsRunning   bool
	numOfTilesX int
	numOfTilesY int
}

const (
	ruleUP    = 0
	ruleRIGHT = 1
	ruleDOWN  = 2
	ruleLEFT  = 3
)

func NewWfc(numOfTilesX, numOfTilesY int) *Wfc {
	return &Wfc{
		Tiles:       make([]Tile, numOfTilesX*numOfTilesY),
		numOfTilesX: numOfTilesX,
		numOfTilesY: numOfTilesY,
	}
}

// TODO: use json file
var tileOptions = map[string][]int{
	"tileGrass1.png":                     {0, 0, 0, 0}, // 0 grass
	"tileGrass2.png":                     {0, 0, 0, 0},
	"tileGrass_roadCornerLL.png":         {0, 0, 1, 1}, // 1 road with grass
	"tileGrass_roadCornerLR.png":         {0, 1, 1, 0},
	"tileGrass_roadCornerUL.png":         {1, 0, 0, 1},
	"tileGrass_roadCornerUR.png":         {1, 1, 0, 0},
	"tileGrass_roadCrossing.png":         {1, 1, 1, 1},
	"tileGrass_roadCrossingRound.png":    {1, 1, 1, 1},
	"tileGrass_roadEast.png":             {0, 1, 0, 1},
	"tileGrass_roadNorth.png":            {1, 0, 1, 0},
	"tileGrass_roadSplitE.png":           {1, 1, 1, 0},
	"tileGrass_roadSplitN.png":           {1, 1, 0, 1},
	"tileGrass_roadSplitS.png":           {0, 1, 1, 1},
	"tileGrass_roadSplitW.png":           {1, 0, 1, 1},
	"tileGrass_roadTransitionE.png":      {4, 3, 4, 1},
	"tileGrass_roadTransitionE_dirt.png": {4, 3, 4, 1},
	"tileGrass_roadTransitionN.png":      {3, 6, 1, 6},
	"tileGrass_roadTransitionN_dirt.png": {3, 6, 1, 6},
	"tileGrass_roadTransitionS.png":      {1, 8, 3, 8},
	"tileGrass_roadTransitionS_dirt.png": {1, 8, 3, 8},
	"tileGrass_roadTransitionW.png":      {5, 1, 5, 3},
	"tileGrass_roadTransitionW_dirt.png": {5, 1, 5, 3},
	"tileGrass_transitionE.png":          {4, 2, 4, 0},
	"tileGrass_transitionN.png":          {2, 6, 0, 6},
	"tileGrass_transitionS.png":          {0, 8, 2, 8},
	"tileGrass_transitionW.png":          {5, 0, 5, 2},
	"tileSand1.png":                      {2, 2, 2, 2},
	"tileSand2.png":                      {2, 2, 2, 2},
	"tileSand_roadCornerLL.png":          {2, 2, 3, 3},
	"tileSand_roadCornerLR.png":          {2, 3, 3, 2},
	"tileSand_roadCornerUL.png":          {3, 2, 2, 3},
	"tileSand_roadCornerUR.png":          {3, 3, 2, 2},
	"tileSand_roadCrossing.png":          {3, 3, 3, 3},
	"tileSand_roadCrossingRound.png":     {3, 3, 3, 3},
	"tileSand_roadEast.png":              {2, 3, 2, 3},
	"tileSand_roadNorth.png":             {3, 2, 3, 2},
	"tileSand_roadSplitE.png":            {3, 3, 3, 2},
	"tileSand_roadSplitN.png":            {3, 3, 2, 3},
	"tileSand_roadSplitS.png":            {2, 3, 3, 3},
	"tileSand_roadSplitW.png":            {3, 2, 3, 3},
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

// StringInSlice checks if a given string is present in a slice of strings.
//
// Parameters:
// - a: the string to search for.
// - slice: the slice of strings to search in.
//
// Returns:
// - bool: true if the string is found in the slice, false otherwise.
func (wfc *Wfc) StringInSlice(a string, slice []string) bool {
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
		if wfc.StringInSlice(o, options) {
			filtered = append(filtered, o)
		}
	}
	return filtered
}

// LookAndFilter filters options based on a set of rules.
//
// Parameters:
// - ruleIndexToProcess: the index of the rule to process.
// - ruleIndexToWatch: the index of the rule to watch.
// - optionsToProcess: the options to process.
// - optionsToWatch: the options to watch.
//
// Returns:
// - []string: the filtered options.
func (wfc *Wfc) LookAndFilter(ruleIndexToProcess, ruleIndexToWatch int, optionsToProcess, optionsToWatch []string) []string {
	rules := make([]int, 0, 5) // random capacity
	for _, optname := range optionsToWatch {
		rule := tileOptions[optname][ruleIndexToWatch]
		rules = append(rules, rule)
	}

	newoptions := make([]string, 0, 5) // random capacity
	for k, v := range tileOptions {
		if wfc.IntInSlice(v[ruleIndexToProcess], rules) {
			newoptions = append(newoptions, k)
		}
	}

	return wfc.FilterOptions(optionsToProcess, newoptions)
}

// func LookAndFilter2(tileId, direction, adjacentTileId string)

// ResetTilesOptions resets the options of each Tile in the provided slice to all available options.
//
// tiles: a pointer to a slice of Tiles that need their options reset.
func (wfc *Wfc) Reset() {
	// create a slice of all the options available
	initialOptions := make([]string, len(tileOptions))
	i := 0
	for k := range tileOptions {
		initialOptions[i] = k
		i++
	}

	// setup tiles with all the options enabled and reset the image name
	for i := 0; i < len(wfc.Tiles); i++ {
		wfc.Tiles[i] = Tile{
			ImageName: "",
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
	minEntropy := len(tileOptions)
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
		ImageName: randomOption,
		Collapsed: true,
	}
}

// Iterate iterates the wave function collapse algorithm.
//
// Parameters:
// - game: a pointer to a Game instance.
//
// Returns:
// - bool: true if the game is not rendered, false otherwise.
func (wfc *Wfc) Iterate(numOfTilesX, numOfTilesY int) bool {

	// pick the minimum entropy indexes
	leastEntropyIndexes := wfc.LeastEntropyCells()

	if len(leastEntropyIndexes) == 0 {
		log.Println("Playfiled is rendered. No more collapsable cells.", "tiles involved", wfc.Tiles)
		return false
	} else {
		wfc.CollapseCell(leastEntropyIndexes[rand.Intn(len(leastEntropyIndexes))])
		// scan all the cells to filter the corresponding options
		for y := 0; y < numOfTilesY; y++ {
			for x := 0; x < numOfTilesX; x++ {
				index := y*numOfTilesX + x
				if len(wfc.Tiles[index].Options) == 0 {
					// we did not found any option, let's restart
					log.Println("No more options found.. restarting!")
					wfc.Reset()
				}

				if !wfc.Tiles[index].Collapsed {
					// Look UP
					if y > 0 {
						wfc.Tiles[index].Options = wfc.LookAndFilter(ruleUP, ruleDOWN, wfc.Tiles[index].Options, wfc.Tiles[(y-1)*numOfTilesX+x].Options)
					}
					// Look RIGHT
					if x < numOfTilesX-1 {
						wfc.Tiles[index].Options = wfc.LookAndFilter(ruleRIGHT, ruleLEFT, wfc.Tiles[index].Options, wfc.Tiles[y*numOfTilesX+x+1].Options)
					}
					// Look DOWN
					if y < numOfTilesY-1 {
						wfc.Tiles[index].Options = wfc.LookAndFilter(ruleDOWN, ruleUP, wfc.Tiles[index].Options, wfc.Tiles[(y+1)*numOfTilesX+x].Options)
					}
					// Look LEFT
					if x > 0 {
						wfc.Tiles[index].Options = wfc.LookAndFilter(ruleLEFT, ruleRIGHT, wfc.Tiles[index].Options, wfc.Tiles[y*numOfTilesX+x-1].Options)
					}
				}
			}
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

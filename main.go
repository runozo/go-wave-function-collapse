package main

import (
	"flag"
	"fmt"
	_ "image/png"
	"log"
	"log/slog"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/runozo/go-wave-function-collapse/assets"
)

const (
	screenWidth  = 1960
	screenHeight = 1088
	tileWidth    = 64
	tileHeight   = 64
	ruleUP       = 0
	ruleRIGHT    = 1
	ruleDOWN     = 2
	ruleLEFT     = 3
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

type Tile struct {
	collapsed bool
	image     *ebiten.Image
	options   []string
}

type Game struct {
	width       int
	height      int
	assets      *assets.Assets
	isRendered  bool
	numOfTilesX int
	numOfTilesY int
	tiles       []Tile
}

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

// filterOptions filters the original options based on the provided options slice.
//
// It takes in two parameters:
// - orig []string: the original options slice
// - options []string: the options to filter by
// Returns []string: the filtered options slice
func filterOptions(orig, options []string) []string {
	var filtered []string
	for _, o := range orig {
		if stringInSlice(o, options) {
			filtered = append(filtered, o)
		}
	}
	return filtered
}

// stringInSlice checks if a string is present in a slice of strings.
//
// It takes a string to search for and a slice of strings to search in and returns a boolean.
func stringInSlice(a string, slice []string) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}

// intInSlice checks if an integer is in a slice of integers.
//
// a int - the integer to check for in the slice
// list []int - the slice of integers to search
// bool - true if the integer is found in the slice, false otherwise
func intInSlice(a int, slice []int) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}

func resetTilesOptions(tiles *[]Tile) {
	// create a slice of all the options available
	initialOptions := make([]string, len(tileOptions))
	i := 0
	for k := range tileOptions {
		initialOptions[i] = k
		i++
	}

	// setup tiles with all the options enabled and a black square as image
	for i := 0; i < len(*tiles); i++ {
		(*tiles)[i].options = initialOptions
		(*tiles)[i].image = ebiten.NewImage(tileWidth, tileHeight)
		(*tiles)[i].collapsed = false
	}
}

// getMinEntropyIndexes returns the indexes of cells with the minimum entropy.
//
// It takes a pointer to a 2D slice of strings, cells, as input.
// The function iterates through each cell in the cells slice and checks if the length of the cell is greater than 1 and less than the current minimum entropy.
// If so, it updates the minimum entropy and resets the minEntropyIndexes slice to contain only the current index.
// If the length of the cell is equal to the current minimum entropy, the index is appended to the minEntropyIndexes slice.
// Finally, the function returns the minEntropyIndexes slice.
//
// Parameters:
// - cells: a pointer to a 2D slice of strings representing the cells
//
// Return type:
// - []int: a slice of integers representing the indexes of cells with the minimum entropy
func getMinEntropyIndexes(tiles *[]Tile) []int {
	minEntropy := 32767
	minEntropyIndexes := make([]int, 0)
	for i, tile := range *tiles {
		if !tile.collapsed {
			cellEntropy := len(tile.options)
			// slog.Info("Entropy", "index", i, "entropy", cellEntropy)
			if cellEntropy > 1 && cellEntropy < minEntropy {
				minEntropy = cellEntropy
				minEntropyIndexes = []int{i}
			} else if cellEntropy > 1 && cellEntropy == minEntropy {
				minEntropyIndexes = append(minEntropyIndexes, i)
			}
		}
	}
	return minEntropyIndexes
}

// collapseRandomCellWithMinEntropy collapses a random cell with the minimum entropy.
//
// Parameters:
// - tiles: a pointer to a slice of Tile representing the game tiles
// - minEntropyIndexes: a pointer to a slice of integers representing the indexes of cells with the minimum entropy
//
// Return type:
// - int: the index of the collapsed cell
func collapseRandomCellWithMinEntropy(tiles *[]Tile, minEntropyIndexes *[]int) int {
	// collapse random cell with least entropy
	index := (*minEntropyIndexes)[rand.Intn(len(*minEntropyIndexes))]

	(*tiles)[index].options = []string{(*tiles)[index].options[rand.Intn(len((*tiles)[index].options))]}
	return index
}

// lookAndFilter applies a rule-based filtering to the optionsToProcess slice
//
// It takes two integer rule indexes, two slices of strings (optionsToProcess and optionsToWatch) as parameters
// Returns a slice of strings
func lookAndFilter(ruleIndexToProcess, ruleIndexToWatch int, optionsToProcess, optionsToWatch []string) []string {
	rules := make([]int, 0, 5) // random capacity
	for _, optname := range optionsToWatch {
		rule := tileOptions[optname][ruleIndexToWatch]
		rules = append(rules, rule)
	}

	newoptions := make([]string, 0, 5) // random capacity
	for k, v := range tileOptions {
		if intInSlice(v[ruleIndexToProcess], rules) {
			newoptions = append(newoptions, k)
		}
	}

	return filterOptions(optionsToProcess, newoptions)
}

func renderPlayfield(game *Game) {
	startTime := time.Now()
	defer func() {
		endTime := time.Now()
		duration := endTime.Sub(startTime)
		slog.Info("Rendering of playfield took", "duration", duration)
	}()

	for !game.isRendered {
		// pick the minimum entropy indexes
		minEntropyIndexes := getMinEntropyIndexes(&game.tiles)

		if len(minEntropyIndexes) <= 0 {
			slog.Info("Playfiled is rendered. No more collapsable cells.", "tiles", len(game.tiles))
			for i := 0; i < len(game.tiles); i++ {
				if !game.tiles[i].collapsed {
					game.tiles[i].image = game.assets.GetSprite(game.tiles[i].options[0])
					game.tiles[i].collapsed = true
				}
			}
			game.isRendered = true
		} else {
			collapsedIndex := collapseRandomCellWithMinEntropy(&game.tiles, &minEntropyIndexes)
			game.tiles[collapsedIndex].image = game.assets.GetSprite(game.tiles[collapsedIndex].options[0])
			game.tiles[collapsedIndex].collapsed = true

			for y := 0; y < game.numOfTilesY; y++ {
				for x := 0; x < game.numOfTilesX; x++ {
					index := y*game.numOfTilesX + x
					if len(game.tiles[index].options) == 0 {
						// we did not found any options, let's restart
						slog.Info("Restarting!")
						resetTilesOptions(&game.tiles)
					}

					if !game.tiles[index].collapsed {
						// Look UP
						if y > 0 {
							game.tiles[index].options = lookAndFilter(ruleUP, ruleDOWN, game.tiles[index].options, game.tiles[(y-1)*game.numOfTilesX+x].options)
						}
						// Look RIGHT
						if x < game.numOfTilesX-1 {
							game.tiles[index].options = lookAndFilter(ruleRIGHT, ruleLEFT, game.tiles[index].options, game.tiles[y*game.numOfTilesX+x+1].options)
						}
						// Look DOWN
						if y < game.numOfTilesY-1 {
							game.tiles[index].options = lookAndFilter(ruleDOWN, ruleUP, game.tiles[index].options, game.tiles[(y+1)*game.numOfTilesX+x].options)
						}
						// Look LEFT
						if x > 0 {
							game.tiles[index].options = lookAndFilter(ruleLEFT, ruleRIGHT, game.tiles[index].options, game.tiles[y*game.numOfTilesX+x-1].options)
						}
					}
				}
			}
		}
	}
}

func (g *Game) Update() error {
	renderPlayfield(g)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	var i int
	for y := 0; y < screenHeight; y += tileHeight {
		for x := 0; x < screenWidth; x += tileWidth {
			if x%tileWidth == 0 && y%tileHeight == 0 && i < len(g.tiles) {
				ops := &ebiten.DrawImageOptions{}
				ops.GeoM.Translate(float64(x), float64(y))
				screen.DrawImage(g.tiles[i].image, ops)
				i++
			}
		}
	}

	// text.Draw(screen, fmt.Sprintf("CURSOR KEYS: move tank. SPACE: shoot. T: new random tank"), nil, 10, 10, color.Black)
	ebitenutil.DebugPrint(screen, fmt.Sprintf("SPACEBAR: generate new map"))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	g := &Game{
		assets:      assets.NewAssets(),
		width:       screenWidth,
		height:      screenHeight,
		isRendered:  false,
		numOfTilesX: screenWidth/tileWidth + 1,
		numOfTilesY: screenHeight/tileHeight + 1,
		tiles:       make([]Tile, (screenWidth/tileWidth+1)*(screenHeight/tileHeight+1)),
	}

	// init tiles
	resetTilesOptions(&(g.tiles))

	// init screen
	ebiten.SetFullscreen(true)
	err := ebiten.RunGame(g)
	if err != nil {
		panic(err)
	}

}

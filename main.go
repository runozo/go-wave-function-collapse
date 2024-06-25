package main

import (
	"flag"
	_ "image/png"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
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

// stringInSlice checks if a given string is present in a slice of strings.
//
// Parameters:
// - a: the string to search for.
// - slice: the slice of strings to search in.
//
// Returns:
// - bool: true if the string is found in the slice, false otherwise.
func stringInSlice(a string, slice []string) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}

// intInSlice checks if a given integer is present in a slice of integers.
//
// Parameters:
// - a: the integer to search for.
// - slice: the slice of integers to search in.
//
// Returns:
// - bool: true if the integer is found in the slice, false otherwise.
func intInSlice(a int, slice []int) bool {
	for _, b := range slice {
		if b == a {
			return true
		}
	}
	return false
}

// resetTilesOptions resets the options of each Tile in the provided slice to all available options.
//
// tiles: a pointer to a slice of Tiles that need their options reset.
func resetTilesOptions(tiles *[]Tile) {
	// create a slice of all the options available
	initialOptions := make([]string, len(tileOptions))
	i := 0
	for k := range tileOptions {
		initialOptions[i] = k
		i++
	}

	// setup tiles with all the options enabled and a black square as image
	black_square := ebiten.NewImage(tileWidth, tileHeight)
	for i := 0; i < len(*tiles); i++ {
		(*tiles)[i] = Tile{
			image:     black_square,
			collapsed: false,
			options:   initialOptions,
		}
	}
}

// getLeastEntropyIndexes returns a slice of integers representing the indexes of the tiles with the minimum entropy.
//
// Parameters:
// - tiles: a pointer to a slice of Tile structs representing the tiles.
//
// Return:
// - []int: a slice of integers representing the indexes of the tiles with the minimum entropy.
func getLeastEntropyIndexes(tiles *[]Tile) []int {
	minEntropy := 32767
	minEntropyIndexes := make([]int, 0)
	for index, tile := range *tiles {
		if !tile.collapsed {
			cellEntropy := len(tile.options)
			if cellEntropy > 1 && cellEntropy < minEntropy {
				minEntropy = cellEntropy
				minEntropyIndexes = []int{index}
			} else if cellEntropy > 1 && cellEntropy == minEntropy {
				minEntropyIndexes = append(minEntropyIndexes, index)
			}
		}
	}
	return minEntropyIndexes
}

// collapseRandomCellWithLeastEntropy collapses a random cell with the least entropy.
//
// Parameters:
// - tiles: a pointer to a slice of Tile structs representing the tiles.
// - minEntropyIndexes: a pointer to a slice of integers representing the indexes of the tiles with the minimum entropy.
//
// Returns:
// - int: the index of the collapsed tile.
func collapseRandomCellWithLeastEntropy(game *Game, minEntropyIndexes *[]int) {
	// collapse random cell with least entropy
	randomIndex := (*minEntropyIndexes)[rand.Intn(len(*minEntropyIndexes))]
	randomOption := game.tiles[randomIndex].options[rand.Intn(len(game.tiles[randomIndex].options))]
	game.tiles[randomIndex] = Tile{
		options:   []string{randomOption},
		image:     game.assets.GetSprite(randomOption),
		collapsed: true,
	}
}

// lookAndFilter filters options based on a set of rules.
//
// Parameters:
// - ruleIndexToProcess: the index of the rule to process.
// - ruleIndexToWatch: the index of the rule to watch.
// - optionsToProcess: the options to process.
// - optionsToWatch: the options to watch.
//
// Returns:
// - []string: the filtered options.
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

func iterateWaveFunctionCollapse(game *Game) {
	if !game.isRendered {
		// pick the minimum entropy indexes
		leastEntropyIndexes := getLeastEntropyIndexes(&game.tiles)

		if len(leastEntropyIndexes) == 0 {
			// collapse last remaining cells
			for i := 0; i < len(game.tiles); i++ {
				if !game.tiles[i].collapsed {
					game.tiles[i].image = game.assets.GetSprite(game.tiles[i].options[0])
					game.tiles[i].collapsed = true
				}
			}
			game.isRendered = true
			log.Println("Playfiled is rendered. No more collapsable cells.", "tiles", len(game.tiles))
		} else {
			collapseRandomCellWithLeastEntropy(game, &leastEntropyIndexes)
			// scan all the cells to filter the corresponding options
			for y := 0; y < game.numOfTilesY; y++ {
				for x := 0; x < game.numOfTilesX; x++ {
					index := y*game.numOfTilesX + x
					if len(game.tiles[index].options) == 0 {
						// we did not found any option, let's restart
						log.Println("No more options found.. restarting!")
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
	if ebiten.IsKeyPressed(ebiten.KeySpace) && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		resetTilesOptions(&(g.tiles))
		g.isRendered = false
	}
	iterateWaveFunctionCollapse(g)

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	var i int
	for y := 0; y < screenHeight; y += tileHeight {
		for x := 0; x < screenWidth; x += tileWidth {
			if i < len(g.tiles) {
				ops := &ebiten.DrawImageOptions{}
				ops.GeoM.Translate(float64(x), float64(y))
				screen.DrawImage(g.tiles[i].image, ops)
				i++
			}
		}
	}

	ebitenutil.DebugPrint(screen, "SPACEBAR: generate new map")
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

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

package main

import (
	"flag"
	_ "image/png"
	"log"
	"os"
	"runtime/pprof"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/runozo/go-wave-function-collapse/assets"
	"github.com/runozo/go-wave-function-collapse/wfc"
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

type Game struct {
	width       int
	height      int
	assets      *assets.Assets
	isRendered  chan bool
	numOfTilesX int
	numOfTilesY int
	tiles       []wfc.Tile
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeySpace) && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		go wfc.StartRendering(&g.tiles, g.numOfTilesX, g.numOfTilesY, g.isRendered)
	}

	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		os.Exit(0)
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	var i int

	for y := 0; y < screenHeight; y += tileHeight {
		for x := 0; x < screenWidth; x += tileWidth {
			ops := &ebiten.DrawImageOptions{}
			ops.GeoM.Translate(float64(x), float64(y))
			if g.tiles[i].ImageName != "" {
				screen.DrawImage(g.assets.GetSprite(g.tiles[i].ImageName), ops)
			} else {
				screen.DrawImage(ebiten.NewImage(tileWidth, tileHeight), ops)
			}
			i++
		}
	}

	ebitenutil.DebugPrint(screen, "SPACEBAR: generate new map  ESC: quit")
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
		isRendered:  make(chan bool),
		numOfTilesX: screenWidth/tileWidth + 1,
		numOfTilesY: screenHeight/tileHeight + 1,
		tiles:       make([]wfc.Tile, (screenWidth/tileWidth+1)*(screenHeight/tileHeight+1)),
	}

	// init screen
	ebiten.SetFullscreen(true)

	go wfc.StartRendering(&g.tiles, g.numOfTilesX, g.numOfTilesY, g.isRendered)

	err := ebiten.RunGame(g)

	if err != nil {
		panic(err)
	}
}

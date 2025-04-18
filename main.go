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
	width  int
	height int
	assets *assets.AssetsJSON
	wfc    *wfc.Wfc
}

func (g *Game) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeySpace) && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		go g.wfc.StartRender()
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
			if g.wfc.Tiles[i].ImageName != "" {
				screen.DrawImage(g.assets.GetSpriteJSON(g.wfc.Tiles[i].Name), ops)
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
	as := assets.NewAssetsJSON()
	g := &Game{
		assets: as,
		width:  screenWidth,
		height: screenHeight,
		wfc:    wfc.NewWfc(screenWidth/tileWidth+1, screenHeight/tileHeight+1, as.TileEntries),
	}

	// init screen
	ebiten.SetFullscreen(true)

	go g.wfc.StartRender()

	err := ebiten.RunGame(g)

	if err != nil {
		panic(err)
	}
}

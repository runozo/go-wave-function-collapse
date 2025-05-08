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
	width      int
	height     int
	assets     *assets.Assets
	wfc        *wfc.Wfc
	gif        *os.File
	iterations int
}

func (g *Game) Update() error {
	if g.iterations > 0 && !g.wfc.IsRunning {
		go g.wfc.StartRender()
		g.iterations--
		if g.iterations == 0 {
			os.Exit(0)
		}
	} else if g.iterations < 0 {
		if ebiten.IsKeyPressed(ebiten.KeySpace) && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			go g.wfc.StartRender()
		}

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
			if g.wfc.Tiles[i].Name != "" {
				screen.DrawImage(g.assets.GetSprite(g.wfc.Tiles[i].Name), ops)
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
var savegif = flag.String("gif", "", "render animation to a gif file")
var iterations = flag.Int("iterations", -1, "number of iterations before exit")

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

	var gif *os.File
	if *savegif != "" {
		var err error
		gif, err = os.Create(*savegif)
		if err != nil {
			log.Fatal(err)
		}
	}

	as := assets.NewAssets()
	g := &Game{
		assets:     as,
		width:      screenWidth,
		height:     screenHeight,
		wfc:        wfc.NewWfc(screenWidth/tileWidth+1, screenHeight/tileHeight+1, as.TileEntries),
		iterations: *iterations,
		gif:        gif,
	}

	// init screen
	ebiten.SetFullscreen(true)

	go g.wfc.StartRender()

	err := ebiten.RunGame(g)

	if err != nil {
		panic(err)
	}
}

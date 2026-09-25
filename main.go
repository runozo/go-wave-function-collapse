package main

import (
	"embed"
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

	// stepsPerFrame is how many cells are collapsed per rendered frame when the
	// generation is driven from the game loop. This keeps the collapse visible
	// on single-threaded targets (WebAssembly), where a background goroutine
	// would otherwise run to completion before the next frame is drawn.
	stepsPerFrame = 2
)

type Game struct {
	width      int
	height     int
	assets     *assets.Assets
	wfc        *wfc.Wfc
	iterations int

	// sprites and blank are precomputed once so Draw does not allocate or do
	// map lookups on every frame, for every cell.
	sprites map[string]*ebiten.Image
	blank   *ebiten.Image
}

//go:embed data/*
var embedfs embed.FS

func (g *Game) Update() error {
	if g.iterations > 0 && !g.wfc.IsRunning {
		go g.wfc.StartRender()
		g.iterations--
		if g.iterations == 0 {
			os.Exit(0)
		}
	} else if g.iterations < 0 {
		// Drive the generation from the game loop, a few cells per frame, so the
		// collapse is animated on every platform (including WebAssembly).
		if g.wfc.IsRunning {
			for i := 0; i < stepsPerFrame && g.wfc.IsRunning; i++ {
				g.wfc.Step()
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) && !g.wfc.IsRunning {
			g.wfc.BeginRender()
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
			if name := g.wfc.Tiles[i].Name; name != "" {
				if sprite := g.sprites[name]; sprite != nil {
					screen.DrawImage(sprite, ops)
				}
			} else {
				screen.DrawImage(g.blank, ops)
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

	// as := assets.NewAssets("data"+string(os.PathSeparator)+"allSprites_default.png", "data"+string(os.PathSeparator)+"mapped_tiles.json")
	tilesheetData, err := embedfs.ReadFile("data/allSprites_default.png")
	if err != nil {
		log.Fatal(err)
	}

	mappingData, err := embedfs.ReadFile("data/mapped_tiles.json")
	if err != nil {
		log.Fatal(err)
	}

	as := assets.NewAssets(tilesheetData, mappingData)

	// Precompute every sprite once; Draw then just indexes the map.
	sprites := make(map[string]*ebiten.Image, len(as.TileEntries))
	for name := range as.TileEntries {
		sprites[name] = as.GetSprite(name)
	}

	g := &Game{
		assets:     as,
		width:      screenWidth,
		height:     screenHeight,
		wfc:        wfc.NewWfc(screenWidth/tileWidth+1, screenHeight/tileHeight+1, as.TileEntries),
		iterations: *iterations,
		sprites:    sprites,
		blank:      ebiten.NewImage(tileWidth, tileHeight),
	}

	// init screen
	ebiten.SetFullscreen(true)

	if g.iterations < 0 {
		// Interactive mode: start the generation; Update advances it frame by frame.
		g.wfc.BeginRender()
	}

	err = ebiten.RunGame(g)

	if err != nil {
		panic(err)
	}
}

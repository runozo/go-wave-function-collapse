package assets

import (
	"embed"
	"encoding/xml"
	"image"
	_ "image/png"
	"io/fs"
	"log"
	"log/slog"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
)

//go:embed allSprites_default.png
//go:embed allSprites_default.xml
var assets embed.FS

const (
	spriteSheet  = "allSprites_default.png"
	xmlSpriteMap = "allSprites_default.xml"
)

type Assets struct {
	SpriteSheet       *ebiten.Image
	SpriteSheetXMLMap SpriteMap
}

type SubTexture struct {
	XMLName xml.Name `xml:"SubTexture"`
	Name    string   `xml:"name,attr"`
	X       int      `xml:"x,attr"`
	Y       int      `xml:"y,attr"`
	Width   int      `xml:"width,attr"`
	Height  int      `xml:"height,attr"`
}

type SpriteMap struct {
	XMLName     xml.Name     `xml:"TextureAtlas"`
	ImagePath   string       `xml:"imagePath,attr"`
	SubTextures []SubTexture `xml:"SubTexture"`
}

func NewAssets() *Assets {
	var spriteSheet = mustLoadImage(spriteSheet)
	var spriteMap = mustLoadXMLSpriteMap(xmlSpriteMap)

	return &Assets{
		SpriteSheet:       spriteSheet,
		SpriteSheetXMLMap: spriteMap,
	}
}

func (a *Assets) GetSprite(name string) *ebiten.Image {
	for i := 0; i < len(a.SpriteSheetXMLMap.SubTextures); i++ {
		subTexture := a.SpriteSheetXMLMap.SubTextures[i]
		if subTexture.Name == name {
			return a.SpriteSheet.SubImage(
				image.Rect(
					subTexture.X,
					subTexture.Y,
					subTexture.X+subTexture.Width,
					subTexture.Y+subTexture.Height,
				)).(*ebiten.Image)
		}
	}
	slog.Info("Sprite not found", "name", name)
	return nil
}

func (a *Assets) GetRandomTankBody() *ebiten.Image {
	bodies := []string{"tankBody_bigRed.png", "tankBody_bigRed_outline.png", "tankBody_blue.png", "tankBody_blue_outline.png", "tankBody_dark.png", "tankBody_darkLarge.png", "tankBody_darkLarge_outline.png", "tankBody_dark_outline.png", "tankBody_green.png", "tankBody_green_outline.png", "tankBody_huge.png", "tankBody_huge_outline.png", "tankBody_red.png", "tankBody_red_outline.png", "tankBody_sand.png", "tankBody_sand_outline.png", "tank_bigRed.png", "tank_blue.png", "tank_dark.png", "tank_darkLarge.png", "tank_green.png", "tank_huge.png", "tank_red.png", "tank_sand.png"}
	name := bodies[rand.Intn(len(bodies))]
	log.Printf("new tank body: %s\n", name)
	return a.GetSprite(name)
}

func (a *Assets) GetRandomTankBarrell() *ebiten.Image {
	barrels := []string{"tankDark_barrel1.png", "tankDark_barrel1_outline.png", "tankDark_barrel2.png", "tankDark_barrel2_outline.png", "tankDark_barrel3.png", "tankDark_barrel3_outline.png", "tankGreen_barrel1.png", "tankGreen_barrel1_outline.png", "tankGreen_barrel2.png", "tankGreen_barrel2_outline.png", "tankGreen_barrel3.png", "tankGreen_barrel3_outline.png", "tankRed_barrel1.png", "tankRed_barrel1_outline.png", "tankRed_barrel2.png", "tankRed_barrel2_outline.png", "tankRed_barrel3.png", "tankRed_barrel3_outline.png", "tankSand_barrel1.png", "tankSand_barrel1_outline.png", "tankSand_barrel2.png", "tankSand_barrel2_outline.png", "tankSand_barrel3.png", "tankSand_barrel3_outline.png"}
	name := barrels[rand.Intn(len(barrels))]
	log.Printf("new tank barrell: %s\n", name)
	return a.GetSprite(name)
}

func mustLoadImage(name string) *ebiten.Image {
	dat, err := assets.Open(name)
	if err != nil {
		panic(err)
	}

	img, _, err := image.Decode(dat)
	if err != nil {
		panic(err)
	}

	return ebiten.NewImageFromImage(img)
}

func mustLoadFont(name string) font.Face {
	f, err := assets.ReadFile(name)
	if err != nil {
		panic(err)
	}

	tt, err := opentype.Parse(f)
	if err != nil {
		panic(err)
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    48,
		DPI:     72,
		Hinting: font.HintingVertical,
	})
	if err != nil {
		panic(err)
	}

	return face
}

func mustLoadXMLSpriteMap(name string) SpriteMap {
	byteValue, _ := fs.ReadFile(assets, name)
	var s SpriteMap
	if err := xml.Unmarshal(byteValue, &s); err != nil {
		panic(err)
	}
	// fmt.Println(s)
	return s
}

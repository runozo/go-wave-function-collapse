package assets

import (
	"embed"
	"encoding/xml"
	"image"
	_ "image/png"
	"io/fs"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
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

// GetSprite retrieves the sprite image by name from the assets.
//
// Parameters:
// - name: the name of the sprite to retrieve.
// Returns:
// - *ebiten.Image: the sprite image corresponding to the given name.
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
	log.Println("Sprite not found", "name", name)
	return nil
}

// mustLoadImage loads an image from the given name and returns it as an *ebiten.Image.
//
// Parameters:
// - name: the name of the image file to load.
//
// Returns:
// - *ebiten.Image: the loaded image.
func mustLoadImage(name string) *ebiten.Image {
	dat, err := assets.Open(name)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(dat)
	if err != nil {
		log.Fatal(err)
	}

	return ebiten.NewImageFromImage(img)
}

// mustLoadXMLSpriteMap loads and parses an XML sprite map file from the assets with the given name.
//
// Parameters:
// - name: the name of the XML sprite map file to load.
// Returns:
// - SpriteMap: the parsed SpriteMap object.
func mustLoadXMLSpriteMap(name string) SpriteMap {
	byteValue, _ := fs.ReadFile(assets, name)
	var s SpriteMap
	if err := xml.Unmarshal(byteValue, &s); err != nil {
		log.Fatal(err)
	}
	// fmt.Println(s)
	return s
}

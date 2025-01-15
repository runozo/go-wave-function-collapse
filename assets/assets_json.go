package assets

import (
	"embed"
	"encoding/json"
	"image"
	_ "image/png"
	"io/fs"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// xmlSpriteMap = "allSprites_default.xml"
	jsonSpriteMap = "mapped_tiles.json"
	spriteSheet   = "allSprites_default.png"
	// xmlSpriteMap = "allSprites_default.xml"
	xmlSpriteMap = "mapped_tiles.xml"
)

//go:embed allSprites_default.png
//go:embed allSprites_default.xml
//go:embed mapped_tiles.json
var assets embed.FS

type AssetsJSON struct {
	SpriteSheet *ebiten.Image
	TileEntries map[string]tileEntry
}

type tileEntry struct {
	Name        string              `json:"name"`
	ImageName   string              `json:"image_name"`
	Options     map[string][]string `json:"options"`
	Connections map[string][]string `json:"connections"` // up right down left above below
	Type        string              `json:"type"`
	X           int                 `json:"x"`
	Y           int                 `json:"y"`
	Width       int                 `json:"width"`
	Height      int                 `json:"height"`
	Weight      int                 `json:"weight"`
}

func NewAssetsJSON() *AssetsJSON {
	var spriteSheet = mustLoadImage(spriteSheet)
	var spriteMap = mustLoadJSONSpriteMap(jsonSpriteMap)

	return &AssetsJSON{
		SpriteSheet: spriteSheet,
		TileEntries: spriteMap,
	}
}

// GetSprite retrieves the sprite image by name from the assets.
//
// Parameters:
// - name: the name of the sprite to retrieve.
// Returns:
// - *ebiten.Image: the sprite image corresponding to the given name.
func (a *AssetsJSON) GetSpriteJSON(name string) *ebiten.Image {
	log.Println("Getting sprite", name)
	subTexture, ok := a.TileEntries[name]
	if ok {
		return a.SpriteSheet.SubImage(
			image.Rect(
				subTexture.X,
				subTexture.Y,
				subTexture.X+subTexture.Width,
				subTexture.Y+subTexture.Height,
			)).(*ebiten.Image)
	}

	log.Println("Sprite not found", "name", name)
	return nil
}

func mustLoadJSONSpriteMap(name string) map[string]tileEntry {
	log.Println("Loading sprite map", "name", name)
	byteValue, err := fs.ReadFile(assets, name)
	if err != nil {
		log.Fatal(err)
	}

	// json is a slice of TileEntry
	var tileEntries []tileEntry
	if err := json.Unmarshal(byteValue, &tileEntries); err != nil {
		log.Fatal(err)
	}

	// transform tileEntries in a map
	tileEntriesMap := make(map[string]tileEntry)
	for i := 0; i < len(tileEntries); i++ {
		tileEntriesMap[tileEntries[i].ImageName] = tileEntries[i]
	}

	return tileEntriesMap
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
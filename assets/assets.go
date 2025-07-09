package assets

import (
	"encoding/json"
	"image"
	_ "image/png"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type Assets struct {
	SpriteSheet *ebiten.Image
	TileEntries map[string]TileEntry
}

type TileEntry struct {
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

func NewAssets(spriteSheetFileName, jsonSpriteMapFileName string) *Assets {
	var spriteSheet = mustLoadImage(spriteSheetFileName)
	var spriteMap = mustLoadJSONSpriteMap(jsonSpriteMapFileName)

	return &Assets{
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
func (a *Assets) GetSprite(name string) *ebiten.Image {
	// log.Println("Getting sprite", name)
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

func mustLoadJSONSpriteMap(name string) map[string]TileEntry {
	log.Println("Loading sprite map", "name", name)
	byteValue, err := os.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}

	// json is a slice of TileEntry
	var tileEntries []TileEntry
	if err := json.Unmarshal(byteValue, &tileEntries); err != nil {
		log.Fatal(err)
	}

	// transform tileEntries in a map
	tileEntriesMap := make(map[string]TileEntry)
	for i := 0; i < len(tileEntries); i++ {
		if tileEntries[i].ImageName[:len("tile")] == "tile" {
			tileEntriesMap[tileEntries[i].Name] = tileEntries[i]
		}
	}
	log.Println("Loaded sprite map", "len", len(tileEntriesMap))
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
	dat, err := os.Open(name)
	if err != nil {
		log.Fatal(err)
	}

	img, _, err := image.Decode(dat)
	if err != nil {
		log.Fatal(err)
	}

	return ebiten.NewImageFromImage(img)
}

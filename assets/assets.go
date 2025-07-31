package assets

import (
	"bytes"
	"encoding/json"
	"image"
	_ "image/png"
	"log"

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

// NewAssets takes a byte slice of a sprite sheet image and a byte slice of json data that describes the tiles in the sprite sheet.
// It returns a pointer to an Assets struct containing the sprite sheet image and a map of tile names to TileEntries.
func NewAssets(spriteSheetData, jsonData []byte) *Assets {
	var spriteSheet = parseSpriteSheet(spriteSheetData)
	var spriteMap = parseJSONData(jsonData)

	return &Assets{
		SpriteSheet: spriteSheet,
		TileEntries: spriteMap,
	}
}

// GetSprite returns the sub-image of the sprite sheet corresponding to the given
// tile name. If the tile name is not found, it logs an error and returns nil.
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

	log.Println("GetSprite: sprite not found", name)

	return nil
}

// parseJSONData takes a json byte slice and returns a map of tileEntries
// each tileEntry is identified by its name and contains all the informations
// needed to render the sprite (position, size, type, etc.)
// the function logs when it is done and returns the map
func parseJSONData(jsonData []byte) map[string]TileEntry {
	// json is a slice of TileEntry
	var tileEntries []TileEntry
	if err := json.Unmarshal(jsonData, &tileEntries); err != nil {
		log.Fatal(err)
	}

	// transform tileEntries in a map
	tileEntriesMap := make(map[string]TileEntry)
	for i := 0; i < len(tileEntries); i++ {
		tileEntriesMap[tileEntries[i].Name] = tileEntries[i]
	}

	log.Println("Loaded sprite map", "len", len(tileEntriesMap))

	return tileEntriesMap
}

// parseSpriteSheet takes a byte slice of a sprite sheet image and returns an
// *ebiten.Image of it. If the image can't be decoded, it logs a fatal error.
func parseSpriteSheet(spritesheetData []byte) *ebiten.Image {
	img, _, err := image.Decode(bytes.NewReader(spritesheetData))
	if err != nil {
		log.Fatal(err)
	}

	return ebiten.NewImageFromImage(img)
}

package main

import (
	"testing"

	"github.com/runozo/go-wave-function-collapse/assets"
)

func TestIntInSlice(t *testing.T) {
	cases := []struct {
		inSlice []int
		inInt   int
		want    bool
	}{
		{[]int{0, 1, 2, 3}, 0, true},
		{[]int{0, 1, 2, 3}, 1, true},
		{[]int{0, 1, 2, 3}, 2, true},
		{[]int{0, 1, 2, 3}, 3, true},
		{[]int{0, 1, 2, 3}, 4, false},
		{[]int{}, 4, false},
	}
	for _, c := range cases {
		// init tiles
		got := IntInSlice(c.inInt, c.inSlice)
		if got != c.want {
			t.Errorf("IntInSlice(%d, %v) == %v, want %v", c.inInt, c.inSlice, got, c.want)
		}
	}
}

func TestIterateWaveFunctionCollapse(t *testing.T) {
	cases := []struct {
		in   *Game
		want bool
	}{
		{&Game{
			assets:      assets.NewAssets(),
			width:       screenWidth,
			height:      screenHeight,
			isRendered:  false,
			numOfTilesX: screenWidth/tileWidth + 1,
			numOfTilesY: screenHeight/tileHeight + 1,
			tiles:       ResetTilesOptions(make([]Tile, (screenWidth/tileWidth+1)*(screenHeight/tileHeight+1))),
		}, true},
		{&Game{
			assets:      assets.NewAssets(),
			width:       screenWidth,
			height:      screenHeight,
			isRendered:  true,
			numOfTilesX: screenWidth/tileWidth + 1,
			numOfTilesY: screenHeight/tileHeight + 1,
			tiles:       make([]Tile, (screenWidth/tileWidth+1)*(screenHeight/tileHeight+1)),
		}, false},
	}
	for _, c := range cases {
		// init tiles
		got := IterateWaveFunctionCollapse(c.in)
		if got != c.want {
			t.Errorf("IterateWaveFunctionCollapse(%v) == %v, want %v", c.in, got, c.want)
		}
	}
}

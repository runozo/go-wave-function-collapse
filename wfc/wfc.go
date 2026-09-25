package wfc

import (
	"log"
	"math/bits"
	"math/rand"
	"runtime"

	"github.com/runozo/go-wave-function-collapse/assets"
)

// Tile is a cell of the wave function.
//
// The set of still-possible tiles is stored as a bitmask over tile IDs: bit i
// set means the tile with ID i is still possible for this cell. With the
// current tile set (40 ground tiles) a single uint64 holds the whole
// superposition, so filtering and union become a handful of machine-word
// operations instead of allocating and scanning string slices.
type Tile struct {
	Collapsed bool
	Name      string
	Mask      uint64
}

// IterateResult describes the outcome of a single Iterate call.
type IterateResult int

const (
	// Iterating means the generation is still in progress and more cells can be collapsed.
	Iterating IterateResult = iota
	// Rendered means every cell has been collapsed and the map is complete.
	Rendered
	// Contradiction means at least one uncollapsed cell has no options left.
	Contradiction
)

// DefaultMaxRestarts is the default number of times StartRender restarts a
// generation after hitting a contradiction before giving up.
const DefaultMaxRestarts = 50

// MaxTiles is the maximum number of tiles supported by the bitmask
// representation (one uint64 word).
const MaxTiles = 64

// Directions used by the compatibility tables.
const (
	dirUp = iota
	dirRight
	dirDown
	dirLeft
	numDirections
)

var dirNames = [numDirections]string{"up", "right", "down", "left"}

type Wfc struct {
	Tiles          []Tile
	TileEntries    map[string]assets.TileEntry
	IsRunning      bool
	IsRendered     bool
	TotalTiles     int
	ProcessedTiles int
	MaxRestarts    int

	numOfTilesX int
	numOfTilesY int
	attempts    int

	// Precomputed tile metadata, indexed by tile ID.
	names       []string
	nameToID    map[string]int
	weights     []int
	compat      [][numDirections]uint64 // compat[id][dir] = tiles allowed in direction dir from id
	initialMask uint64

	// Worklist used by propagate to run arc-consistency from a collapsed cell.
	queue   []int
	inQueue []bool

	// entropyBuffer is reused by LeastEntropyCellIndexes to avoid allocating a
	// new slice on every collapse.
	entropyBuffer []int
}

// NewWfc returns a new WFC with the given number of tiles in the X and Y directions, and the given tile entries.
//
// Only "ground" tiles with at least four direction options are considered, and
// the compatibility between their sides is precomputed into bitmasks once, at
// construction time.
//
// Parameters:
// - numOfTilesX: the number of tiles in the X direction.
// - numOfTilesY: the number of tiles in the Y direction.
// - tileEntries: a map of tile names to TileEntries.
//
// Returns:
// - *Wfc: a pointer to a new WFC.
func NewWfc(numOfTilesX, numOfTilesY int, tileEntries map[string]assets.TileEntry) *Wfc {
	wfc := &Wfc{
		Tiles:       make([]Tile, numOfTilesX*numOfTilesY),
		TileEntries: tileEntries,
		numOfTilesX: numOfTilesX,
		numOfTilesY: numOfTilesY,
		TotalTiles:  numOfTilesX * numOfTilesY,
		IsRunning:   false,
		IsRendered:  false,
		MaxRestarts: DefaultMaxRestarts,
	}
	wfc.buildIndex()
	wfc.Reset()
	return wfc
}

// buildIndex assigns an integer ID to every usable tile and precomputes the
// per-direction compatibility masks.
func (wfc *Wfc) buildIndex() {
	wfc.nameToID = make(map[string]int, len(wfc.TileEntries))
	wfc.names = nil
	wfc.weights = nil
	wfc.compat = nil
	wfc.initialMask = 0

	for name, entry := range wfc.TileEntries {
		if entry.Type != "ground" || len(entry.Options) < 4 {
			continue
		}
		if len(wfc.names) >= MaxTiles {
			log.Fatalf("wfc: too many tiles, the bitmask representation supports at most %d", MaxTiles)
		}
		id := len(wfc.names)
		wfc.nameToID[name] = id
		wfc.names = append(wfc.names, name)
		wfc.weights = append(wfc.weights, entry.Weight)
		wfc.compat = append(wfc.compat, [numDirections]uint64{})
		wfc.initialMask |= 1 << uint(id)
	}

	for name, entry := range wfc.TileEntries {
		id, ok := wfc.nameToID[name]
		if !ok {
			continue
		}
		for dir := 0; dir < numDirections; dir++ {
			var mask uint64
			for _, neighbor := range entry.Options[dirNames[dir]] {
				if nid, ok := wfc.nameToID[neighbor]; ok {
					mask |= 1 << uint(nid)
				}
			}
			wfc.compat[id][dir] = mask
		}
	}

	wfc.queue = wfc.queue[:0]
	wfc.inQueue = make([]bool, len(wfc.Tiles))
}

// Reset resets all the tiles in the WFC to their initial state, with all options available.
// It does not reset the precomputed tile metadata.
func (wfc *Wfc) Reset() {
	for i := range wfc.Tiles {
		wfc.Tiles[i] = Tile{Mask: wfc.initialMask}
	}
	wfc.ProcessedTiles = 0

	if len(wfc.inQueue) != len(wfc.Tiles) {
		wfc.inQueue = make([]bool, len(wfc.Tiles))
	} else {
		for i := range wfc.inQueue {
			wfc.inQueue[i] = false
		}
	}
	wfc.queue = wfc.queue[:0]
}

// LeastEntropyCellIndexes returns the indexes of the cells with the least entropy.
//
// Entropy is the number of still-possible tiles (population count of the
// bitmask). All uncollapsed cells sharing the minimum are returned.
//
// The returned slice is backed by an internal, reused buffer: callers must
// consume it before calling the method again.
//
// Returns:
// - []int: the indexes of the cells with the least entropy.
func (wfc *Wfc) LeastEntropyCellIndexes() []int {
	wfc.entropyBuffer = wfc.entropyBuffer[:0]
	minEntropy := MaxTiles + 1
	for i := range wfc.Tiles {
		if wfc.Tiles[i].Collapsed {
			continue
		}
		entropy := bits.OnesCount64(wfc.Tiles[i].Mask)
		switch {
		case entropy < minEntropy:
			minEntropy = entropy
			wfc.entropyBuffer = append(wfc.entropyBuffer[:0], i)
		case entropy == minEntropy:
			wfc.entropyBuffer = append(wfc.entropyBuffer, i)
		}
	}
	return wfc.entropyBuffer
}

// randomOptionID selects a tile ID for the cell at the given index using a
// weighted random choice over its remaining options. It returns -1 if the cell
// has no options left (a contradiction).
func (wfc *Wfc) randomOptionID(index int) int {
	mask := wfc.Tiles[index].Mask
	if mask == 0 {
		return -1
	}

	totalWeight := 0
	for m := mask; m != 0; m &= m - 1 {
		totalWeight += wfc.weights[bits.TrailingZeros64(m)]
	}

	if totalWeight > 0 {
		randomWeight := rand.Intn(totalWeight)
		accumulated := 0
		for m := mask; m != 0; m &= m - 1 {
			id := bits.TrailingZeros64(m)
			accumulated += wfc.weights[id]
			if randomWeight < accumulated {
				return id
			}
		}
	}

	// All weights are zero (or negative): fall back to a uniform choice.
	remaining := rand.Intn(bits.OnesCount64(mask))
	for m := mask; m != 0; m &= m - 1 {
		if remaining == 0 {
			return bits.TrailingZeros64(m)
		}
		remaining--
	}
	return -1
}

// RandomOptionWithWeight selects a random option for the tile at the given
// index, taking into account the weights of each option.
//
// Parameters:
// - index: the index of the tile for which to select an option.
//
// Returns:
//   - string: the randomly selected option based on weights, or "" if the cell
//     has no options left.
func (wfc *Wfc) RandomOptionWithWeight(index int) string {
	id := wfc.randomOptionID(index)
	if id < 0 {
		return ""
	}
	return wfc.names[id]
}

// CollapseCell collapses a cell to a single, weighted-random option.
//
// Parameters:
// - index: the index of the cell to collapse.
func (wfc *Wfc) CollapseCell(index int) {
	id := wfc.randomOptionID(index)
	if id < 0 {
		// Nothing to collapse: leave the cell as-is so the contradiction stays
		// detectable instead of turning it into a collapsed, empty-named tile.
		return
	}
	wfc.Tiles[index] = Tile{
		Collapsed: true,
		Name:      wfc.names[id],
		Mask:      1 << uint(id),
	}
}

// enqueue adds a cell to the propagation worklist if it is not already queued
// and not collapsed.
func (wfc *Wfc) enqueue(index int) {
	if wfc.Tiles[index].Collapsed || wfc.inQueue[index] {
		return
	}
	wfc.inQueue[index] = true
	wfc.queue = append(wfc.queue, index)
}

// enqueueNeighbors queues the (up to four) orthogonal neighbors of a cell.
func (wfc *Wfc) enqueueNeighbors(index int) {
	x := index % wfc.numOfTilesX
	y := index / wfc.numOfTilesX
	if y > 0 {
		wfc.enqueue(index - wfc.numOfTilesX)
	}
	if x < wfc.numOfTilesX-1 {
		wfc.enqueue(index + 1)
	}
	if y < wfc.numOfTilesY-1 {
		wfc.enqueue(index + wfc.numOfTilesX)
	}
	if x > 0 {
		wfc.enqueue(index - 1)
	}
}

// neighborAllowed returns the tiles allowed for a cell adjacent to
// neighborIndex, where dir is the direction from the neighbor to that cell.
func (wfc *Wfc) neighborAllowed(neighborIndex, dir int) uint64 {
	var allowed uint64
	for m := wfc.Tiles[neighborIndex].Mask; m != 0; m &= m - 1 {
		allowed |= wfc.compat[bits.TrailingZeros64(m)][dir]
	}
	return allowed
}

// allowedMask intersects the options of a cell with the constraints imposed by
// its four neighbors.
func (wfc *Wfc) allowedMask(index int) uint64 {
	x := index % wfc.numOfTilesX
	y := index / wfc.numOfTilesX
	mask := wfc.Tiles[index].Mask
	if y > 0 {
		mask &= wfc.neighborAllowed(index-wfc.numOfTilesX, dirDown)
	}
	if x < wfc.numOfTilesX-1 {
		mask &= wfc.neighborAllowed(index+1, dirLeft)
	}
	if y < wfc.numOfTilesY-1 {
		mask &= wfc.neighborAllowed(index+wfc.numOfTilesX, dirUp)
	}
	if x > 0 {
		mask &= wfc.neighborAllowed(index-1, dirRight)
	}
	return mask
}

// clearQueue empties the worklist, clearing its bookkeeping flags.
func (wfc *Wfc) clearQueue() {
	for _, index := range wfc.queue {
		wfc.inQueue[index] = false
	}
	wfc.queue = wfc.queue[:0]
}

// propagate runs arc-consistency starting from the neighbors of seed: every
// time a cell loses options, its neighbors are re-checked. It returns false as
// soon as a cell is left with no options (a contradiction).
//
// This replaces the previous full-grid sweep: instead of re-elaborating every
// cell after each collapse, only the cells actually affected are revisited.
func (wfc *Wfc) propagate(seed int) bool {
	wfc.queue = wfc.queue[:0]
	wfc.enqueueNeighbors(seed)

	for head := 0; head < len(wfc.queue); head++ {
		index := wfc.queue[head]
		wfc.inQueue[index] = false

		newMask := wfc.allowedMask(index)
		if newMask == wfc.Tiles[index].Mask {
			continue
		}
		wfc.Tiles[index].Mask = newMask
		if newMask == 0 {
			wfc.clearQueue()
			return false
		}
		wfc.enqueueNeighbors(index)
	}

	wfc.queue = wfc.queue[:0]
	return true
}

// Iterate collapses one least-entropy cell and propagates the constraints from
// it until arc-consistency, detecting contradictions as soon as they appear.
//
// Returns:
// - IterateResult: Rendered when complete, Contradiction when stuck, Iterating otherwise.
func (wfc *Wfc) Iterate() IterateResult {
	leastEntropyIndexes := wfc.LeastEntropyCellIndexes()

	if len(leastEntropyIndexes) == 0 {
		wfc.IsRunning = false
		wfc.IsRendered = true
		return Rendered
	}

	if wfc.Tiles[leastEntropyIndexes[0]].Mask == 0 {
		// At least one cell has no options left: the constraint propagation
		// reached an impossible state.
		return Contradiction
	}

	collapseIndex := leastEntropyIndexes[rand.Intn(len(leastEntropyIndexes))]
	wfc.CollapseCell(collapseIndex)
	wfc.ProcessedTiles++

	if !wfc.propagate(collapseIndex) {
		return Contradiction
	}
	return Iterating
}

// BeginRender prepares a new generation without running it.
//
// It sets the running state, clears the rendered flag, resets the restart
// counter and resets the grid to its initial superposition. Call Step to
// advance the generation, or StartRender to run it to completion.
func (wfc *Wfc) BeginRender() {
	wfc.IsRunning = true
	wfc.IsRendered = false
	wfc.attempts = 0
	wfc.Reset()
}

// Step advances the generation by a single Iterate call, transparently
// restarting from scratch when a contradiction is hit.
//
// It returns the outcome of the last Iterate: Iterating while the generation
// is still in progress, Rendered when the map is complete or when MaxRestarts
// has been exhausted. This is the incremental building block used by game loops
// (e.g. Ebitengine's Update) to animate the collapse frame by frame, which is
// required on single-threaded targets such as WebAssembly.
func (wfc *Wfc) Step() IterateResult {
	switch wfc.Iterate() {
	case Rendered:
		return Rendered
	case Contradiction:
		if wfc.attempts >= wfc.MaxRestarts {
			log.Println("wfc: max restarts reached, giving up")
			wfc.IsRendered = true
			wfc.IsRunning = false
			return Rendered
		}
		wfc.attempts++
		log.Printf("wfc: contradiction detected, restarting (%d/%d)", wfc.attempts, wfc.MaxRestarts)
		wfc.Reset()
		return Iterating
	default:
		return Iterating
	}
}

// StartRender initializes and starts the rendering process using the Wave Function Collapse algorithm.
//
// This method first checks if the rendering is already running. If so, it logs a message and returns.
// If not running, it begins a new generation and iteratively collapses cells with the least entropy
// until no more collapsable cells are available or the rendering process is stopped.
//
// If the propagation reaches a contradiction (a cell with no options left), the generation is
// restarted from scratch, up to MaxRestarts times. When the limit is reached the map is left as-is
// and the method gives up without panicking.
func (wfc *Wfc) StartRender() {
	if wfc.IsRunning {
		log.Println("wfc is already running")
		return
	}
	wfc.BeginRender()
	for wfc.Step() == Iterating {
		if !wfc.IsRunning {
			return
		}
		runtime.Gosched()
	}
}

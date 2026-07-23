// Package frame adds multi-frame support on top of internal/layer: each
// frame owns its own independent layer.Store (so a frame can itself have
// multiple layers), and frames are addressed by position (0-based index)
// rather than name. Frame 0 always wraps the canvas/history pair the
// handler was constructed with, so behavior is unchanged from before
// frames existed until a second frame is added and selected.
package frame

import (
	"fmt"
	"image/color"
	"math"
	"sync"

	"pxcli/internal/canvas"
	"pxcli/internal/history"
	"pxcli/internal/layer"
)

type Error struct {
	Code    string
	Message string
}

func (e Error) Error() string {
	if e.Message == "" {
		return e.Code
	}
	return e.Code + ": " + e.Message
}

// Store manages an ordered list of frames and tracks which one is active.
type Store struct {
	mu     sync.Mutex
	width  int
	height int
	frames []*layer.Store
	active int
}

// New creates a store whose frame 0 wraps the provided canvas/history pair.
func New(base *canvas.Canvas, baseHistory *history.Manager) *Store {
	return &Store{
		width:  base.Width(),
		height: base.Height(),
		frames: []*layer.Store{layer.New(base, baseHistory)},
		active: 0,
	}
}

// Add creates a new blank frame (a fresh layer.Store with its own "base"
// layer) and returns its index. The new frame does not become active.
func (s *Store) Add() (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, err := canvas.New(s.width, s.height)
	if err != nil {
		return 0, err
	}
	s.frames = append(s.frames, layer.New(c, history.New(c)))
	return len(s.frames) - 1, nil
}

// Count returns the number of frames.
func (s *Store) Count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.frames)
}

// Select sets the active frame by index.
func (s *Store) Select(index int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= len(s.frames) {
		return outOfRangeErr(index, len(s.frames))
	}
	s.active = index
	return nil
}

// ActiveIndex returns the index of the currently active frame.
func (s *Store) ActiveIndex() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// Active returns the currently active frame's layer store.
func (s *Store) Active() *layer.Store {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.frames[s.active]
}

// at returns the layer store for a frame index, bounds-checked. Callers
// must not hold s.mu.
func (s *Store) at(index int) (*layer.Store, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if index < 0 || index >= len(s.frames) {
		return nil, outOfRangeErr(index, len(s.frames))
	}
	return s.frames[index], nil
}

// Ghost renders an onion-skin composite for the active frame: the target
// frame's flattened content dimmed to opacity (0-1) sits underneath, with
// the active frame's own flattened content drawn on top at full strength.
// It is a pure query — nothing is mutated.
func (s *Store) Ghost(targetIndex int, opacity float64) (*canvas.Canvas, error) {
	target, err := s.at(targetIndex)
	if err != nil {
		return nil, err
	}
	active := s.Active()

	width, height := s.width, s.height
	activeFlat := active.Flatten()
	targetFlat := target.Flatten()

	out, err := canvas.New(width, height)
	if err != nil {
		return nil, err
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			ghostPixel, _ := targetFlat.GetPixel(x, y)
			dimmed := dim(ghostPixel, opacity)
			topPixel, _ := activeFlat.GetPixel(x, y)
			_ = out.SetPixel(x, y, alphaOver(topPixel, dimmed))
		}
	}
	return out, nil
}

// Sheet tiles every frame's flattened content into a single canvas, cols
// frames per row (left-to-right, top-to-bottom), padding any incomplete
// final row with transparent pixels. cols is clamped to [1, frame count].
func (s *Store) Sheet(cols int) (*canvas.Canvas, error) {
	s.mu.Lock()
	width, height := s.width, s.height
	frames := make([]*layer.Store, len(s.frames))
	copy(frames, s.frames)
	s.mu.Unlock()

	if cols <= 0 || cols > len(frames) {
		cols = len(frames)
	}
	rows := (len(frames) + cols - 1) / cols

	sheet, err := canvas.New(width*cols, height*rows)
	if err != nil {
		return nil, err
	}
	for i, f := range frames {
		flat := f.Flatten()
		pixels, err := flat.CopyRegion(0, 0, width, height)
		if err != nil {
			return nil, err
		}
		col, row := i%cols, i/cols
		if err := sheet.PasteRegion(col*width, row*height, width, height, pixels); err != nil {
			return nil, err
		}
	}
	return sheet, nil
}

func dim(c color.RGBA, opacity float64) color.RGBA {
	if opacity < 0 {
		opacity = 0
	}
	if opacity > 1 {
		opacity = 1
	}
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: uint8(math.Round(float64(c.A) * opacity))}
}

// alphaOver composites top over bottom using standard alpha "over".
func alphaOver(top, bottom color.RGBA) color.RGBA {
	topA := float64(top.A) / 255
	bottomA := float64(bottom.A) / 255
	outA := topA + bottomA*(1-topA)
	if outA == 0 {
		return color.RGBA{}
	}
	blend := func(topC, bottomC uint8) uint8 {
		composed := float64(topC)*topA + float64(bottomC)*bottomA*(1-topA)
		return uint8(math.Round(composed / outA))
	}
	return color.RGBA{
		R: blend(top.R, bottom.R),
		G: blend(top.G, bottom.G),
		B: blend(top.B, bottom.B),
		A: uint8(math.Round(outA * 255)),
	}
}

func outOfRangeErr(index, count int) error {
	return Error{Code: "invalid_frame", Message: fmt.Sprintf("frame index %d out of range (0-%d)", index, count-1)}
}

// Package layer manages named, same-size canvas buffers that composite
// together (visible ones, in creation order, "base" first) when the canvas
// is exported, so an agent can draw a background and a sprite on separate
// buffers without one destroying the other.
package layer

import (
	"fmt"
	"image/color"
	"math"
	"sync"

	"pxcli/internal/canvas"
	"pxcli/internal/history"
)

// Base is the always-present, reserved name for the original canvas passed
// to New. It cannot be re-added and is always first in creation order.
const Base = "base"

// Error represents a layer management error with a code and message.
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

// Entry is a single layer: its own canvas with its own independent
// undo/redo history, plus whether it is included when layers are flattened.
type Entry struct {
	Canvas  *canvas.Canvas
	History *history.Manager
	Visible bool
}

// Store holds the base layer plus any additional named layers, and tracks
// which one is active for drawing commands.
type Store struct {
	mu     sync.Mutex
	width  int
	height int
	layers map[string]*Entry
	order  []string
	active string
}

// New creates a layer store whose "base" layer wraps the provided canvas
// and history (typically the daemon's original single canvas, so existing
// behavior is unchanged until a layer is added and selected).
func New(base *canvas.Canvas, baseHistory *history.Manager) *Store {
	return &Store{
		width:  base.Width(),
		height: base.Height(),
		layers: map[string]*Entry{
			Base: {Canvas: base, History: baseHistory, Visible: true},
		},
		order:  []string{Base},
		active: Base,
	}
}

// Add creates a new blank, visible layer. Layer canvases match the base
// canvas's dimensions.
func (s *Store) Add(name string) error {
	if name == "" {
		return Error{Code: "invalid_args", Message: "layer name is required"}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.layers[name]; exists {
		return Error{Code: "invalid_args", Message: fmt.Sprintf("layer %q already exists", name)}
	}
	c, err := canvas.New(s.width, s.height)
	if err != nil {
		return err
	}
	s.layers[name] = &Entry{Canvas: c, History: history.New(c), Visible: true}
	s.order = append(s.order, name)
	return nil
}

// List returns layer names in creation order ("base" first).
func (s *Store) List() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, len(s.order))
	copy(out, s.order)
	return out
}

// Select sets the active layer for drawing commands.
func (s *Store) Select(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.layers[name]; !ok {
		return notFoundErr(name)
	}
	s.active = name
	return nil
}

// ActiveName returns the currently selected layer's name.
func (s *Store) ActiveName() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// Active returns the currently selected layer's entry.
func (s *Store) Active() *Entry {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.layers[s.active]
}

// SetVisible marks whether a layer is included when layers are flattened.
func (s *Store) SetVisible(name string, visible bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry, ok := s.layers[name]
	if !ok {
		return notFoundErr(name)
	}
	entry.Visible = visible
	return nil
}

// Flatten composites every visible layer, in creation order, onto a new
// canvas using standard "source over destination" alpha blending, and
// returns the result. The store's own layers are left untouched.
func (s *Store) Flatten() *canvas.Canvas {
	s.mu.Lock()
	width, height := s.width, s.height
	entries := make([]*Entry, 0, len(s.order))
	for _, name := range s.order {
		entries = append(entries, s.layers[name])
	}
	s.mu.Unlock()

	out, _ := canvas.New(width, height)
	for _, entry := range entries {
		if !entry.Visible {
			continue
		}
		pixels, _ := entry.Canvas.CopyRegion(0, 0, width, height)
		compositeOver(out, pixels, width, height)
	}
	return out
}

func compositeOver(dst *canvas.Canvas, src []color.RGBA, width, height int) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			top := src[y*width+x]
			if top.A == 0 {
				continue
			}
			bottom, _ := dst.GetPixel(x, y)
			_ = dst.SetPixel(x, y, alphaOver(top, bottom))
		}
	}
}

// alphaOver blends top over bottom using the standard Porter-Duff "over"
// operator on straight (non-premultiplied) alpha.
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

func notFoundErr(name string) error {
	return Error{Code: "invalid_layer", Message: fmt.Sprintf("layer %q does not exist", name)}
}

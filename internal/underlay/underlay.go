// Package underlay adds support for importing a user-supplied local image
// (PNG or JPEG) as a low-opacity, non-drawable reference underneath the
// canvas: something to trace proportions/silhouette against. It is
// intentionally not a layer.Store layer — set_pixel/fill_rect/line/etc.
// never touch it, since drawing commands only ever address the active
// frame's active layer, and the underlay is a separate field entirely.
package underlay

import (
	"image"
	"image/color"
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"math"
	"os"
	"sync"

	"pxcli/internal/canvas"
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

// Store holds at most one imported reference image, resized to the canvas
// dimensions and dimmed to the requested opacity at import time.
type Store struct {
	mu      sync.Mutex
	width   int
	height  int
	image   *canvas.Canvas
	version int
}

// New creates an empty underlay store sized to match the canvas.
func New(width, height int) *Store {
	return &Store{width: width, height: height}
}

// Import decodes a local PNG/JPEG file, resizes it (nearest-neighbor) to
// the canvas dimensions, dims it to opacity (0-1), and stores it as the
// underlay, replacing any previous import. Only local files are read; no
// network fetching is supported.
func (s *Store) Import(path string, opacity float64) error {
	if opacity < 0 || opacity > 1 {
		return Error{Code: "invalid_args", Message: "opacity must be between 0 and 1"}
	}
	file, err := os.Open(path)
	if err != nil {
		return Error{Code: "io", Message: err.Error()}
	}
	defer file.Close()

	src, _, err := image.Decode(file)
	if err != nil {
		return Error{Code: "io", Message: "unable to decode image: " + err.Error()}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	resized := resizeNearest(src, s.width, s.height)
	dimmed, err := canvas.New(s.width, s.height)
	if err != nil {
		return err
	}
	for i, px := range resized {
		x, y := i%s.width, i/s.width
		px.A = uint8(math.Round(float64(px.A) * opacity))
		_ = dimmed.SetPixel(x, y, px)
	}
	s.image = dimmed
	s.version++
	return nil
}

// HasImage reports whether an underlay has been imported.
func (s *Store) HasImage() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.image != nil
}

// Version increases every time Import succeeds; used by render callers to
// detect an underlay change even when the canvas itself hasn't changed.
func (s *Store) Version() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.version
}

// CompositeUnder returns a new canvas with the underlay (if any) drawn
// beneath top's own content (top wins wherever it is opaque). Returns a
// copy of top unchanged if no underlay has been imported.
func (s *Store) CompositeUnder(top *canvas.Canvas) (*canvas.Canvas, error) {
	s.mu.Lock()
	underlayImg := s.image
	s.mu.Unlock()

	width, height := top.Width(), top.Height()
	out, err := canvas.New(width, height)
	if err != nil {
		return nil, err
	}
	topPixels, err := top.CopyRegion(0, 0, width, height)
	if err != nil {
		return nil, err
	}
	if underlayImg == nil {
		if err := out.PasteRegion(0, 0, width, height, topPixels); err != nil {
			return nil, err
		}
		return out, nil
	}
	underPixels, err := underlayImg.CopyRegion(0, 0, width, height)
	if err != nil {
		return nil, err
	}
	for i := range topPixels {
		x, y := i%width, i/width
		_ = out.SetPixel(x, y, alphaOver(topPixels[i], underPixels[i]))
	}
	return out, nil
}

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

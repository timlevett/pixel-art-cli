// Package clipboard holds named, in-memory pixel regions captured with
// "copy" so they can be stamped elsewhere with "paste".
package clipboard

import (
	"fmt"
	"image/color"
	"sync"
)

// DefaultName is the clipboard slot used when copy/paste omit an explicit name.
const DefaultName = "default"

// Error represents a clipboard error with a code and message.
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

// Region is a captured rectangle of pixels in row-major order.
type Region struct {
	Width  int
	Height int
	Pixels []color.RGBA
}

// Store holds named clipboard regions.
type Store struct {
	mu      sync.RWMutex
	regions map[string]Region
}

// New creates an empty clipboard store.
func New() *Store {
	return &Store{regions: make(map[string]Region)}
}

// Set defines or replaces a named clipboard region.
func (s *Store) Set(name string, region Region) error {
	if name == "" {
		return Error{Code: "invalid_args", Message: "clipboard name is required"}
	}
	if region.Width <= 0 || region.Height <= 0 {
		return Error{Code: "invalid_args", Message: "clipboard region must have positive dimensions"}
	}
	if len(region.Pixels) != region.Width*region.Height {
		return Error{Code: "invalid_args", Message: "clipboard region pixel count does not match dimensions"}
	}
	pixels := make([]color.RGBA, len(region.Pixels))
	copy(pixels, region.Pixels)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.regions[name] = Region{Width: region.Width, Height: region.Height, Pixels: pixels}
	return nil
}

// Get returns a copy of a named clipboard region.
func (s *Store) Get(name string) (Region, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	region, ok := s.regions[name]
	if !ok {
		return Region{}, Error{Code: "invalid_clipboard", Message: fmt.Sprintf("clipboard %q is empty; copy into it first", name)}
	}
	pixels := make([]color.RGBA, len(region.Pixels))
	copy(pixels, region.Pixels)
	return Region{Width: region.Width, Height: region.Height, Pixels: pixels}, nil
}

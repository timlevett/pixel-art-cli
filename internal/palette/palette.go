// Package palette manages named, ordered color sets that drawing commands
// can reference instead of repeating raw hex values.
package palette

import (
	"fmt"
	"image/color"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// activeAlias is the reserved palette name that resolves to whichever
// palette was most recently selected with Use.
const activeAlias = "p"

// Error represents a palette error with a code and message.
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

// Store holds named palettes and tracks which one is active for the "p:"
// shorthand reference.
type Store struct {
	mu       sync.RWMutex
	palettes map[string][]color.RGBA
	active   string
}

// New creates an empty palette store.
func New() *Store {
	return &Store{palettes: make(map[string][]color.RGBA)}
}

// Add defines or replaces a named palette with the provided ordered colors.
func (s *Store) Add(name string, colors []color.RGBA) error {
	if name == "" {
		return Error{Code: "invalid_args", Message: "palette name is required"}
	}
	if name == activeAlias {
		return Error{Code: "invalid_args", Message: fmt.Sprintf("palette name %q is reserved for the active-palette shorthand", activeAlias)}
	}
	if len(colors) == 0 {
		return Error{Code: "invalid_args", Message: "palette requires at least one color"}
	}
	stored := make([]color.RGBA, len(colors))
	copy(stored, colors)

	s.mu.Lock()
	defer s.mu.Unlock()
	s.palettes[name] = stored
	return nil
}

// List returns all defined palette names, sorted alphabetically.
func (s *Store) List() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	names := make([]string, 0, len(s.palettes))
	for name := range s.palettes {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Colors returns a copy of the ordered colors for a named palette.
func (s *Store) Colors(name string) ([]color.RGBA, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	colors, ok := s.palettes[name]
	if !ok {
		return nil, notDefinedErr(name)
	}
	out := make([]color.RGBA, len(colors))
	copy(out, colors)
	return out, nil
}

// Use selects the named palette as the target for "p:<index>" references.
func (s *Store) Use(name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.palettes[name]; !ok {
		return notDefinedErr(name)
	}
	s.active = name
	return nil
}

// Slot returns the color at the given index for a named palette.
func (s *Store) Slot(name string, index int) (color.RGBA, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	colors, ok := s.palettes[name]
	if !ok {
		return color.RGBA{}, notDefinedErr(name)
	}
	if index < 0 || index >= len(colors) {
		return color.RGBA{}, Error{Code: "invalid_palette", Message: fmt.Sprintf("palette %q has no slot %d (size %d)", name, index, len(colors))}
	}
	return colors[index], nil
}

// ResolveRef parses a "<name>:<index>" reference (the reserved name "p"
// means the active palette selected via Use) and looks up the color.
// matched is false when ref does not look like a palette reference at all,
// in which case the caller should try another color format (e.g. hex).
func (s *Store) ResolveRef(ref string) (value color.RGBA, matched bool, err error) {
	name, index, ok := splitRef(ref)
	if !ok {
		return color.RGBA{}, false, nil
	}
	if name == activeAlias {
		s.mu.RLock()
		active := s.active
		s.mu.RUnlock()
		if active == "" {
			return color.RGBA{}, true, Error{Code: "invalid_palette", Message: "no active palette; call palette use <name> first or reference <name>:<index>"}
		}
		name = active
	}
	value, err = s.Slot(name, index)
	return value, true, err
}

func notDefinedErr(name string) error {
	return Error{Code: "invalid_palette", Message: fmt.Sprintf("palette %q is not defined", name)}
}

func splitRef(ref string) (string, int, bool) {
	idx := strings.LastIndex(ref, ":")
	if idx <= 0 || idx == len(ref)-1 {
		return "", 0, false
	}
	name := ref[:idx]
	index, err := strconv.Atoi(ref[idx+1:])
	if err != nil || index < 0 {
		return "", 0, false
	}
	return name, index, true
}

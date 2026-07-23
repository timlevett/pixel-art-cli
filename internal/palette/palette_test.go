package palette

import (
	"errors"
	"image/color"
	"testing"
)

func TestAddAndColors(t *testing.T) {
	s := New()
	colors := []color.RGBA{{R: 255, A: 255}, {G: 255, A: 255}}
	if err := s.Add("fire", colors); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	got, err := s.Colors("fire")
	if err != nil {
		t.Fatalf("Colors() error = %v", err)
	}
	if len(got) != 2 || got[0] != colors[0] || got[1] != colors[1] {
		t.Fatalf("Colors() = %v, want %v", got, colors)
	}
}

func TestAddReplacesExisting(t *testing.T) {
	s := New()
	_ = s.Add("fire", []color.RGBA{{R: 255, A: 255}})
	_ = s.Add("fire", []color.RGBA{{G: 255, A: 255}, {B: 255, A: 255}})
	got, err := s.Colors("fire")
	if err != nil {
		t.Fatalf("Colors() error = %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("Colors() len = %d, want 2 (replace, not append)", len(got))
	}
}

func TestAddErrors(t *testing.T) {
	s := New()
	if err := s.Add("", []color.RGBA{{A: 255}}); !isCode(err, "invalid_args") {
		t.Fatalf("Add(empty name) error = %v, want invalid_args", err)
	}
	if err := s.Add("fire", nil); !isCode(err, "invalid_args") {
		t.Fatalf("Add(no colors) error = %v, want invalid_args", err)
	}
	if err := s.Add("p", []color.RGBA{{A: 255}}); !isCode(err, "invalid_args") {
		t.Fatalf("Add(reserved name) error = %v, want invalid_args", err)
	}
}

func TestColorsUndefined(t *testing.T) {
	s := New()
	if _, err := s.Colors("missing"); !isCode(err, "invalid_palette") {
		t.Fatalf("Colors(missing) error = %v, want invalid_palette", err)
	}
}

func TestList(t *testing.T) {
	s := New()
	_ = s.Add("b", []color.RGBA{{A: 255}})
	_ = s.Add("a", []color.RGBA{{A: 255}})
	got := s.List()
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("List() = %v, want [a b] (sorted)", got)
	}
}

func TestUseAndResolveActiveShorthand(t *testing.T) {
	s := New()
	colors := []color.RGBA{{R: 1, A: 255}, {R: 2, A: 255}, {R: 3, A: 255}}
	_ = s.Add("sprite", colors)
	if err := s.Use("sprite"); err != nil {
		t.Fatalf("Use() error = %v", err)
	}
	value, matched, err := s.ResolveRef("p:2")
	if err != nil || !matched {
		t.Fatalf("ResolveRef(p:2) = %v, %v, %v", value, matched, err)
	}
	if value != colors[2] {
		t.Fatalf("ResolveRef(p:2) = %v, want %v", value, colors[2])
	}
}

func TestUseUndefinedPalette(t *testing.T) {
	s := New()
	if err := s.Use("missing"); !isCode(err, "invalid_palette") {
		t.Fatalf("Use(missing) error = %v, want invalid_palette", err)
	}
}

func TestResolveRefNoActivePalette(t *testing.T) {
	s := New()
	_, matched, err := s.ResolveRef("p:0")
	if !matched {
		t.Fatalf("ResolveRef(p:0) matched = false, want true")
	}
	if !isCode(err, "invalid_palette") {
		t.Fatalf("ResolveRef(p:0) error = %v, want invalid_palette", err)
	}
}

func TestResolveRefNamedPalette(t *testing.T) {
	s := New()
	colors := []color.RGBA{{R: 9, A: 255}}
	_ = s.Add("sprite", colors)
	value, matched, err := s.ResolveRef("sprite:0")
	if err != nil || !matched || value != colors[0] {
		t.Fatalf("ResolveRef(sprite:0) = %v, %v, %v", value, matched, err)
	}
}

func TestResolveRefOutOfRange(t *testing.T) {
	s := New()
	_ = s.Add("sprite", []color.RGBA{{A: 255}})
	_, matched, err := s.ResolveRef("sprite:5")
	if !matched || !isCode(err, "invalid_palette") {
		t.Fatalf("ResolveRef(sprite:5) matched=%v err=%v, want matched=true invalid_palette", matched, err)
	}
}

func TestResolveRefNotAPaletteReference(t *testing.T) {
	s := New()
	cases := []string{"#ff0000", "red", "", "sprite:", "sprite:-1", "sprite:abc", ":3"}
	for _, ref := range cases {
		_, matched, err := s.ResolveRef(ref)
		if matched {
			t.Fatalf("ResolveRef(%q) matched = true, want false (err=%v)", ref, err)
		}
	}
}

func isCode(err error, code string) bool {
	var perr Error
	if !errors.As(err, &perr) {
		return false
	}
	return perr.Code == code
}

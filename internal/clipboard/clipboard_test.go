package clipboard

import (
	"errors"
	"image/color"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	s := New()
	pixels := []color.RGBA{{R: 1, A: 255}, {R: 2, A: 255}, {R: 3, A: 255}, {R: 4, A: 255}}
	region := Region{Width: 2, Height: 2, Pixels: pixels}
	if err := s.Set("sprite", region); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	got, err := s.Get("sprite")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Width != 2 || got.Height != 2 || len(got.Pixels) != 4 {
		t.Fatalf("Get() = %+v, want matching dims to %+v", got, region)
	}
	for i, p := range pixels {
		if got.Pixels[i] != p {
			t.Fatalf("Get().Pixels[%d] = %v, want %v", i, got.Pixels[i], p)
		}
	}
}

func TestSetReplacesExisting(t *testing.T) {
	s := New()
	_ = s.Set("sprite", Region{Width: 1, Height: 1, Pixels: []color.RGBA{{R: 1, A: 255}}})
	_ = s.Set("sprite", Region{Width: 2, Height: 1, Pixels: []color.RGBA{{R: 2, A: 255}, {R: 3, A: 255}}})
	got, err := s.Get("sprite")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got.Width != 2 || got.Height != 1 {
		t.Fatalf("Get() dims = %dx%d, want 2x1 (replace, not merge)", got.Width, got.Height)
	}
}

func TestSetErrors(t *testing.T) {
	s := New()
	if err := s.Set("", Region{Width: 1, Height: 1, Pixels: []color.RGBA{{}}}); !isCode(err, "invalid_args") {
		t.Fatalf("Set(empty name) error = %v, want invalid_args", err)
	}
	if err := s.Set("a", Region{Width: 0, Height: 1, Pixels: nil}); !isCode(err, "invalid_args") {
		t.Fatalf("Set(zero width) error = %v, want invalid_args", err)
	}
	if err := s.Set("a", Region{Width: 2, Height: 2, Pixels: []color.RGBA{{}}}); !isCode(err, "invalid_args") {
		t.Fatalf("Set(mismatched pixel count) error = %v, want invalid_args", err)
	}
}

func TestGetUndefined(t *testing.T) {
	s := New()
	if _, err := s.Get("missing"); !isCode(err, "invalid_clipboard") {
		t.Fatalf("Get(missing) error = %v, want invalid_clipboard", err)
	}
}

func TestGetReturnsIndependentCopy(t *testing.T) {
	s := New()
	pixels := []color.RGBA{{R: 1, A: 255}}
	_ = s.Set("sprite", Region{Width: 1, Height: 1, Pixels: pixels})

	got, err := s.Get("sprite")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	got.Pixels[0] = color.RGBA{R: 99, A: 255}

	got2, err := s.Get("sprite")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if got2.Pixels[0] != (color.RGBA{R: 1, A: 255}) {
		t.Fatalf("mutating a Get() result affected stored region: %v", got2.Pixels[0])
	}
}

func isCode(err error, code string) bool {
	var cerr Error
	if !errors.As(err, &cerr) {
		return false
	}
	return cerr.Code == code
}

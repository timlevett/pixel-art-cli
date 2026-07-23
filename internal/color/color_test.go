package color

import (
	"errors"
	"image/color"
	"testing"
)

func TestParseShortHex(t *testing.T) {
	got, err := Parse("#f00")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := color.RGBA{R: 255, G: 0, B: 0, A: 255}
	if got != want {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestParseInvalidColor(t *testing.T) {
	_, err := Parse("#12")
	if err == nil {
		t.Fatalf("expected error")
	}
	var cerr Error
	if !errors.As(err, &cerr) {
		t.Fatalf("expected color Error, got %T", err)
	}
	if cerr.Code != "invalid_color" {
		t.Fatalf("expected code invalid_color, got %q", cerr.Code)
	}

	_, err = Parse("not-a-color")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.As(err, &cerr) {
		t.Fatalf("expected color Error, got %T", err)
	}
	if cerr.Code != "invalid_color" {
		t.Fatalf("expected code invalid_color, got %q", cerr.Code)
	}
}

func TestFormatTransparent(t *testing.T) {
	got := Format(color.RGBA{R: 0, G: 0, B: 0, A: 0})
	if got != "#00000000" {
		t.Fatalf("expected #00000000, got %q", got)
	}
}

func TestBlendEndpoints(t *testing.T) {
	a := color.RGBA{R: 0, G: 10, B: 20, A: 255}
	b := color.RGBA{R: 100, G: 110, B: 120, A: 0}
	if got := Blend(a, b, 0); got != a {
		t.Fatalf("Blend(ratio=0) = %v, want %v", got, a)
	}
	if got := Blend(a, b, 1); got != b {
		t.Fatalf("Blend(ratio=1) = %v, want %v", got, b)
	}
}

func TestBlendMidpoint(t *testing.T) {
	black := color.RGBA{R: 0, G: 0, B: 0, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	got := Blend(black, white, 0.5)
	want := color.RGBA{R: 128, G: 128, B: 128, A: 255}
	if got != want {
		t.Fatalf("Blend midpoint = %v, want %v", got, want)
	}
}

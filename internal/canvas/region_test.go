package canvas

import (
	"image/color"
	"testing"
)

func TestCanvasCopyPasteRegion(t *testing.T) {
	c, err := New(8, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	red := color.RGBA{R: 255, A: 255}
	green := color.RGBA{G: 255, A: 255}
	_ = c.SetPixel(1, 1, red)
	_ = c.SetPixel(2, 1, green)
	_ = c.SetPixel(1, 2, green)
	_ = c.SetPixel(2, 2, red)

	pixels, err := c.CopyRegion(1, 1, 2, 2)
	if err != nil {
		t.Fatalf("unexpected copy error: %v", err)
	}
	if len(pixels) != 4 {
		t.Fatalf("expected 4 pixels, got %d", len(pixels))
	}
	want := []color.RGBA{red, green, green, red}
	for i, w := range want {
		if pixels[i] != w {
			t.Fatalf("pixel[%d] = %v, want %v", i, pixels[i], w)
		}
	}

	if err := c.PasteRegion(5, 5, 2, 2, pixels); err != nil {
		t.Fatalf("unexpected paste error: %v", err)
	}
	for i, p := range [][2]int{{5, 5}, {6, 5}, {5, 6}, {6, 6}} {
		got, err := c.GetPixel(p[0], p[1])
		if err != nil {
			t.Fatalf("unexpected get error at %v: %v", p, err)
		}
		if got != want[i] {
			t.Fatalf("pasted pixel at %v = %v, want %v", p, got, want[i])
		}
	}

	// Original region is untouched by paste.
	got, err := c.GetPixel(1, 1)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if got != red {
		t.Fatalf("expected original region unchanged, got %v at (1,1)", got)
	}
}

func TestCanvasCopyRegionErrors(t *testing.T) {
	c, err := New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := c.CopyRegion(0, 0, 0, 1); err == nil {
		t.Fatalf("expected invalid_args error for zero width")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "invalid_args" {
		t.Fatalf("expected invalid_args, got %v", err)
	}
	if _, err := c.CopyRegion(3, 3, 2, 2); err == nil {
		t.Fatalf("expected out_of_bounds error")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "out_of_bounds" {
		t.Fatalf("expected out_of_bounds, got %v", err)
	}
}

func TestCanvasPasteRegionErrors(t *testing.T) {
	c, err := New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.PasteRegion(3, 3, 2, 2, make([]color.RGBA, 4)); err == nil {
		t.Fatalf("expected out_of_bounds error")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "out_of_bounds" {
		t.Fatalf("expected out_of_bounds, got %v", err)
	}
	if err := c.PasteRegion(0, 0, 2, 2, make([]color.RGBA, 3)); err == nil {
		t.Fatalf("expected invalid_args error for pixel count mismatch")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "invalid_args" {
		t.Fatalf("expected invalid_args, got %v", err)
	}
}

func TestCanvasMoveRegion(t *testing.T) {
	c, err := New(8, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	blue := color.RGBA{B: 255, A: 255}
	_ = c.FillRect(1, 1, 2, 2, blue)

	if err := c.MoveRegion(1, 1, 2, 2, 3, 3); err != nil {
		t.Fatalf("unexpected move error: %v", err)
	}

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			got, err := c.GetPixel(x, y)
			if err != nil {
				t.Fatalf("unexpected get error at (%d,%d): %v", x, y, err)
			}
			inDest := x >= 4 && x <= 5 && y >= 4 && y <= 5
			if inDest {
				if got != blue {
					t.Fatalf("expected blue at destination (%d,%d), got %v", x, y, got)
				}
			} else if got != (color.RGBA{}) {
				t.Fatalf("expected transparent at (%d,%d) after move, got %v", x, y, got)
			}
		}
	}
}

func TestCanvasMoveRegionOverlapping(t *testing.T) {
	c, err := New(8, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	red := color.RGBA{R: 255, A: 255}
	_ = c.FillRect(1, 1, 3, 3, red)

	// Move by (1,1): destination overlaps source.
	if err := c.MoveRegion(1, 1, 3, 3, 1, 1); err != nil {
		t.Fatalf("unexpected move error: %v", err)
	}
	for y := 2; y <= 4; y++ {
		for x := 2; x <= 4; x++ {
			got, err := c.GetPixel(x, y)
			if err != nil {
				t.Fatalf("unexpected get error at (%d,%d): %v", x, y, err)
			}
			if got != red {
				t.Fatalf("expected red at overlapping destination (%d,%d), got %v", x, y, got)
			}
		}
	}
	// Row/col 1 (outside the new position) should be cleared.
	if got, _ := c.GetPixel(1, 1); got != (color.RGBA{}) {
		t.Fatalf("expected (1,1) cleared after overlapping move, got %v", got)
	}
}

func TestCanvasMoveRegionErrors(t *testing.T) {
	c, err := New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Source out of bounds.
	if err := c.MoveRegion(3, 3, 2, 2, 0, 0); err == nil {
		t.Fatalf("expected out_of_bounds error for source")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "out_of_bounds" {
		t.Fatalf("expected out_of_bounds, got %v", err)
	}
	// Destination out of bounds.
	if err := c.MoveRegion(0, 0, 2, 2, 3, 3); err == nil {
		t.Fatalf("expected out_of_bounds error for destination")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "out_of_bounds" {
		t.Fatalf("expected out_of_bounds, got %v", err)
	}
}

func TestCanvasMirrorRegionHorizontal(t *testing.T) {
	c, err := New(4, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	left := color.RGBA{R: 255, A: 255}
	right := color.RGBA{B: 255, A: 255}
	_ = c.SetPixel(0, 0, left)
	_ = c.SetPixel(1, 0, right)

	if err := c.MirrorRegion(0, 0, 2, 1, "horizontal"); err != nil {
		t.Fatalf("unexpected mirror error: %v", err)
	}
	got0, _ := c.GetPixel(0, 0)
	got1, _ := c.GetPixel(1, 0)
	if got0 != right || got1 != left {
		t.Fatalf("horizontal mirror = (%v,%v), want (%v,%v)", got0, got1, right, left)
	}
}

func TestCanvasMirrorRegionVertical(t *testing.T) {
	c, err := New(2, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	top := color.RGBA{R: 255, A: 255}
	bottom := color.RGBA{B: 255, A: 255}
	_ = c.SetPixel(0, 0, top)
	_ = c.SetPixel(0, 1, bottom)

	if err := c.MirrorRegion(0, 0, 1, 2, "vertical"); err != nil {
		t.Fatalf("unexpected mirror error: %v", err)
	}
	got0, _ := c.GetPixel(0, 0)
	got1, _ := c.GetPixel(0, 1)
	if got0 != bottom || got1 != top {
		t.Fatalf("vertical mirror = (%v,%v), want (%v,%v)", got0, got1, bottom, top)
	}
}

func TestCanvasMirrorRegionErrors(t *testing.T) {
	c, err := New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.MirrorRegion(0, 0, 2, 2, "diagonal"); err == nil {
		t.Fatalf("expected invalid_args error for unknown axis")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "invalid_args" {
		t.Fatalf("expected invalid_args, got %v", err)
	}
	if err := c.MirrorRegion(3, 3, 2, 2, "horizontal"); err == nil {
		t.Fatalf("expected out_of_bounds error")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "out_of_bounds" {
		t.Fatalf("expected out_of_bounds, got %v", err)
	}
}

package canvas

import (
	"image/color"
	"testing"
)

func TestCanvasCircleOutline(t *testing.T) {
	c, err := New(9, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	red := color.RGBA{R: 255, A: 255}
	if err := c.Circle(4, 4, 3, red, false); err != nil {
		t.Fatalf("unexpected circle error: %v", err)
	}

	for _, p := range [][2]int{{7, 4}, {1, 4}, {4, 7}, {4, 1}} {
		got, err := c.GetPixel(p[0], p[1])
		if err != nil {
			t.Fatalf("unexpected get error at %v: %v", p, err)
		}
		if got != red {
			t.Fatalf("expected red at %v, got %v", p, got)
		}
	}

	center, err := c.GetPixel(4, 4)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if center != (color.RGBA{}) {
		t.Fatalf("expected outline circle to leave center unset, got %v", center)
	}
}

func TestCanvasCircleFilled(t *testing.T) {
	c, err := New(9, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	blue := color.RGBA{B: 255, A: 255}
	if err := c.Circle(4, 4, 3, blue, true); err != nil {
		t.Fatalf("unexpected circle error: %v", err)
	}

	for _, p := range [][2]int{{4, 4}, {6, 4}, {4, 6}, {2, 4}, {4, 2}} {
		got, err := c.GetPixel(p[0], p[1])
		if err != nil {
			t.Fatalf("unexpected get error at %v: %v", p, err)
		}
		if got != blue {
			t.Fatalf("expected blue at %v, got %v", p, got)
		}
	}

	corner, err := c.GetPixel(0, 0)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if corner != (color.RGBA{}) {
		t.Fatalf("expected corner outside filled circle to be unset, got %v", corner)
	}
}

func TestCanvasCircleErrors(t *testing.T) {
	c, err := New(8, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.Circle(4, 4, 0, color.RGBA{}, false); err == nil {
		t.Fatalf("expected invalid_args error for zero radius")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "invalid_args" {
		t.Fatalf("expected invalid_args, got %v", err)
	}

	outOfBounds := [][3]int{
		{1, 4, 3}, // cx-r < 0
		{4, 1, 3}, // cy-r < 0
		{6, 4, 3}, // cx+r >= width
		{4, 6, 3}, // cy+r >= height
	}
	for _, args := range outOfBounds {
		if err := c.Circle(args[0], args[1], args[2], color.RGBA{}, false); err == nil {
			t.Fatalf("expected out_of_bounds error for circle %v", args)
		} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "out_of_bounds" {
			t.Fatalf("expected out_of_bounds for circle %v, got %v", args, err)
		}
	}
}

func TestCanvasEllipseOutline(t *testing.T) {
	c, err := New(11, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	green := color.RGBA{G: 255, A: 255}
	if err := c.Ellipse(5, 4, 4, 3, green, false); err != nil {
		t.Fatalf("unexpected ellipse error: %v", err)
	}

	for _, p := range [][2]int{{9, 4}, {1, 4}, {5, 7}, {5, 1}} {
		got, err := c.GetPixel(p[0], p[1])
		if err != nil {
			t.Fatalf("unexpected get error at %v: %v", p, err)
		}
		if got != green {
			t.Fatalf("expected green at %v, got %v", p, got)
		}
	}

	center, err := c.GetPixel(5, 4)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if center != (color.RGBA{}) {
		t.Fatalf("expected outline ellipse to leave center unset, got %v", center)
	}
}

func TestCanvasEllipseFilled(t *testing.T) {
	c, err := New(11, 9)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	yellow := color.RGBA{R: 255, G: 255, A: 255}
	if err := c.Ellipse(5, 4, 4, 3, yellow, true); err != nil {
		t.Fatalf("unexpected ellipse error: %v", err)
	}

	for _, p := range [][2]int{{5, 4}, {8, 4}, {5, 6}} {
		got, err := c.GetPixel(p[0], p[1])
		if err != nil {
			t.Fatalf("unexpected get error at %v: %v", p, err)
		}
		if got != yellow {
			t.Fatalf("expected yellow at %v, got %v", p, got)
		}
	}

	corner, err := c.GetPixel(0, 0)
	if err != nil {
		t.Fatalf("unexpected get error: %v", err)
	}
	if corner != (color.RGBA{}) {
		t.Fatalf("expected corner outside filled ellipse to be unset, got %v", corner)
	}
}

func TestCanvasEllipseErrors(t *testing.T) {
	c, err := New(8, 8)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	invalidArgs := [][4]int{
		{4, 4, 0, 2},
		{4, 4, 2, 0},
	}
	for _, args := range invalidArgs {
		if err := c.Ellipse(args[0], args[1], args[2], args[3], color.RGBA{}, false); err == nil {
			t.Fatalf("expected invalid_args error for ellipse %v", args)
		} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "invalid_args" {
			t.Fatalf("expected invalid_args for ellipse %v, got %v", args, err)
		}
	}

	outOfBounds := [][4]int{
		{1, 4, 3, 2},
		{4, 1, 2, 3},
		{6, 4, 3, 2},
		{4, 6, 2, 3},
	}
	for _, args := range outOfBounds {
		if err := c.Ellipse(args[0], args[1], args[2], args[3], color.RGBA{}, false); err == nil {
			t.Fatalf("expected out_of_bounds error for ellipse %v", args)
		} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "out_of_bounds" {
			t.Fatalf("expected out_of_bounds for ellipse %v, got %v", args, err)
		}
	}
}

func TestCanvasDitherFillCheckerboard(t *testing.T) {
	c, err := New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	c1 := color.RGBA{R: 255, A: 255}
	c2 := color.RGBA{B: 255, A: 255}
	if err := c.DitherFill(0, 0, 4, 4, c1, c2, ""); err != nil {
		t.Fatalf("unexpected dither_fill error: %v", err)
	}

	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			got, err := c.GetPixel(x, y)
			if err != nil {
				t.Fatalf("unexpected get error at (%d,%d): %v", x, y, err)
			}
			want := c2
			if (x+y)%2 == 0 {
				want = c1
			}
			if got != want {
				t.Fatalf("checkerboard at (%d,%d) = %v, want %v", x, y, got, want)
			}
		}
	}
}

func TestCanvasDitherFillHorizontalAndVertical(t *testing.T) {
	c1 := color.RGBA{R: 255, A: 255}
	c2 := color.RGBA{B: 255, A: 255}

	horiz, err := New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := horiz.DitherFill(0, 0, 4, 4, c1, c2, "horizontal"); err != nil {
		t.Fatalf("unexpected dither_fill error: %v", err)
	}
	for y := 0; y < 4; y++ {
		want := c2
		if y%2 == 0 {
			want = c1
		}
		got, err := horiz.GetPixel(0, y)
		if err != nil {
			t.Fatalf("unexpected get error at row %d: %v", y, err)
		}
		if got != want {
			t.Fatalf("horizontal row %d = %v, want %v", y, got, want)
		}
	}

	vert, err := New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := vert.DitherFill(0, 0, 4, 4, c1, c2, "vertical"); err != nil {
		t.Fatalf("unexpected dither_fill error: %v", err)
	}
	for x := 0; x < 4; x++ {
		want := c2
		if x%2 == 0 {
			want = c1
		}
		got, err := vert.GetPixel(x, 0)
		if err != nil {
			t.Fatalf("unexpected get error at col %d: %v", x, err)
		}
		if got != want {
			t.Fatalf("vertical col %d = %v, want %v", x, got, want)
		}
	}
}

func TestCanvasDitherFillErrors(t *testing.T) {
	c, err := New(4, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := c.DitherFill(0, 0, 0, 1, color.RGBA{}, color.RGBA{}, ""); err == nil {
		t.Fatalf("expected invalid_args error for zero width")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "invalid_args" {
		t.Fatalf("expected invalid_args, got %v", err)
	}

	if err := c.DitherFill(0, 0, 2, 2, color.RGBA{}, color.RGBA{}, "diagonal"); err == nil {
		t.Fatalf("expected invalid_args error for unknown pattern")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "invalid_args" {
		t.Fatalf("expected invalid_args for unknown pattern, got %v", err)
	}

	if err := c.DitherFill(3, 3, 3, 1, color.RGBA{}, color.RGBA{}, ""); err == nil {
		t.Fatalf("expected out_of_bounds error")
	} else if canvasErr, ok := err.(Error); !ok || canvasErr.Code != "out_of_bounds" {
		t.Fatalf("expected out_of_bounds, got %v", err)
	}
}

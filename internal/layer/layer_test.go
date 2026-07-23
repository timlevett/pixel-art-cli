package layer

import (
	"errors"
	"image/color"
	"testing"

	"pxcli/internal/canvas"
	"pxcli/internal/history"
)

func newTestStore(t *testing.T, w, h int) *Store {
	t.Helper()
	c, err := canvas.New(w, h)
	if err != nil {
		t.Fatalf("canvas.New() error = %v", err)
	}
	return New(c, history.New(c))
}

func TestNewHasBaseLayerActive(t *testing.T) {
	s := newTestStore(t, 4, 4)
	if got := s.ActiveName(); got != Base {
		t.Fatalf("ActiveName() = %q, want %q", got, Base)
	}
	if got := s.List(); len(got) != 1 || got[0] != Base {
		t.Fatalf("List() = %v, want [%q]", got, Base)
	}
	if !s.Active().Visible {
		t.Fatalf("base layer should be visible by default")
	}
}

func TestAddAndSelect(t *testing.T) {
	s := newTestStore(t, 4, 4)
	if err := s.Add("sprite"); err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if got := s.List(); len(got) != 2 || got[0] != Base || got[1] != "sprite" {
		t.Fatalf("List() = %v, want [base sprite]", got)
	}
	if err := s.Select("sprite"); err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if got := s.ActiveName(); got != "sprite" {
		t.Fatalf("ActiveName() = %q, want sprite", got)
	}
	if s.Active().Canvas.Width() != 4 || s.Active().Canvas.Height() != 4 {
		t.Fatalf("new layer canvas dims = %dx%d, want 4x4 (match base)", s.Active().Canvas.Width(), s.Active().Canvas.Height())
	}
}

func TestAddErrors(t *testing.T) {
	s := newTestStore(t, 4, 4)
	if err := s.Add(""); !isCode(err, "invalid_args") {
		t.Fatalf("Add(empty) error = %v, want invalid_args", err)
	}
	if err := s.Add(Base); !isCode(err, "invalid_args") {
		t.Fatalf("Add(base) error = %v, want invalid_args (reserved name)", err)
	}
	_ = s.Add("sprite")
	if err := s.Add("sprite"); !isCode(err, "invalid_args") {
		t.Fatalf("Add(duplicate) error = %v, want invalid_args", err)
	}
}

func TestSelectUndefined(t *testing.T) {
	s := newTestStore(t, 4, 4)
	if err := s.Select("missing"); !isCode(err, "invalid_layer") {
		t.Fatalf("Select(missing) error = %v, want invalid_layer", err)
	}
}

func TestSetVisible(t *testing.T) {
	s := newTestStore(t, 4, 4)
	_ = s.Add("sprite")
	if err := s.SetVisible("sprite", false); err != nil {
		t.Fatalf("SetVisible() error = %v", err)
	}
	if s.layers["sprite"].Visible {
		t.Fatalf("expected sprite layer to be hidden")
	}
	if err := s.SetVisible("missing", true); !isCode(err, "invalid_layer") {
		t.Fatalf("SetVisible(missing) error = %v, want invalid_layer", err)
	}
}

func TestLayersHaveIndependentCanvasesAndHistory(t *testing.T) {
	s := newTestStore(t, 4, 4)
	_ = s.Add("sprite")

	base := s.layers[Base]
	_ = base.Canvas.SetPixel(0, 0, color.RGBA{R: 255, A: 255})

	_ = s.Select("sprite")
	sprite := s.Active()
	got, _ := sprite.Canvas.GetPixel(0, 0)
	if got != (color.RGBA{}) {
		t.Fatalf("expected sprite layer unaffected by base edits, got %v", got)
	}

	_ = sprite.History.Apply(func(c *canvas.Canvas) error {
		return c.SetPixel(1, 1, color.RGBA{B: 255, A: 255})
	})
	if err := sprite.History.Undo(); err != nil {
		t.Fatalf("sprite layer undo error = %v", err)
	}
	got, _ = sprite.Canvas.GetPixel(1, 1)
	if got != (color.RGBA{}) {
		t.Fatalf("expected sprite layer edit undone, got %v", got)
	}
	// Base layer's own history/canvas is untouched by the sprite layer's undo.
	got, _ = base.Canvas.GetPixel(0, 0)
	if got != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("expected base layer edit to survive sprite layer's undo, got %v", got)
	}
}

func TestFlattenComposesVisibleLayersInOrder(t *testing.T) {
	s := newTestStore(t, 2, 1)
	_ = s.Add("overlay")

	base := s.layers[Base]
	_ = base.Canvas.SetPixel(0, 0, color.RGBA{R: 255, A: 255})
	_ = base.Canvas.SetPixel(1, 0, color.RGBA{R: 255, A: 255})

	overlay := s.layers["overlay"]
	_ = overlay.Canvas.SetPixel(0, 0, color.RGBA{G: 255, A: 255}) // opaque: fully replaces base at (0,0)
	// (1,0) left transparent on the overlay: base shows through.

	flattened := s.Flatten()
	got0, _ := flattened.GetPixel(0, 0)
	got1, _ := flattened.GetPixel(1, 0)
	if got0 != (color.RGBA{G: 255, A: 255}) {
		t.Fatalf("flatten(0,0) = %v, want opaque overlay color", got0)
	}
	if got1 != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("flatten(1,0) = %v, want base color showing through transparent overlay", got1)
	}
}

func TestFlattenSkipsHiddenLayers(t *testing.T) {
	s := newTestStore(t, 1, 1)
	_ = s.Add("overlay")
	_ = s.layers["overlay"].Canvas.SetPixel(0, 0, color.RGBA{G: 255, A: 255})
	_ = s.SetVisible("overlay", false)

	flattened := s.Flatten()
	got, _ := flattened.GetPixel(0, 0)
	if got != (color.RGBA{}) {
		t.Fatalf("flatten with hidden overlay = %v, want transparent (overlay excluded)", got)
	}
}

func TestFlattenBlendsSemiTransparentPixels(t *testing.T) {
	s := newTestStore(t, 1, 1)
	_ = s.Add("overlay")
	_ = s.layers[Base].Canvas.SetPixel(0, 0, color.RGBA{R: 0, G: 0, B: 0, A: 255})
	_ = s.layers["overlay"].Canvas.SetPixel(0, 0, color.RGBA{R: 255, G: 255, B: 255, A: 128})

	flattened := s.Flatten()
	got, _ := flattened.GetPixel(0, 0)
	// 128/255 ~= 0.502 alpha over opaque black: each channel ~= 255*0.502 ≈ 128.
	if got.A != 255 {
		t.Fatalf("flatten alpha = %d, want 255 (opaque base beneath)", got.A)
	}
	if got.R < 120 || got.R > 135 {
		t.Fatalf("flatten blended R = %d, want ~128", got.R)
	}
}

func isCode(err error, code string) bool {
	var lerr Error
	if !errors.As(err, &lerr) {
		return false
	}
	return lerr.Code == code
}

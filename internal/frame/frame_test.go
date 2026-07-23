package frame

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

func TestNewHasSingleActiveFrame(t *testing.T) {
	s := newTestStore(t, 4, 4)
	if got := s.Count(); got != 1 {
		t.Fatalf("Count() = %d, want 1", got)
	}
	if got := s.ActiveIndex(); got != 0 {
		t.Fatalf("ActiveIndex() = %d, want 0", got)
	}
}

func TestAddAndSelect(t *testing.T) {
	s := newTestStore(t, 4, 4)
	idx, err := s.Add()
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}
	if idx != 1 {
		t.Fatalf("Add() index = %d, want 1", idx)
	}
	if got := s.Count(); got != 2 {
		t.Fatalf("Count() = %d, want 2", got)
	}
	if err := s.Select(1); err != nil {
		t.Fatalf("Select() error = %v", err)
	}
	if got := s.ActiveIndex(); got != 1 {
		t.Fatalf("ActiveIndex() = %d, want 1", got)
	}
	if s.Active().Active().Canvas.Width() != 4 || s.Active().Active().Canvas.Height() != 4 {
		t.Fatalf("new frame canvas dims mismatch, want 4x4")
	}
}

func TestSelectOutOfRange(t *testing.T) {
	s := newTestStore(t, 4, 4)
	if err := s.Select(1); !isCode(err, "invalid_frame") {
		t.Fatalf("Select(1) error = %v, want invalid_frame", err)
	}
	if err := s.Select(-1); !isCode(err, "invalid_frame") {
		t.Fatalf("Select(-1) error = %v, want invalid_frame", err)
	}
}

func TestFramesHaveIndependentCanvasesAndHistory(t *testing.T) {
	s := newTestStore(t, 4, 4)
	_, _ = s.Add()

	frame0 := s.Active()
	_ = frame0.Active().Canvas.SetPixel(0, 0, color.RGBA{R: 255, A: 255})

	_ = s.Select(1)
	frame1 := s.Active()
	got, _ := frame1.Active().Canvas.GetPixel(0, 0)
	if got != (color.RGBA{}) {
		t.Fatalf("expected frame 1 unaffected by frame 0 edits, got %v", got)
	}
}

func TestGhostBlendsTargetUnderActiveAtOpacity(t *testing.T) {
	s := newTestStore(t, 2, 1)
	_, _ = s.Add()

	// frame 0 (target): opaque red at (0,0), opaque blue at (1,0)
	frame0 := s.Active()
	_ = frame0.Active().Canvas.SetPixel(0, 0, color.RGBA{R: 255, A: 255})
	_ = frame0.Active().Canvas.SetPixel(1, 0, color.RGBA{B: 255, A: 255})

	// frame 1 (active): opaque green at (0,0) only, transparent at (1,0)
	_ = s.Select(1)
	_ = s.Active().Active().Canvas.SetPixel(0, 0, color.RGBA{G: 255, A: 255})

	ghosted, err := s.Ghost(0, 0.5)
	if err != nil {
		t.Fatalf("Ghost() error = %v", err)
	}

	// (0,0): active is opaque, so ghost target should be fully hidden.
	got, _ := ghosted.GetPixel(0, 0)
	if got != (color.RGBA{G: 255, A: 255}) {
		t.Fatalf("Ghost(0,0) = %v, want opaque active green", got)
	}

	// (1,0): active is transparent, so the dimmed target should show
	// through at ~50% alpha, full blue.
	got, _ = ghosted.GetPixel(1, 0)
	if got.A < 120 || got.A > 135 {
		t.Fatalf("Ghost(1,0) alpha = %d, want ~128 (50%% of 255)", got.A)
	}
	if got.B != 255 {
		t.Fatalf("Ghost(1,0) blue = %d, want 255", got.B)
	}
}

func TestGhostUndefinedFrame(t *testing.T) {
	s := newTestStore(t, 2, 2)
	if _, err := s.Ghost(5, 0.5); !isCode(err, "invalid_frame") {
		t.Fatalf("Ghost(5) error = %v, want invalid_frame", err)
	}
}

func TestSheetTilesFramesInGridOrder(t *testing.T) {
	s := newTestStore(t, 2, 2)
	_, _ = s.Add()
	_, _ = s.Add()

	_ = s.frames[0].Active().Canvas.SetPixel(0, 0, color.RGBA{R: 255, A: 255})
	_ = s.frames[1].Active().Canvas.SetPixel(0, 0, color.RGBA{G: 255, A: 255})
	_ = s.frames[2].Active().Canvas.SetPixel(0, 0, color.RGBA{B: 255, A: 255})

	sheet, err := s.Sheet(2)
	if err != nil {
		t.Fatalf("Sheet() error = %v", err)
	}
	if sheet.Width() != 4 || sheet.Height() != 4 {
		t.Fatalf("Sheet() dims = %dx%d, want 4x4 (2 cols x 2 rows of 2x2 frames)", sheet.Width(), sheet.Height())
	}

	// frame 0 top-left tile
	got, _ := sheet.GetPixel(0, 0)
	if got != (color.RGBA{R: 255, A: 255}) {
		t.Fatalf("Sheet frame0 tile = %v, want red", got)
	}
	// frame 1 top-right tile (col 1, row 0 -> offset x=2)
	got, _ = sheet.GetPixel(2, 0)
	if got != (color.RGBA{G: 255, A: 255}) {
		t.Fatalf("Sheet frame1 tile = %v, want green", got)
	}
	// frame 2 wraps to row 1, col 0 -> offset y=2
	got, _ = sheet.GetPixel(0, 2)
	if got != (color.RGBA{B: 255, A: 255}) {
		t.Fatalf("Sheet frame2 tile = %v, want blue", got)
	}
	// padding cell (col 1, row 1) stays transparent
	got, _ = sheet.GetPixel(2, 2)
	if got != (color.RGBA{}) {
		t.Fatalf("Sheet padding tile = %v, want transparent", got)
	}
}

func TestSheetDefaultColsIsFrameCount(t *testing.T) {
	s := newTestStore(t, 2, 2)
	_, _ = s.Add()

	sheet, err := s.Sheet(0)
	if err != nil {
		t.Fatalf("Sheet(0) error = %v", err)
	}
	if sheet.Width() != 4 || sheet.Height() != 2 {
		t.Fatalf("Sheet(0) dims = %dx%d, want 4x2 (all frames in one row)", sheet.Width(), sheet.Height())
	}
}

func isCode(err error, code string) bool {
	var ferr Error
	if !errors.As(err, &ferr) {
		return false
	}
	return ferr.Code == code
}

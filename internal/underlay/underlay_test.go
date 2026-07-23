package underlay

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"pxcli/internal/canvas"
)

func writeTestPNG(t *testing.T, w, h int, fill color.RGBA) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, fill)
		}
	}
	path := filepath.Join(t.TempDir(), "ref.png")
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	return path
}

func TestImportDimsAndResizesToCanvasSize(t *testing.T) {
	path := writeTestPNG(t, 8, 8, color.RGBA{R: 255, A: 255})
	s := New(4, 4)

	if s.HasImage() {
		t.Fatalf("HasImage() before import = true, want false")
	}
	if err := s.Import(path, 0.5); err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if !s.HasImage() {
		t.Fatalf("HasImage() after import = false, want true")
	}

	top, _ := canvas.New(4, 4) // fully transparent, so composited pixel is pure underlay
	composed, err := s.CompositeUnder(top)
	if err != nil {
		t.Fatalf("CompositeUnder() error = %v", err)
	}
	if composed.Width() != 4 || composed.Height() != 4 {
		t.Fatalf("CompositeUnder() dims = %dx%d, want 4x4 (resized to canvas)", composed.Width(), composed.Height())
	}
	got, _ := composed.GetPixel(0, 0)
	if got.R != 255 {
		t.Fatalf("composed red channel = %d, want 255", got.R)
	}
	if got.A < 120 || got.A > 135 {
		t.Fatalf("composed alpha = %d, want ~128 (50%% of 255)", got.A)
	}
}

func TestCompositeUnderTopWinsWhereOpaque(t *testing.T) {
	path := writeTestPNG(t, 2, 2, color.RGBA{R: 255, A: 255})
	s := New(2, 2)
	if err := s.Import(path, 1.0); err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	top, _ := canvas.New(2, 2)
	_ = top.SetPixel(0, 0, color.RGBA{G: 255, A: 255})

	composed, err := s.CompositeUnder(top)
	if err != nil {
		t.Fatalf("CompositeUnder() error = %v", err)
	}
	got, _ := composed.GetPixel(0, 0)
	if got != (color.RGBA{G: 255, A: 255}) {
		t.Fatalf("composed(0,0) = %v, want opaque top green (underlay hidden)", got)
	}
	got, _ = composed.GetPixel(1, 0)
	if got.R != 255 {
		t.Fatalf("composed(1,0) red = %d, want 255 (underlay shows through transparent top)", got.R)
	}
}

func TestCompositeUnderWithoutImportReturnsTopUnchanged(t *testing.T) {
	s := New(2, 2)
	top, _ := canvas.New(2, 2)
	_ = top.SetPixel(0, 0, color.RGBA{B: 255, A: 255})

	composed, err := s.CompositeUnder(top)
	if err != nil {
		t.Fatalf("CompositeUnder() error = %v", err)
	}
	got, _ := composed.GetPixel(0, 0)
	if got != (color.RGBA{B: 255, A: 255}) {
		t.Fatalf("composed(0,0) = %v, want top pixel unchanged", got)
	}
}

func TestImportErrors(t *testing.T) {
	s := New(4, 4)

	if err := s.Import("/nonexistent/path.png", 0.5); !isCode(err, "io") {
		t.Fatalf("Import(missing file) error = %v, want io", err)
	}

	badPath := filepath.Join(t.TempDir(), "not-an-image.png")
	if err := os.WriteFile(badPath, []byte("not a png"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}
	if err := s.Import(badPath, 0.5); !isCode(err, "io") {
		t.Fatalf("Import(bad image) error = %v, want io", err)
	}

	validPath := writeTestPNG(t, 4, 4, color.RGBA{A: 255})
	if err := s.Import(validPath, -0.1); !isCode(err, "invalid_args") {
		t.Fatalf("Import(opacity<0) error = %v, want invalid_args", err)
	}
	if err := s.Import(validPath, 1.1); !isCode(err, "invalid_args") {
		t.Fatalf("Import(opacity>1) error = %v, want invalid_args", err)
	}
}

func TestImportBumpsVersion(t *testing.T) {
	path := writeTestPNG(t, 2, 2, color.RGBA{A: 255})
	s := New(2, 2)
	if got := s.Version(); got != 0 {
		t.Fatalf("Version() before import = %d, want 0", got)
	}
	_ = s.Import(path, 0.5)
	if got := s.Version(); got != 1 {
		t.Fatalf("Version() after first import = %d, want 1", got)
	}
	_ = s.Import(path, 0.5)
	if got := s.Version(); got != 2 {
		t.Fatalf("Version() after second import = %d, want 2", got)
	}
}

func isCode(err error, code string) bool {
	var uerr Error
	if !errors.As(err, &uerr) {
		return false
	}
	return uerr.Code == code
}

package daemon

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"pxcli/internal/canvas"
	"pxcli/internal/underlay"
)

func writeTestReferencePNG(t *testing.T, w, h int, fill color.RGBA) string {
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

func TestUnderlayRenderSourceDirtyOnFirstCheck(t *testing.T) {
	grid, _ := canvas.New(2, 2)
	store := underlay.New(2, 2)
	source := newUnderlayRenderSource(grid, store)

	if !source.Dirty() {
		t.Fatalf("Dirty() before any render = false, want true (must always render once)")
	}
	source.RenderSnapshot()
	if source.Dirty() {
		t.Fatalf("Dirty() after render with no changes = true, want false")
	}
}

func TestUnderlayRenderSourceDirtyAfterUnderlayImport(t *testing.T) {
	grid, _ := canvas.New(2, 2)
	store := underlay.New(2, 2)
	source := newUnderlayRenderSource(grid, store)
	source.RenderSnapshot()
	if source.Dirty() {
		t.Fatalf("Dirty() after initial render = true, want false")
	}

	path := writeTestReferencePNG(t, 2, 2, color.RGBA{R: 255, A: 255})
	if err := store.Import(path, 0.5); err != nil {
		t.Fatalf("Import() error = %v", err)
	}
	if !source.Dirty() {
		t.Fatalf("Dirty() after underlay import (grid unchanged) = false, want true")
	}
	source.RenderSnapshot()
	if source.Dirty() {
		t.Fatalf("Dirty() after re-render = true, want false")
	}
}

func TestUnderlayRenderSourceDirtyOnGridChange(t *testing.T) {
	grid, _ := canvas.New(2, 2)
	store := underlay.New(2, 2)
	source := newUnderlayRenderSource(grid, store)
	source.RenderSnapshot()

	_ = grid.SetPixel(0, 0, color.RGBA{G: 255, A: 255})
	if !source.Dirty() {
		t.Fatalf("Dirty() after grid change = false, want true")
	}
}

func TestUnderlayRenderSourceSnapshotCompositesUnderlayBeneathGrid(t *testing.T) {
	grid, _ := canvas.New(2, 1)
	_ = grid.SetPixel(0, 0, color.RGBA{G: 255, A: 255}) // opaque, hides underlay

	store := underlay.New(2, 1)
	path := writeTestReferencePNG(t, 2, 1, color.RGBA{R: 255, A: 255})
	if err := store.Import(path, 1.0); err != nil {
		t.Fatalf("Import() error = %v", err)
	}

	source := newUnderlayRenderSource(grid, store)
	snap := source.RenderSnapshot()
	if snap.Width != 2 || snap.Height != 1 {
		t.Fatalf("RenderSnapshot dims = %dx%d, want 2x1", snap.Width, snap.Height)
	}
	// pixel 0: grid opaque green wins
	if snap.Pixels[1] != 255 || snap.Pixels[0] != 0 {
		t.Fatalf("pixel 0 = rgba(%d,%d,%d,%d), want opaque green", snap.Pixels[0], snap.Pixels[1], snap.Pixels[2], snap.Pixels[3])
	}
	// pixel 1: grid transparent, underlay red shows through
	if snap.Pixels[4] != 255 {
		t.Fatalf("pixel 1 red = %d, want 255 (underlay showing through transparent grid)", snap.Pixels[4])
	}
}

func TestUnderlayRenderSourceSnapshotWithoutImportMatchesGrid(t *testing.T) {
	grid, _ := canvas.New(1, 1)
	_ = grid.SetPixel(0, 0, color.RGBA{B: 255, A: 255})
	store := underlay.New(1, 1)

	source := newUnderlayRenderSource(grid, store)
	snap := source.RenderSnapshot()
	if snap.Pixels[2] != 255 {
		t.Fatalf("pixel blue = %d, want 255 (no underlay, grid passthrough)", snap.Pixels[2])
	}
}

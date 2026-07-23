package daemon

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pxcli/internal/protocol"
)

func writeTestReferenceImage(t *testing.T, w, h int, fill color.RGBA) string {
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

func TestHandlerImportReferenceDefaultOpacity(t *testing.T) {
	handler := newTestHandler(t, 2, 2)
	path := writeTestReferenceImage(t, 2, 2, color.RGBA{R: 255, A: 255})

	resp := handler.Handle(protocol.Request{Command: "import_reference", Args: []string{path}})
	if resp != "ok" {
		t.Fatalf("import_reference = %q, want ok", resp)
	}
	if !handler.underlay.HasImage() {
		t.Fatalf("expected underlay to have an image after import")
	}
}

func TestHandlerImportReferenceExplicitOpacity(t *testing.T) {
	handler := newTestHandler(t, 2, 2)
	path := writeTestReferenceImage(t, 2, 2, color.RGBA{R: 255, A: 255})

	resp := handler.Handle(protocol.Request{Command: "import_reference", Args: []string{path, "0.8"}})
	if resp != "ok" {
		t.Fatalf("import_reference with opacity = %q, want ok", resp)
	}
}

func TestHandlerImportReferenceErrors(t *testing.T) {
	handler := newTestHandler(t, 2, 2)

	resp := handler.Handle(protocol.Request{Command: "import_reference", Args: []string{"/nonexistent/ref.png"}})
	if !strings.HasPrefix(resp, "err io ") {
		t.Fatalf("import_reference missing file = %q, want err io", resp)
	}

	path := writeTestReferenceImage(t, 2, 2, color.RGBA{A: 255})
	resp = handler.Handle(protocol.Request{Command: "import_reference", Args: []string{path, "2.0"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("import_reference opacity out of range = %q, want invalid_args", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "import_reference", Args: []string{}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("import_reference no args = %q, want invalid_args", resp)
	}
}

func TestHandlerDrawingCommandsDoNotTouchUnderlay(t *testing.T) {
	handler := newTestHandler(t, 2, 2)
	path := writeTestReferenceImage(t, 2, 2, color.RGBA{R: 255, A: 255})
	handler.Handle(protocol.Request{Command: "import_reference", Args: []string{path, "1.0"}})

	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#00ff00"}})

	// A plain export must not include the underlay: (0,0) was drawn green
	// (opaque, wins regardless), but (1,0) was left untouched and should
	// stay fully transparent, not show the red reference through.
	path2 := writeTempPNGPath(t)
	handler.Handle(protocol.Request{Command: "export", Args: []string{path2}})
	got := decodePNG(t, path2)
	_, _, _, a := got.At(1, 0).RGBA()
	if a != 0 {
		t.Fatalf("plain export (1,0) alpha = %d, want 0 (underlay excluded from normal export)", a)
	}
}

func TestHandlerExportDebugIncludesUnderlay(t *testing.T) {
	handler := newTestHandler(t, 2, 1)
	path := writeTestReferenceImage(t, 2, 1, color.RGBA{R: 255, A: 255})
	handler.Handle(protocol.Request{Command: "import_reference", Args: []string{path, "1.0"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#00ff00"}})

	debugPath := writeTempPNGPath(t)
	resp := handler.Handle(protocol.Request{Command: "export_debug", Args: []string{debugPath}})
	if resp != "ok" {
		t.Fatalf("export_debug = %q, want ok", resp)
	}
	got := decodePNG(t, debugPath)
	r, g, _, _ := got.At(0, 0).RGBA()
	if r>>8 != 0 || g>>8 != 255 {
		t.Fatalf("export_debug(0,0) = rgb(%d,%d,_), want opaque drawn green (underlay hidden underneath)", r>>8, g>>8)
	}
	r, _, _, a := got.At(1, 0).RGBA()
	if a == 0 || r>>8 != 255 {
		t.Fatalf("export_debug(1,0) = rgb(%d,_,_) alpha=%d, want visible red reference showing through", r>>8, a)
	}

	// Regular export at the same state must still exclude the underlay.
	plainPath := writeTempPNGPath(t)
	handler.Handle(protocol.Request{Command: "export", Args: []string{plainPath}})
	plainGot := decodePNG(t, plainPath)
	_, _, _, plainA := plainGot.At(1, 0).RGBA()
	if plainA != 0 {
		t.Fatalf("plain export(1,0) alpha = %d, want 0 (underlay still excluded)", plainA)
	}
}

func TestHandlerExportDebugWithoutImportMatchesPlainExport(t *testing.T) {
	handler := newTestHandler(t, 1, 1)
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#123456"}})

	debugPath := writeTempPNGPath(t)
	resp := handler.Handle(protocol.Request{Command: "export_debug", Args: []string{debugPath}})
	if resp != "ok" {
		t.Fatalf("export_debug without import = %q, want ok", resp)
	}
	got := decodePNG(t, debugPath)
	r, g, b, _ := got.At(0, 0).RGBA()
	if r>>8 != 0x12 || g>>8 != 0x34 || b>>8 != 0x56 {
		t.Fatalf("export_debug(0,0) = rgb(%02x,%02x,%02x), want 123456", r>>8, g>>8, b>>8)
	}
}

package daemon

import (
	"os"
	"strings"
	"testing"

	pxcolor "pxcli/internal/color"
	"pxcli/internal/protocol"
)

func TestHandlerLayerAddListSelect(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "layer_list", Args: nil})
	if resp != "ok base" {
		t.Fatalf("layer_list before any add = %q, want %q", resp, "ok base")
	}

	resp = handler.Handle(protocol.Request{Command: "layer_add", Args: []string{"sprite"}})
	if resp != "ok" {
		t.Fatalf("layer_add = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "layer_list", Args: nil})
	if resp != "ok base,sprite" {
		t.Fatalf("layer_list after add = %q, want %q", resp, "ok base,sprite")
	}

	resp = handler.Handle(protocol.Request{Command: "layer_select", Args: []string{"sprite"}})
	if resp != "ok" {
		t.Fatalf("layer_select = %q, want ok", resp)
	}
}

func TestHandlerLayerAddErrors(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "layer_add", Args: []string{"base"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("layer_add base = %q, want invalid_args (reserved)", resp)
	}

	handler.Handle(protocol.Request{Command: "layer_add", Args: []string{"sprite"}})
	resp = handler.Handle(protocol.Request{Command: "layer_add", Args: []string{"sprite"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("layer_add duplicate = %q, want invalid_args", resp)
	}
}

func TestHandlerLayerSelectUndefined(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	resp := handler.Handle(protocol.Request{Command: "layer_select", Args: []string{"missing"}})
	if !strings.HasPrefix(resp, "err invalid_layer ") {
		t.Fatalf("layer_select missing = %q, want invalid_layer", resp)
	}
}

func TestHandlerDrawingTargetsActiveLayer(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "layer_add", Args: []string{"sprite"}})
	handler.Handle(protocol.Request{Command: "layer_select", Args: []string{"sprite"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#00ff00"}})

	// Active layer (sprite) shows the new color.
	resp := handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"0", "0"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#00ff00"))
	if resp != want {
		t.Fatalf("get_pixel on sprite layer = %q, want %q", resp, want)
	}

	// Base layer is untouched by drawing on the sprite layer.
	handler.Handle(protocol.Request{Command: "layer_select", Args: []string{"base"}})
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"0", "0"}})
	want = "ok " + pxcolor.Format(mustParse(t, "#ff0000"))
	if resp != want {
		t.Fatalf("get_pixel on base layer after drawing on sprite layer = %q, want %q (base unaffected)", resp, want)
	}
}

func TestHandlerUndoRedoAreScopedToActiveLayer(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "layer_add", Args: []string{"sprite"}})
	handler.Handle(protocol.Request{Command: "layer_select", Args: []string{"sprite"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"1", "1", "#00ff00"}})

	resp := handler.Handle(protocol.Request{Command: "undo", Args: nil})
	if resp != "ok" {
		t.Fatalf("undo on sprite layer = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"1", "1"}})
	if resp != "ok #00000000" {
		t.Fatalf("get_pixel after undo on sprite layer = %q, want cleared", resp)
	}

	// Base layer's own history still has its edit; undoing on sprite must not touch it.
	handler.Handle(protocol.Request{Command: "layer_select", Args: []string{"base"}})
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"0", "0"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#ff0000"))
	if resp != want {
		t.Fatalf("get_pixel on base after sprite-layer undo = %q, want %q (base history independent)", resp, want)
	}
}

func TestHandlerLayerVisible(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "layer_add", Args: []string{"sprite"}})

	resp := handler.Handle(protocol.Request{Command: "layer_visible", Args: []string{"sprite", "false"}})
	if resp != "ok" {
		t.Fatalf("layer_visible = %q, want ok", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "layer_visible", Args: []string{"missing", "true"}})
	if !strings.HasPrefix(resp, "err invalid_layer ") {
		t.Fatalf("layer_visible missing = %q, want invalid_layer", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "layer_visible", Args: []string{"sprite", "not-a-bool"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("layer_visible bad bool = %q, want invalid_args", resp)
	}
}

func TestHandlerExportFlattensVisibleLayers(t *testing.T) {
	handler := newTestHandler(t, 2, 1)
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"1", "0", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "layer_add", Args: []string{"overlay"}})
	handler.Handle(protocol.Request{Command: "layer_select", Args: []string{"overlay"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#00ff00"}})

	path := writeTempPNGPath(t)
	resp := handler.Handle(protocol.Request{Command: "export", Args: []string{path}})
	if resp != "ok" {
		t.Fatalf("export = %q, want ok", resp)
	}

	got := decodePNG(t, path)
	r, g, b, _ := got.At(0, 0).RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 {
		t.Fatalf("exported (0,0) = rgb(%d,%d,%d), want overlay green to win (opaque, on top)", r>>8, g>>8, b>>8)
	}
	r, g, b, _ = got.At(1, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Fatalf("exported (1,0) = rgb(%d,%d,%d), want base red to show through (overlay transparent there)", r>>8, g>>8, b>>8)
	}
}

func TestHandlerExportSkipsHiddenLayers(t *testing.T) {
	handler := newTestHandler(t, 1, 1)
	handler.Handle(protocol.Request{Command: "layer_add", Args: []string{"overlay"}})
	handler.Handle(protocol.Request{Command: "layer_select", Args: []string{"overlay"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#00ff00"}})
	handler.Handle(protocol.Request{Command: "layer_visible", Args: []string{"overlay", "false"}})

	path := writeTempPNGPath(t)
	handler.Handle(protocol.Request{Command: "export", Args: []string{path}})

	got := decodePNG(t, path)
	r, g, b, a := got.At(0, 0).RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0 {
		t.Fatalf("exported hidden-overlay pixel = rgba(%d,%d,%d,%d), want fully transparent", r, g, b, a)
	}
}

func writeTempPNGPath(t *testing.T) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "layer-export-*.png")
	if err != nil {
		t.Fatalf("CreateTemp() error = %v", err)
	}
	path := f.Name()
	_ = f.Close()
	return path
}

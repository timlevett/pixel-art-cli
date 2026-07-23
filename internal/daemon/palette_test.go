package daemon

import (
	"image/color"
	"strings"
	"testing"

	"pxcli/internal/canvas"
	pxcolor "pxcli/internal/color"
	"pxcli/internal/history"
	"pxcli/internal/protocol"
)

func newTestHandler(t *testing.T, w, h int) *Handler {
	t.Helper()
	target, err := canvas.New(w, h)
	if err != nil {
		t.Fatalf("canvas.New() error = %v", err)
	}
	return NewHandler(history.New(target), nil)
}

func TestHandlerPaletteAddAndList(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"fire", "#ff0000", "#ffa500"}})
	if resp != "ok" {
		t.Fatalf("palette_add = %q, want ok", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "palette_list", Args: nil})
	if resp != "ok fire" {
		t.Fatalf("palette_list = %q, want %q", resp, "ok fire")
	}

	resp = handler.Handle(protocol.Request{Command: "palette_list", Args: []string{"fire"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#ff0000")) + "," + pxcolor.Format(mustParse(t, "#ffa500"))
	if resp != want {
		t.Fatalf("palette_list fire = %q, want %q", resp, want)
	}
}

func TestHandlerPaletteAddReplacesExisting(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"fire", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"fire", "#111111", "#222222"}})

	resp := handler.Handle(protocol.Request{Command: "palette_list", Args: []string{"fire"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#111111")) + "," + pxcolor.Format(mustParse(t, "#222222"))
	if resp != want {
		t.Fatalf("palette_list fire = %q, want %q (replace, not append)", resp, want)
	}
}

func TestHandlerPaletteAddErrors(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"fire"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("palette_add with no colors = %q, want invalid_args", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"fire", "not-a-color"}})
	if !strings.HasPrefix(resp, "err invalid_color ") {
		t.Fatalf("palette_add with bad color = %q, want invalid_color", resp)
	}
}

func TestHandlerPaletteListUndefined(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	resp := handler.Handle(protocol.Request{Command: "palette_list", Args: []string{"missing"}})
	if !strings.HasPrefix(resp, "err invalid_palette ") {
		t.Fatalf("palette_list missing = %q, want invalid_palette", resp)
	}
}

func TestHandlerPaletteUse(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"fire", "#ff0000", "#00ff00"}})

	resp := handler.Handle(protocol.Request{Command: "palette_use", Args: []string{"fire"}})
	if resp != "ok" {
		t.Fatalf("palette_use = %q, want ok", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "palette_use", Args: []string{"missing"}})
	if !strings.HasPrefix(resp, "err invalid_palette ") {
		t.Fatalf("palette_use missing = %q, want invalid_palette", resp)
	}
}

func TestHandlerSetPixelAcceptsNamedPaletteReference(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"fire", "#ff0000", "#00ff00"}})

	resp := handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"1", "1", "fire:1"}})
	if resp != "ok" {
		t.Fatalf("set_pixel with palette ref = %q, want ok", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"1", "1"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#00ff00"))
	if resp != want {
		t.Fatalf("get_pixel after palette-ref set_pixel = %q, want %q", resp, want)
	}
}

func TestHandlerSetPixelAcceptsActivePaletteShorthand(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"fire", "#ff0000", "#00ff00"}})
	handler.Handle(protocol.Request{Command: "palette_use", Args: []string{"fire"}})

	resp := handler.Handle(protocol.Request{Command: "fill_rect", Args: []string{"0", "0", "2", "2", "p:0"}})
	if resp != "ok" {
		t.Fatalf("fill_rect with p: shorthand = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"0", "0"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#ff0000"))
	if resp != want {
		t.Fatalf("get_pixel after p: fill_rect = %q, want %q", resp, want)
	}
}

func TestHandlerColorArgUndefinedPaletteSlotIsInvalidColor(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"fire", "#ff0000"}})

	resp := handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "fire:5"}})
	if !strings.HasPrefix(resp, "err invalid_color ") {
		t.Fatalf("set_pixel with out-of-range palette slot = %q, want invalid_color", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "line", Args: []string{"0", "0", "1", "1", "missing:0"}})
	if !strings.HasPrefix(resp, "err invalid_color ") {
		t.Fatalf("line with undefined palette = %q, want invalid_color", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "clear", Args: []string{"p:0"}})
	if !strings.HasPrefix(resp, "err invalid_color ") {
		t.Fatalf("clear with unset active palette shorthand = %q, want invalid_color", resp)
	}
}

func TestHandlerScriptSupportsPaletteReferences(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"fire", "#ff0000", "#00ff00"}})

	resp := handler.HandleScript([]string{"set_pixel 0 0 fire:0", "set_pixel 1 1 fire:1"})
	if resp != "ok" {
		t.Fatalf("HandleScript with palette refs = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"1", "1"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#00ff00"))
	if resp != want {
		t.Fatalf("get_pixel after script palette ref = %q, want %q", resp, want)
	}
}

func mustParse(t *testing.T, hex string) color.RGBA {
	t.Helper()
	value, err := pxcolor.Parse(hex)
	if err != nil {
		t.Fatalf("pxcolor.Parse(%q) error = %v", hex, err)
	}
	return value
}

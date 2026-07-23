package daemon

import (
	"strings"
	"testing"

	pxcolor "pxcli/internal/color"
	"pxcli/internal/protocol"
)

func TestHandlerCircleOutlineAndFilled(t *testing.T) {
	handler := newTestHandler(t, 9, 9)

	resp := handler.Handle(protocol.Request{Command: "circle", Args: []string{"4", "4", "3", "#ff0000"}})
	if resp != "ok" {
		t.Fatalf("circle = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"7", "4"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#ff0000"))
	if resp != want {
		t.Fatalf("get_pixel after circle outline = %q, want %q", resp, want)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"4", "4"}})
	if resp != "ok #00000000" {
		t.Fatalf("get_pixel center of outline circle = %q, want unset", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "circle", Args: []string{"4", "4", "3", "#00ff00", "fill"}})
	if resp != "ok" {
		t.Fatalf("filled circle = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"4", "4"}})
	want = "ok " + pxcolor.Format(mustParse(t, "#00ff00"))
	if resp != want {
		t.Fatalf("get_pixel center after filled circle = %q, want %q", resp, want)
	}
}

func TestHandlerCircleErrors(t *testing.T) {
	handler := newTestHandler(t, 8, 8)

	resp := handler.Handle(protocol.Request{Command: "circle", Args: []string{"4", "4", "0", "red"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("circle zero radius = %q, want invalid_args", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "circle", Args: []string{"6", "4", "3", "red"}})
	if !strings.HasPrefix(resp, "err out_of_bounds ") {
		t.Fatalf("circle out of bounds = %q, want out_of_bounds", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "circle", Args: []string{"4", "4", "3", "not-a-color"}})
	if !strings.HasPrefix(resp, "err invalid_color ") {
		t.Fatalf("circle bad color = %q, want invalid_color", resp)
	}
}

func TestHandlerEllipseOutlineAndFilled(t *testing.T) {
	handler := newTestHandler(t, 11, 9)

	resp := handler.Handle(protocol.Request{Command: "ellipse", Args: []string{"5", "4", "4", "3", "#0000ff"}})
	if resp != "ok" {
		t.Fatalf("ellipse = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"9", "4"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#0000ff"))
	if resp != want {
		t.Fatalf("get_pixel after ellipse outline = %q, want %q", resp, want)
	}

	resp = handler.Handle(protocol.Request{Command: "ellipse", Args: []string{"5", "4", "4", "3", "#ffff00", "fill"}})
	if resp != "ok" {
		t.Fatalf("filled ellipse = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"5", "4"}})
	want = "ok " + pxcolor.Format(mustParse(t, "#ffff00"))
	if resp != want {
		t.Fatalf("get_pixel center after filled ellipse = %q, want %q", resp, want)
	}
}

func TestHandlerEllipseErrors(t *testing.T) {
	handler := newTestHandler(t, 8, 8)

	resp := handler.Handle(protocol.Request{Command: "ellipse", Args: []string{"4", "4", "0", "2", "red"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("ellipse zero radius = %q, want invalid_args", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "ellipse", Args: []string{"6", "4", "3", "2", "red"}})
	if !strings.HasPrefix(resp, "err out_of_bounds ") {
		t.Fatalf("ellipse out of bounds = %q, want out_of_bounds", resp)
	}
}

func TestHandlerDitherFill(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "dither_fill", Args: []string{"0", "0", "4", "4", "#ff0000", "#0000ff"}})
	if resp != "ok" {
		t.Fatalf("dither_fill = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"0", "0"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#ff0000"))
	if resp != want {
		t.Fatalf("get_pixel(0,0) after checkerboard dither_fill = %q, want %q", resp, want)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"1", "0"}})
	want = "ok " + pxcolor.Format(mustParse(t, "#0000ff"))
	if resp != want {
		t.Fatalf("get_pixel(1,0) after checkerboard dither_fill = %q, want %q", resp, want)
	}

	resp = handler.Handle(protocol.Request{Command: "dither_fill", Args: []string{"0", "0", "4", "4", "#ff0000", "#0000ff", "vertical"}})
	if resp != "ok" {
		t.Fatalf("dither_fill vertical = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"1", "0"}})
	want = "ok " + pxcolor.Format(mustParse(t, "#0000ff"))
	if resp != want {
		t.Fatalf("get_pixel(1,0) after vertical dither_fill = %q, want %q", resp, want)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"1", "1"}})
	if resp != want {
		t.Fatalf("get_pixel(1,1) after vertical dither_fill = %q, want %q (column parity, not row)", resp, want)
	}
}

func TestHandlerDitherFillErrors(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "dither_fill", Args: []string{"0", "0", "0", "1", "red", "blue"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("dither_fill zero width = %q, want invalid_args", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "dither_fill", Args: []string{"0", "0", "2", "2", "red", "blue", "diagonal"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("dither_fill bad pattern = %q, want invalid_args", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "dither_fill", Args: []string{"3", "3", "3", "1", "red", "blue"}})
	if !strings.HasPrefix(resp, "err out_of_bounds ") {
		t.Fatalf("dither_fill out of bounds = %q, want out_of_bounds", resp)
	}
}

func TestHandlerShapesAcceptPaletteReferences(t *testing.T) {
	handler := newTestHandler(t, 9, 9)
	handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"sprite", "#ff0000", "#00ff00"}})

	resp := handler.Handle(protocol.Request{Command: "circle", Args: []string{"4", "4", "3", "sprite:1", "fill"}})
	if resp != "ok" {
		t.Fatalf("circle with palette ref = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"4", "4"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#00ff00"))
	if resp != want {
		t.Fatalf("get_pixel after circle with palette ref = %q, want %q", resp, want)
	}
}

func TestHandlerScriptSupportsShapes(t *testing.T) {
	handler := newTestHandler(t, 9, 9)

	resp := handler.HandleScript([]string{
		"circle 4 4 3 #ff0000 fill",
		"dither_fill 0 0 2 2 #00ff00 #0000ff",
	})
	if resp != "ok" {
		t.Fatalf("HandleScript with shapes = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"4", "4"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#ff0000"))
	if resp != want {
		t.Fatalf("get_pixel after script circle = %q, want %q", resp, want)
	}
}

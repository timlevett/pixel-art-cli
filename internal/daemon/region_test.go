package daemon

import (
	"strconv"
	"strings"
	"testing"

	pxcolor "pxcli/internal/color"
	"pxcli/internal/protocol"
)

func TestHandlerCopyPaste(t *testing.T) {
	handler := newTestHandler(t, 8, 8)
	handler.Handle(protocol.Request{Command: "fill_rect", Args: []string{"1", "1", "2", "2", "#ff0000"}})

	resp := handler.Handle(protocol.Request{Command: "copy", Args: []string{"1", "1", "2", "2"}})
	if resp != "ok" {
		t.Fatalf("copy = %q, want ok", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "paste", Args: []string{"5", "5"}})
	if resp != "ok" {
		t.Fatalf("paste = %q, want ok", resp)
	}

	for _, p := range [][2]int{{5, 5}, {6, 5}, {5, 6}, {6, 6}} {
		resp := handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{strconv.Itoa(p[0]), strconv.Itoa(p[1])}})
		want := "ok " + pxcolor.Format(mustParse(t, "#ff0000"))
		if resp != want {
			t.Fatalf("get_pixel(%d,%d) after paste = %q, want %q", p[0], p[1], resp, want)
		}
	}

	// Original region is unchanged.
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"1", "1"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#ff0000"))
	if resp != want {
		t.Fatalf("get_pixel(1,1) after paste elsewhere = %q, want %q (source unchanged)", resp, want)
	}
}

func TestHandlerCopyPasteNamedClipboard(t *testing.T) {
	handler := newTestHandler(t, 8, 8)
	handler.Handle(protocol.Request{Command: "fill_rect", Args: []string{"0", "0", "1", "1", "#00ff00"}})
	handler.Handle(protocol.Request{Command: "copy", Args: []string{"0", "0", "1", "1", "sprite"}})

	resp := handler.Handle(protocol.Request{Command: "paste", Args: []string{"3", "3", "sprite"}})
	if resp != "ok" {
		t.Fatalf("paste named clipboard = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"3", "3"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#00ff00"))
	if resp != want {
		t.Fatalf("get_pixel after named paste = %q, want %q", resp, want)
	}

	// Default clipboard is untouched by the named copy.
	resp = handler.Handle(protocol.Request{Command: "paste", Args: []string{"0", "5"}})
	if !strings.HasPrefix(resp, "err invalid_clipboard ") {
		t.Fatalf("paste from empty default clipboard = %q, want invalid_clipboard", resp)
	}
}

func TestHandlerCopyErrors(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "copy", Args: []string{"0", "0", "0", "1"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("copy zero width = %q, want invalid_args", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "copy", Args: []string{"3", "3", "2", "2"}})
	if !strings.HasPrefix(resp, "err out_of_bounds ") {
		t.Fatalf("copy out of bounds = %q, want out_of_bounds", resp)
	}
}

func TestHandlerPasteUndefinedClipboard(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	resp := handler.Handle(protocol.Request{Command: "paste", Args: []string{"0", "0"}})
	if !strings.HasPrefix(resp, "err invalid_clipboard ") {
		t.Fatalf("paste with empty default clipboard = %q, want invalid_clipboard", resp)
	}
}

func TestHandlerMove(t *testing.T) {
	handler := newTestHandler(t, 8, 8)
	handler.Handle(protocol.Request{Command: "fill_rect", Args: []string{"1", "1", "2", "2", "#0000ff"}})

	resp := handler.Handle(protocol.Request{Command: "move", Args: []string{"1", "1", "2", "2", "3", "3"}})
	if resp != "ok" {
		t.Fatalf("move = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"4", "4"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#0000ff"))
	if resp != want {
		t.Fatalf("get_pixel at destination after move = %q, want %q", resp, want)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"1", "1"}})
	if resp != "ok #00000000" {
		t.Fatalf("get_pixel at source after move = %q, want cleared", resp)
	}
}

func TestHandlerMoveErrors(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	resp := handler.Handle(protocol.Request{Command: "move", Args: []string{"0", "0", "2", "2", "3", "3"}})
	if !strings.HasPrefix(resp, "err out_of_bounds ") {
		t.Fatalf("move destination out of bounds = %q, want out_of_bounds", resp)
	}
}

func TestHandlerMirror(t *testing.T) {
	handler := newTestHandler(t, 4, 2)
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"1", "0", "#0000ff"}})

	resp := handler.Handle(protocol.Request{Command: "mirror", Args: []string{"0", "0", "2", "1", "horizontal"}})
	if resp != "ok" {
		t.Fatalf("mirror = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"0", "0"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#0000ff"))
	if resp != want {
		t.Fatalf("get_pixel(0,0) after horizontal mirror = %q, want %q", resp, want)
	}
}

func TestHandlerMirrorErrors(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	resp := handler.Handle(protocol.Request{Command: "mirror", Args: []string{"0", "0", "2", "2", "diagonal"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("mirror bad axis = %q, want invalid_args", resp)
	}
}

func TestHandlerRegionOpsIntegrateWithUndo(t *testing.T) {
	handler := newTestHandler(t, 8, 8)
	handler.Handle(protocol.Request{Command: "fill_rect", Args: []string{"1", "1", "2", "2", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "copy", Args: []string{"1", "1", "2", "2"}})
	handler.Handle(protocol.Request{Command: "paste", Args: []string{"5", "5"}})

	resp := handler.Handle(protocol.Request{Command: "undo", Args: nil})
	if resp != "ok" {
		t.Fatalf("undo after paste = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"5", "5"}})
	if resp != "ok #00000000" {
		t.Fatalf("get_pixel(5,5) after undoing paste = %q, want cleared", resp)
	}
}

func TestHandlerScriptSupportsPasteMoveMirror(t *testing.T) {
	handler := newTestHandler(t, 8, 8)
	handler.Handle(protocol.Request{Command: "fill_rect", Args: []string{"0", "0", "2", "2", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "copy", Args: []string{"0", "0", "2", "2"}})

	resp := handler.HandleScript([]string{
		"paste 4 4",
		"mirror 4 4 2 2 horizontal",
		"move 4 4 2 2 1 1",
	})
	if resp != "ok" {
		t.Fatalf("HandleScript with paste/mirror/move = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"5", "5"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#ff0000"))
	if resp != want {
		t.Fatalf("get_pixel after script paste/mirror/move = %q, want %q", resp, want)
	}
}


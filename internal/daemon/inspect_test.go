package daemon

import (
	"strings"
	"testing"

	pxcolor "pxcli/internal/color"
	"pxcli/internal/protocol"
)

func TestHandlerInspectWholeCanvas(t *testing.T) {
	handler := newTestHandler(t, 2, 2)
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"1", "1", "#0000ff"}})

	resp := handler.Handle(protocol.Request{Command: "inspect", Args: nil})
	row0 := pxcolor.Format(mustParse(t, "#ff0000")) + "," + "#00000000"
	row1 := "#00000000" + "," + pxcolor.Format(mustParse(t, "#0000ff"))
	want := "ok " + row0 + ";" + row1
	if resp != want {
		t.Fatalf("inspect whole canvas = %q, want %q", resp, want)
	}
}

func TestHandlerInspectRegion(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "fill_rect", Args: []string{"1", "1", "2", "2", "#00ff00"}})

	resp := handler.Handle(protocol.Request{Command: "inspect", Args: []string{"1", "1", "2", "2"}})
	green := pxcolor.Format(mustParse(t, "#00ff00"))
	want := "ok " + green + "," + green + ";" + green + "," + green
	if resp != want {
		t.Fatalf("inspect region = %q, want %q", resp, want)
	}
}

func TestHandlerInspectErrors(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "inspect", Args: []string{"0", "0", "1"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("inspect wrong arg count = %q, want invalid_args", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "inspect", Args: []string{"3", "3", "2", "2"}})
	if !strings.HasPrefix(resp, "err out_of_bounds ") {
		t.Fatalf("inspect out of bounds = %q, want out_of_bounds", resp)
	}
}

func TestHandlerInspectDoesNotAffectHistory(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "inspect", Args: nil})

	resp := handler.Handle(protocol.Request{Command: "undo", Args: nil})
	if !strings.HasPrefix(resp, "err no_history ") {
		t.Fatalf("undo after inspect = %q, want no_history (inspect should not touch canvas history)", resp)
	}
}

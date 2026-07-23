package daemon

import (
	"strings"
	"testing"

	"pxcli/internal/protocol"
)

func TestHandlerBlend(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "blend", Args: []string{"#000000", "#ffffff", "0.5"}})
	if resp != "ok #808080ff" {
		t.Fatalf("blend midpoint = %q, want %q", resp, "ok #808080ff")
	}

	resp = handler.Handle(protocol.Request{Command: "blend", Args: []string{"#000000", "#ffffff", "0"}})
	if resp != "ok #000000ff" {
		t.Fatalf("blend ratio=0 = %q, want %q", resp, "ok #000000ff")
	}

	resp = handler.Handle(protocol.Request{Command: "blend", Args: []string{"#000000", "#ffffff", "1"}})
	if resp != "ok #ffffffff" {
		t.Fatalf("blend ratio=1 = %q, want %q", resp, "ok #ffffffff")
	}
}

func TestHandlerBlendAcceptsPaletteReferences(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "palette_add", Args: []string{"sprite", "#000000", "#ffffff"}})

	resp := handler.Handle(protocol.Request{Command: "blend", Args: []string{"sprite:0", "sprite:1", "0.5"}})
	if resp != "ok #808080ff" {
		t.Fatalf("blend with palette refs = %q, want %q", resp, "ok #808080ff")
	}
}

func TestHandlerBlendDoesNotMutateCanvasOrHistory(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "blend", Args: []string{"#000000", "#ffffff", "0.5"}})

	resp := handler.Handle(protocol.Request{Command: "undo", Args: nil})
	if !strings.HasPrefix(resp, "err no_history ") {
		t.Fatalf("undo after blend = %q, want no_history (blend should not touch canvas history)", resp)
	}
}

func TestHandlerBlendErrors(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "blend", Args: []string{"#000000", "#ffffff", "1.5"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("blend ratio out of range = %q, want invalid_args", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "blend", Args: []string{"#000000", "#ffffff", "not-a-number"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("blend non-numeric ratio = %q, want invalid_args", resp)
	}

	resp = handler.Handle(protocol.Request{Command: "blend", Args: []string{"not-a-color", "#ffffff", "0.5"}})
	if !strings.HasPrefix(resp, "err invalid_color ") {
		t.Fatalf("blend bad color = %q, want invalid_color", resp)
	}
}

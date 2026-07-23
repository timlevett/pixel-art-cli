package daemon

import (
	"strings"
	"testing"

	pxcolor "pxcli/internal/color"
	"pxcli/internal/protocol"
)

func TestHandlerFrameAddListSelect(t *testing.T) {
	handler := newTestHandler(t, 4, 4)

	resp := handler.Handle(protocol.Request{Command: "frame_list", Args: nil})
	if resp != "ok 0" {
		t.Fatalf("frame_list before any add = %q, want %q", resp, "ok 0")
	}

	resp = handler.Handle(protocol.Request{Command: "frame_add", Args: nil})
	if resp != "ok 1" {
		t.Fatalf("frame_add = %q, want ok 1", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "frame_list", Args: nil})
	if resp != "ok 0,1" {
		t.Fatalf("frame_list after add = %q, want %q", resp, "ok 0,1")
	}

	resp = handler.Handle(protocol.Request{Command: "frame_select", Args: []string{"1"}})
	if resp != "ok" {
		t.Fatalf("frame_select = %q, want ok", resp)
	}
}

func TestHandlerFrameSelectOutOfRange(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	resp := handler.Handle(protocol.Request{Command: "frame_select", Args: []string{"1"}})
	if !strings.HasPrefix(resp, "err invalid_frame ") {
		t.Fatalf("frame_select out of range = %q, want invalid_frame", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "frame_select", Args: []string{"not-a-number"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("frame_select non-numeric = %q, want invalid_args", resp)
	}
}

func TestHandlerDrawingTargetsActiveFrame(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "frame_add", Args: nil})
	handler.Handle(protocol.Request{Command: "frame_select", Args: []string{"1"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#00ff00"}})

	resp := handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"0", "0"}})
	want := "ok " + pxcolor.Format(mustParse(t, "#00ff00"))
	if resp != want {
		t.Fatalf("get_pixel on frame 1 = %q, want %q", resp, want)
	}

	handler.Handle(protocol.Request{Command: "frame_select", Args: []string{"0"}})
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"0", "0"}})
	want = "ok " + pxcolor.Format(mustParse(t, "#ff0000"))
	if resp != want {
		t.Fatalf("get_pixel on frame 0 after drawing on frame 1 = %q, want %q (frame 0 unaffected)", resp, want)
	}
}

func TestHandlerFramesHaveIndependentLayersAndUndo(t *testing.T) {
	handler := newTestHandler(t, 4, 4)
	handler.Handle(protocol.Request{Command: "frame_add", Args: nil})
	handler.Handle(protocol.Request{Command: "frame_select", Args: []string{"1"}})

	resp := handler.Handle(protocol.Request{Command: "layer_list", Args: nil})
	if resp != "ok base" {
		t.Fatalf("layer_list on fresh frame 1 = %q, want %q (own independent layer stack)", resp, "ok base")
	}

	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"1", "1", "#0000ff"}})
	resp = handler.Handle(protocol.Request{Command: "undo", Args: nil})
	if resp != "ok" {
		t.Fatalf("undo on frame 1 = %q, want ok", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "get_pixel", Args: []string{"1", "1"}})
	if resp != "ok #00000000" {
		t.Fatalf("get_pixel after undo on frame 1 = %q, want cleared", resp)
	}
}

func TestHandlerFrameGhostBlendsDimmedTargetUnderActive(t *testing.T) {
	handler := newTestHandler(t, 2, 1)
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"1", "0", "#0000ff"}})
	handler.Handle(protocol.Request{Command: "frame_add", Args: nil})
	handler.Handle(protocol.Request{Command: "frame_select", Args: []string{"1"}})
	handler.Handle(protocol.Request{Command: "set_pixel", Args: []string{"0", "0", "#00ff00"}})

	resp := handler.Handle(protocol.Request{Command: "frame_ghost", Args: []string{"0", "0.5"}})
	if !strings.HasPrefix(resp, "ok ") {
		t.Fatalf("frame_ghost = %q, want ok response", resp)
	}
	rows := strings.Split(strings.TrimPrefix(resp, "ok "), ";")
	if len(rows) != 1 {
		t.Fatalf("frame_ghost rows = %d, want 1", len(rows))
	}
	cells := strings.Split(rows[0], ",")
	if len(cells) != 2 {
		t.Fatalf("frame_ghost cells = %d, want 2", len(cells))
	}
	if cells[0] != pxcolor.Format(mustParse(t, "#00ff00")) {
		t.Fatalf("frame_ghost(0,0) = %q, want opaque active green (target hidden underneath)", cells[0])
	}
	if cells[1] == pxcolor.Format(mustParse(t, "#0000ffff")) || cells[1] == pxcolor.Format(mustParse(t, "#00000000")) {
		t.Fatalf("frame_ghost(1,0) = %q, want dimmed blue (neither fully opaque nor fully transparent)", cells[1])
	}
}

func TestHandlerFrameGhostErrors(t *testing.T) {
	handler := newTestHandler(t, 2, 2)
	resp := handler.Handle(protocol.Request{Command: "frame_ghost", Args: []string{"5"}})
	if !strings.HasPrefix(resp, "err invalid_frame ") {
		t.Fatalf("frame_ghost undefined target = %q, want invalid_frame", resp)
	}
	resp = handler.Handle(protocol.Request{Command: "frame_ghost", Args: []string{"0", "1.5"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("frame_ghost opacity out of range = %q, want invalid_args", resp)
	}
}

func TestHandlerExportSheetTilesFrames(t *testing.T) {
	handler := newTestHandler(t, 2, 2)
	handler.Handle(protocol.Request{Command: "fill_rect", Args: []string{"0", "0", "2", "2", "#ff0000"}})
	handler.Handle(protocol.Request{Command: "frame_add", Args: nil})
	handler.Handle(protocol.Request{Command: "frame_select", Args: []string{"1"}})
	handler.Handle(protocol.Request{Command: "fill_rect", Args: []string{"0", "0", "2", "2", "#00ff00"}})

	path := writeTempPNGPath(t)
	resp := handler.Handle(protocol.Request{Command: "export_sheet", Args: []string{path}})
	if resp != "ok" {
		t.Fatalf("export_sheet = %q, want ok", resp)
	}

	got := decodePNG(t, path)
	if got.Bounds().Dx() != 4 || got.Bounds().Dy() != 2 {
		t.Fatalf("export_sheet size = %dx%d, want 4x2 (2 frames side by side, default cols)", got.Bounds().Dx(), got.Bounds().Dy())
	}
	r, g, _, _ := got.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 {
		t.Fatalf("export_sheet frame0 tile = rgb(%d,%d,_), want red", r>>8, g>>8)
	}
	r, g, _, _ = got.At(2, 0).RGBA()
	if r>>8 != 0 || g>>8 != 255 {
		t.Fatalf("export_sheet frame1 tile = rgb(%d,%d,_), want green", r>>8, g>>8)
	}
}

func TestHandlerExportSheetRespectsColsAndPads(t *testing.T) {
	handler := newTestHandler(t, 1, 1)
	handler.Handle(protocol.Request{Command: "frame_add", Args: nil})
	handler.Handle(protocol.Request{Command: "frame_add", Args: nil})

	path := writeTempPNGPath(t)
	resp := handler.Handle(protocol.Request{Command: "export_sheet", Args: []string{path, "2"}})
	if resp != "ok" {
		t.Fatalf("export_sheet with cols = %q, want ok", resp)
	}
	got := decodePNG(t, path)
	if got.Bounds().Dx() != 2 || got.Bounds().Dy() != 2 {
		t.Fatalf("export_sheet cols=2 size = %dx%d, want 2x2 (3 frames wrapped, padded)", got.Bounds().Dx(), got.Bounds().Dy())
	}
}

func TestHandlerExportSheetErrors(t *testing.T) {
	handler := newTestHandler(t, 2, 2)
	resp := handler.Handle(protocol.Request{Command: "export_sheet", Args: []string{"/tmp/out.png", "0"}})
	if !strings.HasPrefix(resp, "err invalid_args ") {
		t.Fatalf("export_sheet cols=0 = %q, want invalid_args", resp)
	}
}

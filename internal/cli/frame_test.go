package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"pxcli/internal/client"
)

func TestFrameCommands_FormatRequests(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantRequest string
	}{
		{name: "add", args: []string{"frame", "add"}, wantRequest: "frame_add"},
		{name: "list", args: []string{"frame", "list"}, wantRequest: "frame_list"},
		{name: "select", args: []string{"frame", "select", "1"}, wantRequest: "frame_select 1"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			stub := &stubClient{response: client.Response{Raw: "ok"}}
			restore := drawNewClient
			drawNewClient = func(socketPath string) (requestSender, error) {
				return stub, nil
			}
			t.Cleanup(func() {
				drawNewClient = restore
			})

			buf := &bytes.Buffer{}
			cmd := NewRootCmd("dev")
			cmd.SetOut(buf)
			cmd.SetErr(io.Discard)
			cmd.SetArgs(tt.args)

			if err := cmd.Execute(); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(stub.requests) != 1 || stub.requests[0] != tt.wantRequest {
				t.Fatalf("expected request %q, got %v", tt.wantRequest, stub.requests)
			}
		})
	}
}

func TestFrameSelectCmd_NonIntegerIndex(t *testing.T) {
	called := false
	restore := drawNewClient
	drawNewClient = func(socketPath string) (requestSender, error) {
		called = true
		return nil, fmt.Errorf("client should not be created")
	}
	t.Cleanup(func() {
		drawNewClient = restore
	})

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"frame", "select", "one"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for non-integer index")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created for invalid args")
	}
}

func TestFrameGhostCmd_RendersGridLikeInspect(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "ok #ff0000ff,#00000000"}}
	restore := drawNewClient
	drawNewClient = func(socketPath string) (requestSender, error) {
		return stub, nil
	}
	t.Cleanup(func() {
		drawNewClient = restore
	})

	buf := &bytes.Buffer{}
	cmd := NewRootCmd("dev")
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"frame", "ghost", "0"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.requests) != 1 || stub.requests[0] != "frame_ghost 0" {
		t.Fatalf("unexpected requests: %v", stub.requests)
	}
	want := "#ff0000ff #00000000\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestFrameGhostCmd_WithOpacityFormatsRequest(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "ok #ff0000ff"}}
	restore := drawNewClient
	drawNewClient = func(socketPath string) (requestSender, error) {
		return stub, nil
	}
	t.Cleanup(func() {
		drawNewClient = restore
	})

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"frame", "ghost", "0", "0.5"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.requests) != 1 || stub.requests[0] != "frame_ghost 0 0.5" {
		t.Fatalf("unexpected requests: %v", stub.requests)
	}
}

func TestFrameAddCmd_DaemonResponsePassthrough(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "ok 1"}}
	restore := drawNewClient
	drawNewClient = func(socketPath string) (requestSender, error) {
		return stub, nil
	}
	t.Cleanup(func() {
		drawNewClient = restore
	})

	buf := &bytes.Buffer{}
	cmd := NewRootCmd("dev")
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"frame", "add"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "ok 1" {
		t.Fatalf("expected new frame index passthrough, got %q", buf.String())
	}
}

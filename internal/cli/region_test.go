package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"pxcli/internal/client"
)

func TestRegionCommands_FormatRequests(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantRequest string
	}{
		{
			name:        "copy_default",
			args:        []string{"copy", "1", "1", "2", "2"},
			wantRequest: "copy 1 1 2 2",
		},
		{
			name:        "copy_named",
			args:        []string{"copy", "1", "1", "2", "2", "sprite"},
			wantRequest: "copy 1 1 2 2 sprite",
		},
		{
			name:        "paste_default",
			args:        []string{"paste", "5", "5"},
			wantRequest: "paste 5 5",
		},
		{
			name:        "paste_named",
			args:        []string{"paste", "5", "5", "sprite"},
			wantRequest: "paste 5 5 sprite",
		},
		{
			name:        "move",
			args:        []string{"move", "1", "1", "2", "2", "3", "3"},
			wantRequest: "move 1 1 2 2 3 3",
		},
		{
			name:        "mirror_horizontal",
			args:        []string{"mirror", "0", "0", "2", "2", "horizontal"},
			wantRequest: "mirror 0 0 2 2 horizontal",
		},
		{
			name:        "mirror_vertical",
			args:        []string{"mirror", "0", "0", "2", "2", "vertical"},
			wantRequest: "mirror 0 0 2 2 vertical",
		},
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
			if len(stub.requests) != 1 {
				t.Fatalf("expected 1 request, got %d", len(stub.requests))
			}
			if stub.requests[0] != tt.wantRequest {
				t.Fatalf("expected request %q, got %q", tt.wantRequest, stub.requests[0])
			}
			if strings.TrimSpace(buf.String()) != "ok" {
				t.Fatalf("expected ok output, got %q", buf.String())
			}
		})
	}
}

func TestCopyCmd_InvalidSize(t *testing.T) {
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
	cmd.SetArgs([]string{"copy", "0", "0", "-1", "2"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for invalid size")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created for invalid args")
	}
}

func TestMoveCmd_WrongArgCount(t *testing.T) {
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
	cmd.SetArgs([]string{"move", "1", "1", "2", "2", "3"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for wrong arg count")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created for invalid args")
	}
}

func TestMirrorCmd_InvalidAxis(t *testing.T) {
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
	cmd.SetArgs([]string{"mirror", "0", "0", "2", "2", "diagonal"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for invalid axis")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created for invalid args")
	}
}

func TestPasteCmd_DaemonErrorPassthrough(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "err invalid_clipboard clipboard \"default\" is empty"}}
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
	cmd.SetArgs([]string{"paste", "0", "0"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "invalid_clipboard") {
		t.Fatalf("expected daemon error passthrough, got %q", buf.String())
	}
}

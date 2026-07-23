package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"pxcli/internal/client"
)

func TestPaletteCommands_FormatRequests(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantRequest string
	}{
		{
			name:        "add",
			args:        []string{"palette", "add", "fire", "#ff0000", "#ffa500"},
			wantRequest: "palette_add fire #ff0000 #ffa500",
		},
		{
			name:        "list_all",
			args:        []string{"palette", "list"},
			wantRequest: "palette_list",
		},
		{
			name:        "list_named",
			args:        []string{"palette", "list", "fire"},
			wantRequest: "palette_list fire",
		},
		{
			name:        "use",
			args:        []string{"palette", "use", "fire"},
			wantRequest: "palette_use fire",
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

func TestPaletteAddCmd_RequiresNameAndColor(t *testing.T) {
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
	cmd.SetArgs([]string{"palette", "add", "fire"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for missing colors")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created for invalid args")
	}
}

func TestPaletteListCmd_TooManyArgs(t *testing.T) {
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
	cmd.SetArgs([]string{"palette", "list", "fire", "extra"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for too many args")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created for invalid args")
	}
}

func TestPaletteUseCmd_DaemonErrorPassthrough(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "err invalid_palette \"missing\" is not defined"}}
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
	cmd.SetArgs([]string{"palette", "use", "missing"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "invalid_palette") {
		t.Fatalf("expected daemon error passthrough, got %q", buf.String())
	}
}

package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"pxcli/internal/client"
)

func TestShapeCommands_FormatRequests(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantRequest string
	}{
		{
			name:        "circle_outline",
			args:        []string{"circle", "4", "4", "3", "red"},
			wantRequest: "circle 4 4 3 red",
		},
		{
			name:        "circle_filled",
			args:        []string{"circle", "4", "4", "3", "red", "fill"},
			wantRequest: "circle 4 4 3 red fill",
		},
		{
			name:        "ellipse_outline",
			args:        []string{"ellipse", "5", "4", "4", "3", "blue"},
			wantRequest: "ellipse 5 4 4 3 blue",
		},
		{
			name:        "ellipse_filled",
			args:        []string{"ellipse", "5", "4", "4", "3", "blue", "fill"},
			wantRequest: "ellipse 5 4 4 3 blue fill",
		},
		{
			name:        "dither_fill_default",
			args:        []string{"dither_fill", "0", "0", "4", "4", "#ff0000", "#0000ff"},
			wantRequest: "dither_fill 0 0 4 4 #ff0000 #0000ff",
		},
		{
			name:        "dither_fill_pattern",
			args:        []string{"dither_fill", "0", "0", "4", "4", "#ff0000", "#0000ff", "vertical"},
			wantRequest: "dither_fill 0 0 4 4 #ff0000 #0000ff vertical",
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

func TestCircleCmd_InvalidRadius(t *testing.T) {
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
	cmd.SetArgs([]string{"circle", "4", "4", "0", "red"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for zero radius")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created for invalid args")
	}
}

func TestEllipseCmd_WrongArgCount(t *testing.T) {
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
	cmd.SetArgs([]string{"ellipse", "5", "4", "4", "blue"})

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

func TestDitherFillCmd_InvalidSize(t *testing.T) {
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
	cmd.SetArgs([]string{"dither_fill", "0", "0", "-1", "2", "red", "blue"})

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

package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"pxcli/internal/client"
)

func TestLayerCommands_FormatRequests(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		wantRequest string
	}{
		{name: "add", args: []string{"layer", "add", "sprite"}, wantRequest: "layer_add sprite"},
		{name: "list", args: []string{"layer", "list"}, wantRequest: "layer_list"},
		{name: "select", args: []string{"layer", "select", "sprite"}, wantRequest: "layer_select sprite"},
		{name: "visible_true", args: []string{"layer", "visible", "sprite", "true"}, wantRequest: "layer_visible sprite true"},
		{name: "visible_false", args: []string{"layer", "visible", "sprite", "false"}, wantRequest: "layer_visible sprite false"},
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
			if strings.TrimSpace(buf.String()) != "ok" {
				t.Fatalf("expected ok output, got %q", buf.String())
			}
		})
	}
}

func TestLayerVisibleCmd_InvalidBool(t *testing.T) {
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
	cmd.SetArgs([]string{"layer", "visible", "sprite", "yes"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for invalid bool")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created for invalid args")
	}
}

func TestLayerAddCmd_DaemonErrorPassthrough(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "err invalid_args layer \"base\" is reserved"}}
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
	cmd.SetArgs([]string{"layer", "add", "base"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "reserved") {
		t.Fatalf("expected daemon error passthrough, got %q", buf.String())
	}
}

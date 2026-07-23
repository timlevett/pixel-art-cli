package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"pxcli/internal/client"
)

func TestInspectCmd_WholeCanvasRendersGrid(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "ok #ff0000ff,#00000000;#00000000,#0000ffff"}}
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
	cmd.SetArgs([]string{"inspect"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.requests) != 1 || stub.requests[0] != "inspect" {
		t.Fatalf("unexpected requests: %v", stub.requests)
	}
	want := "#ff0000ff #00000000\n#00000000 #0000ffff\n"
	if buf.String() != want {
		t.Fatalf("output = %q, want %q", buf.String(), want)
	}
}

func TestInspectCmd_RegionFormatsRequest(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "ok #ff0000ff"}}
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
	cmd.SetArgs([]string{"inspect", "1", "1", "2", "2"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.requests) != 1 || stub.requests[0] != "inspect 1 1 2 2" {
		t.Fatalf("unexpected requests: %v", stub.requests)
	}
}

func TestInspectCmd_DaemonErrorPassthrough(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "err out_of_bounds region (3,3) size 2x2 outside canvas"}}
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
	cmd.SetArgs([]string{"inspect", "3", "3", "2", "2"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "err out_of_bounds") {
		t.Fatalf("expected daemon error passthrough, got %q", buf.String())
	}
}

func TestInspectCmd_WrongArgCount(t *testing.T) {
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
	cmd.SetArgs([]string{"inspect", "1", "2", "3"})

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

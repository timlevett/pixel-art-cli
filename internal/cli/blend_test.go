package cli

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"testing"

	"pxcli/internal/client"
)

func TestBlendCmd_FormatsRequest(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "ok #808080ff"}}
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
	cmd.SetArgs([]string{"blend", "#000000", "#ffffff", "0.5"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.requests) != 1 || stub.requests[0] != "blend #000000 #ffffff 0.5" {
		t.Fatalf("unexpected requests: %v", stub.requests)
	}
	if strings.TrimSpace(buf.String()) != "ok #808080ff" {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestBlendCmd_WrongArgCount(t *testing.T) {
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
	cmd.SetArgs([]string{"blend", "#000000", "#ffffff"})

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

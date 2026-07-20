package cli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"pxcli/internal/client"
)

type stubScriptClient struct {
	lines    []string
	response client.Response
	err      error
}

func (s *stubScriptClient) SendScript(lines []string) (client.Response, error) {
	s.lines = lines
	return s.response, s.err
}

func TestScriptCmd_ReadsFileAndSendsLines(t *testing.T) {
	dir := t.TempDir()
	scriptPath := dir + "/art.pxs"
	content := "# a comment\n\nset_pixel 0 0 red\nfill_rect 1 1 2 2 blue\n"
	if err := os.WriteFile(scriptPath, []byte(content), 0o644); err != nil {
		t.Fatalf("unexpected error writing script file: %v", err)
	}

	stub := &stubScriptClient{response: client.Response{Raw: "ok"}}
	restore := scriptNewClient
	scriptNewClient = func(socketPath string) (scriptSender, error) {
		return stub, nil
	}
	t.Cleanup(func() {
		scriptNewClient = restore
	})

	buf := &bytes.Buffer{}
	cmd := NewRootCmd("dev")
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"script", scriptPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	wantLines := []string{"# a comment", "", "set_pixel 0 0 red", "fill_rect 1 1 2 2 blue"}
	if len(stub.lines) != len(wantLines) {
		t.Fatalf("expected lines %v, got %v", wantLines, stub.lines)
	}
	for i, want := range wantLines {
		if stub.lines[i] != want {
			t.Fatalf("expected line %d to be %q, got %q", i, want, stub.lines[i])
		}
	}
	if strings.TrimSpace(buf.String()) != "ok" {
		t.Fatalf("expected ok output, got %q", buf.String())
	}
}

func TestScriptCmd_ReadsStdinWhenNoFileGiven(t *testing.T) {
	stub := &stubScriptClient{response: client.Response{Raw: "ok"}}
	restore := scriptNewClient
	scriptNewClient = func(socketPath string) (scriptSender, error) {
		return stub, nil
	}
	t.Cleanup(func() {
		scriptNewClient = restore
	})

	buf := &bytes.Buffer{}
	cmd := NewRootCmd("dev")
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetIn(strings.NewReader("set_pixel 0 0 red\n"))
	cmd.SetArgs([]string{"script"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stub.lines) != 1 || stub.lines[0] != "set_pixel 0 0 red" {
		t.Fatalf("expected stdin line to be forwarded, got %v", stub.lines)
	}
}

func TestScriptCmd_PropagatesDaemonError(t *testing.T) {
	stub := &stubScriptClient{err: client.Error{Code: "invalid_args", Message: "line 2: x must be an integer"}}
	restore := scriptNewClient
	scriptNewClient = func(socketPath string) (scriptSender, error) {
		return stub, nil
	}
	t.Cleanup(func() {
		scriptNewClient = restore
	})

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetIn(strings.NewReader("set_pixel x 0 red\n"))
	cmd.SetArgs([]string{"script"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "err invalid_args") || !strings.Contains(err.Error(), "line 2") {
		t.Fatalf("expected invalid_args error with line number, got %q", err.Error())
	}
}

func TestScriptCmd_MissingFile(t *testing.T) {
	called := false
	restore := scriptNewClient
	scriptNewClient = func(socketPath string) (scriptSender, error) {
		called = true
		return nil, fmt.Errorf("client should not be created")
	}
	t.Cleanup(func() {
		scriptNewClient = restore
	})

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"script", "/nonexistent/path/art.pxs"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for missing file")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
	if called {
		t.Fatalf("expected client not to be created when file is missing")
	}
}

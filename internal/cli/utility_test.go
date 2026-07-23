package cli

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pxcli/internal/client"
	"pxcli/internal/testutil"
)

func TestGetPixelCmd_PrintsColor(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	pidPath := filepath.Join(dir, "pxcli.pid")

	restorePID := daemonPIDPath
	daemonPIDPath = pidPath
	t.Cleanup(func() {
		daemonPIDPath = restorePID
	})
	t.Cleanup(func() {
		if _, err := os.Stat(socketPath); err == nil {
			_, _ = sendRequest(socketPath, "stop\n")
		}
	})

	daemonCmd := NewRootCmd("dev")
	daemonCmd.SetOut(io.Discard)
	daemonCmd.SetErr(io.Discard)
	daemonCmd.SetArgs([]string{"daemon", "--headless", "--size", "4x4", "--socket", socketPath})

	errCh := make(chan error, 1)
	go func() {
		errCh <- daemonCmd.Execute()
	}()

	waitForPath(t, socketPath)

	cli, err := client.New(socketPath)
	if err != nil {
		t.Fatalf("unexpected client error: %v", err)
	}
	if _, err := cli.Send("set_pixel 0 0 #00ff00"); err != nil {
		t.Fatalf("unexpected set_pixel error: %v", err)
	}

	buf := &bytes.Buffer{}
	getCmd := NewRootCmd("dev")
	getCmd.SetOut(buf)
	getCmd.SetErr(io.Discard)
	getCmd.SetArgs([]string{"--socket", socketPath, "get_pixel", "0", "0"})

	if err := getCmd.Execute(); err != nil {
		t.Fatalf("unexpected get_pixel error: %v", err)
	}
	if strings.TrimSpace(buf.String()) != "ok #00ff00ff" {
		t.Fatalf("expected ok #00ff00ff output, got %q", buf.String())
	}

	if _, err := cli.Send("stop"); err != nil {
		t.Fatalf("failed to stop daemon: %v", err)
	}

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("unexpected daemon error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for daemon to stop")
	}
}

func TestExportCmd_ResolvesAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected getwd error: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("unexpected chdir error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	expected, err := filepath.Abs("out.png")
	if err != nil {
		t.Fatalf("unexpected abs error: %v", err)
	}

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
	cmd.SetArgs([]string{"export", "out.png"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected export error: %v", err)
	}
	if len(stub.requests) != 1 {
		t.Fatalf("expected 1 request, got %d", len(stub.requests))
	}
	if stub.requests[0] != "export "+expected {
		t.Fatalf("expected export request %q, got %q", "export "+expected, stub.requests[0])
	}
	if strings.TrimSpace(buf.String()) != "ok" {
		t.Fatalf("expected ok output, got %q", buf.String())
	}
}

func TestExportSheetCmd_ResolvesAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected getwd error: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("unexpected chdir error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	expected, err := filepath.Abs("sheet.png")
	if err != nil {
		t.Fatalf("unexpected abs error: %v", err)
	}

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
	cmd.SetArgs([]string{"export_sheet", "sheet.png"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected export_sheet error: %v", err)
	}
	if len(stub.requests) != 1 || stub.requests[0] != "export_sheet "+expected {
		t.Fatalf("expected export_sheet request %q, got %v", "export_sheet "+expected, stub.requests)
	}
}

func TestExportSheetCmd_ColsFlagFormatsRequest(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "ok"}}
	restore := drawNewClient
	drawNewClient = func(socketPath string) (requestSender, error) {
		return stub, nil
	}
	t.Cleanup(func() {
		drawNewClient = restore
	})

	dir := t.TempDir()
	path := filepath.Join(dir, "sheet.png")

	cmd := NewRootCmd("dev")
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"export_sheet", path, "--cols", "3"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected export_sheet error: %v", err)
	}
	if len(stub.requests) != 1 || stub.requests[0] != "export_sheet "+path+" 3" {
		t.Fatalf("expected cols in request, got %v", stub.requests)
	}
}

func TestExportSheetCmd_PropagatesDaemonError(t *testing.T) {
	stub := &stubClient{err: client.Error{Code: "invalid_args", Message: "cols must be a positive integer"}}
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
	cmd.SetArgs([]string{"export_sheet", "sheet.png"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error propagated from daemon")
	}
	if !strings.Contains(err.Error(), "err invalid_args") {
		t.Fatalf("expected invalid_args error, got %q", err.Error())
	}
}

func TestImportReferenceCmd_DefaultOpacityOmittedFromRequest(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "ok"}}
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
	cmd.SetArgs([]string{"import_reference", "ref.png"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.requests) != 1 || stub.requests[0] != "import_reference ref.png" {
		t.Fatalf("expected request without opacity, got %v", stub.requests)
	}
}

func TestImportReferenceCmd_ExplicitOpacityFormatsRequest(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "ok"}}
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
	cmd.SetArgs([]string{"import_reference", "ref.png", "--opacity", "0.8"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stub.requests) != 1 || stub.requests[0] != "import_reference ref.png 0.8" {
		t.Fatalf("expected request with opacity, got %v", stub.requests)
	}
}

func TestImportReferenceCmd_DaemonErrorPassthrough(t *testing.T) {
	stub := &stubClient{response: client.Response{Raw: "err io unable to decode image: unknown format"}}
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
	cmd.SetArgs([]string{"import_reference", "ref.png"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "unable to decode image") {
		t.Fatalf("expected daemon error passthrough, got %q", buf.String())
	}
}

func TestExportDebugCmd_ResolvesAbsolutePath(t *testing.T) {
	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("unexpected getwd error: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("unexpected chdir error: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(wd)
	})

	expected, err := filepath.Abs("debug.png")
	if err != nil {
		t.Fatalf("unexpected abs error: %v", err)
	}

	stub := &stubClient{response: client.Response{Raw: "ok"}}
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
	cmd.SetArgs([]string{"export_debug", "debug.png"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("unexpected export_debug error: %v", err)
	}
	if len(stub.requests) != 1 || stub.requests[0] != "export_debug "+expected {
		t.Fatalf("expected export_debug request %q, got %v", "export_debug "+expected, stub.requests)
	}
}

func TestExportCmd_PropagatesIOError(t *testing.T) {
	stub := &stubClient{err: client.Error{Code: "io", Message: "permission denied"}}
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
	cmd.SetArgs([]string{"export", "out.png"})

	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for io failure")
	}
	if !strings.Contains(err.Error(), "err io") {
		t.Fatalf("expected err io message, got %q", err.Error())
	}
}

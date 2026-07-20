package daemon

import (
	"bufio"
	"errors"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pxcli/internal/protocol"
	"pxcli/internal/testutil"
)

type stubHandler struct {
	response string
}

func (s stubHandler) Handle(request protocol.Request) string {
	return s.response
}

type stubScriptHandler struct {
	stubHandler
	receivedLines []string
	scriptResp    string
}

func (s *stubScriptHandler) HandleScript(lines []string) string {
	s.receivedLines = lines
	return s.scriptResp
}

func TestServerRespondsSingleLine(t *testing.T) {
	socketPath := filepath.Join(testutil.TempDir(t), "pxcli.sock")
	server, err := NewServer(socketPath, stubHandler{response: "ok"})
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}
	done := startServer(t, server)
	t.Cleanup(func() {
		stopServer(t, server, done)
	})

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("unexpected error connecting to socket: %v", err)
	}
	defer conn.Close()

	if _, err := io.WriteString(conn, "clear\n"); err != nil {
		t.Fatalf("unexpected error writing request: %v", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("unexpected error reading response: %v", err)
	}
	if line != "ok\n" {
		t.Fatalf("expected response %q, got %q", "ok\n", line)
	}

	assertConnClosed(t, conn)
}

func TestServerInvalidRequestClosesConnection(t *testing.T) {
	socketPath := filepath.Join(testutil.TempDir(t), "pxcli.sock")
	server, err := NewServer(socketPath, stubHandler{response: "ok"})
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}
	done := startServer(t, server)
	t.Cleanup(func() {
		stopServer(t, server, done)
	})

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("unexpected error connecting to socket: %v", err)
	}
	defer conn.Close()

	if _, err := io.WriteString(conn, "   \n"); err != nil {
		t.Fatalf("unexpected error writing request: %v", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("unexpected error reading response: %v", err)
	}
	if !strings.HasPrefix(line, "err invalid_command ") {
		t.Fatalf("expected invalid_command error, got %q", line)
	}

	assertConnClosed(t, conn)
}

func TestServerHandlesScript(t *testing.T) {
	socketPath := filepath.Join(testutil.TempDir(t), "pxcli.sock")
	handler := &stubScriptHandler{scriptResp: "ok"}
	server, err := NewServer(socketPath, handler)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}
	done := startServer(t, server)
	t.Cleanup(func() {
		stopServer(t, server, done)
	})

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("unexpected error connecting to socket: %v", err)
	}
	defer conn.Close()

	if _, err := io.WriteString(conn, "script\nset_pixel 0 0 red\nfill_rect 1 1 2 2 blue\n"); err != nil {
		t.Fatalf("unexpected error writing request: %v", err)
	}
	closeWrite(t, conn)

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("unexpected error reading response: %v", err)
	}
	if line != "ok\n" {
		t.Fatalf("expected response %q, got %q", "ok\n", line)
	}

	wantLines := []string{"set_pixel 0 0 red", "fill_rect 1 1 2 2 blue"}
	if len(handler.receivedLines) != len(wantLines) {
		t.Fatalf("expected %d lines, got %v", len(wantLines), handler.receivedLines)
	}
	for i, want := range wantLines {
		if handler.receivedLines[i] != want {
			t.Fatalf("expected line %d to be %q, got %q", i, want, handler.receivedLines[i])
		}
	}

	assertConnClosed(t, conn)
}

func TestServerScriptUnsupportedByHandler(t *testing.T) {
	socketPath := filepath.Join(testutil.TempDir(t), "pxcli.sock")
	server, err := NewServer(socketPath, stubHandler{response: "ok"})
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}
	done := startServer(t, server)
	t.Cleanup(func() {
		stopServer(t, server, done)
	})

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("unexpected error connecting to socket: %v", err)
	}
	defer conn.Close()

	if _, err := io.WriteString(conn, "script\n"); err != nil {
		t.Fatalf("unexpected error writing request: %v", err)
	}
	closeWrite(t, conn)

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("unexpected error reading response: %v", err)
	}
	if !strings.HasPrefix(line, "err invalid_command ") {
		t.Fatalf("expected invalid_command error, got %q", line)
	}
}

func TestServerScriptRejectsArgs(t *testing.T) {
	socketPath := filepath.Join(testutil.TempDir(t), "pxcli.sock")
	handler := &stubScriptHandler{scriptResp: "ok"}
	server, err := NewServer(socketPath, handler)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}
	done := startServer(t, server)
	t.Cleanup(func() {
		stopServer(t, server, done)
	})

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		t.Fatalf("unexpected error connecting to socket: %v", err)
	}
	defer conn.Close()

	if _, err := io.WriteString(conn, "script extra\n"); err != nil {
		t.Fatalf("unexpected error writing request: %v", err)
	}

	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("unexpected error reading response: %v", err)
	}
	if !strings.HasPrefix(line, "err invalid_args ") {
		t.Fatalf("expected invalid_args error, got %q", line)
	}
}

func closeWrite(t *testing.T, conn net.Conn) {
	t.Helper()
	unixConn, ok := conn.(interface{ CloseWrite() error })
	if !ok {
		t.Fatalf("connection does not support CloseWrite")
	}
	if err := unixConn.CloseWrite(); err != nil {
		t.Fatalf("unexpected error closing write side: %v", err)
	}
}

func startServer(t *testing.T, server *Server) <-chan error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- server.Serve()
	}()
	return done
}

func stopServer(t *testing.T, server *Server, done <-chan error) {
	t.Helper()
	_ = server.Close()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected server error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("server did not shut down in time")
	}
}

func assertConnClosed(t *testing.T, conn net.Conn) {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	var buf [1]byte
	n, err := conn.Read(buf[:])
	if n != 0 || !errors.Is(err, io.EOF) {
		t.Fatalf("expected EOF after response, got n=%d err=%v", n, err)
	}
}

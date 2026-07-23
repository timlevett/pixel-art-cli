package client

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"pxcli/internal/testutil"
)

func TestClientSendOK(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to listen on unix socket: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	serverErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		line, err := reader.ReadString('\n')
		if err != nil {
			serverErr <- err
			return
		}
		if strings.TrimRight(line, "\r\n") != "clear" {
			serverErr <- fmt.Errorf("expected request clear, got %q", strings.TrimRight(line, "\r\n"))
			return
		}
		if _, err := io.WriteString(conn, "ok\n"); err != nil {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	client, err := New(socketPath, WithDialTimeout(200*time.Millisecond), WithWriteTimeout(200*time.Millisecond), WithReadTimeout(200*time.Millisecond))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	resp, err := client.Send("clear")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Payload != "" {
		t.Fatalf("expected empty payload, got %q", resp.Payload)
	}
	if resp.Raw != "ok" {
		t.Fatalf("expected raw response ok, got %q", resp.Raw)
	}

	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatalf("server error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not finish")
	}
}

func TestClientSendMissingSocket(t *testing.T) {
	socketPath := filepath.Join(testutil.TempDir(t), "missing.sock")
	client, err := New(socketPath, WithDialTimeout(50*time.Millisecond))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.Send("clear")
	if err == nil {
		t.Fatalf("expected error")
	}
	var cerr Error
	if !errors.As(err, &cerr) {
		t.Fatalf("expected client error, got %v", err)
	}
	if cerr.Code != "daemon_not_running" {
		t.Fatalf("expected daemon_not_running, got %q", cerr.Code)
	}
}

func TestClientSendTimeout(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to listen on unix socket: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	serverDone := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverDone <- err
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		_, _ = reader.ReadString('\n')
		time.Sleep(200 * time.Millisecond)
		serverDone <- nil
	}()

	client, err := New(socketPath, WithDialTimeout(100*time.Millisecond), WithWriteTimeout(100*time.Millisecond), WithReadTimeout(100*time.Millisecond))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	start := time.Now()
	_, err = client.Send("clear")
	elapsed := time.Since(start)
	if err == nil {
		t.Fatalf("expected timeout error")
	}
	var cerr Error
	if !errors.As(err, &cerr) {
		t.Fatalf("expected client error, got %v", err)
	}
	if cerr.Code != "timeout" {
		t.Fatalf("expected timeout error code, got %q", cerr.Code)
	}
	if elapsed > time.Second {
		t.Fatalf("timeout exceeded expected duration: %v", elapsed)
	}

	select {
	case err := <-serverDone:
		if err != nil {
			t.Fatalf("server error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not finish")
	}
}

func TestClientSendScriptOK(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to listen on unix socket: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	serverErr := make(chan error, 1)
	var received []string
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		for {
			line, err := reader.ReadString('\n')
			if len(line) > 0 {
				received = append(received, strings.TrimRight(line, "\r\n"))
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				serverErr <- err
				return
			}
		}
		if _, err := io.WriteString(conn, "ok\n"); err != nil {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	client, err := New(socketPath, WithDialTimeout(200*time.Millisecond), WithScriptWriteTimeout(200*time.Millisecond), WithScriptReadTimeout(200*time.Millisecond))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	resp, err := client.SendScript([]string{"set_pixel 0 0 red", "fill_rect 1 1 2 2 blue"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Raw != "ok" {
		t.Fatalf("expected raw response ok, got %q", resp.Raw)
	}

	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatalf("server error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not finish")
	}

	wantLines := []string{"script", "set_pixel 0 0 red", "fill_rect 1 1 2 2 blue"}
	if len(received) != len(wantLines) {
		t.Fatalf("expected lines %v, got %v", wantLines, received)
	}
	for i, want := range wantLines {
		if received[i] != want {
			t.Fatalf("expected line %d to be %q, got %q", i, want, received[i])
		}
	}
}

func TestClientSendScriptError(t *testing.T) {
	dir := testutil.TempDir(t)
	socketPath := filepath.Join(dir, "pxcli.sock")
	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Fatalf("failed to listen on unix socket: %v", err)
	}
	t.Cleanup(func() { _ = listener.Close() })

	serverErr := make(chan error, 1)
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			serverErr <- err
			return
		}
		defer conn.Close()
		reader := bufio.NewReader(conn)
		for {
			_, err := reader.ReadString('\n')
			if err != nil {
				break
			}
		}
		if _, err := io.WriteString(conn, "err invalid_args line 2: x must be an integer\n"); err != nil {
			serverErr <- err
			return
		}
		serverErr <- nil
	}()

	client, err := New(socketPath, WithDialTimeout(200*time.Millisecond), WithScriptWriteTimeout(200*time.Millisecond), WithScriptReadTimeout(200*time.Millisecond))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	_, err = client.SendScript([]string{"set_pixel 0 0 red", "set_pixel x 0 red"})
	if err == nil {
		t.Fatalf("expected error")
	}
	var cerr Error
	if !errors.As(err, &cerr) {
		t.Fatalf("expected client error, got %v", err)
	}
	if cerr.Code != "invalid_args" {
		t.Fatalf("expected invalid_args code, got %q", cerr.Code)
	}
	if !strings.Contains(cerr.Message, "line 2") {
		t.Fatalf("expected message to reference line 2, got %q", cerr.Message)
	}

	select {
	case err := <-serverErr:
		if err != nil {
			t.Fatalf("server error: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("server did not finish")
	}
}

package backend

import (
	"bytes"
	"io"
	"sync"
	"testing"
)

// fakeSession is an in-memory sshSession, so the terminal adapter can be
// tested without a real SSH server.
type fakeSession struct {
	out     chan []byte
	in      bytes.Buffer
	mu      sync.Mutex
	resized [][2]int
	closed  bool
	once    sync.Once
}

func newFakeSession() *fakeSession {
	return &fakeSession{out: make(chan []byte, 1)}
}

func (f *fakeSession) Output() <-chan []byte { return f.out }

func (f *fakeSession) Input(p []byte) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	_, err := f.in.Write(p)
	return err
}

func (f *fakeSession) Resize(cols, rows int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.resized = append(f.resized, [2]int{cols, rows})
	return nil
}

func (f *fakeSession) Close() {
	f.once.Do(func() {
		f.mu.Lock()
		f.closed = true
		f.mu.Unlock()
		close(f.out)
	})
}

func TestSSHConn(t *testing.T) {
	fs := newFakeSession()
	c := newSSHConn(fs)

	fs.out <- []byte("hello")
	buf := make([]byte, 5)
	if _, err := io.ReadFull(c, buf); err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(buf) != "hello" {
		t.Fatalf("read %q, want hello", buf)
	}

	if n, err := c.Write([]byte("ls\n")); err != nil || n != 3 {
		t.Fatalf("write = %d, %v", n, err)
	}
	fs.mu.Lock()
	gotIn := fs.in.String()
	fs.mu.Unlock()
	if gotIn != "ls\n" {
		t.Fatalf("session input = %q, want ls\\n", gotIn)
	}

	if err := c.Resize(120, 40); err != nil {
		t.Fatalf("resize: %v", err)
	}
	fs.mu.Lock()
	gotResize := fs.resized
	fs.mu.Unlock()
	if len(gotResize) != 1 || gotResize[0] != [2]int{120, 40} {
		t.Fatalf("resize = %v, want [[120 40]]", gotResize)
	}

	if err := c.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	fs.mu.Lock()
	closed := fs.closed
	fs.mu.Unlock()
	if !closed {
		t.Fatal("session was not closed")
	}
	if _, err := c.Read(make([]byte, 1)); err == nil {
		t.Fatal("read after close should fail")
	}
}

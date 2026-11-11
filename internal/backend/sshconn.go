package backend

import (
	"io"
	"sync"
)

// sshSession is the subset of sshexec.Session the terminal adapter needs.
// Keeping it an interface lets the adapter be tested without a network.
type sshSession interface {
	Output() <-chan []byte
	Input([]byte) error
	Resize(cols, rows int) error
	Close()
}

// sshConn adapts an interactive SSH session to the terminal plugin's
// io.ReadWriteCloser. Its optional Resize method keeps the remote PTY in
// step with the view.
type sshConn struct {
	sess sshSession
	pr   *io.PipeReader
	pw   *io.PipeWriter
	once sync.Once
}

func newSSHConn(sess sshSession) *sshConn {
	pr, pw := io.Pipe()
	c := &sshConn{sess: sess, pr: pr, pw: pw}
	go func() {
		for chunk := range sess.Output() {
			if _, err := pw.Write(chunk); err != nil {
				return
			}
		}
		_ = pw.Close()
	}()
	return c
}

func (c *sshConn) Read(p []byte) (int, error) {
	return c.pr.Read(p)
}

func (c *sshConn) Write(p []byte) (int, error) {
	if err := c.sess.Input(p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func (c *sshConn) Close() error {
	c.once.Do(func() {
		c.sess.Close()
		_ = c.pr.Close()
		_ = c.pw.Close()
	})
	return nil
}

func (c *sshConn) Resize(cols, rows int) error {
	return c.sess.Resize(cols, rows)
}

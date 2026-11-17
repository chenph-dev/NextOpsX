package app

import (
	"io"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/egoist/mygo/ui"
)

// CompletionState is the typed-prefix command bubble shown over the terminal.
type CompletionState struct {
	Open          bool
	Buffer        string
	Prefix        string
	Items         []CommandItem
	Selected      int
	SuppressUntil time.Time
}

// UpdateCompletion filters the merged history/favorites list by the current
// terminal input line. Programmatic sends through the terminal are observed
// too; the bubble only opens for a plausible command prefix.
func (a *App) UpdateCompletion(buffer string) {
	if a.Term == nil {
		a.CloseCompletion()
		return
	}
	if !a.Completion.SuppressUntil.IsZero() && time.Now().Before(a.Completion.SuppressUntil) {
		a.CloseCompletion()
		return
	}
	a.Completion.Buffer = buffer
	prefix := strings.TrimSpace(buffer)
	if len([]rune(prefix)) < 2 || strings.ContainsAny(prefix, "|;&<>`$\\\r\n") {
		a.CloseCompletion()
		return
	}
	lower := strings.ToLower(prefix)
	var items []CommandItem
	for _, item := range a.CommandItems() {
		command := strings.ToLower(a.CommandItemText(item))
		title := strings.ToLower(item.Title)
		if strings.HasPrefix(command, lower) || (title != "" && strings.HasPrefix(title, lower)) {
			items = append(items, item)
			if len(items) >= 30 {
				break
			}
		}
	}
	if len(items) == 0 {
		a.CloseCompletion()
		return
	}
	a.Completion.Open = true
	a.Completion.Prefix = prefix
	a.Completion.Items = items
	if a.Completion.Selected < 0 || a.Completion.Selected >= len(items) {
		a.Completion.Selected = 0
	}
}

// suppressCompletion hides the bubble while the app itself writes to the
// terminal, so those bytes do not immediately reopen it.
func (a *App) suppressCompletion() {
	a.Completion.SuppressUntil = time.Now().Add(400 * time.Millisecond)
	a.CloseCompletion()
}

// CloseCompletion hides the command bubble.
func (a *App) CloseCompletion() {
	a.Completion.Open = false
	a.Completion.Buffer = ""
	a.Completion.Items = nil
	a.Completion.Selected = 0
	a.Completion.Prefix = ""
}

// AcceptCompletion replaces the typed line with the selected command.
func (a *App) AcceptCompletion(index int) {
	if a.Term == nil || !a.Completion.Open ||
		index < 0 || index >= len(a.Completion.Items) {
		return
	}
	item := a.Completion.Items[index]
	command := a.CommandItemText(item)
	n := len([]rune(a.Completion.Buffer))
	// The tracker sees our own backspaces and command; ignore those updates
	// so accepting does not immediately reopen the bubble.
	a.suppressCompletion()
	if n > 0 {
		a.Term.Send([]byte(strings.Repeat("\x7f", n)))
	}
	a.Term.Send([]byte(command))
}

func (a *App) moveCompletion(delta int) {
	n := len(a.Completion.Items)
	if n == 0 {
		return
	}
	a.Completion.Selected = (a.Completion.Selected + delta + n) % n
}

// HandleTermKey navigates the command bubble and keeps the drawer shortcuts
// working while the terminal has the keyboard. Enter is deliberately left to
// the shell so a password prompt cannot be turned into a selection; Tab
// accepts, Up/Down move, Escape closes.
func (a *App) HandleTermKey(mods ui.Modifiers, key ui.Key) bool {
	switch {
	case (mods == ui.Ctrl|ui.Shift || mods == ui.Super|ui.Shift) && key == ui.KeyJ:
		a.ToggleAIDrawer()
		return true
	case (mods == ui.Ctrl|ui.Shift || mods == ui.Super|ui.Shift) && key == ui.KeyE:
		a.ToggleSFTPDrawer()
		return true
	}
	if !a.Completion.Open {
		return false
	}
	switch key {
	case ui.KeyUp:
		a.moveCompletion(-1)
		return true
	case ui.KeyDown:
		a.moveCompletion(1)
		return true
	case ui.KeyTab:
		a.AcceptCompletion(a.Completion.Selected)
		return true
	case ui.KeyEscape:
		a.CloseCompletion()
		return true
	}
	return false
}

// inputTracker observes what the terminal writes to the SSH session so the
// app can suggest from the current input line. It keeps only a transient
// buffer; typed input is never persisted.
type inputTracker struct {
	conn     io.ReadWriteCloser
	mu       sync.Mutex
	buf      []byte
	esc, csi bool
	onChange func(string)
}

func newInputTracker(conn io.ReadWriteCloser, onChange func(string)) *inputTracker {
	return &inputTracker{conn: conn, onChange: onChange}
}

func (t *inputTracker) Read(p []byte) (int, error) { return t.conn.Read(p) }
func (t *inputTracker) Close() error               { return t.conn.Close() }

// Resize forwards the terminal size to the SSH session. The terminal plugin
// looks for this optional method on its Conn; without it a PTY stays at its
// initial size.
func (t *inputTracker) Resize(cols, rows int) error {
	if r, ok := t.conn.(interface{ Resize(cols, rows int) error }); ok {
		return r.Resize(cols, rows)
	}
	return nil
}

func (t *inputTracker) Write(p []byte) (int, error) {
	n, err := t.conn.Write(p)
	t.observe(p)
	return n, err
}

func (t *inputTracker) observe(p []byte) {
	t.mu.Lock()
	for _, b := range p {
		switch {
		case t.esc:
			switch {
			case b == '[':
				t.csi = true
			case t.csi:
				if b >= 0x40 && b <= 0x7e {
					t.esc, t.csi = false, false
				}
			default:
				t.esc = false
			}
		case b == 0x1b:
			t.esc = true
			t.buf = t.buf[:0]
		case b == '\r' || b == '\n' || b == 0x03 || b == 0x15 || b == '\t':
			t.buf = t.buf[:0]
		case b == 0x7f || b == 0x08:
			if len(t.buf) > 0 {
				_, size := utf8.DecodeLastRune(t.buf)
				t.buf = t.buf[:len(t.buf)-size]
			}
		case b >= 0x20:
			t.buf = append(t.buf, b)
		}
	}
	line := string(t.buf)
	t.mu.Unlock()
	if t.onChange != nil {
		t.onChange(line)
	}
}

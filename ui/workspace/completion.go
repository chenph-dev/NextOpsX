package workspace

import (
	"bytes"
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
func (a *State) UpdateCompletion(buffer string) {
	a.UpdateCompletionLine(buffer, false)
}

// UpdateCompletionLine is UpdateCompletion with the tracker's knowledge of
// whether it still knows the shell line. After Tab completion or an escape
// sequence the line may hold text the tracker did not see, so the bubble
// stays closed instead of guessing how many backspaces to send.
func (a *State) UpdateCompletionLine(buffer string, stale bool) {
	if a.Term == nil {
		a.CloseCompletion()
		return
	}
	if stale || a.TerminalInteractive() {
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
func (a *State) suppressCompletion() {
	a.Completion.SuppressUntil = time.Now().Add(400 * time.Millisecond)
	a.CloseCompletion()
}

// CloseCompletion hides the command bubble.
func (a *State) CloseCompletion() {
	a.Completion.Open = false
	a.Completion.Buffer = ""
	a.Completion.Items = nil
	a.Completion.Selected = 0
	a.Completion.Prefix = ""
}

// AcceptCompletion replaces the typed line with the selected command.
func (a *State) AcceptCompletion(index int) {
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
	// Paste, not Send: a multi-line favorite must not be run as it is typed.
	a.Term.Paste(command)
}

func (a *State) moveCompletion(delta int) {
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
func (a *State) HandleTermKey(mods ui.Modifiers, key ui.Key) bool {
	switch {
	case (mods == ui.Ctrl|ui.Shift || mods == ui.Super|ui.Shift) && key == ui.KeyJ:
		a.ToggleAIDrawer()
		return true
	case (mods == ui.Ctrl|ui.Shift || mods == ui.Super|ui.Shift) && key == ui.KeyE:
		a.ToggleSFTPDrawer()
		return true
	}
	// A full-screen program or a password prompt owns every other key.
	if a.TerminalInteractive() {
		a.CloseCompletion()
		return false
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

// TerminalInteractive reports whether the remote side currently owns the
// keyboard: a full-screen program or a password prompt.
func (a *State) TerminalInteractive() bool {
	return a.termTracker != nil && a.termTracker.Interactive()
}

// inputTracker observes what the terminal writes to the SSH session and what
// the session prints back. The write side feeds command suggestions; the read
// side keeps only enough state to know when suggestions must stay out of the
// way: a full-screen program or a password prompt.
type inputTracker struct {
	conn     io.ReadWriteCloser
	mu       sync.Mutex
	buf      []byte
	esc, csi bool
	stale    bool
	onChange func(string, bool)

	outMu     sync.Mutex
	outTail   []byte
	altScreen bool
	password  bool
}

func newInputTracker(conn io.ReadWriteCloser, onChange func(string, bool)) *inputTracker {
	return &inputTracker{conn: conn, onChange: onChange}
}

func (t *inputTracker) Read(p []byte) (int, error) {
	n, err := t.conn.Read(p)
	if n > 0 {
		t.observeOutput(p[:n])
	}
	return n, err
}

func (t *inputTracker) Close() error { return t.conn.Close() }

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
			t.csi = false
			t.stale = true
			t.buf = t.buf[:0]
		case b == '\r' || b == '\n' || b == 0x03 || b == 0x15:
			t.buf = t.buf[:0]
			t.stale = false
			t.clearPassword()
		case b == '\t':
			t.buf = t.buf[:0]
			t.stale = true
		case b == 0x7f || b == 0x08:
			if !t.stale && len(t.buf) > 0 {
				_, size := utf8.DecodeLastRune(t.buf)
				t.buf = t.buf[:len(t.buf)-size]
			}
		case b >= 0x20:
			t.buf = append(t.buf, b)
		default:
			// Other control keys move or edit the line in ways the
			// tracker cannot follow.
			t.buf = t.buf[:0]
			t.stale = true
		}
	}
	line := string(t.buf)
	stale := t.stale
	t.mu.Unlock()
	if t.onChange != nil {
		t.onChange(line, stale)
	}
}

// Interactive reports whether the remote side owns the keyboard: a
// full-screen program or a password prompt.
func (t *inputTracker) Interactive() bool {
	t.outMu.Lock()
	defer t.outMu.Unlock()
	return t.altScreen || t.password
}

func (t *inputTracker) observeOutput(p []byte) {
	t.outMu.Lock()
	defer t.outMu.Unlock()
	t.outTail = append(t.outTail, p...)
	if len(t.outTail) > 8192 {
		t.outTail = append([]byte(nil), t.outTail[len(t.outTail)-8192:]...)
	}
	enter := lastSequence(t.outTail,
		[]byte("\x1b[?1049h"), []byte("\x1b[?1047h"), []byte("\x1b[?47h"))
	exit := lastSequence(t.outTail,
		[]byte("\x1b[?1049l"), []byte("\x1b[?1047l"), []byte("\x1b[?47l"))
	if enter >= 0 || exit >= 0 {
		t.altScreen = enter > exit
	}
	line := t.outTail
	if i := bytes.LastIndexAny(line, "\r\n"); i >= 0 {
		line = line[i+1:]
	}
	t.password = passwordPrompt(line)
}

func (t *inputTracker) clearPassword() {
	t.outMu.Lock()
	t.password = false
	t.outMu.Unlock()
}

func lastSequence(data []byte, seqs ...[]byte) int {
	last := -1
	for _, seq := range seqs {
		if i := bytes.LastIndex(data, seq); i > last {
			last = i
		}
	}
	return last
}

func passwordPrompt(line []byte) bool {
	s := strings.TrimRight(trimTrailingSGR(string(line)), " \t")
	if s == "" {
		return false
	}
	lower := strings.ToLower(s)
	if !strings.HasSuffix(lower, ":") && !strings.HasSuffix(lower, "：") {
		return false
	}
	return strings.Contains(lower, "password") ||
		strings.Contains(lower, "passphrase") ||
		strings.Contains(lower, "密码") ||
		strings.Contains(lower, "口令")
}

// trimTrailingSGR removes color resets a prompt may print after its colon.
func trimTrailingSGR(s string) string {
	for {
		s = strings.TrimRight(s, " \t")
		i := strings.LastIndex(s, "\x1b[")
		if i < 0 || !strings.HasSuffix(s, "m") {
			return s
		}
		s = s[:i]
	}
}

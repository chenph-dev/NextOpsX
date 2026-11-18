package workspace

import (
	"fmt"
	"strings"

	"nextopsx/i18n"
	historystore "nextopsx/store/history"
	snippetsstore "nextopsx/store/snippets"
)

// SnippetState is the command snippet library and picker state.
type SnippetState struct {
	Items   []snippetsstore.Snippet
	Loading bool
	Err     string

	// Settings editor.
	Selected      string
	Editing       bool
	Draft         snippetsstore.Snippet
	DraftErr      string
	ConfirmDelete string

	// Settings favorites search.
	Query string

	// Terminal command bubble: executed history and favorites, deduplicated.
	PickerOpen     bool
	PickerTab      int
	CommandQuery   string
	History        []historystore.Entry
	HistoryOpen    map[string]bool
	HistoryLoading bool
	HistoryErr     string

	// Focus returns to the terminal after the picker closes.
	PickerWasOpen bool
	FocusTerminal bool

	seq        uint64
	historySeq uint64
}

// LoadSnippets refreshes the command snippet list.
func (a *State) LoadSnippets() {
	a.Snippets.seq++
	seq := a.Snippets.seq
	a.Snippets.Loading = true
	a.Snippets.Err = ""
	if a.Backend == nil {
		a.Snippets.Loading = false
		a.Snippets.Err = i18n.T("settings.err.backendNotInitialized")
		return
	}
	b := a.Backend
	go func() {
		list, err := b.Snippets.ListCommandSnippets()
		a.Post(func() {
			if seq != a.Snippets.seq {
				return
			}
			a.Snippets.Loading = false
			if err != nil {
				a.Snippets.Err = i18n.Tf("snippets.err.loadFailed", err.Error())
				return
			}
			a.Snippets.Items = list
			if a.Snippets.Selected != "" {
				if _, ok := a.snippetByID(a.Snippets.Selected); !ok {
					a.Snippets.Selected = ""
					a.Snippets.Editing = false
					a.Snippets.Draft = snippetsstore.Snippet{}
				}
			}
		})
	}()
}

// LoadCommandHistory refreshes the selected host's executed command history.
func (a *State) LoadCommandHistory() {
	a.Snippets.historySeq++
	seq := a.Snippets.historySeq
	if a.Selected == nil || a.Selected.NodeType != "host" {
		a.Snippets.History = nil
		a.Snippets.HistoryLoading = false
		a.Snippets.HistoryErr = ""
		return
	}
	hostID := a.Selected.ID
	a.Snippets.HistoryLoading = true
	a.Snippets.HistoryErr = ""
	if a.Backend == nil {
		a.Snippets.HistoryLoading = false
		a.Snippets.HistoryErr = i18n.T("settings.err.backendNotInitialized")
		return
	}
	b := a.Backend
	go func() {
		list, err := b.History.ListCommandHistory(hostID, 200)
		a.Post(func() {
			if seq != a.Snippets.historySeq || a.Selected == nil || a.Selected.ID != hostID {
				return
			}
			a.Snippets.HistoryLoading = false
			if err != nil {
				a.Snippets.HistoryErr = i18n.Tf("history.err.load", err.Error())
				return
			}
			a.Snippets.History = list
		})
	}()
}

func (a *State) historyByID(id string) (historystore.Entry, bool) {
	for _, e := range a.Snippets.History {
		if e.ID == id {
			return e, true
		}
	}
	return historystore.Entry{}, false
}

// ToggleHistoryOutput expands or collapses one history entry's output.
func (a *State) ToggleHistoryOutput(id string) {
	if a.Snippets.HistoryOpen == nil {
		a.Snippets.HistoryOpen = map[string]bool{}
	}
	a.Snippets.HistoryOpen[id] = !a.Snippets.HistoryOpen[id]
}

// InsertHistory pastes a historical command into the terminal without running it.
func (a *State) InsertHistory(id string) {
	e, ok := a.historyByID(id)
	if !ok {
		return
	}
	if !a.CanUseSnippet() {
		a.Snippets.Err = i18n.T("snippets.err.notConnected")
		return
	}
	a.suppressCompletion()
	a.Term.Paste(e.Command)
	a.Snippets.PickerOpen = false
	a.Snippets.Err = ""
}

// RunHistory pastes a historical command and presses Enter.
func (a *State) RunHistory(id string) {
	e, ok := a.historyByID(id)
	if !ok {
		return
	}
	if !a.CanUseSnippet() {
		a.Snippets.Err = i18n.T("snippets.err.notConnected")
		return
	}
	a.suppressCompletion()
	a.Term.Paste(e.Command)
	a.Term.Send([]byte("\r"))
	a.Snippets.PickerOpen = false
	a.Snippets.Err = ""
}

// FavoriteHistory saves a history entry as an editable favorite.
func (a *State) FavoriteHistory(id string) {
	e, ok := a.historyByID(id)
	if !ok {
		return
	}
	if a.Backend == nil {
		a.Snippets.Err = i18n.T("settings.err.backendNotInitialized")
		return
	}
	name := snippetName(e.Command)
	b := a.Backend
	go func() {
		saved, err := b.Snippets.SaveCommandSnippet(snippetsstore.Snippet{
			Name: name, Group: i18n.T("history.favoriteGroup"),
			Command: e.Command, Description: i18n.T("history.favoriteDescription"),
		})
		a.Post(func() {
			if err != nil {
				a.Snippets.Err = i18n.Tf("snippets.err.saveFailed", err.Error())
				return
			}
			a.Snippets.Selected = saved.ID
			a.Snippets.Draft = saved
			a.Snippets.Editing = false
			a.Snippets.Err = ""
			a.LoadSnippets()
		})
	}()
}

func snippetName(command string) string {
	line := strings.TrimSpace(command)
	if i := strings.IndexByte(line, '\n'); i >= 0 {
		line = strings.TrimSpace(line[:i])
	}
	if line == "" {
		return i18n.T("history.favoriteFallback")
	}
	r := []rune(line)
	if len(r) > 40 {
		line = string(r[:39]) + "…"
	}
	return line
}

// recordCommand records an app-triggered command for the selected host.
func (a *State) recordCommand(source, command string) {
	if a.Backend == nil || a.Selected == nil || a.Selected.NodeType != "host" {
		return
	}
	hostID := a.Selected.ID
	b := a.Backend
	go func() {
		if err := b.History.SaveCommandHistory(hostID, source, command, nil, ""); err != nil {
			a.Post(func() {
				a.Snippets.HistoryErr = i18n.Tf("history.err.save", err.Error())
			})
		}
	}()
}

func (a *State) snippetByID(id string) (snippetsstore.Snippet, bool) {
	for _, s := range a.Snippets.Items {
		if s.ID == id {
			return s, true
		}
	}
	return snippetsstore.Snippet{}, false
}

// OpenNewSnippet starts an empty snippet in the settings editor.
func (a *State) OpenNewSnippet() {
	a.Snippets.Selected = ""
	a.Snippets.Editing = true
	a.Snippets.Draft = snippetsstore.Snippet{}
	a.Snippets.DraftErr = ""
}

// SelectSnippet opens an existing snippet in the settings editor.
func (a *State) SelectSnippet(id string) {
	s, ok := a.snippetByID(id)
	if !ok {
		return
	}
	a.Snippets.Selected = id
	a.Snippets.Editing = true
	a.Snippets.Draft = s
	a.Snippets.DraftErr = ""
}

// CancelSnippetEdit discards the current editor draft.
func (a *State) CancelSnippetEdit() {
	a.Snippets.Editing = false
	a.Snippets.DraftErr = ""
	if s, ok := a.snippetByID(a.Snippets.Selected); ok {
		a.Snippets.Draft = s
	} else {
		a.Snippets.Draft = snippetsstore.Snippet{}
	}
}

// SaveSnippet validates and persists the editor draft.
func (a *State) SaveSnippet() {
	draft := a.Snippets.Draft
	if strings.TrimSpace(draft.Name) == "" || strings.TrimSpace(draft.Command) == "" {
		a.Snippets.DraftErr = i18n.T("snippets.err.required")
		return
	}
	if a.Backend == nil {
		a.Snippets.DraftErr = i18n.T("settings.err.backendNotInitialized")
		return
	}
	b := a.Backend
	draft.Name = strings.TrimSpace(draft.Name)
	draft.Command = strings.TrimSpace(draft.Command)
	go func() {
		saved, err := b.Snippets.SaveCommandSnippet(draft)
		a.Post(func() {
			if err != nil {
				a.Snippets.DraftErr = i18n.Tf("snippets.err.saveFailed", err.Error())
				return
			}
			a.Snippets.Selected = saved.ID
			a.Snippets.Draft = saved
			a.Snippets.Editing = true
			a.Snippets.DraftErr = ""
			a.LoadSnippets()
		})
	}()
}

// OpenDeleteSnippet asks for confirmation before deleting a snippet.
func (a *State) OpenDeleteSnippet(id string) {
	a.Snippets.ConfirmDelete = id
}

// CancelDeleteSnippet closes the delete confirmation.
func (a *State) CancelDeleteSnippet() {
	a.Snippets.ConfirmDelete = ""
}

// ConfirmDeleteSnippet deletes the snippet awaiting confirmation.
func (a *State) ConfirmDeleteSnippet() {
	id := a.Snippets.ConfirmDelete
	a.Snippets.ConfirmDelete = ""
	if id == "" {
		return
	}
	if a.Backend == nil {
		a.Snippets.Err = i18n.T("settings.err.backendNotInitialized")
		return
	}
	b := a.Backend
	go func() {
		err := b.Snippets.DeleteCommandSnippet(id)
		a.Post(func() {
			if err != nil {
				a.Snippets.Err = i18n.Tf("snippets.err.deleteFailed", err.Error())
				return
			}
			if a.Snippets.Selected == id {
				a.Snippets.Selected = ""
				a.Snippets.Editing = false
				a.Snippets.Draft = snippetsstore.Snippet{}
			}
			a.LoadSnippets()
		})
	}()
}

// ResolveSnippetCommand substitutes variables from the selected asset.
func (a *State) ResolveSnippetCommand(command string) string {
	if a.Selected == nil {
		return command
	}
	n := a.Selected
	return strings.NewReplacer(
		"{{name}}", n.Name,
		"{{host}}", n.Addr,
		"{{addr}}", n.Addr,
		"{{user}}", n.User,
		"{{port}}", fmt.Sprintf("%d", n.Port),
		"{{protocol}}", n.Protocol,
	).Replace(command)
}

// CanUseSnippet reports whether a snippet can be sent to the active terminal.
func (a *State) CanUseSnippet() bool {
	if a.Term == nil || a.Selected == nil || a.TermHostID != a.Selected.ID {
		return false
	}
	if s := a.SessionFor(a.TermHostID); s != nil {
		return a.SessionLive(s)
	}
	return true
}

// Picker filters.
const (
	PickerTabAll = iota
	PickerTabFavorites
	PickerTabHistory
)

// ToggleSnippetPicker opens or closes the terminal command picker.
func (a *State) ToggleSnippetPicker() {
	a.Snippets.PickerOpen = !a.Snippets.PickerOpen
	a.Snippets.Err = ""
	if a.Snippets.PickerOpen {
		// The picker takes focus; an open suggestion bubble would keep
		// refocusing the terminal underneath.
		a.CloseCompletion()
		a.LoadCommandHistory()
		if len(a.Snippets.Items) == 0 {
			a.LoadSnippets()
		}
	}
}

// InsertSnippet pastes a snippet into the active terminal without running it.
func (a *State) InsertSnippet(id string) {
	s, ok := a.snippetByID(id)
	if !ok {
		return
	}
	if !a.CanUseSnippet() {
		a.Snippets.Err = i18n.T("snippets.err.notConnected")
		return
	}
	a.suppressCompletion()
	a.Term.Paste(a.ResolveSnippetCommand(s.Command))
	a.Snippets.PickerOpen = false
	a.Snippets.Err = ""
}

// RunSnippet pastes a snippet into the active terminal and presses Enter.
func (a *State) RunSnippet(id string) {
	s, ok := a.snippetByID(id)
	if !ok {
		return
	}
	if !a.CanUseSnippet() {
		a.Snippets.Err = i18n.T("snippets.err.notConnected")
		return
	}
	command := a.ResolveSnippetCommand(s.Command)
	a.suppressCompletion()
	a.Term.Paste(command)
	a.Term.Send([]byte("\r"))
	a.recordCommand("snippet", command)
	a.Snippets.PickerOpen = false
	a.Snippets.Err = ""
}

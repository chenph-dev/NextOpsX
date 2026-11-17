package app

import (
	"sort"
	"strings"
)

// CommandItem is one command in the terminal bubble, merged from executed
// history and favorites by the command text.
type CommandItem struct {
	Command     string
	Title       string
	Description string
	Source      string // history | favorite | both
	FavoriteID  string
	HistoryID   string
	Ts          int64
	ExitCode    *int
	Output      string
}

// CommandItems returns history and favorites deduplicated by command. Items
// that are both a favorite and previously executed come first.
func (a *App) CommandItems() []CommandItem {
	byCommand := map[string]*CommandItem{}
	for _, e := range a.Snippets.History {
		cmd := strings.TrimSpace(e.Command)
		if cmd == "" {
			continue
		}
		it := byCommand[cmd]
		if it == nil {
			it = &CommandItem{Command: cmd, Source: "history"}
			byCommand[cmd] = it
		}
		if e.Ts >= it.Ts {
			it.HistoryID = e.ID
			it.Ts = e.Ts
			it.ExitCode = e.ExitCode
			it.Output = e.Output
		}
	}
	for _, s := range a.Snippets.Items {
		cmd := strings.TrimSpace(s.Command)
		if cmd == "" {
			continue
		}
		it := byCommand[cmd]
		if it == nil {
			it = &CommandItem{Command: cmd, Source: "favorite"}
			byCommand[cmd] = it
		} else {
			it.Source = "both"
		}
		it.FavoriteID = s.ID
		it.Title = s.Name
		it.Description = s.Description
	}
	out := make([]CommandItem, 0, len(byCommand))
	for _, it := range byCommand {
		out = append(out, *it)
	}
	sort.Slice(out, func(i, j int) bool {
		ri, rj := commandSourceRank(out[i].Source), commandSourceRank(out[j].Source)
		if ri != rj {
			return ri < rj
		}
		if out[i].Ts != out[j].Ts {
			return out[i].Ts > out[j].Ts
		}
		return out[i].Command < out[j].Command
	})
	return out
}

func commandSourceRank(source string) int {
	switch source {
	case "both":
		return 0
	case "favorite":
		return 1
	default:
		return 2
	}
}

// CommandItemText is the command shown and inserted for an item. Favorite
// commands resolve the asset variables; executed history is already concrete.
func (a *App) CommandItemText(item CommandItem) string {
	if item.HistoryID != "" {
		return item.Command
	}
	return a.ResolveSnippetCommand(item.Command)
}

// InsertCommandItem pastes a merged item into the terminal without running it.
func (a *App) InsertCommandItem(item CommandItem) {
	if item.HistoryID != "" {
		a.InsertHistory(item.HistoryID)
		return
	}
	a.InsertSnippet(item.FavoriteID)
}

// RunCommandItem pastes a merged item into the terminal and presses Enter.
func (a *App) RunCommandItem(item CommandItem) {
	if item.HistoryID != "" {
		a.RunHistory(item.HistoryID)
		return
	}
	a.RunSnippet(item.FavoriteID)
}

// FavoriteCommandItem saves a history-only item as a favorite.
func (a *App) FavoriteCommandItem(item CommandItem) {
	if item.FavoriteID != "" {
		return
	}
	a.FavoriteHistory(item.HistoryID)
}

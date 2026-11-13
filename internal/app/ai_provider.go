package app

import (
	"nextopsx/internal/einoagent/session"
	configstore "nextopsx/internal/store/config"
)

// applyAIProvider writes the provider a conversation resolves to into the AI
// panel's runtime state.
func (a *App) applyAIProvider(p configstore.AIProvider, ok bool) {
	if !ok {
		a.AI.ProviderID = ""
		a.AI.ProviderName = ""
		a.AI.ProviderChoice = ""
		a.AI.switchTarget = ""
		a.AI.Model = ""
		a.AI.Configured = false
		return
	}
	a.AI.ProviderID = p.ID
	a.AI.ProviderName = p.Name
	a.AI.ProviderChoice = p.ID
	a.AI.switchTarget = p.ID
	a.AI.Model = p.Model
	a.AI.Configured = p.Model != ""
}

// providerByID finds a provider in the loaded list.
func (a *App) providerByID(id string) (configstore.AIProvider, bool) {
	for _, p := range a.AI.Providers {
		if p.ID == id {
			return p, true
		}
	}
	return configstore.AIProvider{}, false
}

// SwitchAIProvider changes the provider used by the current conversation.
// The next message rebuilds the model graph; while a turn is running or an
// approval is pending the choice is ignored.
//
// Writes are serialized: while a switch is in flight the latest choice is
// kept as switchTarget, and the completion handler issues another write when
// the target moved. This keeps the DB on the last choice even if the user
// clicks several providers quickly.
func (a *App) SwitchAIProvider(id string) {
	sid := a.AI.SessionID
	if a.Backend == nil || sid == "" || id == "" || id == a.AI.ProviderID {
		a.AI.ProviderChoice = a.AI.ProviderID
		return
	}
	if a.AI.Busy() || a.AI.State == session.StateAwaitingApproval {
		a.AI.ProviderChoice = a.AI.ProviderID
		return
	}
	picked, found := a.providerByID(id)
	if !found {
		a.AI.ProviderChoice = a.AI.ProviderID
		return
	}
	prev := configstore.AIProvider{ID: a.AI.ProviderID, Name: a.AI.ProviderName, Model: a.AI.Model}
	a.AI.providerSeq++
	a.applyAIProvider(picked, true)
	a.AI.Err = ""
	if a.AI.switching {
		return
	}
	a.runAISwitch(sid, prev)
}

// runAISwitch persists the latest switch target, then checks whether the user
// moved the target while the write was in flight.
func (a *App) runAISwitch(sid string, prev configstore.AIProvider) {
	a.AI.switching = true
	target := a.AI.switchTarget
	b := a.Backend
	go func() {
		err := b.SetConversationAIProvider(sid, target)
		var authoritative configstore.AIProvider
		var authErr error
		if err != nil {
			authoritative, authErr = b.ProviderForConversation(sid)
		}
		a.Post(func() {
			if a.AI.SessionID != sid {
				return
			}
			a.AI.switching = false
			next := a.AI.switchTarget
			if err != nil {
				// 回落到后端保存的权威值，但用户在此期间选择的新目标优先。
				if authErr == nil {
					a.applyAIProvider(authoritative, true)
				} else {
					a.applyAIProvider(prev, prev.ID != "")
				}
				if next != target && next != "" {
					if p, ok := a.providerByID(next); ok {
						a.applyAIProvider(p, true)
						a.runAISwitch(sid, prev)
					}
				} else {
					a.AI.Err = "切换供应商失败：" + err.Error()
				}
				return
			}
			if next != target && next != "" {
				a.runAISwitch(sid, prev)
			}
		})
	}()
}

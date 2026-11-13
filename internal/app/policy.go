package app

import (
	"strings"

	configstore "nextopsx/internal/store/config"
)

// PolicyState is the approval-policy section of the settings page.
type PolicyState struct {
	EnableAuto bool
	ReadOnly   string // 每行一个命令名；空表示使用内置默认白名单
	Loaded     bool   // 成功读取过；未加载时禁止保存，避免用零值覆盖已存策略
	Loading    bool
	Saving     bool
	Err        string
	Notice     string

	seq uint64
}

// LoadApprovalPolicy reads the global approval policy.
func (a *App) LoadApprovalPolicy() {
	a.Policy.seq++
	seq := a.Policy.seq
	a.Policy.Loading = true
	a.Policy.Err = ""
	a.Policy.Notice = ""
	if a.Backend == nil {
		a.Policy.Loading = false
		a.Policy.Err = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		p, err := b.GetApprovalPolicy()
		a.Post(func() {
			if seq != a.Policy.seq {
				return
			}
			a.Policy.Loading = false
			if err != nil {
				a.Policy.Err = "读取审批策略失败：" + err.Error()
				return
			}
			a.Policy.Loaded = true
			a.Policy.EnableAuto = p.EnableAuto
			a.Policy.ReadOnly = strings.Join(p.ReadOnlyList, "\n")
		})
	}()
}

// SaveApprovalPolicy persists the global approval policy.
func (a *App) SaveApprovalPolicy() {
	if a.Backend == nil {
		a.Policy.Err = "后端未初始化"
		return
	}
	if !a.Policy.Loaded {
		a.Policy.Err = "审批策略尚未加载，请先重试"
		return
	}
	// 使保存开始前就已发出的读取失效，避免旧快照覆盖表单。
	a.Policy.seq++
	a.Policy.Err = ""
	a.Policy.Notice = ""
	a.Policy.Saving = true
	p := configstore.ApprovalPolicy{
		EnableAuto:   a.Policy.EnableAuto,
		ReadOnlyList: parseWhitelist(a.Policy.ReadOnly),
	}
	b := a.Backend
	go func() {
		err := b.SaveApprovalPolicy(p)
		a.Post(func() {
			a.Policy.Saving = false
			if err != nil {
				a.Policy.Err = "保存审批策略失败：" + err.Error()
				return
			}
			// 保存期间可能又发起了读取；再推进一次使它们失效。
			a.Policy.seq++
			a.Policy.Loaded = true
			a.Policy.Notice = "审批策略已保存"
		})
	}()
}

// parseWhitelist splits on lines and commas only, never on spaces: a whole
// command line must stay one entry so it cannot silently broaden matching to
// its first word (matching uses the first token of the executed command).
func parseWhitelist(text string) []string {
	fields := strings.FieldsFunc(text, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\r'
	})
	seen := make(map[string]struct{}, len(fields))
	out := make([]string, 0, len(fields))
	for _, f := range fields {
		f = strings.TrimSpace(f)
		if f == "" {
			continue
		}
		if _, ok := seen[f]; ok {
			continue
		}
		seen[f] = struct{}{}
		out = append(out, f)
	}
	return out
}

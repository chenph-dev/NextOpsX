package connector

import "strings"

// IsQuery 判断 SQL 首关键字是否为查询类（走 QueryRunner.Query 返回行集）。
// 仅按首关键字粗分；WITH 开头的写语句（如 CTE + UPDATE）会保守走 Query，
// 但 Query 对无行结果返回空 Rows，不影响正确性。
func IsQuery(sqlText string) bool {
	switch firstKeyword(sqlText) {
	case "select", "show", "desc", "describe", "explain", "with", "pragma", "values":
		return true
	}
	return false
}

// firstKeyword 提取 SQL 的首个关键字（小写），跳过前置注释与括号。
func firstKeyword(s string) string {
	for {
		s = strings.TrimSpace(s)
		switch {
		case strings.HasPrefix(s, "--"), strings.HasPrefix(s, "#"):
			if i := strings.IndexByte(s, '\n'); i >= 0 {
				s = s[i+1:]
				continue
			}
			return ""
		case strings.HasPrefix(s, "/*"):
			if i := strings.Index(s, "*/"); i >= 0 {
				s = s[i+2:]
				continue
			}
			return ""
		case strings.HasPrefix(s, "("):
			s = strings.TrimSpace(s[1:])
			continue
		}
		break
	}
	if i := strings.IndexAny(s, " \t\r\n"); i >= 0 {
		s = s[:i]
	}
	return strings.ToLower(s)
}

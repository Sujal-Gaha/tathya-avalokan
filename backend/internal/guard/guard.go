package guard

import (
	"regexp"
	"strings"
)

var (
	blockCommentRe = regexp.MustCompile(`(?s)/\*.*?\*/`)
	lineCommentRe  = regexp.MustCompile(`--[^\n]*`)

	mutatingKeywords = map[string]struct{}{
		"insert":   {},
		"update":   {},
		"delete":   {},
		"drop":     {},
		"alter":    {},
		"truncate": {},
		"create":   {},
		"replace":  {},
	}
)

// ExtractFirstSQLKeyword safely strips block comments (/* */), line comments (--),
// leading parentheses, semicolons, and whitespace, returning the first lowercased word token.
func ExtractFirstSQLKeyword(sql string) string {
	stripped := blockCommentRe.ReplaceAllString(sql, "")
	stripped = lineCommentRe.ReplaceAllString(stripped, "")
	stripped = strings.TrimLeft(stripped, " \t\n\r();")
	tokens := strings.Fields(stripped)
	if len(tokens) == 0 {
		return ""
	}
	return strings.ToLower(strings.Trim(tokens[0], "();,"))
}

// IsMutating checks if a SQL statement starts with or contains mutating operations.
// It detects parenthesis wrapping (e.g. `(INSERT ...)`), mutating CTEs (`WITH d AS (DELETE ...)`),
// and multi-statement queries containing mutating keywords.
func IsMutating(sql string) (bool, string) {
	kw := ExtractFirstSQLKeyword(sql)
	if _, isMutating := mutatingKeywords[kw]; isMutating {
		return true, kw
	}

	// Check for CTE with mutating statement or multi-statement bypasses
	stripped := blockCommentRe.ReplaceAllString(sql, "")
	stripped = lineCommentRe.ReplaceAllString(stripped, "")
	tokens := strings.Fields(stripped)

	hasSemicolon := strings.Contains(sql, ";")
	isCTE := (kw == "with")

	if isCTE || hasSemicolon {
		for _, rawToken := range tokens {
			t := strings.ToLower(strings.Trim(rawToken, "();,"))
			if _, isMut := mutatingKeywords[t]; isMut {
				return true, t
			}
		}
	}

	return false, kw
}

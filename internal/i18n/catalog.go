package i18n

import (
	"encoding/json"
	"strings"
	"unicode"
)

// CollectLocalizedTexts returns non-empty unique strings from a locale map.
func CollectLocalizedTexts(m LocalizedMap) []string {
	if len(m) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(m))
	out := make([]string, 0, len(m))
	for _, text := range m {
		text = strings.TrimSpace(text)
		if text == "" {
			continue
		}
		if _, ok := seen[text]; ok {
			continue
		}
		seen[text] = struct{}{}
		out = append(out, text)
	}
	return out
}

// BuildSearchDocument joins catalog text used for cross-locale search indexing.
func BuildSearchDocument(parts ...string) string {
	seen := make(map[string]struct{})
	var b strings.Builder
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString(part)
	}
	return b.String()
}

// ParseSearchAliases unmarshals a JSONB alias list.
func ParseSearchAliases(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return nil
	}
	var aliases []string
	if err := json.Unmarshal(raw, &aliases); err != nil {
		return nil
	}
	out := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		alias = strings.TrimSpace(alias)
		if alias != "" {
			out = append(out, alias)
		}
	}
	return out
}

// NormalizeSearchQuery prepares user input for ILIKE / trigram matching across scripts.
func NormalizeSearchQuery(query string) string {
	query = strings.TrimSpace(query)
	if query == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(query))
	for _, r := range query {
		switch r {
		case '\u200c', '\u200f', '\u202a', '\u202b', '\u202c', '\u202d', '\u202e':
			continue
		case '\u064a': // Arabic yeh → Persian yeh
			b.WriteRune('\u06cc')
		case '\u0643': // Arabic kaf → Persian kaf
			b.WriteRune('\u06a9')
		default:
			if r >= '0' && r <= '9' {
				b.WriteRune(r)
				continue
			}
			if r >= '۰' && r <= '۹' {
				b.WriteRune('0' + (r - '۰'))
				continue
			}
			b.WriteRune(r)
		}
	}
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, b.String())
}

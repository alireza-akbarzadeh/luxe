package i18n

import "context"

const DefaultLocale = "en"

// SupportedLocales matches luxe-front cookie locales (en, es, fa).
var SupportedLocales = map[string]struct{}{
	"en": {},
	"es": {},
	"fa": {},
}

// ParseAcceptLanguage picks the first supported locale from an Accept-Language header.
func ParseAcceptLanguage(header string) string {
	if header == "" {
		return DefaultLocale
	}

	for _, part := range splitAcceptLanguage(header) {
		tag := part
		if idx := indexByte(tag, ';'); idx >= 0 {
			tag = tag[:idx]
		}
		tag = lowerTrim(tag)
		if tag == "" {
			continue
		}
		if _, ok := SupportedLocales[tag]; ok {
			return tag
		}
		if dash := indexByte(tag, '-'); dash >= 0 {
			base := tag[:dash]
			if _, ok := SupportedLocales[base]; ok {
				return base
			}
		}
	}

	return DefaultLocale
}

// LocaleFromContext returns the request locale stored by locale middleware.
func LocaleFromContext(ctx context.Context) string {
	if ctx == nil {
		return DefaultLocale
	}
	if v := ctx.Value(localeContextKey); v != nil {
		if s, ok := v.(string); ok {
			if _, supported := SupportedLocales[s]; supported {
				return s
			}
		}
	}
	return DefaultLocale
}

func splitAcceptLanguage(header string) []string {
	var parts []string
	start := 0
	for i := 0; i <= len(header); i++ {
		if i == len(header) || header[i] == ',' {
			part := header[start:i]
			if part != "" {
				parts = append(parts, part)
			}
			start = i + 1
		}
	}
	return parts
}

func lowerTrim(s string) string {
	i, j := 0, len(s)
	for i < j && (s[i] == ' ' || s[i] == '\t') {
		i++
	}
	for j > i && (s[j-1] == ' ' || s[j-1] == '\t') {
		j--
	}
	b := make([]byte, j-i)
	for k := i; k < j; k++ {
		c := s[k]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[k-i] = c
	}
	return string(b)
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

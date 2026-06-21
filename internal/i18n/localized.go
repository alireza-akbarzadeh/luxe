package i18n

import (
	"context"
	"encoding/json"
)

// LocalizedMap holds user-facing copy keyed by locale tag (en, es, fa).
type LocalizedMap map[string]string

// Resolve picks the best string for the request locale, then English, then fallback.
func (m LocalizedMap) Resolve(ctx context.Context, fallback string) string {
	if len(m) == 0 {
		return fallback
	}

	locale := LocaleFromContext(ctx)
	if text, ok := m[locale]; ok && text != "" {
		return text
	}
	if text, ok := m[DefaultLocale]; ok && text != "" {
		return text
	}

	for _, text := range m {
		if text != "" {
			return text
		}
	}

	return fallback
}

// MergeLocalized merges incoming admin edits with stored translations.
// fallback (usually English label field) always wins for the default locale.
func MergeLocalized(existing, incoming LocalizedMap, fallback string) LocalizedMap {
	out := LocalizedMap{}
	for locale, text := range existing {
		if text != "" {
			out[locale] = text
		}
	}
	for locale, text := range incoming {
		if text != "" {
			out[locale] = text
		}
	}
	if fallback != "" {
		out[DefaultLocale] = fallback
	}
	return out
}

// LocalizedMapFromFallback builds a map when only the default string is provided.
func LocalizedMapFromFallback(fallback string, incoming LocalizedMap) LocalizedMap {
	if len(incoming) == 0 {
		if fallback == "" {
			return nil
		}
		return LocalizedMap{DefaultLocale: fallback}
	}
	return MergeLocalized(nil, incoming, fallback)
}

// ParseLocalizedField unmarshals JSON that is either a plain string or a locale map.
func ParseLocalizedField(raw json.RawMessage) (LocalizedMap, string) {
	if len(raw) == 0 {
		return nil, ""
	}

	var asString string
	if err := json.Unmarshal(raw, &asString); err == nil {
		if asString == "" {
			return nil, ""
		}
		return LocalizedMap{DefaultLocale: asString}, asString
	}

	var asMap LocalizedMap
	if err := json.Unmarshal(raw, &asMap); err != nil {
		return nil, ""
	}

	fallback := asMap[DefaultLocale]
	if fallback == "" {
		for _, text := range asMap {
			if text != "" {
				fallback = text
				break
			}
		}
	}

	return asMap, fallback
}

// MarshalLocalizedField encodes a locale map for JSONB storage.
func MarshalLocalizedField(m LocalizedMap) json.RawMessage {
	if len(m) == 0 {
		return nil
	}
	raw, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return raw
}

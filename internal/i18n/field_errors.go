package i18n

import (
	"context"

	"github.com/alireza-akbarzadeh/luxe/internal/constants"
)

// TranslateFieldError localizes a validation tag or canonical English field error.
func TranslateFieldError(ctx context.Context, value string) string {
	if english, ok := constants.ValidationTagMessages[value]; ok {
		return Translate(ctx, english)
	}
	return Translate(ctx, value)
}

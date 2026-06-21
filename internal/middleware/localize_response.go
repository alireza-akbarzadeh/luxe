package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/alireza-akbarzadeh/luxe/internal/i18n"
	"github.com/gin-gonic/gin"
)

type responseBuffer struct {
	gin.ResponseWriter
	body   bytes.Buffer
	status int
}

func (w *responseBuffer) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(b)
}

func (w *responseBuffer) WriteHeader(statusCode int) {
	w.status = statusCode
}

// LocalizeResponse translates JSON `message` and validation `errors` fields on the way out.
// Covers typed dto.* responses written with c.JSON as well as utils.Response helpers.
func LocalizeResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		buffer := &responseBuffer{ResponseWriter: c.Writer}
		c.Writer = buffer
		c.Next()

		if buffer.body.Len() == 0 {
			return
		}

		status := buffer.status
		if status == 0 {
			status = http.StatusOK
		}

		contentType := buffer.Header().Get("Content-Type")
		if !strings.Contains(contentType, "application/json") {
			buffer.ResponseWriter.WriteHeader(status)
			_, _ = buffer.ResponseWriter.Write(buffer.body.Bytes())
			return
		}

		var payload map[string]any
		if err := json.Unmarshal(buffer.body.Bytes(), &payload); err != nil {
			buffer.ResponseWriter.WriteHeader(status)
			_, _ = buffer.ResponseWriter.Write(buffer.body.Bytes())
			return
		}

		ctx := c.Request.Context()
		if msg, ok := payload["message"].(string); ok && msg != "" {
			payload["message"] = i18n.Translate(ctx, msg)
		}
		localizeErrorsPayload(ctx, payload)

		out, err := json.Marshal(payload)
		if err != nil {
			buffer.ResponseWriter.WriteHeader(status)
			_, _ = buffer.ResponseWriter.Write(buffer.body.Bytes())
			return
		}

		buffer.Header().Set("Content-Length", strconv.Itoa(len(out)))
		buffer.ResponseWriter.WriteHeader(status)
		_, _ = buffer.ResponseWriter.Write(out)
	}
}

func localizeErrorsPayload(ctx context.Context, payload map[string]any) {
	raw, ok := payload["errors"]
	if !ok || raw == nil {
		return
	}

	switch errors := raw.(type) {
	case map[string]any:
		for field, value := range errors {
			if text, ok := value.(string); ok {
				errors[field] = i18n.TranslateFieldError(ctx, text)
			}
		}
	case []any:
		for _, item := range errors {
			row, ok := item.(map[string]any)
			if !ok {
				continue
			}
			if text, ok := row["message"].(string); ok {
				row["message"] = i18n.TranslateFieldError(ctx, text)
			}
		}
	}
}

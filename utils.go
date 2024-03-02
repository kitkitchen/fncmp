package fncmp

import (
	"strings"
)

func sanitizeHTML(html string) string {
	// sanitized := template.HTMLEscapeString(html)
	sanitized := strings.ReplaceAll(html, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\t", "")
	return sanitized
}

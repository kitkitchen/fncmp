package main

// Sanitize html
import (
	"html/template"
	"strings"
)

func SanitizeInput(input string) string {
	sanitized := template.HTMLEscapeString(input)
	sanitized = strings.ReplaceAll(sanitized, "&#34;", "")
	sanitized = strings.ReplaceAll(sanitized, "&#39;", "'")
	sanitized = strings.ReplaceAll(sanitized, "&#96;", "`")
	sanitized = strings.ReplaceAll(sanitized, "&#x60;", "`")
	sanitized = strings.ReplaceAll(sanitized, "&#x27;", "'")
	sanitized = strings.ReplaceAll(sanitized, "&#x2F;", "/")
	sanitized = strings.ReplaceAll(sanitized, "&#x2f;", "/")
	sanitized = strings.ReplaceAll(sanitized, "&#x3D;", "=")
	sanitized = strings.ReplaceAll(sanitized, "&#x3d;", "=")
	sanitized = strings.ReplaceAll(sanitized, "&#x3E;", ">")
	sanitized = strings.ReplaceAll(sanitized, "&#x3e;", ">")
	sanitized = strings.ReplaceAll(sanitized, "&#x3C;", "<")
	sanitized = strings.ReplaceAll(sanitized, "&#x3c;", "<")
	sanitized = strings.ReplaceAll(sanitized, "&#x22;", "")

	return sanitized
}
func SanitizeHTML(html string) string {
	// sanitized := template.HTMLEscapeString(html)
	sanitized := strings.ReplaceAll(html, "\n", "")
	sanitized = strings.ReplaceAll(sanitized, "\t", "")
	return sanitized
}

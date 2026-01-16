package security

import (
	"html"
	"net/url"
)

// EscapeHTML escapes HTML special characters to prevent XSS
func EscapeHTML(input string) string {
	return html.EscapeString(input)
}

// EscapeURL escapes URL special characters
func EscapeURL(input string) string {
	return url.QueryEscape(input)
}

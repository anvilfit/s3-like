package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

func swapDiacritics(s string) string {
	t := transform.Chain(
		norm.NFD,
		runes.Remove(runes.In(unicode.Mn)),
		norm.NFC,
	)
	result, _, _ := transform.String(t, s)
	return result
}

func SanitizeFilename(filename string) string {
	filename = swapDiacritics(filename)

	filename = strings.ReplaceAll(filename, " ", "-")

	re := regexp.MustCompile(`[^a-zA-Z0-9_.-]+`)
	cleaned := re.ReplaceAllString(filename, "")

	cleaned = strings.Trim(cleaned, ".")

	return cleaned
}

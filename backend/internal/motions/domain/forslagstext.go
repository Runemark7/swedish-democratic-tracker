package domain

import (
	"html"
	"regexp"
	"strings"
)

var lineEndRe = regexp.MustCompile(`[ \t\x{00a0}]+\n`)

// PlainForslagstext turns a förslagspunkt's text as Riksdagen serves it
// ("Riksdagen<br />a) godk&auml;nner &auml;ndrade ramar ...&#160;") into the
// plain text a reader sees: <br> becomes a line break, any other tag is
// dropped, and entities become their characters (å, ä, ö, non-breaking space).
// Soft hyphens, which are only hyphenation hints, are removed. The wording is
// otherwise untouched.
//
// The result is text, not HTML: an escaped "&lt;" comes out as a literal "<"
// and must be shown as text, never inserted as markup.
func PlainForslagstext(s string) string {
	s = brRe.ReplaceAllString(s, "\n")
	s = tagRe.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "­", "")
	s = lineEndRe.ReplaceAllString(s, "\n")
	return strings.Trim(s, " \t\r\n ")
}

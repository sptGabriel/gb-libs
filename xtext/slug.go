package xtext

import (
	"regexp"
	"strings"
)

var slugifyRegex = regexp.MustCompile(`[^a-zA-Z0-9]+`)

func Slugify(s string) string { return CustomSlugify(slugifyRegex, s) }

func CustomSlugify(regex *regexp.Regexp, s string) string {
	s = strings.Trim(s, " ")
	s = strings.Trim(s, "-")
	s = RemoveDiacritics(s)
	s = strings.ToLower(s)
	s = regex.ReplaceAllString(s, "-")
	return strings.Trim(s, "-")
}

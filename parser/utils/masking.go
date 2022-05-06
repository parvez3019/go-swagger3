package utils

import "strings"

type Masker struct {
	banStrings []string
}

func NewMasker(banStrings []string) *Masker {
	return &Masker{
		banStrings: banStrings,
	}
}

func (m *Masker) sanitizeString(s string) string {
	for _, bannedStrings := range m.banStrings {
		s = strings.Replace(s, bannedStrings, "", 1)
	}

	return s
}

func (m *Masker) ReplaceBackslash(origin string) string {
	return ReplaceBackslash(m.sanitizeString(origin))
}

func (m *Masker) AddSchemaRefLinkPrefix(name string) string {
	return AddSchemaRefLinkPrefix(m.sanitizeString(name))
}

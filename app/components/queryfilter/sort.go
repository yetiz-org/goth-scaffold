package queryfilter

import "strings"

// SortField represents a single sort directive parsed from the s= parameter.
type SortField struct {
	Field string
	Desc  bool // true = DESC, false = ASC
}

// ParseSort parses a sort string like "+year,-isrc" into a []SortField.
// Fields without a prefix are treated as ASC.
// Returns nil for empty input.
func ParseSort(s string) []SortField {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}

	parts := strings.Split(s, ",")
	result := make([]SortField, 0, len(parts))

	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}

		desc := false
		switch {
		case strings.HasPrefix(p, "-"):
			desc = true
			p = p[1:]
		case strings.HasPrefix(p, "+"):
			p = p[1:]
		}

		if p != "" {
			result = append(result, SortField{Field: p, Desc: desc})
		}
	}

	return result
}

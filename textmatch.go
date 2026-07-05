// Package textmatch provides text matching helpers.
package textmatch

import "strings"

// ContainsNormalized scores whether s contains substr after normalization.
// It returns the highest percent from a matching normalizer, or 0 when no
// normalizer matches.
func ContainsNormalized(s string, substr string, normalizers ...Normalizer) int {
	if substr == "" {
		return 100
	}
	if len(normalizers) == 0 {
		normalizedS, sBits := defaultNormalizeAudit(s)
		normalizedSubstr, substrBits := defaultNormalizeAudit(substr)
		if strings.Contains(normalizedS, normalizedSubstr) {
			return defaultNormalizerPercent(sBits | substrBits)
		}
		return 0
	}

	best := 0
	for _, normalizer := range normalizers {
		if normalizer.Normalize == nil {
			continue
		}
		if strings.Contains(normalizer.Normalize(s), normalizer.Normalize(substr)) && normalizer.Percent > best {
			best = normalizer.Percent
		}
	}
	return best
}

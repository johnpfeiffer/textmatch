package textmatch

import (
	"strings"
	"testing"
)

// runContainsCases is a shared table runner for the default-normalizer path of
// ContainsNormalized. Each table below targets a specific concern (happy path,
// a single normalization rule, percent scoring, negatives, edge cases) so a
// failure points squarely at the behavior that regressed.
func runContainsCases(t *testing.T, testCases []struct {
	name     string
	s        string
	substr   string
	expected int
}) {
	t.Helper()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ContainsNormalized(tc.s, tc.substr)
			if result != tc.expected {
				t.Errorf("ContainsNormalized(%q, %q) = %d, want %d", tc.s, tc.substr, result, tc.expected)
			}
		})
	}
}

func TestContainsNormalizedHappyPath(t *testing.T) {
	runContainsCases(t, []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{name: "simple contains", s: "hotdog", substr: "dog", expected: 100},
		{name: "prefix", s: "hotdoghouse", substr: "hot", expected: 100},
		{name: "middle", s: "abshotdef", substr: "hot", expected: 100},
		{name: "suffix", s: "doghousehot", substr: "hot", expected: 100},
		{name: "identical strings", s: "hotdog", substr: "hotdog", expected: 100},
	})
}

func TestContainsNormalizedCaseFolding(t *testing.T) {
	runContainsCases(t, []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{name: "needle uppercase", s: "hotdog", substr: "Dog", expected: 100},
		{name: "haystack mixed case", s: "hotdoG", substr: "dog", expected: 100},
		{name: "haystack uppercase", s: "HOTDOG", substr: "dog", expected: 100},
	})
}

// TestContainsNormalizedWhitespace covers whitespace handling that does NOT
// require NFKD (tabs, runs of spaces, trimming). Such matches score 100.
// Whitespace chars that NFKD rewrites (NBSP, em space, narrow NBSP) are covered
// in TestContainsNormalizedNormalizationRules because they score 95.
func TestContainsNormalizedWhitespace(t *testing.T) {
	runContainsCases(t, []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{name: "collapsed spaces", s: "my hot     dog", substr: "hot dog", expected: 100},
		{name: "tab to space", s: "my hot\tdog", substr: "hot dog", expected: 100},
		{name: "trimmed", s: " hot dog ", substr: "hot dog", expected: 100},
		{name: "newlines", s: "hot\n\ndog", substr: "hot dog", expected: 100},
	})
}

// TestContainsNormalizedPercentScoring isolates each 95-level normalization
// rule through ContainsNormalized and pins the resulting score, using a clean
// needle so the sBits|substrBits union is exercised one-sided. The 100
// baseline for case folding and whitespace is covered by TestDefaultNormalizerPercent.
func TestContainsNormalizedPercentScoring(t *testing.T) {
	runContainsCases(t, []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{name: "punctuation only -> 95", s: "hot-dog", substr: "hot dog", expected: 95},
		{name: "strip format only -> 95", s: "hot\u200Bdog", substr: "hotdog", expected: 95},
		{name: "NFKD only -> 95", s: "\uFF41\uFF42\uFF43", substr: "abc", expected: 95},
		{name: "strip mark only -> 95", s: "cafe\u0301", substr: "cafe", expected: 95},
		{name: "casefold plus punctuation -> 95", s: "Don\u2019t", substr: "dont", expected: 95},
	})
}

// TestContainsNormalizedNormalizationRules exercises one normalization rule per
// row and pins the exact score, so a rule's percent cannot silently regress.
// Most rules score 95; the ligature scores 100 because case folding itself
// resolves U+FB01 to "fi" (so NFKD never runs).
func TestContainsNormalizedNormalizationRules(t *testing.T) {
	runContainsCases(t, []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{name: "non-breaking space", s: "hot\u00A0dog", substr: "hot dog", expected: 95},
		{name: "narrow non-breaking space", s: "100\u202F000", substr: "100 000", expected: 95},
		{name: "em space", s: "hot\u2003dog", substr: "hot dog", expected: 95},
		{name: "en dash", s: "well\u2013known", substr: "well known", expected: 95},
		{name: "em dash", s: "thing\u2014other", substr: "thing other", expected: 95},
		{name: "minus sign", s: "x\u2212y", substr: "x y", expected: 95},
		{name: "soft hyphen", s: "nor\u00ADmal", substr: "normal", expected: 95},
		{name: "apostrophe", s: "don\u2019t", substr: "dont", expected: 95},
		{name: "single quotes", s: "\u2018quoted\u2019", substr: "quoted", expected: 95},
		{name: "double quotes", s: "\u201Cyes\u201D", substr: "yes", expected: 95},
		{name: "backticks", s: "`hello`", substr: "hello", expected: 95},
		{name: "precomposed vs decomposed", s: "caf\u00E9", substr: "caf\u0065\u0301", expected: 95},
		{name: "ligature (resolved by casefold)", s: "\uFB01nance", substr: "finance", expected: 100},
		{name: "spanish tilde", s: "jalape\u00F1o", substr: "jalapeno", expected: 95},
		{name: "fullwidth uppercase", s: "\uFF21\uFF22\uFF23", substr: "abc", expected: 95},
	})
}

func TestContainsNormalizedNegativeCases(t *testing.T) {
	runContainsCases(t, []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{name: "spaces prevent join", s: "my hot dog", substr: "hotdog", expected: 0},
		{name: "no match", s: "hot cat", substr: "dog", expected: 0},
		{name: "needle longer than haystack", s: "hot", substr: "hotdog", expected: 0},
		{name: "disjoint", s: "hello", substr: "world", expected: 0},
		{name: "empty haystack non-empty needle", s: "", substr: "dog", expected: 0},
	})
}

func TestContainsNormalizedEdgeCases(t *testing.T) {
	runContainsCases(t, []struct {
		name     string
		s        string
		substr   string
		expected int
	}{
		{name: "empty needle", s: "hot cat", substr: "", expected: 100},
		{name: "empty needle and haystack", s: "", substr: "", expected: 100},
	})
}

func firstWordLower(s string) string {
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return ""
	}
	return strings.ToLower(fields[0])
}

func TestContainsNormalizedCustomNormalizer(t *testing.T) {
	firstWord := Normalizer{Percent: 87, Normalize: firstWordLower}
	testCases := []struct {
		name        string
		s           string
		substr      string
		normalizers []Normalizer
		expected    int
	}{
		{name: "custom normalizer matches", s: "hot dog", substr: "hot cat", normalizers: []Normalizer{firstWord}, expected: 87},
		{name: "custom normalizer no match", s: "hot dog", substr: "cat dog", normalizers: []Normalizer{firstWord}, expected: 0},
		{name: "nil normalize is skipped", s: "hot dog", substr: "hot cat", normalizers: []Normalizer{{Percent: 50, Normalize: nil}, firstWord}, expected: 87},
		{name: "matching normalizer with percent 0 returns 0", s: "hot dog", substr: "hot cat", normalizers: []Normalizer{{Percent: 0, Normalize: firstWordLower}}, expected: 0},
		// Supplying any custom normalizer replaces the default entirely, so a
		// non-matching custom normalizer yields 0 even when the default would match
		// ("caf\u00e9" -> "cafe" contains "cafe"). ToUpper does not strip the mark.
		{name: "custom normalizer replaces default entirely", s: "caf\u00E9", substr: "cafe", normalizers: []Normalizer{{Percent: 80, Normalize: strings.ToUpper}}, expected: 0},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ContainsNormalized(tc.s, tc.substr, tc.normalizers...)
			if result != tc.expected {
				t.Errorf("ContainsNormalized(%q, %q) = %d, want %d", tc.s, tc.substr, result, tc.expected)
			}
		})
	}
}

func TestContainsNormalizedMultipleCustomNormalizers(t *testing.T) {
	lower := Normalizer{Percent: 75, Normalize: strings.ToLower}
	lowerDash := Normalizer{
		Percent: 92,
		Normalize: func(s string) string {
			return strings.ReplaceAll(strings.ToLower(s), "-", " ")
		},
	}
	upper := Normalizer{Percent: 60, Normalize: strings.ToUpper}
	nilHigh := Normalizer{Percent: 99} // nil Normalize, must be skipped

	t.Run("highest matching percent wins", func(t *testing.T) {
		if result := ContainsNormalized("hot-dog stand", "hot dog", nilHigh, lower, lowerDash); result != 92 {
			t.Errorf("got %d, want 92", result)
		}
	})

	t.Run("best match is order independent", func(t *testing.T) {
		orders := [][]Normalizer{
			{nilHigh, lower, lowerDash},
			{lowerDash, lower, nilHigh},
			{lower, nilHigh, lowerDash},
		}
		for i, order := range orders {
			if result := ContainsNormalized("hot-dog stand", "hot dog", order...); result != 92 {
				t.Errorf("order %d: got %d, want 92", i, result)
			}
		}
	})

	t.Run("no normalizer matches returns 0", func(t *testing.T) {
		if result := ContainsNormalized("hot-dog stand", "hot dog", lower, upper); result != 0 {
			t.Errorf("got %d, want 0", result)
		}
	})
}

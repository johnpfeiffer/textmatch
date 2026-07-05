package textmatch

import (
	"testing"

	"golang.org/x/text/transform"
)

func TestDefaultNormalizer(t *testing.T) {
	testCases := []struct {
		rule     string
		s        string
		expected string
	}{
		{rule: "empty", s: "", expected: ""},
		{rule: "casefold", s: "Dog", expected: "dog"},
		{rule: "NFKD compatibility", s: "\uFB01nance", expected: "finance"},
		{rule: "NFKD width", s: "\uFF21\uFF22\uFF23", expected: "abc"},
		{rule: "strip format zero width space", s: "hot\u200Bdog", expected: "hotdog"},
		{rule: "strip format zero width non-joiner", s: "hot\u200Cdog", expected: "hotdog"},
		{rule: "strip format zero width joiner", s: "hot\u200Ddog", expected: "hotdog"},
		{rule: "strip format byte order mark", s: "hot\uFEFFdog", expected: "hotdog"},
		{rule: "strip format soft hyphen", s: "nor\u00ADmal", expected: "normal"},
		{rule: "strip format word joiner", s: "hot\u2060dog", expected: "hotdog"},
		{rule: "strip format mongolian vowel separator", s: "hot\u180Edog", expected: "hotdog"},
		{rule: "strip mark precomposed", s: "caf\u00E9", expected: "cafe"},
		{rule: "strip mark decomposed", s: "cafe\u0301", expected: "cafe"},
		{rule: "strip mark spanish", s: "jalape\u00F1o", expected: "jalapeno"},
		{rule: "punctuation dash", s: "well\u2013known", expected: "well known"},
		{rule: "punctuation em dash", s: "thing\u2014other", expected: "thing other"},
		{rule: "punctuation minus", s: "x\u2212y", expected: "x y"},
		{rule: "punctuation slash", s: "hot/dog", expected: "hot dog"},
		{rule: "punctuation underscore", s: "hot_dog", expected: "hot dog"},
		{rule: "punctuation apostrophe", s: "don\u2019t", expected: "dont"},
		{rule: "punctuation single quotes", s: "\u2018quoted\u2019", expected: "quoted"},
		{rule: "punctuation double quotes", s: "\u201Cyes\u201D", expected: "yes"},
		{rule: "punctuation backticks", s: "`hello`", expected: "hello"},
		{rule: "whitespace spaces", s: "hot     dog", expected: "hot dog"},
		{rule: "whitespace tab", s: "hot\tdog", expected: "hot dog"},
		{rule: "whitespace trim", s: " hot \n dog ", expected: "hot dog"},
		{rule: "whitespace non-breaking space", s: "hot\u00A0dog", expected: "hot dog"},
		{rule: "whitespace narrow non-breaking space", s: "100\u202F000", expected: "100 000"},
		{rule: "whitespace em space", s: "hot\u2003dog", expected: "hot dog"},
	}
	for _, tc := range testCases {
		t.Run(tc.rule, func(t *testing.T) {
			assertStringEqual(t, tc.expected, DefaultNormalizer(tc.s))
		})
	}
}

func TestNormalizerTransformersShortDestination(t *testing.T) {
	punctuationTransformer := stringFuncTransformer{transform: canonicalizePunctuationString}
	punctuationTransformer.Reset()
	_, _, err := punctuationTransformer.Transform([]byte{}, []byte("hot/dog"), true)
	if err != transform.ErrShortDst {
		t.Error("\nExpected:", transform.ErrShortDst, "\nReceived:", err)
	}

	whitespaceTransformer := whitespaceCollapseTransformer{}
	whitespaceTransformer.Reset()
	_, _, err = whitespaceTransformer.Transform([]byte{}, []byte("hot dog"), true)
	if err != transform.ErrShortDst {
		t.Error("\nExpected:", transform.ErrShortDst, "\nReceived:", err)
	}
}

func TestWhitespaceCollapseTransformer(t *testing.T) {
	testCases := []struct {
		rule     string
		s        string
		expected string
	}{
		{rule: "empty", s: "", expected: ""},
		{rule: "no whitespace", s: "hotdog", expected: "hotdog"},
		{rule: "spaces", s: "hot     dog", expected: "hot dog"},
		{rule: "tabs", s: "hot\tdog", expected: "hot dog"},
		{rule: "newlines", s: "hot\n\ndog", expected: "hot dog"},
		{rule: "mixed whitespace", s: "hot \t\n dog", expected: "hot dog"},
		{rule: "leading whitespace", s: " \t hot dog", expected: "hot dog"},
		{rule: "trailing whitespace", s: "hot dog \t ", expected: "hot dog"},
		{rule: "only whitespace", s: " \t\n ", expected: ""},
		{rule: "non-breaking space", s: "hot\u00A0dog", expected: "hot dog"},
		{rule: "narrow non-breaking space", s: "100\u202F000", expected: "100 000"},
		{rule: "em space", s: "hot\u2003dog", expected: "hot dog"},
	}
	for _, tc := range testCases {
		t.Run(tc.rule, func(t *testing.T) {
			result, _, err := transform.String(whitespaceCollapseTransformer{}, tc.s)
			if err != nil {
				t.Fatal(err)
			}
			assertStringEqual(t, tc.expected, result)
		})
	}
}

// TestDefaultNormalizerPercent pins the 100-vs-95 weighting of
// defaultNormalizerPercent directly, independent of ContainsNormalized.
// Case folding and whitespace collapse keep a match at 100; any of NFKD,
// strip-format, strip-mark, or punctuation drops it to 95.
func TestDefaultNormalizerPercent(t *testing.T) {
	testCases := []struct {
		name     string
		bits     uint32
		expected int
	}{
		{name: "no normalization", bits: 0, expected: 100},
		{name: "casefold only", bits: normalizerCaseFold, expected: 100},
		{name: "whitespace only", bits: normalizerWhitespace, expected: 100},
		{name: "casefold and whitespace", bits: normalizerCaseFold | normalizerWhitespace, expected: 100},
		{name: "NFKD only", bits: normalizerNFKD, expected: 95},
		{name: "strip format only", bits: normalizerStripFormat, expected: 95},
		{name: "strip mark only", bits: normalizerStripMark, expected: 95},
		{name: "punctuation only", bits: normalizerPunctuation, expected: 95},
		{name: "casefold plus punctuation", bits: normalizerCaseFold | normalizerPunctuation, expected: 95},
		{name: "all rules combined", bits: normalizerCaseFold | normalizerNFKD | normalizerStripFormat | normalizerStripMark | normalizerPunctuation | normalizerWhitespace, expected: 95},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if result := defaultNormalizerPercent(tc.bits); result != tc.expected {
				t.Errorf("defaultNormalizerPercent(%#b) = %d, want %d", tc.bits, result, tc.expected)
			}
		})
	}
}

func assertStringEqual(t *testing.T, expected string, result string) {
	t.Helper()
	if expected != result {
		t.Error("\nExpected:", expected, "\nReceived:", result)
	}
}

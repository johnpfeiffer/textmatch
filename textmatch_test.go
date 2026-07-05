package textmatch

import (
	"fmt"
	"strings"
	"testing"
)

func TestContainsNormalized(t *testing.T) {
	testCases := []struct {
		s        string
		substr   string
		expected int
	}{
		{s: "hotdog", substr: "dog", expected: 100},
		{s: "hotdog", substr: "Dog", expected: 100},
		{s: "hotdoG", substr: "dog", expected: 100},
		{s: "hotdoghouse", substr: "hot", expected: 100},
		{s: "my hot     dog", substr: "hot dog", expected: 100},
		{s: "my hot\tdog", substr: "hot dog", expected: 100},
		{s: "un caf\u00E9 chaud", substr: "caf\u00E9", expected: 95},
		{s: "un cafe\u0301 chaud", substr: "caf\u00E9", expected: 95},
		{s: "un caf\u00E9 chaud", substr: "cafe", expected: 95},
		{s: "deja vu", substr: "d\u00E9j\u00E0", expected: 95},
		{s: "my hot dog", substr: "hotdog", expected: 0},
		{s: "hot cat", substr: "dog", expected: 0},
		{s: "hot cat", substr: "", expected: 100},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%#v contains normalized %#v", tc.s, tc.substr), func(t *testing.T) {
			result := ContainsNormalized(tc.s, tc.substr)
			if tc.expected != result {
				t.Error("\nExpected:", tc.expected, "\nReceived:", result)
			}
		})
	}
}

func TestContainsNormalizedExamples(t *testing.T) {
	testCases := []struct {
		haystack string
		needle   string
		expected bool
	}{
		{haystack: "hot\u00A0dog", needle: "hot dog", expected: true},
		{haystack: "100\u202F000", needle: "100 000", expected: true},
		{haystack: "hot\u2003dog", needle: "hot dog", expected: true},
		{haystack: "well\u2013known", needle: "well known", expected: true},
		{haystack: "thing\u2014other", needle: "thing other", expected: true},
		{haystack: "x\u2212y", needle: "x y", expected: true},
		{haystack: "nor\u00ADmal", needle: "normal", expected: true},
		{haystack: "don\u2019t", needle: "dont", expected: true},
		{haystack: "\u2018quoted\u2019", needle: "quoted", expected: true},
		{haystack: "\u201Cyes\u201D", needle: "yes", expected: true},
		{haystack: "`hello`", needle: "hello", expected: true},
		{haystack: "caf\u00E9", needle: "caf\u0065\u0301", expected: true},
		{haystack: "\uFB01nance", needle: "finance", expected: true},
		{haystack: "jalape\u00F1o", needle: "jalapeno", expected: true},
		{haystack: "\uFF21\uFF22\uFF23", needle: "abc", expected: true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%#v contains %#v", tc.haystack, tc.needle), func(t *testing.T) {
			result := ContainsNormalized(tc.haystack, tc.needle) > 0
			if tc.expected != result {
				t.Error("\nExpected:", tc.expected, "\nReceived:", result)
			}
		})
	}
}

func TestContainsNormalizedWithCustomNormalizer(t *testing.T) {
	firstWordLower := Normalizer{
		Percent: 87,
		Normalize: func(s string) string {
			fields := strings.Fields(s)
			if len(fields) == 0 {
				return ""
			}
			return strings.ToLower(fields[0])
		},
	}
	result := ContainsNormalized("hot dog", "hot cat", firstWordLower)
	if result != 87 {
		t.Error("\nExpected:", 87, "\nReceived:", result)
	}
	result = ContainsNormalized("hot dog", "cat dog", firstWordLower)
	if result != 0 {
		t.Error("\nExpected:", 0, "\nReceived:", result)
	}
}

func TestContainsNormalizedWithMultipleCustomNormalizers(t *testing.T) {
	normalizers := []Normalizer{
		{Percent: 99},
		{
			Percent: 75,
			Normalize: func(s string) string {
				return strings.ToLower(s)
			},
		},
		{
			Percent: 92,
			Normalize: func(s string) string {
				return strings.ReplaceAll(strings.ToLower(s), "-", " ")
			},
		},
	}
	result := ContainsNormalized("hot-dog stand", "hot dog", normalizers...)
	if result != 92 {
		t.Error("\nExpected:", 92, "\nReceived:", result)
	}
}

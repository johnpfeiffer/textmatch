package textmatch

import (
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// Normalizer defines a normalization strategy and the percent returned when it
// matches.
type Normalizer struct {
	Percent   int
	Normalize func(string) string
}

const (
	normalizerCaseFold uint32 = 1 << iota
	normalizerNFKD
	normalizerStripFormat
	normalizerStripMark
	normalizerPunctuation
	normalizerWhitespace
)

var defaultNormalizerWeights = []struct {
	bit     uint32
	percent int
}{
	{bit: normalizerCaseFold, percent: 100},
	{bit: normalizerNFKD, percent: 95},
	{bit: normalizerStripFormat, percent: 95},
	{bit: normalizerStripMark, percent: 95},
	{bit: normalizerPunctuation, percent: 95},
	{bit: normalizerWhitespace, percent: 100},
}

// DefaultNormalizer lowercases strings, strips invisible format characters and
// diacritics, canonicalizes punctuation, and collapses whitespace.
func DefaultNormalizer(s string) string {
	normalized, _ := defaultNormalizeAudit(s)
	return normalized
}

func defaultNormalizerPercent(bits uint32) int {
	percent := 100
	for _, weight := range defaultNormalizerWeights {
		if bits&weight.bit != 0 && weight.percent < percent {
			percent = weight.percent
		}
	}
	return percent
}

func defaultNormalizeAudit(s string) (string, uint32) {
	var bits uint32
	normalizer := transform.Chain(
		auditTransform(cases.Fold(), normalizerCaseFold, &bits),
		auditTransform(norm.NFKD, normalizerNFKD, &bits),
		auditTransform(runes.Remove(runes.In(unicode.Cf)), normalizerStripFormat, &bits),
		auditTransform(runes.Remove(runes.In(unicode.M)), normalizerStripMark, &bits),
		auditTransform(stringFuncTransformer{transform: canonicalizePunctuationString}, normalizerPunctuation, &bits),
		auditTransform(whitespaceCollapseTransformer{}, normalizerWhitespace, &bits),
	)
	result, _, _ := transform.String(normalizer, s)
	return result, bits
}

type auditedTransformer struct {
	transformer transform.Transformer
	bit         uint32
	bits        *uint32
}

func auditTransform(transformer transform.Transformer, bit uint32, bits *uint32) transform.Transformer {
	return &auditedTransformer{transformer: transformer, bit: bit, bits: bits}
}

func (t *auditedTransformer) Reset() {
	t.transformer.Reset()
}

func (t *auditedTransformer) Transform(dst []byte, src []byte, atEOF bool) (int, int, error) {
	nDst, nSrc, err := t.transformer.Transform(dst, src, atEOF)
	if nSrc > 0 || nDst > 0 {
		input := string(src[:nSrc])
		output := string(dst[:nDst])
		if !runeStringsEqual(input, output) {
			*t.bits |= t.bit
		}
	}
	return nDst, nSrc, err
}

func runeStringsEqual(a string, b string) bool {
	aRunes := []rune(a)
	bRunes := []rune(b)
	if len(aRunes) != len(bRunes) {
		return false
	}
	for i, r := range aRunes {
		if r != bRunes[i] {
			return false
		}
	}
	return true
}

type stringFuncTransformer struct {
	transform func(string) string
}

func (stringFuncTransformer) Reset() {}

func (t stringFuncTransformer) Transform(dst []byte, src []byte, atEOF bool) (int, int, error) {
	result := t.transform(string(src))
	if len(result) > len(dst) {
		return 0, 0, transform.ErrShortDst
	}
	copy(dst, result)
	return len(result), len(src), nil
}

func canonicalizePunctuationString(s string) string {
	var builder strings.Builder
	for _, r := range s {
		switch {
		case isDashLike(r), isSlashLike(r), isUnderscoreLike(r):
			builder.WriteRune(' ')
		case isQuoteLike(r):
			continue
		default:
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

func isDashLike(r rune) bool {
	return unicode.Is(unicode.Pd, r) || r == '\u2212'
}

func isSlashLike(r rune) bool {
	switch r {
	case '/', '\\', '\u2044', '\u2215', '\u29F8', '\uFF0F', '\uFF3C':
		return true
	default:
		return false
	}
}

func isUnderscoreLike(r rune) bool {
	switch r {
	case '_',
		'\u203F',
		'\u2040',
		'\uFE33',
		'\uFE34',
		'\uFE4D',
		'\uFE4E',
		'\uFE4F',
		'\uFF3F':
		return true
	default:
		return false
	}
}

func isQuoteLike(r rune) bool {
	switch r {
	case '\'',
		'"',
		'`',
		'\u00AB',
		'\u00BB',
		'\u2018',
		'\u2019',
		'\u201A',
		'\u201B',
		'\u201C',
		'\u201D',
		'\u201E',
		'\u201F',
		'\u2032',
		'\u2033',
		'\u2039',
		'\u203A',
		'\u275B',
		'\u275C',
		'\u275D',
		'\u275E',
		'\u276E',
		'\u276F',
		'\u300C',
		'\u300D',
		'\u300E',
		'\u300F',
		'\u301D',
		'\u301E',
		'\u301F',
		'\uFE41',
		'\uFE42',
		'\uFE43',
		'\uFE44',
		'\uFF02',
		'\uFF07':
		return true
	default:
		return false
	}
}

type whitespaceCollapseTransformer struct{}

func (whitespaceCollapseTransformer) Reset() {}

func (whitespaceCollapseTransformer) Transform(dst []byte, src []byte, atEOF bool) (int, int, error) {
	var builder strings.Builder
	pendingWhitespace := false
	for _, r := range string(src) {
		if unicode.IsSpace(r) {
			if builder.Len() > 0 {
				pendingWhitespace = true
			}
			continue
		}
		if pendingWhitespace {
			builder.WriteRune(' ')
			pendingWhitespace = false
		}
		builder.WriteRune(r)
	}
	result := builder.String()
	if len(result) > len(dst) {
		return 0, 0, transform.ErrShortDst
	}
	copy(dst, result)
	return len(result), len(src), nil
}

# textmatch

Text matching helpers

From the Needle in a Haystack concept.

## Currently

`ContainsNormalized` - score whether one string contains another using weighted normalizers

*(normalization as a pragmatic addition)*

- `DefaultNormalizer` - casefold, decompose compatibility characters, strip format/mark characters, canonicalize punctuation, and collapse whitespace

## Future

- Find: Position/span extraction
- FindAll
- Near: Proximity
- Coverage
- Multi-needle search

# Tests

`go test ./...`

`go test -cover ./...`

`go test -covermode=count -coverprofile=count.out .`

`go tool cover -html=count.out`

> cool heat map , <https://go.dev/blog/cover>

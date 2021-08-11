package gocsv

type Config struct {
	// Holder is the struct to Marshal/unmarshal.
	Holder interface{}

	// FailIfUnmatchedStructTags indicates whether it is considered an
	// error when there is an unmatched struct tag.
	FailIfUnmatchedStructTags bool

	// FailIfDoubleHeaderNames indicates whether it is considered an
	// error when a header name is repeated in the csv header.
	FailIfDoubleHeaderNames bool

	// ShouldAlignDuplicateHeadersWithStructFieldOrder indicates
	// whether we should align duplicate CSV headers per their
	// alignment in the struct definition.
	ShouldAlignDuplicateHeadersWithStructFieldOrder bool
}

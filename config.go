package gocsv

import (
	"reflect"
)

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

func (c *Config) validate() (*validConfig, error) {
	return &validConfig{
		Config: *c,
		outType: reflect.TypeOf(c.Holder),
	}, nil
}

type validConfig struct {
	Config

	outType reflect.Type
}

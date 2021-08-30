package gocsv

import (
	"reflect"
)

type Config struct {
	// Holder is the type of struct to marshal from/unmarshal info.
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

// validate ensures that a struct was used to create the Unmarshaller, and validates
// CSV headers against the CSV tags in the struct.
func (c *Config) validate(headers []string) (*validConfig, error) {
	concreteType := reflect.TypeOf(c.Holder)
	if concreteType.Kind() == reflect.Ptr {
		concreteType = concreteType.Elem()
	}
	if err := ensureOutInnerType(concreteType); err != nil {
		return nil, err
	}
	structInfo := getStructInfo(concreteType) // Get struct info to get CSV annotations.
	if len(structInfo.Fields) == 0 {
		return nil, ErrNoStructTags
	}
	csvHeadersLabels := make([]*fieldInfo, len(headers)) // Used to store the corresponding header <-> position in CSV
	headerCount := map[string]int{}
	for i, csvColumnHeader := range headers {
		curHeaderCount := headerCount[csvColumnHeader]
		if fieldInfo := getCSVFieldPosition(csvColumnHeader, structInfo, curHeaderCount); fieldInfo != nil {
			csvHeadersLabels[i] = fieldInfo
			if c.ShouldAlignDuplicateHeadersWithStructFieldOrder {
				curHeaderCount++
				headerCount[csvColumnHeader] = curHeaderCount
			}
		}
	}

	if c.FailIfUnmatchedStructTags {
		if err := maybeMissingStructFields(structInfo.Fields, headers); err != nil {
			return nil, err
		}
	}

	if c.FailIfDoubleHeaderNames {
		if err := maybeDoubleHeaderNames(headers); err != nil {
			return nil, err
		}
	}

	return &validConfig{
		Config:                 *c,
		outType:                reflect.TypeOf(c.Holder),
		headers:                headers,
		structInfo:             structInfo,
		fieldInfoMap:           csvHeadersLabels,
		mismatchedHeaders:      mismatchHeaderFields(structInfo.Fields, headers),
		mismatchedStructFields: mismatchStructFields(structInfo.Fields, headers),
	}, nil
}

// validConfig is a Config which has been validated and contains
// metadata about the output struct type.
type validConfig struct {
	Config

	outType reflect.Type

	// headers is a slice of header names in file order.
	headers []string

	structInfo             *structInfo
	fieldInfoMap           []*fieldInfo
	mismatchedHeaders      []string
	mismatchedStructFields []string
}

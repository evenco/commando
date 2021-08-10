// Copyright 2014 Jonathan Picques. All rights reserved.
// Use of this source code is governed by a MIT license
// The license can be found in the LICENSE file.

// The GoCSV package aims to provide easy CSV serialization and deserialization to the golang programming language

package gocsv

import (
)

// FailIfUnmatchedStructTags indicates whether it is considered an error when there is an unmatched
// struct tag.
var FailIfUnmatchedStructTags = false

// FailIfDoubleHeaderNames indicates whether it is considered an error when a header name is repeated
// in the csv header.
var FailIfDoubleHeaderNames = false

// ShouldAlignDuplicateHeadersWithStructFieldOrder indicates whether we should align duplicate CSV
// headers per their alignment in the struct definition.
var ShouldAlignDuplicateHeadersWithStructFieldOrder = false


package commando

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrNoStructTags = errors.New("no csv struct tags found")
)

func mismatchStructFields(structInfo []fieldInfo, headers []string) []string {
	missing := make([]string, 0)
	if len(structInfo) == 0 {
		return missing
	}

	headerMap := make(map[string]struct{}, len(headers))
	for idx := range headers {
		headerMap[headers[idx]] = struct{}{}
	}

	for _, info := range structInfo {
		found := false
		for _, key := range info.keys {
			if _, ok := headerMap[key]; ok {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, info.keys...)
		}
	}
	return missing
}

func mismatchHeaderFields(structInfo []fieldInfo, headers []string) []string {
	missing := make([]string, 0)
	if len(headers) == 0 {
		return missing
	}

	keyMap := make(map[string]struct{})
	for _, info := range structInfo {
		for _, key := range info.keys {
			keyMap[key] = struct{}{}
		}
	}

	for _, header := range headers {
		if _, ok := keyMap[header]; !ok {
			missing = append(missing, header)
		}
	}
	return missing
}

// Check that no header name is repeated twice
func maybeDoubleHeaderNames(headers []string) error {
	headerMap := make(map[string]bool, len(headers))
	for _, v := range headers {
		if _, ok := headerMap[v]; ok {
			return fmt.Errorf("repeated header name: %v", v)
		}
		headerMap[v] = true
	}
	return nil
}

// Check if the outInnerType is of type struct
func ensureOutInnerType(outInnerType reflect.Type) error {
	switch outInnerType.Kind() {
	case reflect.Struct:
		return nil
	}
	return fmt.Errorf("cannot use %q, only struct supported", outInnerType)
}

func getCSVFieldPosition(key string, structInfo *structInfo, curHeaderCount int) *fieldInfo {
	matchedFieldCount := 0
	for _, field := range structInfo.Fields {
		if field.matchesKey(key) {
			if matchedFieldCount >= curHeaderCount {
				return &field
			}
			matchedFieldCount++
		}
	}
	return nil
}

func createNewOutInner(outInnerWasPointer bool, outInnerType reflect.Type) reflect.Value {
	if outInnerWasPointer {
		return reflect.New(outInnerType)
	}
	return reflect.New(outInnerType).Elem()
}

func setInnerField(outInner *reflect.Value, outInnerWasPointer bool, index []int, value string, omitEmpty bool) error {
	oi := *outInner
	if outInnerWasPointer {
		// initialize nil pointer
		if oi.IsNil() {
			setField(oi, "", omitEmpty)
		}
		oi = outInner.Elem()
	}
	// because pointers can be nil need to recurse one index at a time and perform nil check
	if len(index) > 1 {
		nextField := oi.Field(index[0])
		return setInnerField(&nextField, nextField.Kind() == reflect.Ptr, index[1:], value, omitEmpty)
	}
	return setField(oi.FieldByIndex(index), value, omitEmpty)
}

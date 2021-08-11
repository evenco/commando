package gocsv

import (
	"encoding/csv"
	"fmt"
	"reflect"
)

// Unmarshaller is a CSV to struct unmarshaller.
type Unmarshaller struct {
	config                 *validConfig
	reader                 *csv.Reader}

// NewUnmarshaller creates an unmarshaller from a csv.Reader and a struct.
func (c *Config) NewUnmarshaller(reader *csv.Reader) (*Unmarshaller, error) {
	headers, err := reader.Read()
	if err != nil {
		return nil, err
	}

	vc, err := c.validate(headers)
	if err != nil {
		return nil, err
	}

	um := &Unmarshaller{
		reader: reader,
		config: vc,
	}

	return um, nil
}

// Read returns an interface{} whose runtime type is the same as the struct that
// was used to create the Unmarshaller.
func (um *Unmarshaller) Read() (interface{}, error) {
	value, _, err := um.ReadUnmatched()
	return value, err
}

// ReadUnmatched is same as Read(), but returns a map of the columns that didn't match a field in the struct
func (um *Unmarshaller) ReadUnmatched() (interface{}, map[string]string, error) {
	row, err := um.reader.Read()
	if err != nil {
		return nil, nil, err
	}
	return um.unmarshalRow(row)
}

// unmarshalRow converts a CSV row to a struct, based on CSV struct tags.
// If unmatched is non nil, it is populated with any columns that don't map to a struct field
func (um *Unmarshaller) unmarshalRow(row []string) (interface{}, map[string]string, error) {
	unmatched := make(map[string]string)

	isPointer := false
	concreteOutType := um.config.outType
	if um.config.outType.Kind() == reflect.Ptr {
		isPointer = true
		concreteOutType = concreteOutType.Elem()
	}
	outValue := createNewOutInner(isPointer, concreteOutType)
	for j, csvColumnContent := range row {
		if j < len(um.config.fieldInfoMap) && um.config.fieldInfoMap[j] != nil {
			fieldInfo := um.config.fieldInfoMap[j]
			if err := setInnerField(&outValue, isPointer, fieldInfo.IndexChain, csvColumnContent, fieldInfo.omitEmpty); err != nil { // Set field of struct
				return nil, nil, fmt.Errorf("cannot assign field at %v to %s through index chain %v: %v", j, outValue.Type(), fieldInfo.IndexChain, err)
			}
		}
	}
	return outValue.Interface(), unmatched, nil
}

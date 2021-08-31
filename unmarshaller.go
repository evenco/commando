package commando

import (
	"errors"
	"fmt"
	"io"
	"reflect"
)

// Unmarshaller is a CSV to struct unmarshaller.
type Unmarshaller struct {
	config *validConfig
	line   int
	reader Reader
}

// NewUnmarshaller is a convenience function which allocates and
// returns a new Unmarshaller.
func NewUnmarshaller(holder interface{}, reader Reader) (*Unmarshaller, error) {
	return (&Config{Holder: holder}).NewUnmarshaller(reader)
}

// NewUnmarshaller creates an unmarshaller from a Reader and a struct.
func (c *Config) NewUnmarshaller(reader Reader) (*Unmarshaller, error) {
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
		line:   1,
	}

	return um, nil
}

// Read returns an interface{} whose runtime type is the same as the
// struct that was used to create the Unmarshaller.
func (um *Unmarshaller) Read() (interface{}, error) {
	value, _, err := um.ReadUnmatched()
	return value, err
}

// ReadAll returns a slice of structs.
func (um *Unmarshaller) ReadAll(onError func(err error) error) (interface{}, error) {
	out := reflect.MakeSlice(reflect.SliceOf(um.config.outType), 0, 0)
	err := ReadAllCallback(um, func(rec interface{}) error {
		out = reflect.Append(out, reflect.ValueOf(rec))
		return nil
	}, onError)

	return out.Interface(), err
}

// ReadAllCallback calls cb for every record Read() from um.
func ReadAllCallback(um *Unmarshaller, onSuccess func(interface{}) error, onError func(error) error) error {
	for {
		rec, err := um.Read()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			if handlerErr := onError(err); handlerErr != nil {
				return handlerErr
			}
		}

		if err := onSuccess(rec); err != nil {
			return err
		}
	}
	return nil
}

// wrapLine wraps err, including the line the error occurred on.
func wrapLine(err error, line int) error {
	if err != nil {
		return fmt.Errorf("on line %d: %w", line, err)
	}
	return err
}

// ReadUnmatched is same as Read(), but returns a map of the columns
// that didn't match a field in the struct.
func (um *Unmarshaller) ReadUnmatched() (interface{}, map[string]string, error) {
	row, err := um.reader.Read()
	if err != nil {
		return nil, nil, err
	}
	out, unmatched, err := um.unmarshalRow(row)
	if err != nil {
		um.line++
	}
	return out, unmatched, wrapLine(err, um.line)
}

// createNew allocates and returns a new holder to unmarshal data
// into.
func (um *Unmarshaller) createNew() (reflect.Value, bool) {
	isPointer := false
	concreteOutType := um.config.outType
	if um.config.outType.Kind() == reflect.Ptr {
		isPointer = true
		concreteOutType = concreteOutType.Elem()
	}
	outValue := createNewOutInner(isPointer, concreteOutType)
	return outValue, isPointer
}

// unmarshalRow converts a CSV row to a struct, based on CSV struct
// tags.
func (um *Unmarshaller) unmarshalRow(row []string) (interface{}, map[string]string, error) {
	unmatched := make(map[string]string)

	outValue, isPointer := um.createNew()

	for j, csvColumnContent := range row {
		if j < len(um.config.fieldInfoMap) && um.config.fieldInfoMap[j] != nil {
			fieldInfo := um.config.fieldInfoMap[j]
			if err := setInnerField(&outValue, isPointer, fieldInfo.IndexChain, csvColumnContent, fieldInfo.omitEmpty); err != nil { // Set field of struct
				return nil, nil, fmt.Errorf("cannot assign field at %v to %T through index chain %v: %v", j, outValue, fieldInfo.IndexChain, err)
			}
		}
	}
	return outValue.Interface(), unmatched, nil
}

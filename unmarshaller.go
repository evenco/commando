package commando

import (
	"context"
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
	row, err := um.reader.Read()
	if err != nil {
		return nil, err
	}
	out, err := um.unmarshalRow(row)
	if err != nil {
		um.line++
	}
	return out, wrapLine(err, um.line)
}

// ReadAll returns a slice of structs.
func (um *Unmarshaller) ReadAll(ctx context.Context, onError func(ctx context.Context, err error) error) (interface{}, error) {
	out := reflect.MakeSlice(reflect.SliceOf(um.config.outType), 0, 0)
	err := um.ReadAllCallback(ctx, func(_ context.Context, rec interface{}) error {
		out = reflect.Append(out, reflect.ValueOf(rec))
		return nil
	}, onError)

	return out.Interface(), err
}

// ReadAllCallback calls onSuccess for every record Read() from um.
//
// If Read() returns an error, it's passed to onError(), which decides
// whether to continue processing or stop.  If onError() returns nil,
// processing continues; if it returns an error, processing stops and
// its error (not the one returned by Read()) is returned.
//
// If onSuccess() returns an error, processing stops and its error is
// returned.
func (um *Unmarshaller) ReadAllCallback(ctx context.Context,
	onSuccess func(context.Context, interface{}) error,
	onError func(context.Context, error) error,
) error {
	for {
		rec, err := um.Read()
		if errors.Is(err, io.EOF) {
			return nil
		} else if err != nil {
			if handlerErr := onError(ctx, err); handlerErr != nil {
				return handlerErr
			}

			continue
		}

		if err := onSuccess(ctx, rec); err != nil {
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
func (um *Unmarshaller) unmarshalRow(row []string) (interface{}, error) {
	outValue, isPointer := um.createNew()

	for j, csvColumnContent := range row {
		if j < len(um.config.fieldInfoMap) && um.config.fieldInfoMap[j] != nil {
			fieldInfo := um.config.fieldInfoMap[j]
			if err := setInnerField(&outValue, isPointer, fieldInfo.IndexChain, csvColumnContent, fieldInfo.omitEmpty); err != nil { // Set field of struct
				return nil, fmt.Errorf("cannot assign field at %v to %T through index chain %v: %v", j, outValue, fieldInfo.IndexChain, err)
			}
		}
	}
	return outValue.Interface(), nil
}

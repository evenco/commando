package commando

import (
	"encoding/csv"
	"fmt"
	"reflect"
)

// Marshaller is a CSV to struct marshaller.
type Marshaller struct {
	config *validConfig
	writer *csv.Writer
}

// NewMarshaller is a convenience function which allocates and
// returns a new Marshaller.
func NewMarshaller(holder interface{}, writer *csv.Writer) (*Marshaller, error) {
	return (&Config{Holder: holder}).NewMarshaller(writer)
}

// NewMarshaller creates a marshaller from a csv.Writer.  The CSV
// header will be immediately written to writer.
func (c *Config) NewMarshaller(writer *csv.Writer) (*Marshaller, error) {
	vc, err := c.validate(nil)
	if err != nil {
		return nil, err
	}

	m := &Marshaller{
		writer: writer,
		config: vc,
	}

	if err := m.writeHeaders(); err != nil {
		return nil, err
	}
	return m, nil
}

func (m *Marshaller) writeHeaders() error {
	return m.writer.Write(m.config.structInfo.headers())
}

func (m *Marshaller) Write(record interface{}) error {
	if reflect.TypeOf(record) != m.config.outType {
		return fmt.Errorf("Expected %q, but got %q", m.config.outType, reflect.TypeOf(record))
	}

	inValue, inType := getConcreteReflectValueAndType(record) // Get the concrete type
	inInnerWasPointer := inType.Kind() == reflect.Ptr

	csvHeadersLabels := make([]string, len(m.config.structInfo.Fields))
	for i, fieldInfo := range m.config.structInfo.Fields {
		inInnerFieldValue, err := getInnerField(inValue, inInnerWasPointer, fieldInfo.IndexChain) // Get the correct field header <-> position
		if err != nil {
			return err
		}
		csvHeadersLabels[i] = inInnerFieldValue
	}
	return m.writer.Write(csvHeadersLabels)
}

// WriteAll writes every element of values as CSV.
//
// values must be a slice of elements of the configured Holder for
// this Marshaller.
//
// Flush() must be called when writing is complete.
func (m *Marshaller) WriteAll(values interface{}) error {
	v := reflect.ValueOf(values)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	if v.Kind() != reflect.Slice && reflect.TypeOf(v.Elem()) != m.config.outType {
		return fmt.Errorf("Expected []%s, but got %T", m.config.outType, values)
	}

	n := v.Len()
	for i := 0; i < n; i++ {
		sv := v.Index(i).Interface()
		if err := m.Write(sv); err != nil {
			return err
		}
	}

	return nil
}

func (m *Marshaller) Flush() error {
	m.writer.Flush()
	return m.writer.Error()
}

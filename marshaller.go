package gocsv

import (
	"encoding/csv"
	"reflect"
)

// Marshaller is a CSV to struct marshaller.
type Marshaller struct {
	config *validConfig
	writer *csv.Writer
}

// NewMarshaller creates a marshaller from a csv.Writer.
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

	inValue, inType := getConcreteReflectValueAndType(record) // Get the concrete type
	if err := ensureStructOrPtr(inType); err != nil {
		return err
	}
	inInnerWasPointer := inType.Kind() == reflect.Ptr

	csvHeadersLabels := make([]string, len(m.config.structInfo.Fields))
	for j, fieldInfo := range m.config.structInfo.Fields {
		// csvHeadersLabels[j] = ""
		inInnerFieldValue, err := getInnerField(inValue, inInnerWasPointer, fieldInfo.IndexChain) // Get the correct field header <-> position
		if err != nil {
			return err
		}
		csvHeadersLabels[j] = inInnerFieldValue
	}
	return m.writer.Write(csvHeadersLabels)
}

func (m *Marshaller) Flush() error {
	m.writer.Flush()
	return m.writer.Error()
}

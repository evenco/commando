package commando

import (
	"encoding/csv"
	"io"
	"strings"
	"testing"
)

func TestUnmarshallerLongRow(t *testing.T) {
	type sample struct {
		FieldA string `csv:"field_a"`
		FieldB string `csv:"field_b"`
	}
	const csvContents = `field_a,field_b
a,b
c,d,e
`

	reader := csv.NewReader(strings.NewReader(csvContents))
	reader.FieldsPerRecord = -1
	c := &Config{Holder: sample{}}
	um, err := c.NewUnmarshaller(reader)
	if err != nil {
		t.Fatalf("Error calling NewUnmarshaller: %#v", err)
	}

	obj, err := um.Read()
	if err != nil {
		t.Fatalf("Error calling Read(): %#v", err)
	}
	if obj.(sample).FieldA != "a" || obj.(sample).FieldB != "b" {
		t.Fatalf("Unepxected result from Read(): %#v", obj)
	}

	obj, err = um.Read()
	if err != nil {
		t.Fatalf("Error calling Read(): %#v", err)
	}
	if obj.(sample).FieldA != "c" || obj.(sample).FieldB != "d" {
		t.Fatalf("Unepxected result from Read(): %#v", obj)
	}

	obj, err = um.Read()
	if err != io.EOF {
		t.Fatalf("Unepxected result from Read(): (%#v, %#v)", obj, err)
	}
}

func Test_ReadAll(t *testing.T) {
    t.Parallel()

	type sample struct {
		FieldA string `csv:"field_a"`
		FieldB string `csv:"field_b"`
	}
	const csvContents = `field_a,field_b
a,b
c,d
`

	um, err := NewUnmarshaller(sample{}, csv.NewReader(strings.NewReader(csvContents)))
	if err != nil {
		t.Fatalf("Failed to allocate Unmarshaller: %s", err.Error())
	}

	out, err := um.ReadAll(StopOnError)
	if err != nil {
		t.Fatalf("Failed to allocate ReadAll(): %s", err.Error())
	}
	switch samples := out.(type) {
	case []sample:
		if len(samples) != 2 {
			t.Fatalf("Expected length 2, but got: %v", samples)
		}
	default:
		t.Fatalf("Expected []sample, but got %T", out)
	}

	// With pointers

	um, err = NewUnmarshaller(&sample{}, csv.NewReader(strings.NewReader(csvContents)))
	if err != nil {
		t.Fatalf("Failed to allocate Unmarshaller: %s", err.Error())
	}

	out, err = um.ReadAll(StopOnError)
	if err != nil {
		t.Fatalf("Failed to allocate ReadAll(): %s", err.Error())
	}
	switch samples := out.(type) {
	case []*sample:
		if len(samples) != 2 {
			t.Fatalf("Expected length 2, but got: %v", samples)
		}
	default:
		t.Fatalf("Expected []sample, but got %T", out)
	}
}

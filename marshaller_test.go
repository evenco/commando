package gocsv

import (
	"encoding/csv"
	"testing"
	"bytes"
)

func TestMarshaller(t *testing.T) {
	type sample struct {
		FieldA string `csv:"field_a"`
		FieldB string `csv:"field_b"`
	}

	out := new(bytes.Buffer)

	c := &Config{Holder: sample{}}
	cw := csv.NewWriter(out)
	m, err := c.NewMarshaller(cw)
	if err != nil {
		t.Fatalf("Error calling NewMarshaller: %#v", err)
	}

	s := sample{FieldA: "a", FieldB: "b"}

	if err := m.Write(s); err != nil {
		t.Fatalf("Error calling Write(): %#v", err)
	}

	if err := m.Flush(); err != nil {
		t.Fatalf("Error calling Flush(): %#v", err)
	}

	csv := out.String()
	expected := `field_a,field_b
a,b
`
	if csv != expected {
		t.Fatal("Got unexpected CSV output")
	}
}

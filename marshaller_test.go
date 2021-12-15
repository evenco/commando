package commando

import (
	"bytes"
	"encoding/csv"
	"testing"
)

func TestMarshaller(t *testing.T) {
	t.Parallel()

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

func TestMarshaller_WriteAll(t *testing.T) {
	t.Parallel()

	type sample struct {
		FieldA string `csv:"field_a"`
		FieldB string `csv:"field_b"`
	}

	out := new(bytes.Buffer)

	m, err := NewMarshaller(sample{}, csv.NewWriter(out))
	if err != nil {
		t.Fatalf("Error calling NewMarshaller: %#v", err)
	}

	// Make sure a type error is returned
	ws := []int{1, 2, 3}
	if err := m.WriteAll(ws); err == nil {
		t.Fatalf("Expected a type error")
	}

	s := []sample{
		{FieldA: "a", FieldB: "b"},
		{FieldA: "c", FieldB: "d"},
		{FieldA: "A", FieldB: "B"}}

	if err := m.WriteAll(s); err != nil {
		t.Fatalf("Error calling WriteAll(): %#v", err)
	}
	m.Flush()

	csv := out.String()
	expected := `field_a,field_b
a,b
c,d
A,B
`
	if csv != expected {
		t.Fatalf("Got unexpected CSV output:\n%q\n", csv)
	}
}

func TestMarshaller_WriteAll_SliceAliass(t *testing.T) {
	t.Parallel()

	type sample struct {
		FieldA string `csv:"field_a"`
		FieldB string `csv:"field_b"`
	}

	type samples []sample

	out := new(bytes.Buffer)

	m, err := NewMarshaller(sample{}, csv.NewWriter(out))
	if err != nil {
		t.Fatalf("Error calling NewMarshaller: %#v", err)
	}

	// Make sure a type error is returned
	ws := []int{1, 2, 3}
	if err := m.WriteAll(ws); err == nil {
		t.Fatalf("Expected a type error")
	}

	s := samples{
		{FieldA: "a", FieldB: "b"},
		{FieldA: "c", FieldB: "d"},
		{FieldA: "A", FieldB: "B"}}

	if err := m.WriteAll(s); err != nil {
		t.Fatalf("Error calling WriteAll(): %#v", err)
	}
	m.Flush()

	csv := out.String()
	expected := `field_a,field_b
a,b
c,d
A,B
`
	if csv != expected {
		t.Fatalf("Got unexpected CSV output:\n%q\n", csv)
	}
}

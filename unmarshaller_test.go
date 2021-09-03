package commando

import (
	"context"
	"encoding/csv"
	"io"
	"strings"
	"testing"
)

const (
	csvContents = `field_a,field_b
a,b
c,d
`

	brokenCSV = csvContents + `e,f,g
h,i,j
k,l
`
)

type sample struct {
	FieldA string `csv:"field_a"`
	FieldB string `csv:"field_b"`
}

func TestUnmarshallerLongRow(t *testing.T) {
	csvText := `field_a,field_b
a,b
c,d,e
`
	reader := csv.NewReader(strings.NewReader(csvText))
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

func Test_Read_ErrorLineNumbers(t *testing.T) {
    t.Parallel()

	um, err := NewUnmarshaller(sample{}, csv.NewReader(strings.NewReader(brokenCSV)))
	if err != nil {
		t.Fatalf("Failed to allocate Unmarshaller: %s", err.Error())
	}


	// Lines 4 & 5 have errors.

	var rec interface{}
	// Header is line 1
	_, err = um.Read()			// line 2
	_, err = um.Read()			// line 3
	_, err = um.Read()			// line 4

	// Line 5 has the first error
	rec, err = um.Read()
	if err == nil {
		t.Fatal("Expected error")
	} else if strings.Index(err.Error(), "line 5") == -1 {
		t.Fatal("Expected the error to mention line 5")
	}
	if rec != nil {
		t.Fatal("Expected no record")
	}
}

func Test_ReadAll(t *testing.T) {
    t.Parallel()

	ctx := context.Background()

	um, err := NewUnmarshaller(sample{}, csv.NewReader(strings.NewReader(csvContents)))
	if err != nil {
		t.Fatalf("Failed to allocate Unmarshaller: %s", err.Error())
	}

	out, err := um.ReadAll(ctx, StopOnError)
	if err != nil {
		t.Fatalf("Failed to ReadAll(): %s", err.Error())
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

	out, err = um.ReadAll(ctx, StopOnError)
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

	// Handling errors

	ignoreErrors := func(_ context.Context, _ error) error {
		return nil
	}

	um, err = NewUnmarshaller(&sample{}, csv.NewReader(strings.NewReader(brokenCSV)))
	if err != nil {
		t.Fatalf("Failed to allocate Unmarshaller: %s", err.Error())
	}

	out, err = um.ReadAll(ctx, ignoreErrors)
	if err != nil {
		t.Fatalf("Failed to allocate ReadAll(): %s", err.Error())
	}
	switch samples := out.(type) {
	case []*sample:
		if len(samples) != 3 {
			t.Fatalf("Expected length 2, but got: %v", samples)
		}
	default:
		t.Fatalf("Expected []sample, but got %T", out)
	}
}

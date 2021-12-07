package commando

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	csvContents = `field_a,field_b
a,b
c,d
`

	brokenCSV = csvContents + `e,f,g
h,i,j
k,l
m,n,o
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
	require.NoError(t, err)

	obj, err := um.Read()
	require.NoError(t, err)
	assert.Equal(t, sample{"a", "b"}, obj, "Unexpected result from Read()")

	obj, err = um.Read()
	require.NoError(t, err)
	assert.Equal(t, sample{"c", "d"}, obj, "Unexpected result from Read()")

	obj, err = um.Read()
	require.Equal(t, io.EOF, err, "Expected io.EOF")
}

func Test_Read_ErrorLineNumbers(t *testing.T) {
	t.Parallel()

	um, err := NewUnmarshaller(sample{}, csv.NewReader(strings.NewReader(brokenCSV)))
	require.NoError(t, err)

	// Lines 4 & 5 have errors.

	var rec interface{}
	// Header is line 1
	_, err = um.Read() // line 2
	assert.NoError(t, err)
	_, err = um.Read() // line 3
	assert.NoError(t, err)

	// Line 4 has the first error
	rec, err = um.Read()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 4", "Expected the error to mention line number")
	assert.Nil(t, rec, "Expected no record")

	// Line 5 is also broken
	rec, err = um.Read()
	require.Error(t, err, "Expected error")
	assert.Contains(t, err.Error(), "line 5", "Expected the error to mention line number")
	assert.Nil(t, rec, "Expected no record")

	// Line 6 is good
	rec, err = um.Read()
	require.NoError(t, err, "Expected no error")
	require.Equal(t, sample{"k", "l"}, rec)

	// Line 7 is broken again
	rec, err = um.Read()
	require.Error(t, err, "Expected error")
	assert.Contains(t, err.Error(), "line 7", "Expected the error to mention line number")
	assert.Nil(t, rec, "Expected no record")
}

func Test_ReadAll_ErrorLineNumbers(t *testing.T) {
	t.Parallel()

	type sample2 struct {
		A float64 `csv:"a"`
		B string  `csv:"b"`
	}

	brokenCSV := `a,b
1.0,a
2.0-,b
3.3,c
`

	ctx := context.Background()
	um, err := NewUnmarshaller(sample2{}, csv.NewReader(strings.NewReader(brokenCSV)))
	require.NoError(t, err, "Failed to allocate Unmarshaller")

	_, err = um.ReadAll(ctx, StopOnError)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "line 3", "Expected the error to mention line number")
}

func Test_ReadAll_LogErrorLineNumbers(t *testing.T) {
	t.Parallel()

	var errs []error
	logError := func(_ context.Context, err error) error {
		errs = append(errs, err)
		return nil
	}

	ctx := context.Background()
	um, err := NewUnmarshaller(sample{}, csv.NewReader(strings.NewReader(brokenCSV)))
	require.NoError(t, err)

	out, err := um.ReadAll(ctx, logError)
	require.NoError(t, err)
	expected := []sample{
		{"a", "b"},
		{"c", "d"},
		{"k", "l"},
	}
	require.Equal(t, expected, out)

	assert.Len(t, errs, 3)
	assert.Contains(t, errs[0].Error(), "line 4")
	assert.Contains(t, errs[1].Error(), "line 5")
	assert.Contains(t, errs[2].Error(), "line 7")
}

func Test_ReadAll(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	um, err := NewUnmarshaller(sample{}, csv.NewReader(strings.NewReader(csvContents)))
	require.NoError(t, err)

	out, err := um.ReadAll(ctx, StopOnError)
	require.NoError(t, err)

	switch samples := out.(type) {
	case []sample:
		assert.Len(t, samples, 2)
	default:
		assert.Fail(t, fmt.Sprintf("Expected []sample, but got %T", out))
	}

	// With pointers

	um, err = NewUnmarshaller(&sample{}, csv.NewReader(strings.NewReader(csvContents)))
	require.NoError(t, err)

	out, err = um.ReadAll(ctx, StopOnError)
	require.NoError(t, err)
	switch samples := out.(type) {
	case []*sample:
		assert.Len(t, samples, 2)
	default:
		assert.Fail(t, fmt.Sprintf("Expected []sample, but got %T", out))
	}

	// Handling errors

	ignoreErrors := func(_ context.Context, _ error) error {
		return nil
	}

	um, err = NewUnmarshaller(&sample{}, csv.NewReader(strings.NewReader(brokenCSV)))
	require.NoError(t, err)

	out, err = um.ReadAll(ctx, ignoreErrors)
	require.NoError(t, err)
	switch samples := out.(type) {
	case []*sample:
		assert.Len(t, samples, 3)
	default:
		assert.Fail(t, fmt.Sprintf("Expected []sample, but got %T", out))
	}
}

func Test_Unmarshaller_Allocation(t *testing.T) {
	t.Parallel()

	exactHeaders := `field_a,field_b`
	overlappingHeaders := `field_b,field_c`
	disjointHeaders := `field_c,field_d`

	var um *Unmarshaller
	var err error
	config := &Config{
		Holder: sample{},
	}

	// An Unmarshaller should be returned if the file headers match the struct.
	um, err = config.NewUnmarshaller(csv.NewReader(strings.NewReader(exactHeaders)))
	require.NoError(t, err, "Unexpected error")
	require.NotNil(t, um, "Expected Unmarshaller")

	// Same behavior if FailIfUnmatchedStructTags is set.
	config.FailIfUnmatchedStructTags = true
	um, err = config.NewUnmarshaller(csv.NewReader(strings.NewReader(exactHeaders)))
	require.NoError(t, err, "Unexpected error")
	require.NotNil(t, um, "Expected Unmarshaller")

	// An Unmarshaller should be returned if a *subset* of the file
	// headers match the struct, and FailIfUnmatchedStructTags = false
	config.FailIfUnmatchedStructTags = false
	um, err = config.NewUnmarshaller(csv.NewReader(strings.NewReader(overlappingHeaders)))
	require.NoError(t, err, "Unexpected error")
	require.NotNil(t, um, "Expected Unmarshaller")

	// An error should be returned if *any* fields don't match and
	// FailIfUnmatchedStructTags = true.
	config.FailIfUnmatchedStructTags = true
	um, err = config.NewUnmarshaller(csv.NewReader(strings.NewReader(overlappingHeaders)))
	require.Error(t, err, "Expected error")
	require.Nil(t, um, "Expected no Unmarshaller")

	// An error should be returned if none of the file headers match
	// the struct, no matter what FailIfUnmatchedStructTags is set to
	config.FailIfUnmatchedStructTags = false
	um, err = config.NewUnmarshaller(csv.NewReader(strings.NewReader(disjointHeaders)))
	require.Error(t, err, "Expected error")
	require.Nil(t, um, "Expected no Unmarshaller")

	config.FailIfUnmatchedStructTags = true
	um, err = config.NewUnmarshaller(csv.NewReader(strings.NewReader(disjointHeaders)))
	require.Error(t, err, "Expected error")
	require.Nil(t, um, "Expected no Unmarshaller")
}

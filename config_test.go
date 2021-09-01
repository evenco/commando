package commando

import (
	"testing"
)

func Test_Config_Validate(t *testing.T) {
    t.Parallel()

	var err error
	headers := []string{}
	_, err = (&Config{Holder: "whops"}).validate(headers)

	if err == nil {
		t.Fatal("Expected an error.")
	}

	_, err = (&Config{Holder: 123}).validate(headers)
	if err == nil {
		t.Fatal("Expected an error.")
	}
}

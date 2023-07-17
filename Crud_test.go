package crud

import (
	"testing"
)

func TestIErrorThrownWhenFuncRowsNotProvided(t *testing.T) {
	_, err := NewCrud(CrudConfig{
		// Endpoint: "TESTENDPOINT",
	})

	if err == nil {
		t.Error("Error MUST NOT be nil")
	}

	expected := "FuncRows function is required"
	if err.Error() != expected {
		t.Error("Error MUST be "+expected+" , but found: ", err.Error())
	}

	// if crud.endpoint != "TESTENDPOINT" {
	// 	t.Error("Crud endpoint MUST be TESTENDPOINT, found:", crud.endpoint)
	// }
}

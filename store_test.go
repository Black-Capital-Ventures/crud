package crud_test

import (
	"testing"

	"github.com/black-capital-ventures/crud"
)

func (t testOutput) GetInstance() crud.StorageOutput {
	return testOutput{}
}

type testOutput struct {
	Column1 string `crud:"column1"`
}

func TestSetField(t *testing.T) {
	instance := &testOutput{}

	// Test case: Successfully setting an exported field's value
	err := crud.SetField(instance, "Column1", "new value")
	requireEqual(t, nil, err, "Expected no error setting an exported field")
	requireEqual(t, "new value", instance.Column1, "The ExportedField should have been updated to 'new value'")

	// Test case: Attempting to set a non-existing field
	err = crud.SetField(instance, "NonExistingField", "value")
	requireNotEqual(t, nil, err, "Expected an error setting a non-existing field")
}

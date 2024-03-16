package crud_test

import (
	"database/sql"
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/black-capital-ventures/crud"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type (
	userInput struct {
		Name string
		Age  int
	}
)
type userOutput struct {
	ID   int    `crud:"id"`
	Name string `crud:"name"`
	Age  int    `crud:"age"`
}

type testOutput struct {
	Column1 string    `crud:"column1"`
	ID      uuid.UUID `crud:"id"`
}

func TestMain(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:@localhost:5432/test?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	// Create a new table
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id serial primary key, name text, age int)")
	if err != nil {
		log.Fatal(err)
	}

	// Insert a new user
	store := crud.NewStore[userInput, *userOutput](db)

	input := userInput{Name: "John Doe", Age: 30}
	output := userOutput{}
	err = store.QueryRow(
		"INSERT INTO users (name, age) VALUES ($1, $2) RETURNING id, name, age",
		input,
		&output,
	)
	if err != nil {
		log.Fatal(err)
	}

	expected := &userOutput{Name: "John Doe", Age: 30}

	assert.Equal(t, expected.Age, output.Age)
	assert.Equal(t, expected.Name, output.Name)
}

func (u userInput) GetArgs() []interface{} {
	return []interface{}{u.Name, u.Age}
}

func TestSetField(t *testing.T) {
	instance := &testOutput{}

	// Test case: Successfully setting an exported field's value
	err := crud.SetField(instance, "Column1", "new value")
	require.Equal(t, nil, err, "Expected no error setting an exported field")
	require.Equal(t, "new value", instance.Column1, "The ExportedField should have been updated to 'new value'")

	// Test case: Attempting to set a non-existing field
	err = crud.SetField(instance, "NonExistingField", "value")
	require.NotEqual(t, nil, err, "Expected an error setting a non-existing field")

	// test case: complex data type
	id := uuid.New()
	// sql will return the data as a string
	err = crud.SetField(instance, "ID", id.String())
	require.Equal(t, nil, err, "Expected no error setting an uuid field")
}

func TestScan(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	defer db.Close()

	id := uuid.New()

	tests := map[string]struct {
		mock   func()
		assert func(*testing.T, *testOutput, error)
	}{
		"no rows returned": {
			mock: func() {
				mock.ExpectQuery("SELECT").WillReturnRows(sqlmock.NewRows(nil))
			},
			assert: func(t *testing.T, output *testOutput, err error) {
				require.NotNil(t, err, "Expected an error when no rows are returned")
			},
		},
		"error getting columns": {
			mock: func() {
				rows := sqlmock.NewRows([]string{"column1"}).AddRow("testValue")
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			assert: func(t *testing.T, output *testOutput, err error) {
				require.Nil(t, err, "Expected no error when scanning rows")
				require.Equal(t, "testValue", output.Column1, "Expected Column1 to be 'testValue'")
			},
		},
		"Incompatible Instance and Rows": {
			mock: func() {
				rows := sqlmock.NewRows([]string{"random column"}).AddRow("testValue")
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			assert: func(t *testing.T, output *testOutput, err error) {
				require.NotNil(t, err, "Expected an error when instance and rows are incompatible")
			},
		},
		"Partial Data Match": {
			mock: func() {
				rows := sqlmock.NewRows([]string{"column1", "no match"}).AddRow("testValue", "no match")
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			assert: func(t *testing.T, output *testOutput, err error) {
				require.NotNil(t, err, "Expected an error when instance and rows are incompatible")
			},
		},
		"Full Data Match with Complex Types": {
			mock: func() {
				rows := sqlmock.NewRows([]string{"column1", "id"}).AddRow("testValue", id)
				mock.ExpectQuery("SELECT").WillReturnRows(rows)
			},
			assert: func(t *testing.T, output *testOutput, err error) {
				require.Nil(t, err, "Expected no error when scanning rows")
				require.Equal(t, "testValue", output.Column1, "Expected Column1 to be 'testValue'")
				require.Equal(t, id.String(), output.ID.String(), "Expected ID to match the value from the row")
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			_ = name //nice for debugging

			tc.mock()

			rows, _ := db.Query("SELECT")

			defer rows.Close()

			output := &testOutput{}
			err = crud.Scan(output, rows)
			tc.assert(t, output, err)
		})
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
	}
}

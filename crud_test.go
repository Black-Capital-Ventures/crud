package crud_test

import (
	"database/sql"
	"log"
	"testing"

	"github.com/black-capital-ventures/crud"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type (
	userInput struct {
		Name        string
		Age         int
		NullableInt *int
		// NullableBig *int64
		FK         uuid.UUID
		NullableFK uuid.UUID
	}
	userOutput struct {
		ID          uuid.UUID `crud:"id"`
		Name        string    `crud:"name"`
		Age         int       `crud:"age"`
		NullableInt *int      `crud:"nullable_int"`
		// NullableBig *int64     `crud:"nullable_big"`
		FK         uuid.UUID  `crud:"fk"`
		NullableFK *uuid.UUID `crud:"nullable_fk"`
	}

	inputT  = userInput
	outputT = *userOutput
	arrange func(expected *userOutput, input *userInput, output *userOutput)
	act     func(store *crud.Store[inputT, outputT], input userInput, output *userOutput) error
	assert  func(t *testing.T, err error, expected userOutput, output *userOutput)
)

func (u userInput) GetArgs() []interface{} {
	return []interface{}{u.Name, u.Age, u.NullableInt, u.FK, u.NullableFK}
}

func setUp(t *testing.T, db *sql.DB) {
	_, err := db.Exec(
		"CREATE TABLE users (id uuid primary key default gen_random_uuid()," +
			" name text, age  int not null,nullable_int int, fk uuid not null, nullable_fk uuid)",
	)
	if err != nil {
		t.Fatalf("error creating table: %v", err)
	}
}

func TestQueryRow(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:@localhost:5432/test?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}
	// Insert a new user
	store := crud.NewStore[inputT, outputT](db)

	cleanUp(t, db)
	setUp(t, db)

	var (
		input    = userInput{}
		output   = &userOutput{}
		expected = userOutput{}
		fkID     = uuid.New()
		_int     = 10
		cases    = []struct {
			name    string
			arrange arrange
			act     act
			assert  assert
		}{
			{
				name: "insert user and scan",
				arrange: func(expected *userOutput, input *userInput, output *userOutput) {
					*expected = userOutput{Name: "John Doe", Age: _int, NullableInt: &_int, FK: fkID, NullableFK: &fkID}
					*input = userInput{Name: "John Doe", Age: _int, NullableInt: &_int, FK: fkID, NullableFK: fkID}
					*output = userOutput{}
				},
				act: func(store *crud.Store[inputT, outputT], input userInput, output *userOutput) error {
					return store.QueryRow(
						"INSERT INTO users (name, age, nullable_int, fk, nullable_fk) VALUES ($1, $2, $3, $4, $5) RETURNING *",
						input,
						output,
					)
				},
				assert: func(t *testing.T, err error, expected userOutput, output *userOutput) {
					require.Nil(t, err)
					require.Equal(t, expected.Age, output.Age, "age")
					require.Equal(t, expected.Name, output.Name, "name")
					require.NotEqual(t, uuid.Nil, output.ID, "id")
					require.Equal(t, expected.NullableFK, output.NullableFK, "nullable_fk")
					require.Equal(t, expected.FK, output.FK, "fk")
					require.Equal(t, expected.NullableInt, output.NullableInt, "nullable_int")
					// require.Equal(t, expected.NullableBig, output.NullableBig, "nullable_big")
				},
			},
		}
	)

	for _, tt := range cases {
		tt.arrange(&expected, &input, output)
		err = tt.act(store, input, output)
		tt.assert(t, err, expected, output)
	}
}

func cleanUp(t *testing.T, db *sql.DB) {
	_, err := db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		t.Fatalf("error dropping table: %v", err)
	}
}

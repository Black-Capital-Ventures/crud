package crud_test

import (
	"database/sql"
	"fmt"
	"log"
	"testing"

	"github.com/black-capital-ventures/crud"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

type (
	userInput struct {
		Name string
		Age  int
	}
	userOutput struct {
		ID   uuid.UUID  `crud:"id"`
		Name string     `crud:"name"`
		Age  int        `crud:"age"`
		FK   *uuid.UUID `crud:"fk"`
	}
)

func TestMain(t *testing.T) {
	db, err := sql.Open("postgres", "postgres://postgres:@localhost:5432/test?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec("DROP TABLE IF EXISTS users")
	if err != nil {
		log.Fatal(fmt.Errorf("error dropping table: %w", err))
	}

	// Create a new table
	_, err = db.Exec(
		"CREATE TABLE users (id uuid primary key default gen_random_uuid(), name text, age int, fk uuid)")
	if err != nil {
		log.Fatal(fmt.Errorf("error creating table: %w", err))
	}

	// Insert a new user
	store := crud.NewStore[userInput, *userOutput](db)

	input := userInput{Name: "John Doe", Age: 30}
	output := userOutput{}
	err = store.QueryRow(
		"INSERT INTO users (name, age) VALUES ($1, $2) RETURNING id, name, age, fk",
		input,
		&output,
	)
	if err != nil {
		log.Fatal(err)
	}

	expected := &userOutput{Name: "John Doe", Age: 30}

	require.Equal(t, expected.Age, output.Age)
	require.Equal(t, expected.Name, output.Name)
	require.Equal(t, expected.FK, output.FK)

	fkID := uuid.New()
	_, err = db.Exec("update users set fk = $1", fkID)
	if err != nil {
		log.Fatal(err)
	}

	err = store.QueryRow(
		"SELECT * FROM users WHERE name = $1 AND age = $2",
		input,
		&output,
	)
	if err != nil {
		log.Fatal(err)
	}

	expected = &userOutput{Name: "John Doe", Age: 30, FK: &fkID}

	require.Equal(t, expected.Age, output.Age)
	require.Equal(t, expected.Name, output.Name)
	require.Equal(t, *expected.FK, *output.FK)
}

func (u userInput) GetArgs() []interface{} {
	return []interface{}{u.Name, u.Age}
}

package crud_test

import (
	"database/sql"
	"log"
	"testing"

	"github.com/black-capital-ventures/crud"
	_ "github.com/lib/pq"
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
	store := crud.NewStore[*userInput, *userOutput](db)

	input := userInput{Name: "John Doe", Age: 30}
	output := userOutput{}
	err = store.Create(
		"INSERT INTO users (name, age) VALUES ($1, $2) RETURNING id, name, age",
		&input,
		&output,
	)
	if err != nil {
		log.Fatal(err)
	}

	expected := &userOutput{Name: "John Doe", Age: 30}

	assertEqual(t, expected.Age, output.Age)
	assertEqual(t, expected.Name, output.Name)
}

func assertEqual(t *testing.T, expected, actual interface{}, msg ...string) {
	t.Helper()

	if expected != actual {
		t.Errorf("expected %v, got %v %v", expected, actual, msg)
	}
}

func requireEqual(t *testing.T, expected, actual interface{}, msg ...string) {
	t.Helper()

	if expected != actual {
		t.Fatalf("expected %v, got %v %v", expected, actual, msg)
	}
}

func requireNotEqual(t *testing.T, forbidden, actual interface{}, msg ...string) {
	t.Helper()

	if forbidden == actual {
		t.Fatalf("expected %v to differ from %v %v", actual, forbidden, msg)
	}
}

func (u userInput) GetArgs() []interface{} {
	return []interface{}{u.Name, u.Age}
}

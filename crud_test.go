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
	userType struct {
		ID   int
		Name string
		Age  int
	}
)

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
	store := crud.NewStore[userInput, userType](db)

	user, err := store.Create(
		"INSERT INTO users (name, age) VALUES ($1, $2) RETURNING id, name, age",
		userInput{Name: "John Doe", Age: 30},
	)
	if err != nil {
		log.Fatal(err)
	}

	expected := userType{Name: "John Doe", Age: 30}

	assert(t, expected.Age, user.Age)
	assert(t, expected.Name, user.Name)
}

func assert(t *testing.T, expected, actual interface{}, msg ...string) {
	t.Helper()

	if expected != actual {
		t.Errorf("expected %v, got %v %v", expected, actual, msg)
	}
}

func (u userType) GetInstance() crud.StorageOutput {
	return userType{}
}

func (u userType) Scan(rows *sql.Rows) (crud.StorageOutput, error) {
	if !rows.Next() {
		return u, nil
	}

	err := rows.Scan(&u.ID, &u.Name, &u.Age)
	if err != nil {
		return u, err
	}

	return u, nil
}

func (u userInput) GetArgs() []interface{} {
	return []interface{}{u.Name, u.Age}
}

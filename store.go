package crud

import (
	"fmt"

	"database/sql"
)

type Store[in StorageInput, out StorageOutput] struct {
	*sql.DB
}

func NewStore[in StorageInput, out StorageOutput](db *sql.DB) *Store[in, out] {
	return &Store[in, out]{DB: db}
}

func (s Store[in, out]) Create(query string, input in) (output out, err error) {
	rows, err := s.Query(
		query,
		input.GetArgs()...,
	)
	if err != nil {
		return output, fmt.Errorf("error creating %T: %w", output, err)
	}

	defer rows.Close()

	instance := output.GetInstance()
	output_, err := instance.Scan(rows)

	if err != nil {
		return output, fmt.Errorf("error scanning %T: %w", output, err)
	}

	return output_.(out), nil
}

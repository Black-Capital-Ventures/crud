package crud

import (
	"fmt"
	"reflect"

	"database/sql"
)

type Store[in StorageInput, out any] struct {
	*sql.DB
}

func NewStore[in StorageInput, out any](db *sql.DB) *Store[in, out] {
	return &Store[in, out]{DB: db}
}

func (s Store[in, out]) Scan(instance any, rows *sql.Rows) (output out, err error) {
	if !rows.Next() {
		return output, fmt.Errorf("no rows returned")
	}

	// scan rows into map
	columns, err := rows.Columns()
	if err != nil {
		return output, fmt.Errorf("error getting columns: %w", err)
	}

	// create a slice of interface{} to hold the values of each column
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range columns {
		valuePtrs[i] = &values[i]
	}

	err = rows.Scan(valuePtrs...)

	if err != nil {
		return output, fmt.Errorf("error scanning rows: %w", err)
	}

	// set instance fields
	orderedFieldNames, err := GetColumnsFieldNames(instance, columns)
	if err != nil {
		return output, fmt.Errorf("error getting ordered field names: %w", err)
	}

	for i, fieldName := range orderedFieldNames {
		err = SetField(instance, fieldName, values[i])
		if err != nil {
			return output, fmt.Errorf("error setting field: %w", err)
		}
	}

	return output, nil
}

func SetField(instance any, field string, value interface{}) error {
	// reflect on instance to get its value
	v := reflect.ValueOf(instance)

	if v.Kind() != reflect.Ptr {
		return fmt.Errorf("expected a pointer to a struct type setting field, got %v", v.Kind())
	}

	v = v.Elem()

	// get field by name
	f := v.FieldByName(field)
	if !f.IsValid() {
		return fmt.Errorf("field %s not found", field)
	}

	// set field value
	if f.CanSet() {
		v := reflect.ValueOf(value)
		if f.Type() != v.Type() {
			// check if the value can be converted to the field type
			if v.Type().ConvertibleTo(f.Type()) {
				v = v.Convert(f.Type())
			} else {
				return fmt.Errorf("field %s type mismatch %v != %v", field, f.Type(), v.Type())
			}
		}

		f.Set(v)
	} else {
		return fmt.Errorf("field %s cannot be set", field)
	}

	return nil
}

// GetColumnsFieldNames returns a slice ordered by the order of the columns slice.
// The slice contains the field names of the struct input
// The field names are obtained from the "crud" tag of the struct fields.
// If a field does not have a "crud" tag, it is not included in the returned slice.
// If a field has a "crud" tag that is not present in the columns slice, it is not included in the returned slice.
func GetColumnsFieldNames(instance any, columns []string) ([]string, error) {
	t := reflect.TypeOf(instance)

	// Ensure a pointer to a struct type was passed.
	if t.Kind() != reflect.Ptr {
		return nil, fmt.Errorf("expected a pointer to a struct type, got %v", t.Kind())
	}

	// Get the struct type.
	t = t.Elem()

	columnsFieldNameMap := make([]string, len(columns))
	numFields := t.NumField()
	j := 0
	for i := 0; i < numFields; i++ {
		fieldType := t.Field(i)
		crudTag := fieldType.Tag.Get("crud")

		for _, column := range columns {
			if crudTag == column {
				columnsFieldNameMap[j] = fieldType.Name
			}
		}

		j++
	}

	return columnsFieldNameMap, nil
}

func (s Store[in, out]) Create(query string, input in, output out) (err error) {
	rows, err := s.Query(
		query,
		input.GetArgs()...,
	)
	if err != nil {
		return fmt.Errorf("error creating %T: %w", output, err)
	}

	defer rows.Close()

	output, err = s.Scan(output, rows)
	if err != nil {
		return fmt.Errorf("error scanning %T: %w", output, err)
	}

	return nil
}

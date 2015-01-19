package gomodel

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
)

const fieldTag = "sql"

var snakeRegexp = regexp.MustCompile("([a-z])([A-Z])")

var (
	// ErrNotPtr is returned when the provided value was not a pointer
	ErrNotPtr = errors.New("models not a pointer")

	// ErrNotSlice is returned when the provided value was not a slice (or a
	// pointer to a slice)
	ErrNotSlice = errors.New("models not a slice")

	// ErrNotStruct is returned when the provided value was not a struct (or
	// a slice or pointer to a slice of structs)
	ErrNotStruct = errors.New("models not a struct")
)

// A ScannerFunc unwraps an ordered set of data into the values pointed to
// by the provided arguments.
//
// This will most almost certainly be either sql.Rows.Scan() or sql.Row.Scan()
type ScannerFunc func(...interface{}) error

// Scan maps the provided columns to pointers to fields on the provided model
// and returns any errors encountered during scanning.
func Scan(scanner ScannerFunc, columns []string, model interface{}) error {
	var dests []interface{}

	mvalue := reflect.Indirect(reflect.ValueOf(model))
	for _, column := range columns {
		dests = append(dests, modelDest(mvalue, column))
	}

	return scanner(dests...)
}

// Rows is the interface implemented by objects that scan an ordered set of data.
// This will most likely be a *sql.Rows instance.
type Rows interface {
	Columns() ([]string, error)
	Next() bool
	Scan(...interface{}) error
}

// ScanRows iterates over all rows listed in rows (think *sql.Rows) and creates
// and appends a new T for each row, where models is []T.
//
// The provided models argument must be a pointer to a slice and have an
// element type of either a struct or a pointer to a struct.
func ScanRows(rows Rows, models interface{}) error {
	// Check to ensure a pointer to an array was passed
	if reflect.ValueOf(models).Kind() != reflect.Ptr {
		return ErrNotPtr
	}

	mslice := reflect.Indirect(reflect.ValueOf(models))

	if mslice.Kind() != reflect.Slice {
		return ErrNotSlice
	}

	mType, err := discoverModelType(models)
	if err != nil {
		return err
	}

	// Get the list of columns returned in this row set
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// Loop over the rows as long as we can
	for rows.Next() {
		// Construct a new instance of the model typea. Analogous to:
		// model := &ModelType{}
		model := reflect.New(mType)

		// Attempt to scan into an interface{} representation of the model.
		// If there are any scan errors, return them
		if err := Scan(rows.Scan, columns, model.Interface()); err != nil {
			return err
		}

		// Append the new model to the slice.  Analogous to:
		// models = append(models, &model)
		if mslice.Type().Elem().Kind() != reflect.Ptr {
			model = reflect.Indirect(model)
		}
		mslice.Set(reflect.Append(mslice, model))
	}

	return nil
}

func discoverModelType(models interface{}) (reflect.Type, error) {
	modelv := reflect.Indirect(reflect.ValueOf(models))
	mType := modelv.Type().Elem()

	if mType.Kind() == reflect.Ptr {
		mType = mType.Elem()
	}

	if mType.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	return mType, nil
}

func modelDest(model reflect.Value, column string) interface{} {
	mType := model.Type()

	// Loop over the struct fields and look for one with a matching name
	for i := 0; i < mType.NumField(); i++ {
		fdef := mType.Field(i)

		if fieldIsColumn(fdef, column) {
			return model.Field(i).Addr().Interface()
		}
	}

	// Unknown field -- discard the data
	var discard interface{}
	return &discard
}

func fieldIsColumn(field reflect.StructField, column string) bool {
	// If the field has a tag, it should supercede the field name
	if tag := field.Tag.Get(fieldTag); tag != "" {
		return tag == column
	}

	return snake(field.Name) == column
}

func snake(input string) string {
	return strings.ToLower(snakeRegexp.ReplaceAllString(input, "${1}_${2}"))
}

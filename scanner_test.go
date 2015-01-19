package gomodel

import (
	"database/sql"
	"fmt"
	"reflect"
	"testing"
)

// Some sanity checks -- Ensure that the interfaces we're working with actually
// represent database/sql's structs
var _ Rows = &sql.Rows{}
var _ ScannerFunc = (&sql.Row{}).Scan
var _ ScannerFunc = (&sql.Rows{}).Scan

func TestScanner(t *testing.T) {
	type model struct {
		ID        int64
		Title     string
		ForeignID int64
	}

	called := false
	mymodel := &model{}

	scanner := func(dests ...interface{}) error {
		called = true

		expectations := []interface{}{&mymodel.ID, &mymodel.Title, &mymodel.ForeignID}

		if len(dests) != len(expectations) {
			t.Errorf("want: %d dests, got: %d", len(expectations), len(dests))
		} else {
			for i, expected := range expectations {
				actual := dests[i]

				if actual != expected {
					t.Errorf("want: %v, got: %v", expected, actual)
				}
			}
		}

		return nil
	}

	if err := Scan(scanner, []string{"id", "title", "foreign_id"}, mymodel); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if !called {
		t.Fatalf("Expected scanner to have been called")
	}
}

func TestScannerUnknownFields(t *testing.T) {
	type model struct {
	}

	called := false
	mymodel := &model{}

	scanner := func(dests ...interface{}) error {
		called = true

		if len(dests) != 2 {
			t.Fatalf("want: %d dests, got: %d", 1, len(dests))
		}

		return nil
	}

	if err := Scan(scanner, []string{"test1", "test2"}, mymodel); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if !called {
		t.Fatalf("Expected scanner to have been called")
	}
}

func TestScannerTags(t *testing.T) {
	type model struct {
		GoName string `sql:"sql_name"`
	}

	called := false
	mymodel := &model{}

	scanner := func(dests ...interface{}) error {
		called = true

		if len(dests) != 1 {
			t.Fatalf("want: %d dests, got: %d", 1, len(dests))
		}

		if dests[0] != &mymodel.GoName {
			t.Errorf("want: %v, got: %v", &mymodel.GoName, dests[0])
		}

		return nil
	}

	if err := Scan(scanner, []string{"sql_name"}, mymodel); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if !called {
		t.Fatalf("Expected scanner to have been called")
	}
}

func TestSnake(t *testing.T) {
	expectations := map[string]string{
		"ID":        "id",
		"Title":     "title",
		"ForeignID": "foreign_id",
	}

	for input, expected := range expectations {
		actual := snake(input)

		if actual != expected {
			t.Errorf("want: %s, got: %s", expected, actual)
		}
	}
}

type mockRows struct {
	currentRow int
	columns    []string
	data       [][]interface{}
}

func (m mockRows) Columns() ([]string, error) {
	return m.columns, nil
}

func (m *mockRows) Next() bool {
	m.currentRow++
	return m.currentRow <= len(m.data)
}

func (m mockRows) Scan(dests ...interface{}) error {
	row := m.data[m.currentRow-1]

	for i, src := range row {
		dest := dests[i]

		switch s := src.(type) {
		case string:
			switch d := dest.(type) {
			case *string:
				*d = s
			default:
				return fmt.Errorf("Unknown string type: %s", d)
			}
		case int:
			switch d := dest.(type) {
			case *int64:
				*d = int64(s)
			}
		default:
			return fmt.Errorf("Unknown source type: %s", s)
		}
	}

	return nil
}

func TestScanRows(t *testing.T) {
	type MyModel struct {
		ID       int64
		Username string
	}

	rows := &mockRows{
		columns: []string{"id", "username"},
		data: [][]interface{}{
			{1234, "user_a"},
			{1235, "user_b"},
		},
	}
	expecteds := []MyModel{
		{1234, "user_a"},
		{1235, "user_b"},
	}

	var models []MyModel

	if err := ScanRows(rows, &models); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if len(models) != len(expecteds) {
		t.Fatalf("Expected %d models, got %d", len(expecteds), len(models))
	}

	for i, expected := range expecteds {
		actual := models[i]
		t.Logf("Result: %#v", actual)

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("want: %#v, got: %#v", expected, actual)
		}
	}
}

func TestScanRowsPointerModel(t *testing.T) {
	type MyModel struct {
		ID       int64
		Username string
	}

	rows := &mockRows{
		columns: []string{"id", "username"},
		data: [][]interface{}{
			{1234, "user_a"},
			{1235, "user_b"},
		},
	}
	expecteds := []*MyModel{
		{1234, "user_a"},
		{1235, "user_b"},
	}

	var models []*MyModel

	if err := ScanRows(rows, &models); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if len(models) != len(expecteds) {
		t.Fatalf("Expected %d models, got %d", len(expecteds), len(models))
	}

	for i, expected := range expecteds {
		actual := models[i]
		t.Logf("Result: %#v", actual)

		if !reflect.DeepEqual(actual, expected) {
			t.Errorf("want: %#v, got: %#v", expected, actual)
		}
	}
}

type erows struct {
	cerr error
	serr error
}

func (e erows) Next() bool                 { return true }
func (e erows) Scan(...interface{}) error  { return e.serr }
func (e erows) Columns() ([]string, error) { return []string{}, e.cerr }

func TestScanRowsColumnsError(t *testing.T) {
	var expected = fmt.Errorf("Something horrible!")

	var models []struct{}
	if err := ScanRows(&erows{expected, nil}, &models); err != expected {
		t.Fatalf("want: %s, got: %v", expected, err)
	}
}

func TestScanRowsScanError(t *testing.T) {
	var expected = fmt.Errorf("Fire detected")

	var models []struct{}
	if err := ScanRows(&erows{nil, expected}, &models); err != expected {
		t.Fatalf("want: %s, got: %v", expected, err)
	}
}

func TestScanRowsNonPointer(t *testing.T) {
	err := ScanRows(&mockRows{}, []interface{}{})
	if err == nil {
		t.Fatalf("Expected an error, got %v", err)
	}

	t.Logf("Reported error: %s", err)
}

func TestScanRowsNonSlice(t *testing.T) {
	var models string

	err := ScanRows(&mockRows{}, &models)
	if err == nil {
		t.Fatalf("Expected an error, got %v", err)
	}

	t.Logf("Reported error: %s", err)
}

func TestScanRowsNonStruct(t *testing.T) {
	var models []string

	err := ScanRows(&mockRows{}, &models)
	if err == nil {
		t.Fatalf("Expected an error, got %v", err)
	}

	t.Logf("Reported error: %s", err)
}

// Mocked SQL struct
var db *sql.DB

func ExampleScan() {
	type BlogPost struct {
		ID     int64
		Title  string
		UserID int64
	}

	// Perform the query
	row := db.QueryRow("SELECT id, title, user_id FROM blog_posts WHERE id = ?", 1234)
	columns := []string{"id", "title", "user_id"}
	post := &BlogPost{}
	if err := Scan(row.Scan, columns, post); err != nil {
		panic(err)
	}

	fmt.Printf("[%d] %s by %d", post.ID, post.Title, post.UserID)
}

func ExampleScanRows() {
	type BlogComment struct {
		ID         int64
		BlogPostID int64
		Comment    string
	}

	rows, err := db.Query("SELECT * FROM blog_comments WHERE blog_post_id = ?", 1234)
	if err != nil {
		panic(err)
	}

	var comments []BlogComment

	if err := ScanRows(rows, &comments); err != nil {
		panic(err)
	}
}

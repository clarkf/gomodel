package gomodel

import (
	"testing"
	"time"
)

func mockScanner(...interface{}) error { return nil }

func BenchmarkScan(b *testing.B) {
	type model struct {
		ID        int64
		Title     string
		Body      string
		AuthorID  int64 `sql:"user_id"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	columns := []string{"id", "title", "body", "user_id", "created_at", "updated_at"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		model := &model{}
		Scan(mockScanner, columns, model)
	}
}

type stubRows struct {
	currentRow int
	totalRows  int
	columns    []string
}

func (s *stubRows) Next() bool {
	s.currentRow++
	return s.currentRow <= s.totalRows
}
func (s *stubRows) Columns() ([]string, error) { return s.columns, nil }
func (s *stubRows) Scan(...interface{}) error  { return nil }

func BenchmarkScanRows(b *testing.B) {
	type model struct {
		ID        int64
		Title     string
		Body      string
		AuthorID  int64 `sql:"user_id"`
		CreatedAt time.Time
		UpdatedAt time.Time
	}

	rows := &stubRows{
		totalRows: 10,
		columns:   []string{"id", "title", "body", "user_id", "created_at", "updated_at"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rows.currentRow = 0
		var models []model
		ScanRows(rows, &models)
	}
}

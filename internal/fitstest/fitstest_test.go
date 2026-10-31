package fitstest_test

import (
	"bytes"
	"slices"
	"testing"

	"github.com/astrogo/fitsio"
	"github.com/janearc/nilgai/internal/fitstest"
)

// fractional and whole are the two sizes the test writes.
const (
	// fractional is two and a half, a size with a fraction.
	fractional = 5.0 / 2
	// whole is four, a size with no fraction.
	whole = 2.0 * 2
)

// fromTheTop is the row that a read of the table starts from, row zero.
const fromTheTop = 0

// TestWriteIsReadBack checks that a table written is a table that fitsio
// reads back, with the rows put in it.
func TestWriteIsReadBack(test *testing.T) {
	type row struct {
		Size float32 `fits:"Size"`
	}
	var out bytes.Buffer
	columns := []fitsio.Column{{Name: "Size", Format: "E"}}
	err := fitstest.Write(&out, "T", columns, &row{fractional},
		&row{whole})
	if err != nil {
		test.Fatal(err)
	}
	file, err := fitsio.Open(&out)
	if err != nil {
		test.Fatal(err)
	}
	table, isTable := file.Get("T").(*fitsio.Table)
	if !isTable {
		test.Fatalf("read back %v", file.Get("T"))
	}
	rows, err := table.Read(fromTheTop, table.NumRows())
	if err != nil {
		test.Fatal(err)
	}
	defer rows.Close()
	var got []float32
	for rows.Next() {
		var each row
		if err := rows.Scan(&each); err != nil {
			test.Fatal(err)
		}
		got = append(got, each.Size)
	}
	if want := []float32{fractional, whole}; !slices.Equal(got, want) {
		test.Errorf("rows read back as %v, want %v", got, want)
	}
}

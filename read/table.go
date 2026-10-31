package read

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/janearc/libtheme-css/primitives/bands"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai/band"
)

// firstDataRow numbers the first data row. nilgai.RowError counts rows
// from 1 and does not count the header.
const firstDataRow = 1

// anyCount is the value of csv.Reader's FieldsPerRecord that holds no row
// to its header's length. A short row is then read, and its missing cells
// are refused by name, instead of ending the table.
const anyCount = -1

// tsv reads tab-separated rows and passes over # lines, as the CSC's TAP
// service writes them.
func tsv(in io.Reader) *csv.Reader {
	rows := separated(in)
	rows.Comma = '\t'
	return rows
}

// separated reads comma-separated rows and passes over # lines, as IRSA's
// TAP service writes them.
func separated(in io.Reader) *csv.Reader {
	rows := csv.NewReader(in)
	rows.Comment = '#'
	rows.FieldsPerRecord = anyCount
	return rows
}

// header reads a table's header row and returns where each wanted column
// sits, or an error naming a missing column.
func header(rows *csv.Reader, want []string) (map[string]int, error) {
	names, err := rows.Read()
	if err != nil {
		return nil, fmt.Errorf("no header: %w", err)
	}
	at := map[string]int{}
	for place, column := range names {
		at[column] = place
	}
	for _, column := range want {
		if _, found := at[column]; !found {
			return nil, noColumn(column)
		}
	}
	return at, nil
}

// noColumn returns the error for a table that lacks a column.
func noColumn(column string) error {
	return fmt.Errorf("the table has no column %s", column)
}

// cellsOf returns a row's cells by column name. A row shorter than its
// header has no cells for the missing columns, which are then refused by
// name.
func cellsOf(fields []string, at map[string]int) map[string]string {
	cells := map[string]string{}
	for column, place := range at {
		if place < len(fields) {
			cells[column] = fields[place]
		}
	}
	return cells
}

// numbers reads the named cells as numbers, or returns the column that
// could not be read and why. An empty cell is refused as empty, because
// the catalogue has no value there and nilgai does not make one up.
func numbers(cells map[string]string,
	columns ...string) (map[string]float64, string, error) {
	values := map[string]float64{}
	for _, column := range columns {
		text := cells[column]
		if text == "" {
			return nil, column, errors.New("empty")
		}
		value, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return nil, column, err
		}
		values[column] = value
	}
	return values, "", nil
}

// profileBands returns the first and last libtheme band of the named
// profile. A reader's band edges are then written once, in
// band/data/profiles.json beside its colour.
func profileBands(profiles []radiation.Profile,
	name string) (first, last bands.Band, err error) {
	for _, each := range profiles {
		if each.Name == name {
			return band.BandsBetween(each.Source.Lo, each.Source.Hi)
		}
	}
	return first, last, fmt.Errorf("no profile is named %s", name)
}

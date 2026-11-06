package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// table is what a verb writes: notes for its header, its columns' names,
// and one line a star.
type table struct {
	notes   []string
	columns []string
	lines   [][]string
}

// The tables' own marks, each defined.
const (
	noteMark  = "#"  // defined: a header line starts with it
	separator = "\t" // defined: between a line's fields
)

// joined is a line's fields, tab-separated.
func joined(fields []string) string {
	return strings.Join(fields, separator)
}

// writeTo writes a table to the file at the path given, made anew: the
// table written, flushed and the file closed, in that order, with any
// of their errors.
func writeTo(path string, written table) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	out := bufio.NewWriter(file)
	return errors.Join(writeTable(out, written), out.Flush(), file.Close())
}

// read is a table read back: its lines as fields, found by column name.
type read struct {
	columns map[string]int
	lines   [][]string
}

// readFrom reads the table in the file at the path given: the notes
// skipped, the first other line its columns. It refuses a file it
// cannot read and a table without the columns asked for.
func readFrom(path string, wanted ...string) (*read, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return readTable(file, path, wanted)
}

// readTable reads a table from the reader given, named so in refusals.
func readTable(in io.Reader, name string, wanted []string) (*read,
	error) {
	lines := bufio.NewScanner(in)
	found := &read{}
	for lines.Scan() {
		line := lines.Text()
		if strings.HasPrefix(line, noteMark) {
			continue
		}
		fields := strings.Split(line, separator)
		if found.columns == nil {
			found.columns = map[string]int{}
			for at, column := range fields {
				found.columns[column] = at
			}
			continue
		}
		found.lines = append(found.lines, fields)
	}
	if err := lines.Err(); err != nil {
		return nil, err
	}
	for _, column := range wanted {
		if _, ok := found.columns[column]; !ok {
			return nil, fmt.Errorf("%s has no column %s", name, column)
		}
	}
	return found, nil
}

// number is a line's field in a column, as a number. It refuses a field
// that is not one, or a line too short for the column.
func (table read) number(line []string, column string) (float64, error) {
	at := table.columns[column]
	if at >= len(line) {
		return nothing, fmt.Errorf("a line of %d fields has no %s",
			len(line), column)
	}
	return strconv.ParseFloat(line[at], 64)
}

// text is a line's field in a column as it was written.
func (table read) text(line []string, column string) string {
	at := table.columns[column]
	if at >= len(line) {
		return ""
	}
	return line[at]
}

// screenedFrom reads the stars a screen found from its table.
func screenedFrom(path string) ([]screened, error) {
	table, err := readFrom(path, screenColumns...)
	if err != nil {
		return nil, err
	}
	var stars []screened
	for _, line := range table.lines {
		star, err := table.screenedLine(line)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		stars = append(stars, *star)
	}
	return stars, nil
}

// screenedLine is a star from a line of the screen's table: its row and
// id read whole, since a source id is past what a float holds to the
// unit, and its two passes.
func (table read) screenedLine(line []string) (*screened, error) {
	row, err := table.whole(line, "row")
	if err != nil {
		return nil, err
	}
	id, err := table.whole(line, "source_id")
	if err != nil {
		return nil, err
	}
	nearFrame, err := table.number(line, "near_frame_pc")
	if err != nil {
		return nil, err
	}
	nearModel, err := table.number(line, "near_model_pc")
	if err != nil {
		return nil, err
	}
	return &screened{row: row, id: id, nearFrame: nearFrame,
		nearModel: nearModel}, nil
}

// whole is a line's field in a column, as a whole number. It refuses a
// field that is not one.
func (table read) whole(line []string, column string) (int64, error) {
	return strconv.ParseInt(table.text(line, column), decimal, bits)
}

// How whole numbers are read, each defined.
const (
	decimal = 10 // defined: written in base ten
	bits    = 64 // defined: as wide as an int64
)

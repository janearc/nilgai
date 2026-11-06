// Package fitstest writes small FITS binary tables for reader tests and fixture
// generation. It is internal to Nilgai.
package fitstest

import (
	"io"

	"github.com/astrogo/fitsio"
)

// Write emits an empty primary HDU followed by a named binary table with the
// supplied columns and rows. Each row must be a pointer to a struct whose fits
// tags identify the columns.
func Write(out io.Writer, name string, columns []fitsio.Column,
	rows ...any) error {
	var each [][]any
	for _, row := range rows {
		each = append(each, []any{row})
	}
	return writeRows(out, name, columns, each)
}

// writeRows emits an empty primary HDU followed by a named binary table
// with the supplied columns, each row written from its arguments: one, a
// pointer to a struct, or a value for each column in order.
func writeRows(out io.Writer, name string, columns []fitsio.Column,
	rows [][]any) error {
	file, err := fitsio.Create(out)
	if err != nil {
		return err
	}
	primary, err := fitsio.NewPrimaryHDU(nil)
	if err != nil {
		return err
	}
	if err := file.Write(primary); err != nil {
		return err
	}
	table, err := fitsio.NewTable(name, columns, fitsio.BINARY_TBL)
	if err != nil {
		return err
	}
	for _, each := range rows {
		if err := table.Write(each...); err != nil {
			return err
		}
	}
	if err := file.Write(table); err != nil {
		return err
	}
	if err := table.Close(); err != nil {
		return err
	}
	return file.Close()
}

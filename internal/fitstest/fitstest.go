// Package fitstest writes small FITS binary tables, for the tests of
// nilgai's readers and for the program that cuts their fixture. It lives
// inside internal so that it is not part of the library.
package fitstest

import (
	"io"

	"github.com/astrogo/fitsio"
)

// Write writes a FITS file to out. The file holds an empty primary
// header, then one binary table of the name given, with the columns given
// and a row for each of rows. Each row is a pointer to a struct whose
// fields carry fits tags naming the columns they fill.
func Write(out io.Writer, name string, columns []fitsio.Column,
	rows ...any) error {
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
		if err := table.Write(each); err != nil {
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

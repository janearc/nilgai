package nilgai

import (
	"fmt"
	"io"
	"iter"
)

// RowError says where a catalogue row could not be read. A reader yields
// one for each such row and never drops the row silently.
type RowError struct {
	file   string
	row    int
	column string
	err    error
}

// NewRowError returns a RowError saying that the row of the file, counting
// data rows from 1, could not be read at the column, and why.
func NewRowError(file string, row int, column string, err error) *RowError {
	return &RowError{file: file, row: row, column: column, err: err}
}

// File returns the name the caller gave the input.
func (rowError *RowError) File() string { return rowError.file }

// Row returns the row's number, counting from 1 and counting data rows
// only.
func (rowError *RowError) Row() int { return rowError.row }

// Column returns the catalogue's column name.
func (rowError *RowError) Column() string { return rowError.column }

// Error returns a message naming the file, row and column.
func (rowError *RowError) Error() string {
	return fmt.Sprintf("%s: row %d, column %s: %v", rowError.file,
		rowError.row, rowError.column, rowError.err)
}

// Unwrap returns what was wrong, for errors.Is and errors.As.
func (rowError *RowError) Unwrap() error { return rowError.err }

//go:generate go tool mockgen -destination=mocks/fermi.go . FermiReader

// FermiReader reads the Fermi-LAT 4FGL FITS table, with its 95% position
// ellipses, unassociated sources included.
//
// Read yields one *Source for each row, or nil and a *RowError for a row
// it cannot read, then carries on with the next row. A failure to read the
// input at all is yielded once, as nil and an ordinary error, and ends the
// sequence. The name appears in errors. Read does not open files.
type FermiReader interface {
	Read(name string, in io.Reader) iter.Seq2[*Source, error]
}

//go:generate go tool mockgen -destination=mocks/meerkat.go . MeerKATReader

// MeerKATReader reads the MeerKAT Galactic Centre compact source
// catalogue (Heywood et al. 2022), a FITS table at 1.28 GHz with
// one-sigma position errors. Its Read behaves as FermiReader.Read does.
type MeerKATReader interface {
	Read(name string, in io.Reader) iter.Seq2[*Source, error]
}

//go:generate go tool mockgen -destination=mocks/chandra.go . ChandraReader

// ChandraReader reads a Chandra Source Catalog search, as the CSC's TAP
// service writes it when asked for text/csv: tab-separated, with its
// column descriptions on # lines. Positions carry 95% ellipses. Its Read
// behaves as FermiReader.Read does. read.Chandra is a ChandraReader.
type ChandraReader interface {
	Read(name string, in io.Reader) iter.Seq2[*Source, error]
}

//go:generate go tool mockgen -destination=mocks/neowise.go . NEOWISEReader

// NEOWISEReader reads a Neowise single-exposure cone search, a VOTable in
// which each row is one look. It groups the looks by position and yields
// one LightCurve for each position. It yields failures as
// FermiReader.Read does.
type NEOWISEReader interface {
	Read(name string, in io.Reader) iter.Seq2[*LightCurve, error]
}

package nilgai

import (
	"fmt"
	"io"
	"iter"
)

// RowError locates a catalogue row that could not be parsed. Readers yield one
// per invalid row so callers can report failures while retaining valid sources.
type RowError struct {
	file   string
	row    int
	column string
	err    error
}

// NewRowError records a parsing error at a file, data row, and column. Data
// rows are numbered from one, excluding the header.
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

// Unwrap returns the underlying error for errors.Is and errors.As.
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

// MeerKATReader defines a reader for the Galactic Centre compact source
// catalogue (Heywood et al. 2022), a 1.28 GHz FITS table with one-sigma
// position errors. Read follows FermiReader.Read’s error behaviour. A concrete
// MeerKAT reader is not implemented yet.
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

// NEOWISEReader defines a reader for single-exposure Neowise VOTables. It
// groups observations by position and yields one LightCurve per group, with the
// error behaviour described by FermiReader.Read. The interface is defined here;
// a concrete Neowise reader is not implemented yet.
type NEOWISEReader interface {
	Read(name string, in io.Reader) iter.Seq2[*LightCurve, error]
}

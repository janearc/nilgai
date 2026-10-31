// Command mkfermi writes testdata/4fgl-rows.fit. The file holds a few rows
// of the real 4FGL-DR4 catalogue (gll_psc_v35.fit) with their real values.
// mkfermi writes them to a table of the same name, with the columns
// nilgai reads, in the formats and units those columns have in the real
// file.
//
// mkfermi is not part of the library. Run it by hand when the rows are to
// change:
//
//	go run ./testdata/mkfermi /path/to/gll_psc_v35.fit testdata/4fgl-rows.fit
//
// Run it from the read directory. The rows are named below, and mkfermi
// looks up each by name. A row that is not in the catalogue stops the run.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/astrogo/fitsio"
	"github.com/janearc/nilgai/internal/fitstest"
)

// tableName is the name that 4FGL gives its point source table.
const tableName = "LAT_Point_Source_Catalog"

// columns lists the columns nilgai reads, in the order the real file has
// them.
var columns = []string{"Source_Name", "RAJ2000", "DEJ2000",
	"Conf_95_SemiMajor", "Conf_95_SemiMinor", "Conf_95_PosAng",
	"Energy_Flux100", "Unc_Energy_Flux100", "CLASS1", "ASSOC1", "Flags"}

// wanted lists the rows to copy, in the order they are written, with the
// reason for each.
var wanted = []struct{ name, why string }{
	{"4FGL J1745.6-2859", "the Galactic Centre source: a point source " +
		"with a position ellipse"},
	{"4FGL J0058.0-7245e", "an extended source (the Small Magellanic " +
		"Cloud): no ellipse"},
	{"4FGL J0358.4-5446", "a nova, a point source whose position is " +
		"fixed from outside 4FGL: no ellipse"},
	{"4FGL J0004.0+5715", "unassociated (blank CLASS1) at Galactic " +
		"latitude -5.02: the case Unassociated selects"},
	{"4FGL J0002.8+6217", "associated (PSR, a pulsar) at Galactic " +
		"latitude -0.05: low latitude, not unassociated"},
	{"4FGL J0002.1+6721c", "associated (unk, a source of unknown " +
		"nature) at Galactic latitude 4.94: unk is not blank"},
	{"4FGL J0000.3-7355", "unassociated at Galactic latitude -42.73: " +
		"unassociated, but far from the plane"},
}

// row holds the columns nilgai reads from a 4FGL row.
type row struct {
	Name      string  `fits:"Source_Name"`
	RA        float32 `fits:"RAJ2000"`
	Dec       float32 `fits:"DEJ2000"`
	Major     float32 `fits:"Conf_95_SemiMajor"`
	Minor     float32 `fits:"Conf_95_SemiMinor"`
	Angle     float32 `fits:"Conf_95_PosAng"`
	Flux      float32 `fits:"Energy_Flux100"`
	FluxError float32 `fits:"Unc_Energy_Flux100"`
	Class     string  `fits:"CLASS1"`
	Assoc     string  `fits:"ASSOC1"`
	Flags     int16   `fits:"Flags"`
}

// written holds a row as it is written. fitsio writes a Go string with a
// byte in front of it, which loses the last letter of a name as long as
// its column, such as 4FGL J0058.0-7245e. A byte array is written as it
// is, padded with spaces as the real file pads its text.
type written struct {
	Name      [18]byte `fits:"Source_Name"`
	RA        float32  `fits:"RAJ2000"`
	Dec       float32  `fits:"DEJ2000"`
	Major     float32  `fits:"Conf_95_SemiMajor"`
	Minor     float32  `fits:"Conf_95_SemiMinor"`
	Angle     float32  `fits:"Conf_95_PosAng"`
	Flux      float32  `fits:"Energy_Flux100"`
	FluxError float32  `fits:"Unc_Energy_Flux100"`
	Class     [5]byte  `fits:"CLASS1"`
	Assoc     [28]byte `fits:"ASSOC1"`
	Flags     int16    `fits:"Flags"`
}

// padded copies text into the array and fills what is left with spaces.
func padded(into []byte, text string) {
	for place := range into {
		into[place] = ' '
	}
	copy(into, text)
}

// asWritten returns the row in the form that is ready to write.
func (each row) asWritten() written {
	out := written{RA: each.RA, Dec: each.Dec, Major: each.Major,
		Minor: each.Minor, Angle: each.Angle, Flux: each.Flux,
		FluxError: each.FluxError, Flags: each.Flags}
	padded(out.Name[:], each.Name)
	padded(out.Class[:], each.Class)
	padded(out.Assoc[:], each.Assoc)
	return out
}

// main copies the wanted rows from the first file named to the second.
func main() {
	if len(os.Args) != 3 {
		fmt.Fprintln(os.Stderr, "usage: mkfermi gll_psc_v35.fit out.fit")
		os.Exit(2)
	}
	if err := run(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "mkfermi:", err)
		os.Exit(1)
	}
}

// run reads the catalogue at from and writes the fixture to to.
func run(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()
	file, err := fitsio.Open(in)
	if err != nil {
		return err
	}
	defer file.Close()
	real, isTable := file.Get(tableName).(*fitsio.Table)
	if !isTable {
		return fmt.Errorf("%s has no table %s", from, tableName)
	}
	found, err := find(real)
	if err != nil {
		return err
	}
	return write(to, real, found)
}

// find returns the wanted rows, looked up by name, as the catalogue has
// them.
func find(real *fitsio.Table) ([]row, error) {
	rows, err := real.Read(0, real.NumRows())
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	found := map[string]row{}
	for rows.Next() {
		var each row
		if err := rows.Scan(&each); err != nil {
			return nil, err
		}
		found[strings.TrimSpace(each.Name)] = each
	}
	var chosen []row
	for _, want := range wanted {
		each, there := found[want.name]
		if !there {
			return nil, fmt.Errorf("the catalogue has no %s", want.name)
		}
		chosen = append(chosen, each)
	}
	return chosen, nil
}

// write writes the rows to the path, in a table made of those columns of
// the real table that nilgai reads.
func write(path string, real *fitsio.Table, chosen []row) error {
	var made []fitsio.Column
	for _, name := range columns {
		place := real.Index(name)
		if place < 0 {
			return fmt.Errorf("the catalogue has no column %s", name)
		}
		column := real.Col(place)
		made = append(made, fitsio.Column{Name: column.Name,
			Format: column.Format, Unit: column.Unit})
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()
	var rows []any
	for _, each := range chosen {
		written := each.asWritten()
		rows = append(rows, &written)
	}
	return fitstest.Write(out, tableName, made, rows...)
}

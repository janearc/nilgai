package main

import (
	"errors"
	"fmt"
	"image/png"
	"io"
	"iter"
	"os"
	"strconv"

	"github.com/janearc/libreadme/profile"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai"
	"github.com/janearc/nilgai/band"
	"github.com/janearc/nilgai/read"
	"github.com/janearc/nilgai/sky"
	"github.com/janearc/puzzlemouth/pzm"
)

// showWith returns the show verb, reading the profiles and the searches
// given. The real tool passes band.NewEmbedded() and realSearches.
func showWith(profiles band.ProfileSource,
	searches ...search) func(pzm.Call) error {
	return shower{profiles, searches}.show
}

// shower holds the profiles and the searches that a show call reads.
type shower struct {
	profiles band.ProfileSource
	searches []search
}

// show is the verb. It draws what the call names from the searches.
func (shown shower) show(call pzm.Call) error {
	return show(call, shown.profiles, shown.searches)
}

// show draws the searches the call names as a png on standard output.
// Rows that cannot be read are returned together, after the picture is
// written. The picture then shows what could be read, and the call still
// says what could not.
func show(call pzm.Call, source band.ProfileSource,
	searches []search) error {
	wanted, err := askedFor(call)
	if err != nil {
		return err
	}
	profiles, err := source.Profiles()
	if err != nil {
		return err
	}
	sources, unread, err := readSearches(call, profiles, searches)
	if err != nil {
		return err
	}
	reader, err := profile.Current()
	if err != nil {
		return err
	}
	var skies []*sky.Sky
	for _, width := range wanted.widths {
		panel, err := sky.NewSky(wanted.centre, width, wanted.pixels,
			profiles, reader)
		if err != nil {
			return err
		}
		skies = append(skies, panel)
	}
	picture := sky.Strip(skies, sources, wanted.centre)
	if err := png.Encode(call.Out(), picture); err != nil {
		return err
	}
	return errors.Join(unread...)
}

// reader is a catalogue reader, such as read.Chandra, read.AllWISE or
// read.Fermi.
type reader interface {
	Read(name string, in io.Reader) iter.Seq2[*nilgai.Source, error]
}

// asReader is the table as a reader, or a nil reader and the error. It
// is here because a nil *Chandra put in a reader is not a nil reader.
func asReader[Table reader](table Table, err error) (reader, error) {
	if err != nil {
		return nil, err
	}
	return table, nil
}

// search pairs a flag that show reads a catalogue from with a function
// that makes the reader for what the flag names.
type search struct {
	flag  string
	build func([]radiation.Profile) (reader, error)
}

// realSearches lists the searches show draws, in the order show reads
// them. The fixed order means the same call draws the same picture every
// time. Sources are drawn in this order, the last on top.
func realSearches() []search {
	return []search{
		{"chandra", chandraReader},
		{"allwise", allwiseReader},
		{"fermi", fermiReader},
	}
}

// chandraReader makes the reader for a Chandra Source Catalog search.
func chandraReader(profiles []radiation.Profile) (reader, error) {
	return asReader(read.NewChandra(profiles))
}

// allwiseReader makes the reader for an AllWISE search.
func allwiseReader([]radiation.Profile) (reader, error) {
	return asReader(read.NewAllWISE())
}

// fermiReader makes the reader for the 4FGL file. It needs no profile.
func fermiReader([]radiation.Profile) (reader, error) {
	return asReader(read.NewFermi())
}

// readSearches reads every search the call names. It returns the sources
// and the rows that could not be read, or the error that stopped the
// reading. The call must name at least one search, though the search may
// hold no rows.
func readSearches(call pzm.Call, profiles []radiation.Profile,
	searches []search) (sources []nilgai.Source, unread []error,
	err error) {
	named := false
	for _, each := range searches {
		if !call.Given(each.flag) {
			continue
		}
		named = true
		table, err := each.build(profiles)
		if err != nil {
			return nil, nil, err
		}
		got, missed, err := readFile(table, call.Value(each.flag))
		if err != nil {
			return nil, nil, err
		}
		sources = append(sources, got...)
		unread = append(unread, missed...)
	}
	if !named {
		return nil, nil, errors.New("show needs --chandra, --allwise " +
			"or --fermi")
	}
	return sources, unread, nil
}

// readFile returns every source in the named file and every row that
// could not be read, or the error that stopped the reading.
func readFile(table reader, path string) (sources []nilgai.Source,
	unread []error, err error) {
	in, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer in.Close()
	for source, err := range table.Read(path, in) {
		var row *nilgai.RowError
		switch {
		case errors.As(err, &row):
			unread = append(unread, err)
		case err != nil:
			return nil, nil, err
		default:
			sources = append(sources, *source)
		}
	}
	return sources, unread, nil
}

// asked holds what a show call asked to see:
//   - a centre
//   - the width of each panel in arcseconds
//   - how many pixels across each panel is
type asked struct {
	centre nilgai.Position
	widths []float64
	pixels int
}

// askedFor reads the show call's flags. It needs every one of them.
func askedFor(call pzm.Call) (wanted asked, err error) {
	ra, err := number(call, "ra")
	if err != nil {
		return wanted, err
	}
	dec, err := number(call, "dec")
	if err != nil {
		return wanted, err
	}
	if !call.Given("width") {
		return wanted, fmt.Errorf("show needs --width")
	}
	for _, text := range call.List("width") {
		width, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return wanted, fmt.Errorf("--width: %w", err)
		}
		wanted.widths = append(wanted.widths, width)
	}
	if !call.Given("pixels") {
		return wanted, fmt.Errorf("show needs --pixels")
	}
	wanted.pixels, err = strconv.Atoi(call.Value("pixels"))
	if err != nil {
		return wanted, fmt.Errorf("--pixels: %w", err)
	}
	wanted.centre, err = nilgai.NewPosition(ra, dec)
	return wanted, err
}

// number returns the flag's value as a number. It refuses a flag that is
// not given, or is not a number, and the error names the flag.
func number(call pzm.Call, name string) (value float64, err error) {
	if !call.Given(name) {
		return value, fmt.Errorf("show needs --%s", name)
	}
	value, err = strconv.ParseFloat(call.Value(name), 64)
	if err != nil {
		return value, fmt.Errorf("--%s: %w", name, err)
	}
	return value, nil
}

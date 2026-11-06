package main

import (
	"fmt"
	"io"

	"github.com/janearc/celestial/orbit"
	"github.com/janearc/puzzlemouth/pzm"
)

// screened is a star the screen kept: its row in the product, its id,
// and its nearest pass in parsecs with each estimate of the Sun's speed
// along the rotation.
type screened struct {
	row, id              int64
	nearFrame, nearModel float64
}

// screening is what a screen found: the rows read, kept and refused by
// the frame, and the stars passing within the radius under either
// estimate.
type screening struct {
	rows, kept, refused int64
	found               []screened
}

// screenBatch is the stars waiting for the solver: each once with the
// Sun as published and once at the model's speed.
type screenBatch struct {
	pending          []screened
	framed, modelled []orbit.State
}

// screen is the verb: every kept star of the product run back with each
// estimate of the Sun's speed, and those passing within screen-pc under
// either written to --into.
func (runs runner) screen(call pzm.Call) error {
	by, err := runs.asked(call, "gaia", "into")
	if err != nil {
		return err
	}
	if call.DryRun() {
		fmt.Fprintf(call.Out(), "would read %s, run every star kept at a "+
			"parallax of %g times its error back %g years with each "+
			"estimate of the sun's speed, and write those passing within %g "+
			"pc to %s\n", call.Value("gaia"), by.quality, by.years,
			by.screenPc, call.Value("into"))
		return nil
	}
	catalogue, file, err := opened(call.Value("gaia"))
	if err != nil {
		return err
	}
	defer file.Close()
	at, err := settingFor(runs.read, runs.open, *by)
	if err != nil {
		return err
	}
	defer at.made.Close()
	found, err := at.screen(catalogue, *by)
	if err != nil {
		return err
	}
	return writeTo(call.Value("into"), found.table(*by))
}

// screen runs every kept star of the catalogue back with each estimate
// of the Sun's speed, a batch at a time.
func (at setting) screen(catalogue product, by picked) (*screening,
	error) {
	frame, model, err := at.estimates()
	if err != nil {
		return nil, err
	}
	found := &screening{rows: catalogue.Rows()}
	var batch screenBatch
	for row := range found.rows {
		source, err := catalogue.Source(row)
		if err != nil {
			return nil, err
		}
		if !kept(source, by.quality) {
			continue
		}
		found.kept++
		asFrame, frameErr := source.State(*frame)
		asModel, modelErr := source.State(*model)
		if frameErr != nil || modelErr != nil {
			found.refused++
			continue
		}
		batch.pending = append(batch.pending, screened{row: row,
			id: source.ID()})
		batch.framed = append(batch.framed, *asFrame)
		batch.modelled = append(batch.modelled, *asModel)
		if len(batch.framed)+len(batch.modelled) < at.batch {
			continue
		}
		if batch, err = at.flush(batch, found, by.screenPc); err != nil {
			return nil, err
		}
	}
	if _, err := at.flush(batch, found, by.screenPc); err != nil {
		return nil, err
	}
	return found, nil
}

// flush runs a batch's stars back and adds those passing within the
// radius to what was found, giving back an empty batch.
func (at setting) flush(batch screenBatch, found *screening,
	radius float64) (screenBatch, error) {
	stars := append(batch.framed, batch.modelled...)
	near, err := at.back(stars)
	if err != nil {
		return batch, err
	}
	half := len(batch.pending)
	for which, star := range batch.pending {
		star.nearFrame, star.nearModel = near[which], near[half+which]
		if star.nearFrame < radius || star.nearModel < radius {
			found.found = append(found.found, star)
		}
	}
	return screenBatch{}, nil
}

// table is the screen as a table: a header that says what was read,
// kept and found, then one line a star.
func (found screening) table(by picked) table {
	notes := []string{fmt.Sprintf("firang screen: %d rows, %d kept at a "+
		"parallax of %g times its error, %d refused by the frame, %d "+
		"within %g pc under either sun speed", found.rows, found.kept,
		by.quality, found.refused, len(found.found), by.screenPc)}
	var lines [][]string
	for _, star := range found.found {
		lines = append(lines, []string{fmt.Sprint(star.row),
			fmt.Sprint(star.id), fmt.Sprint(star.nearFrame),
			fmt.Sprint(star.nearModel)})
	}
	return table{notes: notes, columns: screenColumns, lines: lines}
}

// screenColumns are the screen table's columns.
var screenColumns = []string{"row", "source_id", "near_frame_pc",
	"near_model_pc"}

// writeTable writes a table: its notes, each on a line of its own after
// a hash, its columns, and its lines, tab-separated.
func writeTable(out io.Writer, written table) error {
	var lines []string
	for _, note := range written.notes {
		lines = append(lines, "# "+note)
	}
	lines = append(lines, joined(written.columns))
	for _, line := range written.lines {
		lines = append(lines, joined(line))
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(out, line); err != nil {
			return err
		}
	}
	return nil
}

package read

import (
	"errors"

	"github.com/janearc/nilgai"
)

// cell holds one flag column of a row: its name and its value, verbatim.
type cell struct{ name, value string }

// kept returns the measurement with the cells kept as flags and with the
// quality the reader has decided on, as a list of that one measurement,
// or nil on a refusal. The readers hand the list on to the record, which
// refuses an empty one.
//
// nilgai's record refuses:
//   - a flag with no column name
//   - a quality that is not good, doubtful or bad
//
// kept's callers write both in code, so no input can cause either. If a
// reader's code is wrong, kept returns the refusals joined, for the caller
// to return, with a nil list; kept neither hides them nor panics.
func kept(measurement nilgai.Measurement, quality nilgai.Quality,
	cells ...cell) ([]nilgai.Measurement, error) {
	flags := make([]nilgai.Flag, len(cells))
	failed := make([]error, len(cells))
	for place, each := range cells {
		flags[place], failed[place] = nilgai.NewFlag(each.name,
			each.value)
	}
	made, err := measurement.WithReading(quality, flags)
	refused := errors.Join(append(failed, err)...)
	var list []nilgai.Measurement
	if refused == nil {
		list = []nilgai.Measurement{*made}
	}
	return list, refused
}

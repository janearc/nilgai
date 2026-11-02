package nilgai

import "errors"

// Quality is nilgai's reading of a catalogue's flags. Flag keeps the flags
// themselves as they came.
type Quality int

const (
	// qualityUnset means no quality assessment was made. It must not be
	// treated as Good.
	qualityUnset Quality = iota
	// Good means the catalogue's flags raise no doubt.
	Good
	// Doubtful means catalogue flags indicate a possible issue.
	Doubtful
	// Bad means catalogue flags indicate an unreliable measurement.
	Bad
)

// String returns the quality in words, and "invalid" for anything else.
func (quality Quality) String() string {
	switch quality {
	case Good:
		return "good"
	case Doubtful:
		return "doubtful"
	case Bad:
		return "bad"
	}
	return "invalid"
}

// valid says whether the quality is good, doubtful or bad.
func (quality Quality) valid() bool {
	return quality > qualityUnset && quality <= Bad
}

// Flag holds one flag column of a catalogue row, verbatim.
type Flag struct {
	name, value string
}

// NewFlag preserves a catalogue flag. The column name must be nonempty; an
// empty value is allowed because it can carry catalogue meaning.
func NewFlag(name, value string) (flag Flag, err error) {
	if name == "" {
		return flag, errors.New("flag has no column name")
	}
	return Flag{name: name, value: value}, nil
}

// Name returns the catalogue's column name.
func (flag Flag) Name() string { return flag.name }

// Value returns the cell as text, unchanged.
func (flag Flag) Value() string { return flag.value }

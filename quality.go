package nilgai

import "errors"

// Quality is nilgai's reading of a catalogue's flags. Flag keeps the flags
// themselves as they came.
type Quality int

const (
	// qualityUnset is the zero Quality. It means no reading was made, and
	// it must not pass as good.
	qualityUnset Quality = iota
	// Good means the catalogue's flags raise no doubt.
	Good
	// Doubtful means the catalogue's flags say to take care.
	Doubtful
	// Bad means the catalogue's flags say the value is not to be trusted.
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

// NewFlag keeps a flag as the catalogue gave it. The value may be empty,
// because an empty cell is something the catalogue said. The column name
// may not be empty.
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

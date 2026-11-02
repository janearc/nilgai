package band

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/janearc/libtheme-css/spaces/radiation"
)

// profileData is nilgai's false colour: each instrument's band, written in
// its own units, and the slice of visible light it is shown in. The bands
// are cited; the hues are chosen, so they live in data and not in code.
//
//go:embed data/profiles.json
var profileData []byte

// ProfileSource is what a caller of band needs from the profiles: one
// libtheme radiation.Profile for each instrument, or why it has none.
// ProfileFile and Embedded are the two that exist; MockProfileSource, in
// nilgai's mocks, stands in for them in tests.
//
//go:generate -command mock go tool mockgen -package=mock_nilgai
//go:generate mock -destination=../mocks/profiles.go . ProfileSource
type ProfileSource interface {
	Profiles() ([]radiation.Profile, error)
}

// Embedded reads the profiles built into Nilgai from data/profiles.json. Use
// NewEmbedded to construct it.
type Embedded struct{}

// NewEmbedded constructs a source for the built-in profiles.
func NewEmbedded() *Embedded { return &Embedded{} }

// Profiles reads the embedded data/profiles.json.
func (Embedded) Profiles() ([]radiation.Profile, error) {
	return Profiles()
}

// ProfileFile represents data/profiles.json: a provenance note and an ordered
// list of instruments with their colour mappings.
type ProfileFile struct {
	// Source says where the file's bands and hues come from. It is a
	// note for the reader of the file and nilgai does not use it.
	Source string `json:"source"`
	// Entries are the instruments, in the order the file lists them,
	// under the key "profiles".
	Entries []ProfileEntry `json:"profiles"`
}

// ProfileEntry records an instrument’s name, coverage in published units, and
// the visible wavelength range used to display it.
type ProfileEntry struct {
	// Name is the instrument's name, such as "fermi". It is not empty
	// and no other entry has it.
	Name string `json:"name"`
	// Band is the citation for the band. nilgai does not use it.
	Band string `json:"band"`
	// From is one end of the band; From and To may come in either order.
	From Quantity `json:"from"`
	// To is the other end of the band.
	To Quantity `json:"to"`
	// Shown is the range of visible light in nm that the band is drawn
	// in, short end first.
	Shown radiation.Range `json:"shown"`
}

// NewProfileFile decodes JSON shaped like data/profiles.json. Malformed JSON
// and unknown keys return nil and an error identifying the problem.
func NewProfileFile(data []byte) (*ProfileFile, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var file ProfileFile
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("profiles: %w", err)
	}
	return &file, nil
}

// Profiles validates entries and returns radiation profiles in file order.
// Duplicate names and invalid entries return all validation errors together,
// each identifying its profile, with no partial profile list.
func (file *ProfileFile) Profiles() ([]radiation.Profile, error) {
	var refused []error
	seen := map[string]bool{}
	made := make([]radiation.Profile, none, len(file.Entries))
	for _, entry := range file.Entries {
		if seen[entry.Name] {
			refused = append(refused, fmt.Errorf(
				"profile %q is named twice", entry.Name))
			continue
		}
		seen[entry.Name] = true
		profile, err := entry.Profile()
		if err != nil {
			refused = append(refused, fmt.Errorf("profile %q: %w",
				entry.Name, err))
			continue
		}
		made = append(made, *profile)
	}
	if len(refused) != none {
		return nil, errors.Join(refused...)
	}
	return made, nil
}

// Profile validates the entry and constructs its radiation.Profile. An empty
// name, unreadable endpoint, zero-width source band, reversed shown range, or
// shown range outside visible light returns nil and an error. Endpoint errors
// identify which value failed and why.
func (entry ProfileEntry) Profile() (*radiation.Profile, error) {
	if entry.Name == "" {
		return nil, errors.New("it has no name")
	}
	from, err := entry.From.Nanometres()
	if err != nil {
		return nil, fmt.Errorf("its from end: %w", err)
	}
	to, err := entry.To.Nanometres()
	if err != nil {
		return nil, fmt.Errorf("its to end: %w", err)
	}
	if from == to {
		return nil, errors.New("its band has no width")
	}
	shown, visible := entry.Shown, radiation.Visible
	if shown.Lo >= shown.Hi {
		return nil, fmt.Errorf("shown from %v to %v nm runs backwards",
			shown.Lo, shown.Hi)
	}
	if shown.Lo < visible.Lo || shown.Hi > visible.Hi {
		return nil, fmt.Errorf("shown from %v to %v nm leaves visible "+
			"light, %v to %v nm", shown.Lo, shown.Hi, visible.Lo,
			visible.Hi)
	}
	source := radiation.Range{Lo: min(from, to), Hi: max(from, to)}
	return &radiation.Profile{Name: entry.Name, Source: source,
		Render: shown}, nil
}

// Profiles returns the embedded false-colour profiles, one libtheme
// radiation.Profile per instrument, in data/profiles.json order.
func Profiles() ([]radiation.Profile, error) {
	return ReadProfiles(profileData)
}

// ReadProfiles decodes and validates profile JSON by calling NewProfileFile and
// ProfileFile.Profiles.
func ReadProfiles(data []byte) ([]radiation.Profile, error) {
	file, err := NewProfileFile(data)
	if err != nil {
		return nil, err
	}
	return file.Profiles()
}

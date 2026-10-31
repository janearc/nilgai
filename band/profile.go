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

// Embedded is the ProfileSource that reads the profiles built into
// nilgai, from data/profiles.json. NewEmbedded makes one.
type Embedded struct{}

// NewEmbedded is the ProfileSource of nilgai's built-in profiles.
func NewEmbedded() *Embedded { return &Embedded{} }

// Profiles reads the embedded data/profiles.json.
func (Embedded) Profiles() ([]radiation.Profile, error) {
	return Profiles()
}

// ProfileFile is data/profiles.json as it is written: a note on where it
// comes from, and one entry for each instrument in the order the file
// lists them.
type ProfileFile struct {
	// Source says where the file's bands and hues come from. It is a
	// note for the reader of the file and nilgai does not use it.
	Source string `json:"source"`
	// Entries are the instruments, in the order the file lists them,
	// under the key "profiles".
	Entries []ProfileEntry `json:"profiles"`
}

// ProfileEntry is one instrument in data/profiles.json: its name, its
// band in the units of its paper, and the slice of visible light it is
// shown in.
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

// NewProfileFile reads JSON shaped as data/profiles.json is. NewProfileFile
// refuses JSON it cannot read, and a key the file does not have, naming
// the key. NewProfileFile returns nil on a refusal.
func NewProfileFile(data []byte) (*ProfileFile, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var file ProfileFile
	if err := decoder.Decode(&file); err != nil {
		return nil, fmt.Errorf("profiles: %w", err)
	}
	return &file, nil
}

// Profiles checks every entry and gives one radiation.Profile for each,
// in the file's order. Profiles refuses a name used twice, and anything
// ProfileEntry.Profile refuses. Profiles returns every refusal together,
// each naming its profile, and no profiles.
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

// Profile checks the entry and gives its radiation.Profile. Profile
// refuses:
//
//   - no name
//   - an end it cannot read, naming which end and why
//   - a band with no width
//   - a shown range that runs backwards
//   - a shown range that leaves visible light
//
// Profile returns nil on a refusal.
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

// Profiles are nilgai's false colour, one libtheme radiation.Profile per
// instrument, read from the embedded data/profiles.json in the order it
// lists them.
func Profiles() ([]radiation.Profile, error) {
	return ReadProfiles(profileData)
}

// ReadProfiles reads profiles from JSON shaped as data/profiles.json is,
// and gives what NewProfileFile and ProfileFile.Profiles give.
func ReadProfiles(data []byte) ([]radiation.Profile, error) {
	file, err := NewProfileFile(data)
	if err != nil {
		return nil, err
	}
	return file.Profiles()
}

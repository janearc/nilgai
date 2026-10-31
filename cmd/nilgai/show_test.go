package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/astrogo/fitsio"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/nilgai/band"
	"github.com/janearc/nilgai/internal/fitstest"
	mock_nilgai "github.com/janearc/nilgai/mocks"
	"go.uber.org/mock/gomock"
)

//go:embed testdata/*.json
var fixtures embed.FS

// none is the count of nothing, zero.
const none = 0

// one is the count of a single thing.
const one = 1

// columnSpec describes a column of a 4FGL-shaped table as the fixture
// names it.
type columnSpec struct{ Name, Format string }

// fermiRowSpec describes a row of such a table as the fixture writes it,
// where a number may be the string "NaN".
type fermiRowSpec struct {
	Name                                          string
	RA, Dec, Major, Minor, Angle, Flux, FluxError flexible
}

// flexible is a number that a fixture may write as the string "NaN".
type flexible float64

// UnmarshalJSON reads a number, or the string "NaN".
func (number *flexible) UnmarshalJSON(data []byte) error {
	if string(data) == `"NaN"` {
		*number = flexible(math.NaN())
		return nil
	}
	var plain float64
	err := json.Unmarshal(data, &plain)
	*number = flexible(plain)
	return err
}

// call describes a call of show in the fixture:
//   - its name
//   - the arguments after the verb, in which {name} stands for a file
//   - the words wanted on standard error
type call struct {
	Name string
	Code int
	Args []string
	Want []string
}

// exits returns the exit code of a refused call. It is the call's own
// code, or 1 when the fixture names none.
func (each call) exits() int {
	if each.Code == none {
		return exitFailed
	}
	return each.Code
}

// showFixture holds testdata/show.json.
type showFixture struct {
	ChandraHeader, ChandraGood, ChandraBad string
	AllwiseHeader, AllwiseGood, AllwiseBad string
	NotATable                              string
	Columns                                []columnSpec
	FermiGood, FermiExtended               fermiRowSpec
	Pixels                                 int
	Panels                                 struct {
		Widths []string
		Bounds string
	}
	Repeats int
	Draw    []call
	Refused []call
	Unread  []call
}

// loadShow reads the show fixture.
func loadShow(test *testing.T) showFixture {
	test.Helper()
	data, err := fixtures.ReadFile("testdata/show.json")
	if err != nil {
		test.Fatal(err)
	}
	var read showFixture
	if err := json.Unmarshal(data, &read); err != nil {
		test.Fatal(err)
	}
	return read
}

// writeText writes text to a file of the name in the directory and
// returns its path.
func writeText(test *testing.T, directory, name, text string) string {
	test.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(text), os.ModePerm); err != nil {
		test.Fatal(err)
	}
	return path
}

// writeFermi writes a 4FGL-shaped FITS table of the rows to a file of the
// name in the directory and returns its path.
func writeFermi(test *testing.T, read showFixture, directory, name string,
	rows ...fermiRowSpec) string {
	test.Helper()
	var out bytes.Buffer
	columns := make([]fitsio.Column, len(read.Columns))
	written := make([]any, len(rows))
	for place, each := range read.Columns {
		columns[place] = fitsio.Column{Name: each.Name, Format: each.Format}
	}
	for place, each := range rows {
		written[place] = fermiRow(each)
	}
	err := fitstest.Write(&out, "LAT_Point_Source_Catalog", columns,
		written...)
	if err != nil {
		test.Fatal(err)
	}
	return writeText(test, directory, name, out.String())
}

// fermiRow returns the row to write to the table, by 4FGL's column names.
func fermiRow(each fermiRowSpec) any {
	return &struct {
		Name      string  `fits:"Source_Name"`
		RA        float32 `fits:"RAJ2000"`
		Dec       float32 `fits:"DEJ2000"`
		Major     float32 `fits:"Conf_95_SemiMajor"`
		Minor     float32 `fits:"Conf_95_SemiMinor"`
		Angle     float32 `fits:"Conf_95_PosAng"`
		Flux      float32 `fits:"Energy_Flux100"`
		FluxError float32 `fits:"Unc_Energy_Flux100"`
	}{each.Name, float32(each.RA), float32(each.Dec), float32(each.Major),
		float32(each.Minor), float32(each.Angle), float32(each.Flux),
		float32(each.FluxError)}
}

// files writes the fixture's named files to a directory and returns their
// paths, with the path of one file that is not there.
func files(test *testing.T, read showFixture) map[string]string {
	test.Helper()
	directory := test.TempDir()
	return map[string]string{
		"missing": filepath.Join(directory, "missing"),
		"chandra": writeText(test, directory, "chandra",
			read.ChandraHeader+read.ChandraGood),
		"chandraRows": writeText(test, directory, "chandraRows",
			read.ChandraHeader+read.ChandraGood+read.ChandraBad),
		"allwise": writeText(test, directory, "allwise",
			read.AllwiseHeader+read.AllwiseGood),
		"allwiseRows": writeText(test, directory, "allwiseRows",
			read.AllwiseHeader+read.AllwiseGood+read.AllwiseBad),
		"notATable": writeText(test, directory, "notATable",
			read.NotATable),
		"fermi": writeFermi(test, read, directory, "fermi",
			read.FermiGood),
		"fermiExtended": writeFermi(test, read, directory, "fermiExtended",
			read.FermiGood, read.FermiExtended),
	}
}

// argsFor returns a call's arguments after the verb, with each {name}
// replaced by the path of the file of that name.
func argsFor(each call, named map[string]string) []string {
	args := []string{"show"}
	for _, arg := range each.Args {
		for name, path := range named {
			arg = strings.ReplaceAll(arg, "{"+name+"}", path)
		}
		args = append(args, arg)
	}
	return args
}

// runShow runs the call and returns the exit code and what it wrote.
func runShow(each call, named map[string]string) (code int, out, errs []byte) {
	var wrote, said bytes.Buffer
	code = run(argsFor(each, named), nothingToRead(), &wrote, &said, "")
	return code, wrote.Bytes(), said.Bytes()
}

// TestShowRefusesWhatItCannotDraw checks that a call that cannot be
// answered exits 1, writes no picture, and says what was wrong, naming the
// flag or the file.
func TestShowRefusesWhatItCannotDraw(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	for _, each := range read.Refused {
		code, out, errs := runShow(each, named)
		if code != each.exits() || len(out) != none {
			test.Errorf("%s: code %d and %d bytes of picture, want code "+
				"%d", each.Name, code, len(out), each.exits())
		}
		for _, word := range each.Want {
			if !strings.Contains(string(errs), word) {
				test.Errorf("%s: %q does not say %q", each.Name, errs, word)
			}
		}
	}
}

// pictureOf decodes what a call wrote. It fails the test if the output is
// not a png.
func pictureOf(test *testing.T, out []byte) image.Image {
	test.Helper()
	picture, err := png.Decode(bytes.NewReader(out))
	if err != nil {
		test.Fatal(err)
	}
	return picture
}

// inked says whether the picture has more than one colour in it.
func inked(picture image.Image) bool {
	bounds := picture.Bounds()
	corner := picture.At(bounds.Min.X, bounds.Min.Y)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if picture.At(x, y) != corner {
				return true
			}
		}
	}
	return false
}

// TestShowDrawsWhatItIsGiven checks that a call it can answer exits 0,
// says nothing, and writes a png of the size asked for with something
// drawn in it.
func TestShowDrawsWhatItIsGiven(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	for _, each := range read.Draw {
		code, out, errs := runShow(each, named)
		if code != exitRan || len(errs) != none {
			test.Errorf("%s: code %d and %q", each.Name, code, errs)
			continue
		}
		picture := pictureOf(test, out)
		if picture.Bounds().Dx() != read.Pixels ||
			picture.Bounds().Dy() != read.Pixels || !inked(picture) {
			test.Errorf("%s: a picture of %v, inked %v", each.Name,
				picture.Bounds(), inked(picture))
		}
	}
}

// TestShowDrawsNothingWhereThereIsNothing is an anti-test. A call centred
// where the search has nothing draws a picture without any source and does
// not fail. The picture is not the one centred on the source.
func TestShowDrawsNothingWhereThereIsNothing(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	far := call{Args: []string{"--chandra", "{chandra}", "--ra", "10",
		"--dec", "10", "--width", "12", "--pixels", "300"}}
	code, out, errs := runShow(far, named)
	_, near, _ := runShow(read.Draw[none], named)
	if code != exitRan || len(errs) != none || bytes.Equal(out, near) {
		test.Errorf("code %d, %q, the same picture as at the source: %v",
			code, errs, bytes.Equal(out, near))
	}
}

// TestShowNamesUnreadRowsAfterThePicture checks that rows that cannot be
// read are named after the picture is written. The call fails, the
// picture is whole with what could be read, and every row that was lost is
// on standard error.
func TestShowNamesUnreadRowsAfterThePicture(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	for _, each := range read.Unread {
		code, out, errs := runShow(each, named)
		picture := pictureOf(test, out)
		if code != exitFailed || picture.Bounds().Dx() != read.Pixels ||
			!inked(picture) {
			test.Errorf("%s: code %d and a picture of %v, inked %v",
				each.Name, code, picture.Bounds(), inked(picture))
		}
		for _, word := range each.Want {
			if !strings.Contains(string(errs), word) {
				test.Errorf("%s: %q does not say %q", each.Name, errs, word)
			}
		}
	}
}

// TestShowDrawsOnePanelForEachWidth checks that two widths give two
// panels side by side, with the gap between them.
func TestShowDrawsOnePanelForEachWidth(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	args := []string{"--chandra", "{chandra}", "--ra", "266.4168371",
		"--dec", "-29.0078106", "--pixels", "300", "--width",
		strings.Join(read.Panels.Widths, ",")}
	_, out, _ := runShow(call{Args: args}, named)
	if got := pictureOf(test, out).Bounds().String(); got !=
		read.Panels.Bounds {
		test.Errorf("two panels are %s, want %s", got, read.Panels.Bounds)
	}
}

// TestShowDrawsTheSamePictureEachTime checks that the same call draws the
// same picture every time, and that the order the flags are given in does
// not change it.
func TestShowDrawsTheSamePictureEachTime(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	common := []string{"--ra", "266.4168371", "--dec", "-29.0078106",
		"--width", "12", "--pixels", "300"}
	forward := call{Args: append([]string{"--chandra", "{chandra}",
		"--allwise", "{allwise}", "--fermi", "{fermi}"}, common...)}
	backward := call{Args: append([]string{"--fermi", "{fermi}",
		"--allwise", "{allwise}", "--chandra", "{chandra}"}, common...)}
	_, first, _ := runShow(forward, named)
	for range read.Repeats {
		_, again, _ := runShow(forward, named)
		_, reversed, _ := runShow(backward, named)
		if !bytes.Equal(first, again) || !bytes.Equal(first, reversed) {
			test.Fatal("the same call drew a different picture")
		}
	}
}

// TestShowFailsWhenItCannotWrite checks that a picture that cannot be
// written is a failure, and that the call says so.
func TestShowFailsWhenItCannotWrite(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	var said bytes.Buffer
	args := argsFor(read.Draw[none], named)
	code := run(args, nothingToRead(), closedWriter{}, &said, "")
	if code != exitFailed || said.Len() == none {
		test.Errorf("a closed output: code %d and %q", code, said.String())
	}
}

// TestShowFailsWhenAReaderCannotBeMade is an anti-test. A search whose
// reader cannot be made fails the call, and nothing is drawn.
func TestShowFailsWhenAReaderCannotBeMade(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	unreadable := []search{{"chandra", noReader}, {"allwise", noReader},
		{"fermi", noReader}}
	var wrote, said bytes.Buffer
	code := toolWith("", unreadable, band.NewEmbedded()).RunWith(
		argsFor(read.Draw[none], named), nothingToRead(), &wrote, &said)
	out, errs := wrote.Bytes(), said.Bytes()
	if code != exitFailed || len(out) != none ||
		!strings.Contains(string(errs), errNoReader.Error()) {
		test.Errorf("code %d, %d bytes drawn and %q", code, len(out), errs)
	}
}

// TestShowFailsWhenTheProfilesCannotBeRead is the anti-test: profiles that
// cannot be read fail the call, asked for once, and nothing is drawn.
func TestShowFailsWhenTheProfilesCannotBeRead(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	errProfiles := errors.New("no profiles today")
	source := mock_nilgai.NewMockProfileSource(gomock.NewController(test))
	source.EXPECT().Profiles().Return(nil, errProfiles).Times(one)
	var wrote, said bytes.Buffer
	code := toolWith("", realSearches(), source).RunWith(
		argsFor(read.Draw[none], named), nothingToRead(), &wrote, &said)
	if code != exitFailed || wrote.Len() != none ||
		!strings.Contains(said.String(), errProfiles.Error()) {
		test.Errorf("code %d, %d bytes drawn and %q", code, wrote.Len(),
			said.Bytes())
	}
}

// TestShowFailsWhenTheProfilesGiveChandraNoBand: profiles with none for
// Chandra fail the call at the Chandra reader, and nothing is drawn: a
// reader that was refused is not used.
func TestShowFailsWhenTheProfilesGiveChandraNoBand(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	source := mock_nilgai.NewMockProfileSource(gomock.NewController(test))
	source.EXPECT().Profiles().Return([]radiation.Profile{}, nil).Times(one)
	var wrote, said bytes.Buffer
	code := toolWith("", realSearches(), source).RunWith(
		argsFor(read.Draw[none], named), nothingToRead(), &wrote, &said)
	if code != exitFailed || wrote.Len() != none ||
		!strings.Contains(said.String(), "chandra") {
		test.Errorf("code %d, %d bytes drawn and %q", code, wrote.Len(),
			said.Bytes())
	}
}

// TestShowDrawsWithTheProfilesItIsGiven: show draws with the profiles it
// is given, and asks for them once.
func TestShowDrawsWithTheProfilesItIsGiven(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	profiles, err := band.Profiles()
	if err != nil {
		test.Fatal(err)
	}
	source := mock_nilgai.NewMockProfileSource(gomock.NewController(test))
	source.EXPECT().Profiles().Return(profiles, nil).Times(one)
	var wrote, said bytes.Buffer
	code := toolWith("", realSearches(), source).RunWith(
		argsFor(read.Draw[none], named), nothingToRead(), &wrote, &said)
	picture := pictureOf(test, wrote.Bytes())
	if code != exitRan || !inked(picture) {
		test.Errorf("code %d and %q", code, said.Bytes())
	}
}

// errNoReader is the error that noReader returns.
var errNoReader = errors.New("no such reader")

// noReader stands for a reader that cannot be made. It always returns
// errNoReader.
func noReader([]radiation.Profile) (reader, error) {
	return nil, errNoReader
}

// TestShowDrawsWhatIsNamedAndNoMore is an anti-test. What is drawn
// depends on what is named. Three searches draw something that one search
// alone does not.
func TestShowDrawsWhatIsNamedAndNoMore(test *testing.T) {
	read := loadShow(test)
	named := files(test, read)
	_, alone, _ := runShow(read.Draw[none], named)
	_, together, _ := runShow(read.Draw[len(read.Draw)-one], named)
	if bytes.Equal(alone, together) {
		test.Error("one search and three draw the same picture")
	}
}

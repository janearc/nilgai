package main

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
)

// The candidates tests' made-up stars, each defined: a draws table's
// line of numbers, in the draws table's columns after the id.
const (
	chosenID   = 101 // defined: meets every number
	slowID     = 102 // defined: too slow through the galaxy
	seldomID   = 103 // defined: too seldom near
	poorRVID   = 104 // defined: its radial velocity too poorly measured
	oftenNear  = 0.4 // defined: a share of draws near, over the cut
	seldomNear = 0.1 // defined: a share under it
	fastKms    = 467 // defined: km/s through the galaxy, over the cut
	slowKms    = 300 // defined: km/s, under it
	goodRV     = 1.9 // defined: km/s, a radial velocity error under the cut
	poorRV     = 6.8 // defined: km/s, one over it
)

// drawsLine is a draws table's line for a made-up star: its id, its
// share near, its speed through the galaxy and its radial velocity's
// error, the rest picked so the readme's numbers can be read back.
func drawsLine(id int64, share, speed, rvError float64) string {
	return fmt.Sprintf("%d\t120\t130\t%g\t1\t60\t110\t230\t10\t0\t0\t2.7\t"+
		"%g\t466\t9.4\t5.72\t-423.1\t%g\t12.44\t0.68\t269.30578\t-36.5411",
		id, share, speed, rvError)
}

// drawnAt writes a draws table of the lines given and gives its path.
func drawnAt(test *testing.T, lines ...string) string {
	test.Helper()
	text := "# a draws table, for a test\n" + strings.Join(drawsColumns,
		"\t") + "\n" + strings.Join(lines, "\n") + "\n"
	return tableOf(test, text)
}

// the test: of four stars, the one fast enough, near often enough and
// with its radial velocity measured well enough is chosen: its folder is
// written with its readme, its numbers as the table had them, and its id
// goes to standard output.
func TestTheCandidatesAreTheStarsThatMeetEveryNumber(test *testing.T) {
	lines := []string{drawsLine(chosenID, oftenNear, fastKms, goodRV),
		drawsLine(slowID, oftenNear, slowKms, goodRV),
		drawsLine(seldomID, seldomNear, fastKms, goodRV),
		drawsLine(poorRVID, oftenNear, fastKms, poorRV)}
	drawn := drawnAt(test, lines...)
	into := test.TempDir()
	open := NewMockopener(gomock.NewController(test))
	code, out, errs := firangRan(open, published{}, "candidates",
		"--numbers", numbersPath, "--drawn", drawn, "--into", into)
	if code != exitRan || strings.TrimSpace(out) != "101" {
		test.Fatalf("exit %d, wrote %q: %s", code, out, errs)
	}
	folders, err := os.ReadDir(into)
	if err != nil || len(folders) != oneLine {
		test.Fatalf("the candidates wrote %v, %v", folders, err)
	}
	raw, err := os.ReadFile(filepath.Join(into, "101", readme))
	if err != nil {
		test.Fatal(err)
	}
	page := string(raw)
	for _, want := range []string{"# 101\n",
		"place             ra 269.30578, dec -36.54110 degrees",
		"distance          5.72 kpc, parallax 9.4 times its error",
		"radial velocity   -423.1 +- 1.90 km/s", "G magnitude       12.44",
		"from the centre   2.70 kpc now, moving 467 km/s, 466 outward",
		"draws             40% within 100 pc, 100% within 1000 pc",
		"nearest pass      median 110 pc, 16th to 84th percentile 60 to 230 pc"} {
		if !strings.Contains(page, want) {
			test.Errorf("the readme lacks %q:\n%s", want, page)
		}
	}
}

// the anti-test: a draws table the candidates cannot read, or a folder
// they cannot write, stops them with its words.
func TestWhatTheCandidatesCannotReadOrWriteStopsThem(test *testing.T) {
	blocked := filepath.Join(test.TempDir(), "a file")
	if err := os.WriteFile(blocked, nil, fileBit); err != nil {
		test.Fatal(err)
	}
	columns := slices.Clone(drawsColumns)
	columns = slices.Delete(columns, firstStarColumn,
		firstStarColumn+oneLine)
	header := strings.Join(columns, "\t") + "\n"
	chosen := drawsLine(chosenID, oftenNear, fastKms, goodRV)
	worded := strings.Replace(chosen, "466", "outward", oneLine)
	cases := map[string][]string{
		"has no column g_mag": {tableOf(test, header), test.TempDir()},
		"invalid syntax":      {drawnAt(test, worded), test.TempDir()},
		"not a directory":     {drawnAt(test, chosen), blocked},
	}
	for said, paths := range cases {
		open := NewMockopener(gomock.NewController(test))
		code, _, errs := firangRan(open, published{}, "candidates",
			"--numbers", numbersPath, "--drawn", paths[firstLine], "--into",
			paths[secondLine])
		if code == exitRan || !strings.Contains(errs, said) {
			test.Errorf("%s: exit %d, %s", said, code, errs)
		}
	}
}

// firstStarColumn is g_mag's place among the draws table's columns, the
// column the anti-test leaves out.
var firstStarColumn = slices.Index(drawsColumns, "g_mag")

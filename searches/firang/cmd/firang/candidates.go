package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/janearc/puzzlemouth/pzm"
)

// candidateColumns are the draws table's columns a candidate is chosen
// and written by.
var candidateColumns = []string{"source_id", "ra_deg", "dec_deg",
	"distance_kpc", "parallax_over_error", "rv_kms", "rv_error_kms",
	"g_mag", "r_gc_kpc", "v_gc_kms", "v_out_kms", "share_near",
	"share_screen", "p16_pc", "p50_pc", "p84_pc"}

// The candidates' own numbers and marks, each defined.
const (
	fence     = "```" // defined: a block of numbers starts and ends so
	percent   = 100   // defined: a share as a percentage
	folderBit = 0o755 // defined: a folder its owner writes and all read
	fileBit   = 0o644 // defined: a file its owner writes and all read
	readme    = "README.md"
)

// candidate is a star in the draws table, its numbers by column.
type candidate struct {
	id      string
	numbers map[string]float64
}

// candidates is the verb: the stars in the draws table that move fast
// through the galaxy, pass near Sgr A* often across their draws, and
// have well-measured radial velocities, each a folder in --into named
// by its id, the ids to standard output.
func (runs runner) candidates(call pzm.Call) error {
	by, err := runs.asked(call, "drawn", "into")
	if err != nil {
		return err
	}
	if call.DryRun() {
		fmt.Fprintf(call.Out(), "would write a folder in %s for each star "+
			"in %s at %g km/s or more, at least %g of its draws within %g "+
			"pc, its radial velocity good to %g km/s\n", call.Value("into"),
			call.Value("drawn"), by.speed, by.share, by.nearPc, by.rvError)
		return nil
	}
	table, err := readFrom(call.Value("drawn"), candidateColumns...)
	if err != nil {
		return err
	}
	chosen, err := choose(*table, *by)
	if err != nil {
		return err
	}
	for _, each := range chosen {
		if err := each.write(call.Value("into"), *by); err != nil {
			return err
		}
		fmt.Fprintln(call.Out(), each.id)
	}
	return nil
}

// choose is the stars in the table that meet the numbers, in the
// table's order.
func choose(table read, by picked) ([]candidate, error) {
	var chosen []candidate
	for _, line := range table.lines {
		each := candidate{id: table.text(line, "source_id"),
			numbers: map[string]float64{}}
		for _, column := range candidateColumns[oneStar:] {
			value, err := table.number(line, column)
			if err != nil {
				return nil, fmt.Errorf("star %s: %w", each.id, err)
			}
			each.numbers[column] = value
		}
		if each.numbers["v_gc_kms"] >= by.speed &&
			each.numbers["share_near"] >= by.share &&
			each.numbers["rv_error_kms"] <= by.rvError {
			chosen = append(chosen, each)
		}
	}
	return chosen, nil
}

// write writes the candidate's folder in the folder given: its readme,
// with its numbers as the draws found them.
func (each candidate) write(into string, by picked) error {
	folder := filepath.Join(into, each.id)
	if err := os.MkdirAll(folder, folderBit); err != nil {
		return err
	}
	page := strings.Join(each.page(by), "\n") + "\n"
	return os.WriteFile(filepath.Join(folder, readme), []byte(page),
		fileBit)
}

// page is the candidate's readme, line by line.
func (each candidate) page(by picked) []string {
	at := each.numbers
	return []string{"# " + each.id, "",
		"A candidate from firang's draws, unreviewed.", "", fence,
		fmt.Sprintf("place             ra %.5f, dec %.5f degrees",
			at["ra_deg"], at["dec_deg"]),
		fmt.Sprintf("distance          %.2f kpc, parallax %.1f times its "+
			"error", at["distance_kpc"], at["parallax_over_error"]),
		fmt.Sprintf("radial velocity   %.1f +- %.2f km/s", at["rv_kms"],
			at["rv_error_kms"]),
		fmt.Sprintf("G magnitude       %.2f", at["g_mag"]),
		fmt.Sprintf("from the centre   %.2f kpc now, moving %.0f km/s, "+
			"%.0f outward", at["r_gc_kpc"], at["v_gc_kms"], at["v_out_kms"]),
		fmt.Sprintf("draws             %.0f%% within %g pc, %.0f%% within "+
			"%g pc", at["share_near"]*percent, by.nearPc,
			at["share_screen"]*percent, by.screenPc),
		fmt.Sprintf("nearest pass      median %.0f pc, 16th to 84th "+
			"percentile %.0f to %.0f pc", at["p50_pc"], at["p16_pc"],
			at["p84_pc"]),
		fence, "",
		"Next: the archive's quality columns for its radial velocity, then",
		"its spectrum, for the refractory elements against the volatile",
		"ones and for lithium."}
}

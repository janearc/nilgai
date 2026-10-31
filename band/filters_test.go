package band_test

import (
	_ "embed"
	"encoding/json"
	"strings"
	"testing"

	"github.com/janearc/nilgai/band"
)

// shippedFilters holds data/filters.json as nilgai ships it.
//
//go:embed data/filters.json
var shippedFilters []byte

// filtersFixture holds testdata/filters.json.
type filtersFixture struct {
	Shipped []struct {
		Name        string
		LoNm, HiNm  Num
		Approximate bool
	}
	NotW1 struct{ LoNm, HiNm Num }
	Docs  []struct {
		Name       string
		Filter     string
		LoNm, HiNm Num
		Doc        json.RawMessage
	}
	BadDocs []struct {
		Name   string
		Filter string
		Raw    string
		Want   []string
		Doc    json.RawMessage
	}
}

// loadFilters reads the filters fixture.
func loadFilters(test *testing.T) filtersFixture {
	test.Helper()
	var read filtersFixture
	load(test, "filters", &read)
	return read
}

// TestFilterBandsAreTheShippedEdges checks that each shipped filter
// covers the bands that its edges, worked out by hand in the fixture, give.
func TestFilterBandsAreTheShippedEdges(test *testing.T) {
	for _, each := range loadFilters(test).Shipped {
		first, last, err := band.FilterBands(each.Name)
		wantFirst, wantLast, wantErr := band.BandsBetween(float64(each.LoNm),
			float64(each.HiNm))
		if err != nil || wantErr != nil || first != wantFirst ||
			last != wantLast {
			test.Errorf("%s: bands %v to %v (%v), want %v to %v (%v)",
				each.Name, first, last, err, wantFirst, wantLast,
				wantErr)
		}
	}
}

// TestW1IsNotW1AndW2 is an anti-test. It checks that the W1 filter covers
// W1 alone, and not the W1 and W2 bands together that the neowise colour
// profile covers.
func TestW1IsNotW1AndW2(test *testing.T) {
	read := loadFilters(test)
	first, last, err := band.FilterBands("W1")
	both, bothLast, bothErr := band.BandsBetween(float64(read.NotW1.LoNm),
		float64(read.NotW1.HiNm))
	if err != nil || bothErr != nil {
		test.Fatal(err, bothErr)
	}
	if first == both && last == bothLast {
		test.Errorf("W1 covers %v to %v, which is W1 and W2 together",
			first, last)
	}
	if last >= bothLast {
		test.Errorf("W1 ends in band %v, no sooner than W1 and W2 do",
			last)
	}
}

// TestShippedFiltersAreCitedAndHonest checks that every shipped filter has
// a citation and a description, and that its approximate flag matches what
// the fixture knows.
func TestShippedFiltersAreCitedAndHonest(test *testing.T) {
	var shipped struct {
		Filters []struct {
			Name, Band, Cite string
			Approximate      *bool
		}
	}
	if err := json.Unmarshal(shippedFilters, &shipped); err != nil {
		test.Fatal(err)
	}
	known := loadFilters(test).Shipped
	if len(shipped.Filters) != len(known) {
		test.Fatalf("%d filters shipped, the fixture knows %d",
			len(shipped.Filters), len(known))
	}
	for place, each := range shipped.Filters {
		if each.Name != known[place].Name || each.Band == "" ||
			each.Cite == "" || each.Approximate == nil ||
			*each.Approximate != known[place].Approximate {
			test.Errorf("%+v: not cited, not described, or wrong about "+
				"its edges being approximate", each)
		}
	}
}

// TestApproximateFiltersSayTheyAre is an anti-test. It checks that a
// filter's description contains the word "approximate" if and only if its
// approximate flag is set.
func TestApproximateFiltersSayTheyAre(test *testing.T) {
	var shipped struct {
		Filters []struct {
			Name, Band  string
			Approximate bool
		}
	}
	if err := json.Unmarshal(shippedFilters, &shipped); err != nil {
		test.Fatal(err)
	}
	for _, each := range shipped.Filters {
		says := strings.Contains(each.Band, "approximate")
		if says != each.Approximate {
			test.Errorf("%s: approximate is %v and its description says "+
				"%v", each.Name, each.Approximate, says)
		}
	}
}

// TestReadFilterBandsReadsTheFilterAsked checks that ReadFilterBands
// returns the filter it is asked for, whichever order that filter's edges
// are in and whatever other filters sit beside it.
func TestReadFilterBandsReadsTheFilterAsked(test *testing.T) {
	for _, each := range loadFilters(test).Docs {
		first, last, err := band.ReadFilterBands(each.Doc, each.Filter)
		wantFirst, wantLast, wantErr := band.BandsBetween(float64(each.LoNm),
			float64(each.HiNm))
		if err != nil || wantErr != nil || first != wantFirst ||
			last != wantLast {
			test.Errorf("%s: bands %v to %v (%v), want %v to %v (%v)",
				each.Name, first, last, err, wantFirst, wantLast,
				wantErr)
		}
	}
}

// TestReadFilterBandsRefusesBadDocs checks that ReadFilterBands refuses,
// and says why:
//   - a document it cannot read
//   - a filter that is missing, or listed twice
//   - a unit it does not read
//   - an edge of zero
//   - a band with no width
func TestReadFilterBandsRefusesBadDocs(test *testing.T) {
	for _, each := range loadFilters(test).BadDocs {
		data := each.Doc
		if each.Raw != "" {
			data = json.RawMessage(each.Raw)
		}
		_, _, err := band.ReadFilterBands(data, each.Filter)
		if err == nil {
			test.Errorf("%s: read", each.Name)
			continue
		}
		for _, word := range each.Want {
			if !strings.Contains(err.Error(), word) {
				test.Errorf("%s: %q does not say %q", each.Name, err,
					word)
			}
		}
	}
}

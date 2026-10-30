package nilgai_test

import (
	"errors"
	"iter"
	"slices"
	"strings"
	"testing"

	"github.com/janearc/nilgai"
	mocks "github.com/janearc/nilgai/mocks"
	"go.uber.org/mock/gomock"
)

// The mocks satisfy the interfaces they stand for.
var (
	_ nilgai.FermiReader   = (*mocks.MockFermiReader)(nil)
	_ nilgai.MeerKATReader = (*mocks.MockMeerKATReader)(nil)
	_ nilgai.ChandraReader = (*mocks.MockChandraReader)(nil)
	_ nilgai.NEOWISEReader = (*mocks.MockNEOWISEReader)(nil)
)

// yielded is one item a mock reader gives, or the error in its place.
type yielded[Item any] struct {
	item Item
	err  error
}

// replay plays a list of yielded items to the caller of a sequence.
type replay[Item any] struct{ items []yielded[Item] }

// play yields the items in order, stopping when the caller does.
func (played replay[Item]) play(yield func(Item, error) bool) {
	for _, each := range played.items {
		if !yield(each.item, each.err) {
			return
		}
	}
}

// sequence is the iterator a mock reader returns.
func sequence[Item any](items []yielded[Item]) iter.Seq2[Item, error] {
	return replay[Item]{items: items}.play
}

// sourcesAndAnError is a good source followed by a row error.
func sourcesAndAnError(test *testing.T) ([]yielded[nilgai.Source], error) {
	test.Helper()
	read := loadWorld(test)
	band := read.W1.build(test)
	detection := read.density(test, band)
	got, err := source(test, read, []nilgai.Measurement{detection})
	if err != nil {
		test.Fatal(err)
	}
	spec := read.RowError
	bad := nilgai.NewRowError(spec.File, spec.Row, spec.Column,
		errors.New("bad"))
	return []yielded[nilgai.Source]{{item: got}, {err: bad}}, bad
}

// drain reads a whole sequence of sources.
func drain[Item any](
	items iter.Seq2[Item, error],
) (got []Item, errs []error) {
	for item, err := range items {
		got = append(got, item)
		errs = append(errs, err)
	}
	return got, errs
}

// failures is the errors among errs that are not nil.
func failures(errs []error) []error {
	return slices.DeleteFunc(slices.Clone(errs), isNil)
}

// isNil says whether err is nil.
func isNil(err error) bool { return err == nil }

func TestSourceReadersYieldSourcesAndRowErrors(test *testing.T) {
	ctrl := gomock.NewController(test)
	items, bad := sourcesAndAnError(test)
	fermi := mocks.NewMockFermiReader(ctrl)
	meerkat := mocks.NewMockMeerKATReader(ctrl)
	chandra := mocks.NewMockChandraReader(ctrl)
	fermi.EXPECT().Read("in", gomock.Any()).Return(sequence(items))
	meerkat.EXPECT().Read("in", gomock.Any()).Return(sequence(items))
	chandra.EXPECT().Read("in", gomock.Any()).Return(sequence(items))
	in := strings.NewReader("")
	for name, seq := range map[string]iter.Seq2[nilgai.Source, error]{
		"fermi":   fermi.Read("in", in),
		"meerkat": meerkat.Read("in", in),
		"chandra": chandra.Read("in", in),
	} {
		got, errs := drain(seq)
		failed := failures(errs)
		if len(got) != len(items) || !slices.Equal(failed, []error{bad}) {
			test.Errorf("%s: %v", name, errs)
		}
	}
}

func TestNEOWISEReaderYieldsLightCurves(test *testing.T) {
	read := loadWorld(test)
	first, err := curve(test, read, looks(test, read))
	if err != nil {
		test.Fatal(err)
	}
	reader := mocks.NewMockNEOWISEReader(gomock.NewController(test))
	reader.EXPECT().Read("w.xml", gomock.Any()).Return(sequence(
		[]yielded[nilgai.LightCurve]{{item: first}, {item: first}}))
	in := strings.NewReader("")
	seq := reader.Read("w.xml", in)
	got, errs := drain(seq)
	if len(got) != len(errs) || len(failures(errs)) != none ||
		got[none].ID() != first.ID() {
		test.Errorf("%v %v", got, errs)
	}
}

// A reader that cannot read its input says so, and a caller that stops
// at the failure sees nothing after it.
func TestReadersEndOnAnInputFailure(test *testing.T) {
	items, _ := sourcesAndAnError(test)
	broken := errors.New("input ended early")
	reader := mocks.NewMockFermiReader(gomock.NewController(test))
	played := append([]yielded[nilgai.Source]{{err: broken}}, items...)
	reader.EXPECT().Read("in", gomock.Any()).Return(sequence(played))
	var seen []error
	for _, err := range reader.Read("in", strings.NewReader("")) {
		seen = append(seen, err)
		break
	}
	if !slices.Equal(seen, []error{broken}) {
		test.Errorf("saw %v", seen)
	}
}

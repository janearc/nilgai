// Package sky draws the sources that the readers return. It draws a square
// of sky around a centre, with each source as its one-sigma error ellipse
// in an ink for its instrument. Strip puts several skies side by side.
//
// data/sky.json sets how the sky is drawn:
//   - the ground colour
//   - the colour the mark starts from
//   - the widths of the lines, dots and gaps
package sky

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/janearc/libreadme/colour"
	"github.com/janearc/libreadme/profile"
	"github.com/janearc/libtheme-css/spaces/radiation"
	"github.com/janearc/libtheme-css/spaces/srgb"
	"github.com/janearc/nilgai"
)

// skyData holds how nilgai draws the sky: the ground, the colour the mark
// starts from, and the widths of what is drawn. Someone picked these by
// eye, so they live in data and not in code.
//
//go:embed data/sky.json
var skyData []byte

// skyStyle holds data/sky.json as ReadSky reads it.
type skyStyle struct {
	ground, mark, between  string
	contrast               map[string]float64
	within                 float64
	line, dot, points, gap int
	cross                  float64
	pixelsMax              int
}

// readStyle reads a style from JSON shaped like data/sky.json, or returns
// nil and an error when the JSON cannot be read.
func readStyle(data []byte) (*skyStyle, error) {
	var read struct {
		Ground, Mark, Between  string
		Contrast               map[string]float64
		Within                 float64
		Line, Dot, Points, Gap int
		Cross                  float64
		PixelsMax              int
	}
	if err := json.Unmarshal(data, &read); err != nil {
		return nil, fmt.Errorf("sky style: %w", err)
	}
	return &skyStyle{ground: read.Ground, mark: read.Mark,
		between: read.Between, contrast: read.Contrast,
		within: read.Within, line: read.Line, dot: read.Dot,
		points: read.Points, gap: read.Gap, cross: read.Cross,
		pixelsMax: read.PixelsMax}, nil
}

// Sky draws a square of sky around a centre. It draws:
//   - each source as its one-sigma error ellipse in its instrument's ink,
//     with a dot at its position
//   - a cross at the mark, a place known from elsewhere
//
// Make one with NewSky. The mark's contrast against the ground sits in the
// reader's prose band, so the mark reads first.
//
// Each instrument's ink has its own contrast against the ground in
// data/sky.json. That contrast may not fall under the reader's floor for
// marks that are scanned and not read.
type Sky struct {
	centre  nilgai.Position
	scale   float64
	pixels  int
	style   skyStyle
	ground  color.Color
	between color.Color
	mark    color.Color
	inks    map[nilgai.Catalogue]color.Color
}

// inked pairs a catalogue with the profile in band/data/profiles.json that
// gives its ink its hue.
type inked struct {
	catalogue nilgai.Catalogue
	profile   string
}

// profileFor names each catalogue's profile, in the order the inks are
// made. The fixed order means a bad style fails with the same error every
// time.
var profileFor = []inked{{nilgai.Fermi, "fermi"},
	{nilgai.Chandra, "chandra"}, {nilgai.NEOWISE, "neowise"},
	{nilgai.AllWISE, "neowise"}, {nilgai.MeerKAT, "meerkat"}}

// NewSky returns a sky width arcseconds across and pixels wide, centred on
// the position given, or nil on a refusal. Its inks take their hues from
// the profiles given and their contrast from the reader's libreadme
// profile, in the style that data/sky.json sets.
//
// NewSky refuses:
//   - a width or a pixel count that is not above zero
//   - more pixels than the style's pixelsMax allows
//   - an instrument with no contrast in the style
//   - an instrument whose contrast is under the reader's floor
//   - an ink that libreadme cannot bring to its contrast, because a
//     picture that cannot be read is not drawn
//
// NewSky gives no ink to a catalogue whose profile is not among those
// given, and leaves that catalogue's sources out.
func NewSky(centre nilgai.Position, width float64, pixels int,
	profiles []radiation.Profile, reader profile.Profile) (*Sky, error) {
	return ReadSky(skyData, centre, width, pixels, profiles, reader)
}

// ReadSky does what NewSky does, in a style given as JSON shaped like
// data/sky.json. ReadSky also refuses a style it cannot read, and a style
// with a colour it cannot parse, and returns nil on a refusal.
func ReadSky(data []byte, centre nilgai.Position, width float64,
	pixels int, profiles []radiation.Profile,
	reader profile.Profile) (*Sky, error) {
	if !nilgai.Positive(width) || !nilgai.Positive(float64(pixels)) {
		return nil, fmt.Errorf("a sky %v arcsec and %d pixels "+
			"across cannot be drawn", width, pixels)
	}
	style, err := readStyle(data)
	if err != nil {
		return nil, err
	}
	if pixels > style.pixelsMax {
		return nil, fmt.Errorf("%d pixels across is past the "+
			"style's limit of %d pixels", pixels, style.pixelsMax)
	}
	ground, err := colour.Parse(style.ground)
	if err != nil {
		return nil, fmt.Errorf("sky ground: %w", err)
	}
	between, err := colour.Parse(style.between)
	if err != nil {
		return nil, fmt.Errorf("sky between: %w", err)
	}
	surfaces := []colour.RGB{ground.RGB}
	mark, err := markInk(*style, reader, surfaces)
	if err != nil {
		return nil, err
	}
	sky := &Sky{centre: centre, scale: width / float64(pixels),
		pixels: pixels, style: *style, ground: opaque(ground.RGB),
		between: opaque(between.RGB), mark: opaque(*mark),
		inks: map[nilgai.Catalogue]color.Color{}}
	for _, each := range profileFor {
		found, isGiven := profileNamed(profiles, each.profile)
		if !isGiven {
			continue
		}
		ink, err := instrumentInk(*style, reader, found, surfaces)
		if err != nil {
			return nil, err
		}
		sky.inks[each.catalogue] = opaque(*ink)
	}
	return sky, nil
}

// profileNamed returns the profile called name, and whether there is one.
func profileNamed(profiles []radiation.Profile,
	name string) (found radiation.Profile, isGiven bool) {
	for _, each := range profiles {
		if each.Name == name {
			return each, true
		}
	}
	return found, false
}

// markInk returns the mark's colour: its starting colour moved into the
// reader's prose band of contrast, so the mark reads first. On a refusal
// markInk returns nil.
func markInk(style skyStyle, reader profile.Profile,
	surfaces []colour.RGB) (*colour.RGB, error) {
	start, err := colour.Parse(style.mark)
	if err != nil {
		return nil, fmt.Errorf("sky mark: %w", err)
	}
	prose := colour.Band{Low: reader.Contrast.Low,
		High: reader.Contrast.High, Centre: reader.Contrast.Centre}
	ink, reached := prose.Into(start.RGB, surfaces)
	if !reached {
		return nil, fmt.Errorf("the mark cannot reach the "+
			"prose band on %s", style.ground)
	}
	return &ink, nil
}

// instrumentInk returns an instrument's colour: its profile's hue, at the
// lightness giving the contrast data/sky.json sets against the ground. On
// a contrast under the reader's floor for marks scanned and not read,
// instrumentInk refuses and returns nil.
func instrumentInk(style skyStyle, reader profile.Profile,
	each radiation.Profile, surfaces []colour.RGB) (*colour.RGB, error) {
	aim, given := style.contrast[each.Name]
	if !given {
		return nil, fmt.Errorf("%s has no contrast in the "+
			"sky style", each.Name)
	}
	if aim < reader.Contrast.ChipMin {
		return nil, fmt.Errorf("%s's contrast %v is under "+
			"the reader's floor of %v", each.Name, aim,
			reader.Contrast.ChipMin)
	}
	spread := aim * style.within
	band := colour.Band{Low: aim - spread, High: aim + spread, Centre: aim}
	ink, reached := band.Into(startingInk(each), surfaces)
	if !reached {
		return nil, fmt.Errorf("%s cannot reach %v:1 on %s",
			each.Name, aim, style.ground)
	}
	return &ink, nil
}

// startingInk returns a profile's colour at the middle of its band, in
// libreadme's 0 to 255. The middle is the geometric mean of the band's
// ends, which is even in the logarithm, as libtheme spreads its ranges.
func startingInk(each radiation.Profile) colour.RGB {
	middle := math.Sqrt(each.Source.Lo * each.Source.Hi)
	lamps, _ := srgb.FromSwatch(each.Swatch(middle))
	red, green, blue := lamps.Bytes()
	return colour.RGB{R: float64(red), G: float64(green), B: float64(blue)}
}

// opaque converts a libreadme colour, 0 to 255 in each channel, to an
// opaque colour that Go's image packages draw with.
func opaque(each colour.RGB) color.Color {
	return color.RGBA{R: channel(each.R), G: channel(each.G),
		B: channel(each.B), A: math.MaxUint8}
}

// channel rounds one of libreadme's 0 to 255 channels to a byte.
func channel(value float64) uint8 { return uint8(math.Round(value)) }

// Draw returns the picture: the ground, every source whose catalogue has
// an ink, and the mark. Pixels outside the square are not painted, so a
// source wholly outside leaves no mark, and a source whose ellipse crosses
// the edge is drawn in part.
func (sky Sky) Draw(sources []nilgai.Source,
	mark nilgai.Position) *image.RGBA {
	picture := image.NewRGBA(image.Rect(none, none, sky.pixels,
		sky.pixels))
	ground := &image.Uniform{C: sky.ground}
	draw.Draw(picture, picture.Bounds(), ground, image.Point{}, draw.Src)
	for _, source := range sources {
		if ink, inked := sky.inks[source.Catalogue()]; inked {
			sky.ellipse(picture, source, ink)
		}
	}
	sky.cross(picture, mark)
	return picture
}

// Strip returns several skies side by side, left to right as given. Each
// is drawn from the same sources and mark. When the skies differ in width
// the strip shows a zoom.
//
// Strip fills the gap between skies with the first sky's between colour,
// and is as tall as the tallest sky. Given no skies, Strip is empty.
func Strip(skies []*Sky, sources []nilgai.Source,
	mark nilgai.Position) *image.RGBA {
	if len(skies) == none {
		return image.NewRGBA(image.Rectangle{})
	}
	first := skies[none]
	wide, tall := (len(skies)-one)*first.style.gap, none
	for _, sky := range skies {
		wide, tall = wide+sky.pixels, max(tall, sky.pixels)
	}
	strip := image.NewRGBA(image.Rect(none, none, wide, tall))
	between := &image.Uniform{C: first.between}
	draw.Draw(strip, strip.Bounds(), between, image.Point{}, draw.Src)
	left := none
	for _, sky := range skies {
		panel := sky.Draw(sources, mark)
		at := image.Rect(left, none, left+sky.pixels, sky.pixels)
		draw.Draw(strip, at, panel, image.Point{}, draw.Src)
		left += sky.pixels + first.style.gap
	}
	return strip
}

// none is the count of nothing, zero.
const none = 0

// one is the count of a single thing. Between n things in a row there are
// n less one gaps.
const one = 1

// place returns where a position falls on the picture, in pixels from the
// top left. East is to the left and north is up, as the sky is drawn.
func (sky Sky) place(at nilgai.Position) (x, y float64) {
	toward := sky.centre.Dec() * math.Pi / nilgai.DegreesInHalfACircle
	east := (at.RA() - sky.centre.RA()) * math.Cos(toward) *
		nilgai.ArcsecondsPerDegree
	north := (at.Dec() - sky.centre.Dec()) * nilgai.ArcsecondsPerDegree
	middle := nilgai.Half(float64(sky.pixels))
	return middle - east/sky.scale, middle - north/sky.scale
}

// ellipse draws a source's one-sigma ellipse and a dot at its centre. The
// major axis lies along the position angle, measured east of north.
func (sky Sky) ellipse(picture *image.RGBA, source nilgai.Source,
	ink color.Color) {
	x, y := sky.place(source.Position())
	pen(picture, x, y, sky.style.dot, ink)
	ellipse := source.Error()
	major := ellipse.SemiMajor() / sky.scale
	minor := ellipse.SemiMinor() / sky.scale
	angle := ellipse.PositionAngle() * math.Pi /
		nilgai.DegreesInHalfACircle
	for step := range sky.style.points {
		around := float64(step) / float64(sky.style.points) * fullTurn
		alongMajor := major * math.Cos(around)
		alongMinor := minor * math.Sin(around)
		east := alongMajor*math.Sin(angle) + alongMinor*math.Cos(angle)
		north := alongMajor*math.Cos(angle) - alongMinor*math.Sin(angle)
		pen(picture, x-east, y-north, sky.style.line, ink)
	}
}

// fullTurn is a full circle in radians, two pi.
const fullTurn = 2 * math.Pi

// cross draws the mark as two lines through it. Each arm is a share of
// the picture's width, set by the style's cross.
func (sky Sky) cross(picture *image.RGBA, mark nilgai.Position) {
	x, y := sky.place(mark)
	arm := sky.style.cross * float64(sky.pixels)
	for along := -arm; along <= arm; along++ {
		pen(picture, x+along, y, sky.style.line, sky.mark)
		pen(picture, x, y+along, sky.style.line, sky.mark)
	}
}

// pen paints a square of the width given, centred on a point. It does not
// paint pixels outside the picture.
func pen(picture *image.RGBA, x, y float64, width int, ink color.Color) {
	reach := nilgai.Half(float64(width))
	left := int(math.Round(x - reach))
	top := int(math.Round(y - reach))
	for across := range width {
		for down := range width {
			picture.Set(left+across, top+down, ink)
		}
	}
}

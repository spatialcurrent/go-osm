package osm

import (
	"strconv"
	"strings"
)

type Bounds struct {
	MinimumLongitude float64 `xml:"minlon,attr"`
	MinimumLatitude  float64 `xml:"minlat,attr"`
	MaximumLongitude float64 `xml:"maxlon,attr"`
	MaximumLatitude  float64 `xml:"maxlat,attr"`
}

func (b Bounds) BoundingBox() string {
	return strings.Join([]string{
		strconv.FormatFloat(b.MinimumLongitude, 'f', 6, 64),
		strconv.FormatFloat(b.MinimumLatitude, 'f', 6, 64),
		strconv.FormatFloat(b.MaximumLongitude, 'f', 6, 64),
		strconv.FormatFloat(b.MaximumLatitude, 'f', 6, 64),
	}, ",")
}

func (b Bounds) ContainsPoint(lon float64, lat float64) bool {
	return lon >= b.MinimumLatitude && lon <= b.MaximumLongitude && lat >= b.MinimumLatitude && lat <= b.MaximumLatitude
}

func NewBounds(minlon float64, minlat float64, maxlon float64, maxlat float64) *Bounds {
	return &Bounds{
		MinimumLongitude: minlon,
		MinimumLatitude:  minlat,
		MaximumLongitude: maxlon,
		MaximumLatitude:  maxlat,
	}
}

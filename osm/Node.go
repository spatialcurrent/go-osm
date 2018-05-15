package osm

import (
	"math"
)

import (
	"github.com/dhconnelly/rtreego"
	"github.com/golang/geo/s1"
)

type Node struct {
	TaggedElement
	Longitude float64 `xml:"lon,attr" parquet:"name=lon, type=DOUBLE"`
	Latitude  float64 `xml:"lat,attr" parquet:"name=lat, type=DOUBLE"`
}

func (n Node) Bounds() *rtreego.Rect {
	rect, err := rtreego.NewRect(rtreego.Point([]float64{n.Longitude, n.Latitude}), []float64{0, 0})
	if err != nil {
		panic(err)
	}
	return rect
}

func (n Node) Tile(z int) []int {
	x := int((180 + n.Longitude) * (math.Pow(2, float64(z)) / 360))

	lat_rad := s1.Angle(n.Latitude).Radians()
	y := int((1.0 - math.Log(math.Tan(lat_rad)+(1/math.Cos(lat_rad)))/math.Pi) / 2.0 * math.Pow(2, float64(z)))

	return []int{x, y}
}

func NewNode() *Node {
	return &Node{TaggedElement: TaggedElement{Element: Element{}}}
}

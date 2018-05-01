package osm

import (
	"math"
	//"strings"
	//"time"
)

import (
	"github.com/dhconnelly/rtreego"
	"github.com/golang/geo/s1"
	//"github.com/pkg/errors"
)

import (
//"github.com/spatialcurrent/go-dfl/dfl"
//"github.com/spatialcurrent/go-graph/graph"
)

type Node struct {
	TaggedElement
	Longitude float64 `xml:"lon,attr"`
	Latitude  float64 `xml:"lat,attr"`
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

func (n *Node) DropAttributes(output Output) {
	if output.DropVersion {
		n.DropVersion()
	}

	if output.DropTimestamp {
		n.DropTimestamp()
	}

	if output.DropChangeset {
		n.DropChangeset()
	}

	if output.DropUserId {
		n.DropUserId()
	}

	if output.DropUserName {
		n.DropUserName()
	}
}

func NewNode() *Node {
	return &Node{TaggedElement: TaggedElement{Element: Element{}}}
}

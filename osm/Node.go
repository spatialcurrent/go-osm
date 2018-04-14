package osm

import (
	"math"
	"time"
)

import (
	"github.com/golang/geo/s1"
)

type Node struct {
	Id        int64      `xml:"id,attr"`
	Version   int        `xml:"version,attr,omitempty"`
	Timestamp *time.Time `xml:"timestamp,attr,omitempty"`
	Changeset int64      `xml:"changeset,attr,omitempty"`
	UserId    int64      `xml:"uid,attr,omitempty"`
	UserName  string     `xml:"user,attr,omitempty"`
	Longitude float64    `xml:"lon,attr"`
	Latitude  float64    `xml:"lat,attr"`
	Tags      []Tag      `xml:"tag"`
}

func (n Node) HasKey(key string) bool {
	for _, t := range n.Tags {
		if key == t.Key {
			return true
		}
	}
	return false
}

func (n Node) Tile(z int) []int {
	x := int((180 + n.Longitude) * (math.Pow(2, float64(z)) / 360))

	lat_rad := s1.Angle(n.Latitude).Radians()
	y := int((1.0 - math.Log(math.Tan(lat_rad)+(1/math.Cos(lat_rad)))/math.Pi) / 2.0 * math.Pow(2, float64(z)))

	return []int{x, y}
}

func (n *Node) DropVersion() {
	n.Version = 0
}

func (n *Node) DropTimestamp() {
	n.Timestamp = nil
}

func (n *Node) DropChangeset() {
	n.Changeset = int64(0)
}

func (n *Node) DropUid() {
	n.UserId = int64(0)
}

func (n *Node) DropUser() {
	n.UserName = ""
}

func (n *Node) TagsAsMap() map[string]interface{} {
	m := map[string]interface{}{}
	for _, t := range n.Tags {
		m[t.Key] = t.Value
	}
	return m
}

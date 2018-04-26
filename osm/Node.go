package osm

import (
	"math"
	"strings"
	"time"
)

import (
	"github.com/dhconnelly/rtreego"
	"github.com/golang/geo/s1"
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-graph/graph"
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

func (n Node) Feature() graph.Feature {
	return graph.NewFeature(n.Id, n.TagsAsMap(), graph.NewPoint(n.Longitude, n.Latitude))
}

func (n Node) Bounds() *rtreego.Rect {
	rect, err := rtreego.NewRect(rtreego.Point([]float64{n.Longitude, n.Latitude}), []float64{0, 0})
	if err != nil {
		panic(err)
	}
	return rect
}

func (n Node) HasKey(key string) bool {
	for _, t := range n.Tags {
		if key == t.Key {
			return true
		}
	}
	return false
}

func (n Node) HasAnyKey(keys []string) bool {
	for _, k := range keys {
		if n.HasKey(k) {
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

func (n *Node) DropVersion() {
	n.Version = 0
}

func (n *Node) DropTimestamp() {
	n.Timestamp = nil
}

func (n *Node) DropChangeset() {
	n.Changeset = int64(0)
}

func (n *Node) DropUserId() {
	n.UserId = int64(0)
}

func (n *Node) DropUserName() {
	n.UserName = ""
}

func (n *Node) TagsAsMap() map[string]interface{} {
	m := map[string]interface{}{}
	for _, t := range n.Tags {
		m[t.Key] = t.Value
	}
	return m
}

func (n *Node) AddTag(t Tag) {
	n.Tags = append(n.Tags, t)
}

func (n *Node) Evaluate(root dfl.Node, funcs dfl.FunctionMap, dfl_attributes []string, dfl_cache *dfl.Cache) (bool, error) {
	m := n.TagsAsMap()
	key := strings.Join(MapToSlice(m, dfl_attributes), "\n")
	if dfl_cache != nil && dfl_cache.Has(key) {
		return dfl_cache.Get(key), nil
	}
	result, err := root.Evaluate(m, funcs)
	if err != nil {
		return false, errors.Wrap(err, "Unknown error evaluating node.")
	}
	switch result.(type) {
	case bool:
		result_bool := result.(bool)
		if dfl_cache != nil {
			dfl_cache.Set(key, result_bool)
		}
		return result_bool, nil
	default:
		return false, errors.New("Error converting dfl output to bool")
	}
	return false, errors.New("Unknown error evaluating node.")
}

func (n *Node) Keep(fi FilterInput, dfl_cache *dfl.Cache) (bool, error) {
	if fi.HasKeysToKeep() {
		if !n.HasAnyKey(fi.KeysToKeep) {
			return false, nil
		}
	}
	if fi.HasKeysToDrop() {
		if n.HasAnyKey(fi.KeysToDrop) {
			return false, nil
		}
	}
	if !fi.ContainsPoint(n.Longitude, n.Latitude) {
		return false, nil
	}
	if !fi.HasExpression() {
		return true, nil
	}
	return n.Evaluate(fi.Expression, fi.Functions, fi.Attributes, dfl_cache)
}

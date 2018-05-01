package osm

import (
	"github.com/spatialcurrent/go-graph/graph"
)

func NodeToFeature(n *Node, tc *TagsCache) graph.Feature {
	return graph.NewFeature(n.GetId(), tc.Map(n.TagsIndex), graph.NewPoint(n.Longitude, n.Latitude))
}

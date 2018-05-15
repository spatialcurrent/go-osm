package osm

import (
	"github.com/spatialcurrent/go-graph/graph"
)

// NodeToFeature converts a Node to a graph.Feature, which searilizes to GeoJSON.
func NodeToFeature(n *Node, tc *TagsCache) graph.Feature {
	return graph.NewFeature(n.GetId(), tc.Map(n.TagsIndex), graph.NewPoint(n.Longitude, n.Latitude))
}

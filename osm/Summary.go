package osm

import (
	"fmt"
)

// Summary is a struct for storing the results of a summariziation of the data in a Planet.
type Summary struct {
	Bounds         Bounds                    `xml:"bounds,omitempty"` // the maximum bounds of the Planet.  read from input and not validated.
	CountUsers     int                       `xml:"users"`            // number of users in the Planet
	CountNodes     int                       `xml:"nodes"`            // number of nodes in the Planet
	CountWays      int                       `xml:"ways"`             // number of ways in the Planet
	CountRelations int                       `xml:"relations"`        // number of relations in the Planet
	CountKeys      int                       `xml:"keys"`             // total number of unique keys in the Planet
	CountTags      int                       `xml:"tags"`             // total number of unique key=value combinations in the Planet
	CountsByKey    map[string]map[string]int `xml:"by_key"`           // count of nodes, ways, and relations by key
}

// BoundingBox returns the summary's bounds as a bbox (aka "minx,miny,maxx,maxy").
func (s Summary) BoundingBox() string {
	return s.Bounds.BoundingBox()
}

// Print prints the summary to stdout.
func (s Summary) Print() {
	fmt.Println("Bounding Box:", s.BoundingBox())
	fmt.Println("Total Number of Users:", s.CountUsers)
	fmt.Println("Total Number of Nodes:", s.CountNodes)
	fmt.Println("Total Number of Ways:", s.CountWays)
	fmt.Println("Total Number of Relations:", s.CountRelations)
	fmt.Println("Total Number of Keys:", s.CountKeys)
	fmt.Println("Total Number of Tags:", s.CountTags)
	for key, counts := range s.CountsByKey {
		fmt.Println("-----------")
		fmt.Println("Key:", key)
		fmt.Println("Number of Nodes:", counts["nodes"])
		fmt.Println("Number of Ways:", counts["ways"])
		fmt.Println("Number of Relations:", counts["relations"])
	}
}

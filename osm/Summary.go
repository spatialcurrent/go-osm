package osm

import (
	"fmt"
)

type Summary struct {
	Bounds         Bounds                    `xml:"bounds,omitempty"`
	CountNodes     int                       `xml:"nodes"`
	CountWays      int                       `xml:"ways"`
	CountRelations int                       `xml:"relations"`
	CountsByKey    map[string]map[string]int `xml:"by_key"`
}

func (s Summary) BoundingBox() string {
	return s.Bounds.BoundingBox()
}

func (s Summary) Print() {
	fmt.Println("Bounding Box:", s.BoundingBox())
	fmt.Println("Total Number of Nodes:", s.CountNodes)
	fmt.Println("Total Number of Ways:", s.CountWays)
	fmt.Println("Total Number of Relations:", s.CountRelations)
	for key, counts := range s.CountsByKey {
		fmt.Println("-----------")
		fmt.Println("Key:", key)
		fmt.Println("Number of Nodes:", counts["nodes"])
		fmt.Println("Number of Ways:", counts["ways"])
		fmt.Println("Number of Relations:", counts["relations"])
	}
}

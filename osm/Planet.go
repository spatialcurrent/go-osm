package osm

import (
	"encoding/xml"
	"time"
)

type Planet struct {
	XMLName   xml.Name    `xml:"osm"`
	Version   string      `xml:"version,attr,omitempty"`
	Generator string      `xml:"generator,attr,omitempty"`
	Timestamp time.Time   `xml:"timestamp,attr"`
	Bounds    Bounds      `xml:"bounds,omitempty"`
	Nodes     []*Node     `xml:"node"`
	Ways      []*Way      `xml:"way"`
	Relations []*Relation `xml:"relation"`
}

type PlanetFile struct {
	Name *Planet `xml:"osm"`
}

func (p Planet) BoundingBox() string {
	return p.Bounds.BoundingBox()
}

func (p *Planet) FilterNodes(include_keys []string) {

	if len(include_keys) > 0 {
		nodes := make([]*Node, 0)
		for _, n := range p.Nodes {
			keep := false
			for _, k := range include_keys {
				if n.HasKey(k) {
					keep = true
					break
				}
			}
			if keep {
				nodes = append(nodes, n)
			}
		}
		p.Nodes = nodes
	}

}

func (p *Planet) FilterWays(include_keys []string) {

	if len(include_keys) > 0 {
		ways := make([]*Way, 0)
		for _, w := range p.Ways {
			keep := false
			for _, k := range include_keys {
				if w.HasKey(k) {
					keep = true
					break
				}
			}
			if keep {
				ways = append(ways, w)
			}
		}
		p.Ways = ways
	}

}

func (p *Planet) Filter(include_keys []string) {

	if len(include_keys) > 0 {
		p.FilterNodes(include_keys)
		p.FilterWays(include_keys)
	}

}

func (p *Planet) DropRelations() {
	p.Relations = make([]*Relation, 0)
}

func (p *Planet) DropVersion() {
	for _, n := range p.Nodes {
		n.DropVersion()
	}
	for _, w := range p.Ways {
		w.DropVersion()
	}
	for _, r := range p.Relations {
		r.DropVersion()
	}
}

func (p *Planet) DropTimestamp() {
	for _, n := range p.Nodes {
		n.DropTimestamp()
	}
	for _, w := range p.Ways {
		w.DropTimestamp()
	}
	for _, r := range p.Relations {
		r.DropTimestamp()
	}
}

func (p *Planet) DropChangeset() {
	for _, n := range p.Nodes {
		n.DropChangeset()
	}
	for _, w := range p.Ways {
		w.DropChangeset()
	}
	for _, r := range p.Relations {
		r.DropChangeset()
	}
}

func (p *Planet) ConvertWaysToNodes() {

	m := map[int64]int{}
	maxNodeId := int64(0)

	for i, n := range p.Nodes {
		m[n.Id] = i
		if n.Id > maxNodeId {
			maxNodeId = n.Id
		}
	}

	uid := maxNodeId + 1
	for _, w := range p.Ways {

		count := float64(w.NumberOfNodes())
		sum_lon := 0.0
		sum_lat := 0.0

		for _, nr := range w.NodeReferences {
			n := p.Nodes[m[nr.Reference]]
			sum_lon += n.Longitude
			sum_lat += n.Latitude
		}

		n := &Node{
			Id:        uid,
			Version:   w.Version,
			Timestamp: w.Timestamp,
			Changeset: w.Changeset,
			UserId:    w.UserId,
			UserName:  w.UserName,
			Longitude: sum_lon / count,
			Latitude:  sum_lat / count,
			Tags:      w.Tags,
		}
		uid += 1
		p.Nodes = append(p.Nodes, n)

	}

	p.Ways = make([]*Way, 0)

}

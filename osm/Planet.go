package osm

import (
	"encoding/xml"
	"time"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
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

func (p *Planet) FilterNodes(include_keys []string, root dfl.Node, funcs map[string]func(map[string]interface{}, []string) (interface{}, error)) error {

	if len(include_keys) > 0 || root != nil {
		nodes := make([]*Node, 0)
		for _, n := range p.Nodes {
			keep := true
			if len(include_keys) > 0 {
				keep = false
				for _, k := range include_keys {
					if n.HasKey(k) {
						keep = true
						break
					}
				}
			}
			if keep && root != nil {
				dfl_result, err := root.Evaluate(n.TagsAsMap(), funcs)
				if err != nil {
					return err
				}
				keep = dfl_result.(bool)
			}
			if keep {
				nodes = append(nodes, n)
			}
		}
		p.Nodes = nodes
	}

	return nil
}

func (p *Planet) FilterWays(include_keys []string, root dfl.Node, funcs map[string]func(map[string]interface{}, []string) (interface{}, error)) error {

	if len(include_keys) > 0 || root != nil {
		ways := make([]*Way, 0)
		for _, w := range p.Ways {
			keep := true
			if len(include_keys) > 0 {
				keep = false
				for _, k := range include_keys {
					if w.HasKey(k) {
						keep = true
						break
					}
				}
			}
			if keep && root != nil {
				dfl_result, err := root.Evaluate(w.TagsAsMap(), funcs)
				if err != nil {
					return err
				}
				keep = dfl_result.(bool)
			}
			if keep {
				ways = append(ways, w)
			}
		}
		p.Ways = ways
	}
	return nil
}

func (p *Planet) Filter(include_keys []string, root dfl.Node, funcs map[string]func(map[string]interface{}, []string) (interface{}, error)) error {

	err := p.FilterNodes(include_keys, root, funcs)
	if err != nil {
		return err
	}

	err = p.FilterWays(include_keys, root, funcs)
	if err != nil {
		return err
	}

	return nil
}

func (p *Planet) DropRelations() {
	p.Relations = make([]*Relation, 0)
}

func (p *Planet) DropAttributes(drop_version bool, drop_timestamp bool, drop_changeset bool, drop_uid bool, drop_user bool) {
	if drop_version || drop_timestamp || drop_changeset {
		for _, n := range p.Nodes {
			if drop_version {
				n.DropVersion()
			}
			if drop_timestamp {
				n.DropTimestamp()
			}
			if drop_changeset {
				n.DropChangeset()
			}
			if drop_uid {
				n.DropUid()
			}
			if drop_user {
				n.DropUser()
			}
		}
		for _, w := range p.Ways {
			if drop_version {
				w.DropVersion()
			}
			if drop_timestamp {
				w.DropTimestamp()
			}
			if drop_changeset {
				w.DropChangeset()
			}
			if drop_uid {
				w.DropUid()
			}
			if drop_user {
				w.DropUser()
			}
		}
		for _, r := range p.Relations {
			if drop_version {
				r.DropVersion()
			}
			if drop_timestamp {
				r.DropTimestamp()
			}
			if drop_changeset {
				r.DropChangeset()
			}
			if drop_uid {
				r.DropUid()
			}
			if drop_user {
				r.DropUser()
			}
		}
	}
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

func (p Planet) CountNodes(key string) int {
	count := 0
	for _, n := range p.Nodes {
		if n.HasKey(key) {
			count += 1
		}
	}
	return count
}

func (p Planet) CountWays(key string) int {
	count := 0
	for _, w := range p.Ways {
		if w.HasKey(key) {
			count += 1
		}
	}
	return count
}

func (p Planet) CountRelations(key string) int {
	count := 0
	for _, r := range p.Relations {
		if r.HasKey(key) {
			count += 1
		}
	}
	return count
}

func (p Planet) Count(key string) int {
	return p.CountNodes(key) + p.CountWays(key) + p.CountRelations(key)
}

func (p Planet) Summarize(keys []string) Summary {

	countsByKey := map[string]map[string]int{}
	for _, key := range keys {
		countsByKey[key] = map[string]int{
			"nodes":     0,
			"ways":      0,
			"relations": 0,
		}
	}
	for _, n := range p.Nodes {
		for _, key := range keys {
			if n.HasKey(key) {
				countsByKey[key]["nodes"] += 1
			}
		}
	}
	for _, w := range p.Ways {
		for _, key := range keys {
			if w.HasKey(key) {
				countsByKey[key]["ways"] += 1
			}
		}
	}
	for _, r := range p.Relations {
		for _, key := range keys {
			if r.HasKey(key) {
				countsByKey[key]["relations"] += 1
			}
		}
	}
	s := Summary{
		Bounds:         p.Bounds,
		CountNodes:     len(p.Nodes),
		CountWays:      len(p.Ways),
		CountRelations: len(p.Relations),
		CountsByKey:    countsByKey,
	}

	return s
}

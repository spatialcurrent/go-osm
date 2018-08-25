package osm

import (
	"encoding/xml"
	"fmt"
	"reflect"
	"sync"
	"time"
)

import (
	"github.com/dhconnelly/rtreego"
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
	"github.com/spatialcurrent/go-graph/graph"
)

type Planet struct {
	XMLName        xml.Name          `xml:"osm"`
	Version        string            `xml:"version,attr,omitempty"`
	Generator      string            `xml:"generator,attr,omitempty"`
	Timestamp      time.Time         `xml:"timestamp,attr"`
	Bounds         Bounds            `xml:"bounds,omitempty"`
	maxId          uint64            `xml:"-"`
	Nodes          []*Node           `xml:"node"`
	nodesIndex     map[uint64]int    `xml:"-"` // map of node id to location in nodes slice
	Ways           []*Way            `xml:"way"`
	waysIndex      map[uint64]int    `xml:"-"` // map of way id to location in ways slice
	Relations      []*Relation       `xml:"relation"`
	relationsIndex map[uint64]int    `xml:"-"` // map of relation id to position in ways slice
	Tags           *TagsCache        `xml:"-"`
	UserNames      map[uint64]string `xml:"-"` // map of UserName by UserId
	Rtree          *rtreego.Rtree    `xml:"-"`
}

func NewPlanet() *Planet {
	p := &Planet{
		maxId:          uint64(0),
		Nodes:          make([]*Node, 0, 10000),
		nodesIndex:     map[uint64]int{},
		Ways:           make([]*Way, 0, 10000),
		waysIndex:      map[uint64]int{},
		Relations:      make([]*Relation, 0, 10000),
		relationsIndex: map[uint64]int{},
		Tags:           NewTagsCache(),
		UserNames:      map[uint64]string{},
		Rtree:          rtreego.NewTree(2, 25, 50),
	}
	return p
}

func (p *Planet) Init() error {
	return nil
}

func (p *Planet) WayToFeature(w *Way) graph.Feature {

	coordinates := make([][]float64, 0)
	for _, nr := range w.NodeReferences {
		n := p.Nodes[p.nodesIndex[nr.Reference]]
		coordinates = append(coordinates, []float64{n.Longitude, n.Latitude})
	}

	if coordinates[0][0] == coordinates[len(coordinates)][0] && coordinates[0][1] == coordinates[len(coordinates)][1] {
		return graph.NewFeature(
			w.GetId(),
			p.Tags.Map(w.TagsIndex),
			graph.NewPolygon(coordinates))
	}

	return graph.NewFeature(
		w.GetId(),
		p.Tags.Map(w.TagsIndex),
		graph.NewLine(coordinates))
}

func (p *Planet) GetFeatures(output *Output) ([]graph.Feature, error) {

	var dfl_cache *dfl.Cache
	if output.Filter.HasExpression() && output.Filter.UseCache {
		dfl_cache = dfl.NewCache()
	}

	features := make([]graph.Feature, 0)

	if !output.DropNodes {
		for _, n := range p.Nodes {
			keep, err := KeepNode(p, output.Filter, n, dfl_cache)
			if err != nil {
				return features, errors.Wrap(err, "Error filtering node for FeatureCollection")
			}
			if keep {
				features = append(features, NodeToFeature(n, p.Tags))
			}
		}
	}

	uid := p.maxId
	//nodes := make([]Node, 0)
	if !output.DropWays {
		for _, w := range p.Ways {
			keep, err := KeepWay(p, output.Filter, w, dfl_cache)
			if err != nil {
				return features, errors.Wrap(err, "Error filtering way for FeatureCollection.")
			}
			if keep {
				if output.WaysToNodes {
					n, err := p.ConvertWayToNode(w, uid)
					if err != nil {
						fmt.Println(errors.Wrap(err, "Error converting way "+fmt.Sprint(w.Id)+" to node.  Planet has "+fmt.Sprint(len(p.Nodes))+" nodes."))
						continue
					}
					uid += 1
					features = append(features, NodeToFeature(n, p.Tags))
				} else {
					uid += 1
					features = append(features, p.WayToFeature(w))
				}
			}
		}
	}

	return features, nil
}

func (p *Planet) GetFeatureCollection(output *Output) (graph.FeatureCollection, error) {

	features, err := p.GetFeatures(output)
	if err != nil {
		return graph.FeatureCollection{}, errors.Wrap(err, "error getting features from planet")
	}
	return graph.NewFeatureCollection(features), nil
}

func (p Planet) BoundingBox() string {
	return p.Bounds.BoundingBox()
}

func (p *Planet) AddTag(t Tag) uint32 {
	return p.Tags.AddTag(t)
}

func (p *Planet) AddTags(tags []Tag) []uint32 {
	return p.Tags.AddTags(tags)
}

func (p *Planet) GetTag(tagIndex uint32) Tag {
	return p.Tags.Values[int(tagIndex)]
}

func (p *Planet) GetTagsAsMap(tagIndicies []uint32) map[string]interface{} {
	return p.Tags.Map(tagIndicies)
}

func (p *Planet) AddNode(n *Node) error {

	i, ok := p.nodesIndex[n.Id]
	if ok {
		return errors.New("Node with id " + fmt.Sprint(n.Id) + " already exists in the index at position " + fmt.Sprint(i) + ".  This node might be used in multiple input files.")
	}

	p.Nodes = append(p.Nodes, n)
	p.nodesIndex[n.Id] = len(p.Nodes) - 1

	if n.Id > p.maxId {
		p.maxId = n.Id
	}

	return nil
}

func (p *Planet) AddWay(w *Way) error {

	i, ok := p.waysIndex[w.Id]
	if ok {
		return errors.New("Way with id " + fmt.Sprint(w.Id) + " already exists in the index at position " + fmt.Sprint(i) + ".  This way might be present in multiple input files.")
	}

	p.Ways = append(p.Ways, w)
	p.waysIndex[w.Id] = len(p.Ways) - 1

	if w.Id > p.maxId {
		p.maxId = w.Id
	}

	return nil
}

func (p *Planet) AddRelation(r *Relation) error {

	i, ok := p.relationsIndex[r.Id]
	if ok {
		return errors.New("Relation with id " + fmt.Sprint(r.Id) + " already exists in the index at position " + fmt.Sprint(i) + ".  This relation might be present in multiple input files.")
	}

	p.Relations = append(p.Relations, r)
	p.relationsIndex[r.Id] = len(p.Relations) - 1

	if r.Id > p.maxId {
		p.maxId = r.Id
	}

	return nil
}

func (p *Planet) ConvertWayToNode(w *Way, way_node_id uint64) (*Node, error) {
	count := float64(w.NumberOfNodes())
	sum_lon := 0.0
	sum_lat := 0.0

	for _, nr := range w.NodeReferences {
		position, ok := p.nodesIndex[nr.Reference]
		if !ok {
			return &Node{}, errors.New("Node reference " + fmt.Sprint(nr.Reference) + " is for a node that has not been seen yet.")
		}
		if position >= len(p.Nodes) {
			return &Node{}, errors.New("For node with id " + fmt.Sprint(nr.Reference) + ", the position " + fmt.Sprint(position) + " is greater than or equal to the length of nodes " + fmt.Sprint(len(p.Nodes)) + ".")
		}
		n := p.Nodes[position]
		sum_lon += n.Longitude
		sum_lat += n.Latitude
	}

	n := &Node{
		TaggedElement: TaggedElement{
			Element: Element{
				Id:        way_node_id,
				Version:   w.Version,
				Timestamp: w.Timestamp,
				Changeset: w.Changeset,
				UserId:    w.UserId,
				UserName:  w.UserName,
			},
			Tags: w.Tags,
		},
		Longitude: sum_lon / count,
		Latitude:  sum_lat / count,
	}

	return n, nil
}

func (p *Planet) AddWayAsNode(w *Way, node_id uint64) error {
	n, err := p.ConvertWayToNode(w, node_id)
	if err != nil {
		return err
	}
	return p.AddNode(n)
}

func (p *Planet) FilterNodes(fi *Filter, dfl_cache *dfl.Cache) error {

	if fi.HasKeysToKeep() || fi.HasKeysToDrop() || fi.HasExpression() {

		nodes := make([]*Node, 0)
		for _, n := range p.Nodes {
			keep, err := KeepNode(p, fi, n, dfl_cache)
			if err != nil {
				return errors.Wrap(err, "error running keep on node")
			}
			if keep {
				nodes = append(nodes, n)
			}
		}
		p.Nodes = nodes

	}

	return nil
}

func (p *Planet) FilterWays(fi *Filter, dfl_cache *dfl.Cache) error {

	if fi.HasKeysToKeep() || fi.HasKeysToDrop() || fi.HasExpression() {

		ways := make([]*Way, 0)
		for _, w := range p.Ways {
			keep, err := KeepWay(p, fi, w, dfl_cache)
			if err != nil {
				return errors.Wrap(err, "error running keep on way")
			}
			if keep {
				ways = append(ways, w)
			}
		}
		p.Ways = ways

	}

	return nil
}

func (p *Planet) Filter(fi *Filter, dfl_cache *dfl.Cache) error {

	err := p.FilterNodes(fi, dfl_cache)
	if err != nil {
		return err
	}

	err = p.FilterWays(fi, dfl_cache)
	if err != nil {
		return err
	}

	return nil
}

func (p *Planet) DropWays() {
	p.Ways = make([]*Way, 0)
}

func (p *Planet) DropRelations() {
	p.Relations = make([]*Relation, 0)
}

/*
func (p *Planet) DropAttributes(config *Config) {
	if config.HasDrop() {
		for _, n := range p.Nodes {
			n.DropAttributes(config)
		}
		for _, w := range p.Ways {
			w.DropAttributes(config)
		}
		for _, r := range p.Relations {
			r.DropAttributes(config)
		}
	}
}*/

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

func (p *Planet) ConvertWaysToNodes(async bool) {

	m := map[uint64]uint{}
	/*maxId := uint64(0)

	for i, n := range p.Nodes {
		m[n.Id] = uint(i)
		if n.Id > maxId {
			maxId = n.Id
		}
	}*/

	uid := p.maxId
	for _, w := range p.Ways {
		uid += 1

		count := float64(w.NumberOfNodes())
		sum_lon := 0.0
		sum_lat := 0.0

		for _, nr := range w.NodeReferences {
			n := p.Nodes[m[nr.Reference]]
			sum_lon += n.Longitude
			sum_lat += n.Latitude
		}

		n := &Node{
			TaggedElement: TaggedElement{
				Element: Element{
					Id:        uid,
					Version:   w.Version,
					Timestamp: w.Timestamp,
					Changeset: w.Changeset,
					UserId:    w.UserId,
					UserName:  w.UserName,
				},
				Tags: w.Tags,
			},
			Longitude: sum_lon / count,
			Latitude:  sum_lat / count,
		}
		p.AddNode(n)
	}

	p.Ways = make([]*Way, 0)

}

func (p Planet) CountNodes(key string) int {
	count := 0
	for _, n := range p.Nodes {
		m := p.GetTagsAsMap(n.GetTagsIndex())
		if _, ok := m[key]; ok {
			count += 1
		}
	}
	return count
}

func (p Planet) CountWays(key string) int {
	count := 0
	for _, w := range p.Ways {
		m := p.GetTagsAsMap(w.GetTagsIndex())
		if _, ok := m[key]; ok {
			count += 1
		}
	}
	return count
}

func (p Planet) CountRelations(key string) int {
	count := 0
	for _, r := range p.Relations {
		m := p.GetTagsAsMap(r.GetTagsIndex())
		if _, ok := m[key]; ok {
			count += 1
		}
	}
	return count
}

func (p Planet) Count(key string) int {
	return p.CountNodes(key) + p.CountWays(key) + p.CountRelations(key)
}

func WaitThenClose(wg *sync.WaitGroup, c chan<- string) {
	wg.Wait()
	close(c)
}

func AddToChannel(ch chan interface{}, values interface{}) {

	var wg sync.WaitGroup

	s := reflect.ValueOf(values)

	for i := 0; i < s.Len(); i++ {
		wg.Add(1)
		go func(i interface{}, ch chan interface{}, wg *sync.WaitGroup) {
			ch <- i
			wg.Done()
		}(s.Index(i).Interface(), ch, &wg)
	}

	go func(wg *sync.WaitGroup, c chan<- interface{}) {
		wg.Wait()
		close(c)
		fmt.Println("Closed nodes_chan")
	}(&wg, ch)

}

/*func AddNodesToChannel(nodes_chan chan<- *Node, nodes_slice []*Node, wg *sync.WaitGroup) {
	for _, n := range nodes_slice {
		wg.Add(1)
		go func(i *Node, ch chan<- *Node) {
			ch <- i
			wg.Done()
		}(n, nodes_chan)
	}
	go func(wg *sync.WaitGroup, c chan<- string) {
		wg.Wait()
		close(c)
		fmt.Println("Closed nodes_chan")
	}(wg, nodes_chan)
}*/

func (p Planet) Summarize(keys []string) Summary {

	countsByKey := map[string]map[string]int{}
	for _, key := range keys {
		countsByKey[key] = map[string]int{
			"nodes":     p.CountNodes(key),
			"ways":      p.CountWays(key),
			"relations": p.CountRelations(key),
		}
	}

	s := Summary{
		Bounds:         p.Bounds,
		CountUsers:     len(p.UserNames),
		CountNodes:     len(p.Nodes),
		CountWays:      len(p.Ways),
		CountRelations: len(p.Relations),
		CountsByKey:    countsByKey,
		CountKeys:      len(p.Tags.Index),
		CountTags:      len(p.Tags.Values),
	}

	return s
}

// GetWayNodeIdsAsSlice returns a slice of all the IDs of the Nodes that are part of ways.
func (p Planet) GetWayNodeIdsAsSlice() UInt64Slice {
	set_way_nodes := NewUInt64Set()
	for _, w := range p.Ways {
		for _, nr := range w.NodeReferences {
			set_way_nodes.Add(nr.Reference)
		}
	}
	return set_way_nodes.Slice(true)
}

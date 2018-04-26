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
	XMLName   xml.Name       `xml:"osm"`
	Version   string         `xml:"version,attr,omitempty"`
	Generator string         `xml:"generator,attr,omitempty"`
	Timestamp time.Time      `xml:"timestamp,attr"`
	Bounds    Bounds         `xml:"bounds,omitempty"`
	Nodes     []*Node        `xml:"node"`
	Ways      []*Way         `xml:"way"`
	Relations []*Relation    `xml:"relation"`
	NodeIndex map[int64]int  `xml:"-"` // map of NodeId to location in Nodes slice
	MaxNodeId int64          `xml:"-"`
	Rtree     *rtreego.Rtree `xml:"-"`
}

func NewPlanet() *Planet {
	p := &Planet{
		Nodes:     make([]*Node, 0, 10000),
		Ways:      make([]*Way, 0, 10000),
		Relations: make([]*Relation, 0, 10000),
		NodeIndex: map[int64]int{},
		MaxNodeId: int64(0),
		Rtree:     rtreego.NewTree(2, 25, 50),
	}
	return p
}

func (p Planet) FeatureCollection() graph.FeatureCollection {

	fc := graph.FeatureCollection{}
	features := make([]graph.Feature, 0)
	for _, n := range p.Nodes {
		features = append(features, n.Feature())
	}
	fc = graph.NewFeatureCollection(features)

	return fc
}

func (p Planet) BoundingBox() string {
	return p.Bounds.BoundingBox()
}

func (p *Planet) AddNode(n *Node, updateIndex bool) {
	p.Nodes = append(p.Nodes, n)
	if updateIndex {
		if n.Id > p.MaxNodeId {
			p.MaxNodeId = n.Id
		}
		p.NodeIndex[n.Id] = len(p.Nodes) - 1
		//p.Rtree.Insert(n)
	}
}

func (p *Planet) AddWay(w *Way) {
	p.Ways = append(p.Ways, w)
}

func (p *Planet) AddRelation(r *Relation) {
	p.Relations = append(p.Relations, r)
}

func (p *Planet) AddWayAsNode(w *Way) {

	count := float64(w.NumberOfNodes())
	sum_lon := 0.0
	sum_lat := 0.0

	for _, nr := range w.NodeReferences {
		n := p.Nodes[p.NodeIndex[nr.Reference]]
		sum_lon += n.Longitude
		sum_lat += n.Latitude
	}

	n := &Node{
		Id:        p.MaxNodeId + 1,
		Version:   w.Version,
		Timestamp: w.Timestamp,
		Changeset: w.Changeset,
		UserId:    w.UserId,
		UserName:  w.UserName,
		Longitude: sum_lon / count,
		Latitude:  sum_lat / count,
		Tags:      w.Tags,
	}

	p.AddNode(n, true)
}

func (p *Planet) FilterNodes(fi FilterInput, dfl_cache *dfl.Cache) error {

	if fi.HasKeysToKeep() || fi.HasKeysToDrop() || fi.HasExpression() {

		nodes := make([]*Node, 0)
		for _, n := range p.Nodes {
			keep, err := n.Keep(fi, dfl_cache)
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

func (p *Planet) FilterWays(fi FilterInput, dfl_cache *dfl.Cache) error {

	if fi.HasKeysToKeep() || fi.HasKeysToDrop() || fi.HasExpression() {

		ways := make([]*Way, 0)
		for _, w := range p.Ways {
			keep, err := w.Keep(fi, dfl_cache)
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

func (p *Planet) Filter(fi FilterInput, dfl_cache *dfl.Cache) error {

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

func (p *Planet) DropAttributes(output Output) {
	if output.HasDrop() {
		for _, n := range p.Nodes {
			n.DropAttributes(output)
		}
		for _, w := range p.Ways {
			w.DropAttributes(output)
		}
		for _, r := range p.Relations {
			r.DropAttributes(output)
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

func (p *Planet) ConvertWaysToNodes(async bool) {

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

func (p Planet) Summarize(keys []string, async bool) Summary {

	countsByKey := map[string]map[string]int{}
	for _, key := range keys {
		countsByKey[key] = map[string]int{
			"nodes":     0,
			"ways":      0,
			"relations": 0,
		}
	}
	if async {

		keys_chan := make(chan map[string]string, 100000)
		/*
			// Sends nodes to channel
			//nodes_chan := make(chan *Node)
			nodes_chan := make(chan interface{})
			//var nodes_sent_wg sync.WaitGroup
			//go AddToChannel(nodes_chan, p.Nodes)
			go func(ch chan<- interface{}, s []*Node) {
				for _, v := range s {
					ch <- v
				}
				fmt.Println("Closed nodes_chan")
				close(ch)
			}(nodes_chan, p.Nodes)

			// Sends ways to channel
			//ways_chan := make(chan *Way)
			ways_chan := make(chan interface{})
			//go AddToChannel(ways_chan, p.Ways)
			go func(ch chan<- interface{}, s []*Way) {
				for _, v := range s {
					ch <- v
				}
				fmt.Println("Closed ways_chan")
				close(ch)
			}(ways_chan, p.Ways)
		*/

		// Receive nodes from channel
		var nodes_wg sync.WaitGroup
		nodes_wg.Add(1)
		go func(s []*Node, keys_chan chan<- map[string]string, wg *sync.WaitGroup) {
			for _, n := range s {
				//n := v.(*Node)
				//fmt.Println("Received node " + fmt.Sprint(n.Id))
				wg.Add(1)
				go func(n *Node, wg *sync.WaitGroup, keys_chan chan<- map[string]string) {
					for _, key := range keys {
						if n.HasKey(key) {
							keys_chan <- map[string]string{
								"element": "node",
								"key":     key,
							}
						}
					}
					wg.Done()
				}(n, wg, keys_chan)
			}
			wg.Done()
		}(p.Nodes, keys_chan, &nodes_wg)

		// Receive ways from channel
		//ways_keys_chan := make(chan string)
		var ways_wg sync.WaitGroup
		ways_wg.Add(1)
		go func(s []*Way, keys_chan chan<- map[string]string, wg *sync.WaitGroup) {
			for _, w := range s {
				//w := v.(*Way)
				//fmt.Println("Received way " + fmt.Sprint(w.Id))
				wg.Add(1)
				go func(w *Way, wg *sync.WaitGroup, keys_chan chan<- map[string]string) {
					for _, key := range keys {
						if w.HasKey(key) {
							keys_chan <- map[string]string{
								"element": "node",
								"key":     key,
							}
						}
					}
					wg.Done()
				}(w, wg, keys_chan)
			}
			wg.Done()
		}(p.Ways, keys_chan, &ways_wg)

		// Wait for output
		go func(waitgroups []*sync.WaitGroup, c chan<- map[string]string) {
			for _, wg := range waitgroups {
				wg.Wait()
			}
			close(c)
		}([]*sync.WaitGroup{&nodes_wg, &ways_wg}, keys_chan)

		// Aggregate Output
		for k := range keys_chan {
			countsByKey[k["key"]][k["element"]] += 1
		}

	} else {
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
		/*for _, r := range p.Relations {
			for _, key := range keys {
				if r.HasKey(key) {
					countsByKey[key]["relations"] += 1
				}
			}
		}*/
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

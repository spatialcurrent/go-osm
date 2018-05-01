package osm

import (
	//"bytes"
	"encoding/xml"
	"io"
	"io/ioutil"
	//"fmt"
	"time"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

func UnmarshalPlanet(planet *Planet, input_reader io.Reader, stream bool, output Output, fi FilterInput, ways_to_nodes bool) error {

	var dfl_cache *dfl.Cache
	if fi.HasExpression() && fi.UseCache {
		dfl_cache = dfl.NewCache()
	}

	if stream {

		decoder := xml.NewDecoder(input_reader)
		for {

			t, _ := decoder.Token()
			if t == nil {
				break
			}

			switch e := t.(type) {
			case xml.StartElement:
				switch e.Name.Local {
				case "osm":
					planet.XMLName = e.Name
					for _, attr := range e.Attr {
						switch attr.Name.Local {
						case "version":
							planet.Version = attr.Value
						case "generator":
							planet.Generator = attr.Value
						case "timestamp":
							v, err := time.Parse(time.RFC3339, attr.Value)
							if err != nil {
								return errors.Wrap(err, "Error parsing osm timestamp")
							}
							planet.Timestamp = v
						}
					}
				case "bounds":
					b, err := UnmarshalBounds(decoder, e)
					if err != nil {
						return err
					}
					planet.Bounds = b
				case "node":
					n, tags, err := UnmarshalNode(decoder, e, output)
					if err != nil {
						return err
					}
					n.SetTagsIndex(planet.AddTags(tags))
					keep := true
					if output.DropWays && output.DropRelations {
						keep, err = KeepNode(planet, fi, n, dfl_cache)
						if err != nil {
							return err
						}
					}
					if keep {
						planet.AddNode(n, ways_to_nodes)
					}
				case "way":
					if !output.DropWays {
						w, tags, err := UnmarshalWay(decoder, e, output)
						if err != nil {
							return err
						}
						w.SetTagsIndex(planet.AddTags(tags))
						keep, err := KeepWay(planet, fi, w, dfl_cache)
						if err != nil {
							return err
						}
						if keep {
							if ways_to_nodes {
								planet.AddWayAsNode(w)
							} else {
								planet.AddWay(w)
							}
						}
					}
				case "relation":
					if !output.DropRelations {
						r, err := UnmarshalRelation(decoder, e, output)
						if err != nil {
							return err
						}
						planet.AddRelation(&r)
					}
				}
			}

		}

		if !(output.DropWays && output.DropRelations) {
			set_way_nodes := NewInt64Set()
			for _, w := range planet.Ways {
				for _, nr := range w.NodeReferences {
					set_way_nodes.Add(nr.Reference)
					//set_way_nodes[nr.Reference] = struct{}{}
				}
			}
			slice_way_nodes := set_way_nodes.Slice(true)
			nodes := make([]*Node, 0)
			for _, n := range planet.Nodes {
				keep := true
				if slice_way_nodes.Contains(n.Id) {
					nodes = append(nodes, n)
					continue
				}
				keep, err := KeepNode(planet, fi, n, dfl_cache)
				if err != nil {
					return err
				}
				if keep {
					nodes = append(nodes, n)
				}
			}
			planet.Nodes = nodes
		}

		//fmt.Println("DFL Cache:", dfl_cache)

	} else {
		input_bytes, err := ioutil.ReadAll(input_reader)
		if err != nil {
			return errors.Wrap(err, "Error reading input from reader.")
		}
		err = xml.Unmarshal(input_bytes, planet)
		if err != nil {
			return errors.Wrap(err, "Error unmarshalling input.")
		}

		if output.DropWays {
			planet.DropWays()
		}

		if output.DropRelations {
			planet.DropRelations()
		}

		planet.DropAttributes(output)

		planet.Filter(fi, dfl_cache)

		if ways_to_nodes {
			planet.ConvertWaysToNodes(false)
		}
	}
	return nil
}

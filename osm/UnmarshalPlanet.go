package osm

import (
	"bytes"
	"encoding/xml"
	//"fmt"
	"time"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

func UnmarshalPlanet(planet *Planet, input_bytes []byte, stream bool, output Output, filterInput FilterInput, ways_to_nodes bool) error {

	var dfl_cache *dfl.Cache
	if filterInput.HasExpression() && filterInput.UseCache {
		dfl_cache = dfl.NewCache()
	}

	if stream {

		decoder := xml.NewDecoder(bytes.NewReader(input_bytes))
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
					n, err := UnmarshalNode(decoder, e, output)
					if err != nil {
						return err
					}
					keep := true
					if output.DropWays && output.DropRelations {
						keep, err = n.Keep(filterInput, dfl_cache)
						if err != nil {
							return err
						}
					}
					if keep {
						planet.AddNode(&n, ways_to_nodes)
					}
				case "way":
					if !output.DropWays {
						w, err := UnmarshalWay(decoder, e, output)
						if err != nil {
							return err
						}
						keep, err := w.Keep(filterInput, dfl_cache)
						if err != nil {
							return err
						}
						if keep {
							if ways_to_nodes {
								planet.AddWayAsNode(&w)
							} else {
								planet.AddWay(&w)
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
			nodes := make([]*Node, 0)
			for _, n := range planet.Nodes {
				keep, err := n.Keep(filterInput, dfl_cache)
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
		err := xml.Unmarshal(input_bytes, planet)
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

		planet.Filter(filterInput, dfl_cache)

		if ways_to_nodes {
			planet.ConvertWaysToNodes(false)
		}
	}
	return nil
}

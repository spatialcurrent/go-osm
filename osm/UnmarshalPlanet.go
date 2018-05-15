package osm

import (
	//"bytes"
	"encoding/xml"
	//"io"
	//"fmt"
	//"io/ioutil"
	"time"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-composite-logger/compositelogger"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

// UnmarshalPlanet reads an OSM Planet File from an io.Reader and unmarshals the data into the *Planet object.
// Returns an error if any.
func UnmarshalPlanet(p *Planet, input *Input, logger *compositelogger.CompositeLogger) error {

	var dfl_cache *dfl.Cache
	if input.Filter != nil && input.Filter.HasExpression() && input.Filter.UseCache {
		dfl_cache = dfl.NewCache()
	}

	nodes := make([]*Node, 0)
	ways := make([]*Way, 0)
	relations := make([]*Relation, 0)

	decoder := xml.NewDecoder(input.Reader)
	for {

		t, _ := decoder.Token()
		if t == nil {
			break
		}

		switch e := t.(type) {
		case xml.StartElement:
			switch e.Name.Local {
			case "osm":
				p.XMLName = e.Name
				for _, attr := range e.Attr {
					switch attr.Name.Local {
					case "version":
						p.Version = attr.Value
					case "generator":
						p.Generator = attr.Value
					case "timestamp":
						v, err := time.Parse(time.RFC3339, attr.Value)
						if err != nil {
							return errors.Wrap(err, "Error parsing osm timestamp")
						}
						p.Timestamp = v
					}
				}
			case "bounds":
				b, err := UnmarshalBounds(decoder, e)
				if err != nil {
					return err
				}
				p.Bounds = b
			case "node":
				n, user_id, user_name, tags, err := UnmarshalNode(decoder, e, input)
				if err != nil {
					return err
				}
				if user_id > 0 {
					n.UserId = user_id
					if len(user_name) > 0 {
						p.UserNames[n.UserId] = user_name
					}
				}
				n.SetTagsIndex(p.AddTags(tags))
				keep := true
				if input.DropWays && input.DropRelations {
					keep, err = KeepNode(p, input.Filter, n, dfl_cache)
					if err != nil {
						return err
					}
				}
				if keep {
					nodes = append(nodes, n)
				}
			case "way":
				if !input.DropWays {
					w, user_id, user_name, tags, err := UnmarshalWay(decoder, e, input)
					if err != nil {
						return err
					}
					if user_id > 0 {
						w.UserId = user_id
						if len(user_name) > 0 {
							p.UserNames[w.UserId] = user_name
						}
					}
					w.SetTagsIndex(p.AddTags(tags))
					keep, err := KeepWay(p, input.Filter, w, dfl_cache)
					if err != nil {
						return err
					}
					if keep {
						ways = append(ways, w)
					}
				}
			case "relation":
				if !input.DropRelations {
					r, user_id, user_name, tags, err := UnmarshalRelation(decoder, e, input)
					if err != nil {
						return err
					}
					if user_id > 0 {
						r.UserId = user_id
						if len(user_name) > 0 {
							p.UserNames[r.UserId] = user_name
						}
					}
					r.SetTagsIndex(p.AddTags(tags))
					keep, err := KeepRelation(p, input.Filter, r, dfl_cache)
					if err != nil {
						return err
					}
					if keep {
						relations = append(relations, r)
					}
				}
			}
		}

	}

	if input.DropWays && input.DropRelations {
		for _, n := range nodes {
			p.AddNode(n)
		}
	} else {
		set_way_nodes := NewUInt64Set()
		for _, w := range ways {
			for _, nr := range w.NodeReferences {
				set_way_nodes.Add(nr.Reference)
			}
		}
		slice_way_nodes := set_way_nodes.Slice(true)
		valid_nodes := make([]*Node, 0)
		for _, n := range nodes {
			keep := true
			if slice_way_nodes.Contains(n.Id) {
				valid_nodes = append(valid_nodes, n)
				continue
			}
			keep, err := KeepNode(p, input.Filter, n, dfl_cache)
			if err != nil {
				return err
			}
			if keep {
				valid_nodes = append(valid_nodes, n)
			}
		}

		// Add elements to planet
		for _, n := range valid_nodes {
			p.AddNode(n)
		}
		for _, w := range ways {
			p.AddWay(w)
		}
		for _, r := range relations {
			p.AddRelation(r)
		}
	}

	return nil
}

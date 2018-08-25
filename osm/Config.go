// +build !js
package osm

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

type Globals struct {
	Input  map[string]interface{} `hcl:"input,omitempty"`  // input global config
	Output map[string]interface{} `hcl:"output,omitempty"` // output global config
}

type Config struct {
	Globals               Globals        `hcl:"globals,omitempty"`
	InputConfigs          []InputConfig  `hcl:"inputs,omitempty"`
	OutputConfigs         []OutputConfig `hcl:"outputs,omitempty"`
	Inputs                []*Input       `hcl:"-"`
	Outputs               []*Output      `hcl:"-"`
	DropAllNodes          bool           `hcl:"-"`
	DropAllWays           bool           `hcl:"-"`
	DropAllRelations      bool           `hcl:"-"`
	ConvertAllWaysToNodes bool           `hcl:"-"`
	DropAllVersions       bool           `hcl:"-"`
	DropAllTimestamps     bool           `hcl:"-"`
	DropAllChangesets     bool           `hcl:"-"`
	DropAllUserIds        bool           `hcl:"-"`
	DropAllUserNames      bool           `hcl:"-"`
	OutputKeysToKeep      StringSet      `hcl:"-"` //
	OutputKeysToDrop      StringSet      `hcl:"-"` //
}

func (c *Config) Init(ctx map[string]interface{}, funcs *dfl.FunctionMap) error {

	c.Inputs = make([]*Input, len(c.InputConfigs))
	for i, x := range c.InputConfigs {

		input := &Input{
			PlanetResource: &PlanetResource{
				FilteredResource: &FilteredResource{
					Resource: &Resource{
						Uri: x.Uri,
					},
					Filter: x.Filter,
				},
				DropWays:      x.DropWays,
				DropRelations: x.DropRelations,
				DropVersion:   x.DropVersion,
				DropChangeset: x.DropChangeset,
				DropTimestamp: x.DropTimestamp,
				DropUserId:    x.DropUserId,
				DropUserName:  x.DropUserName,
				KeysToKeep:    x.KeysToKeep,
				KeysToDrop:    x.KeysToDrop,
			},
		}

		err := input.Init(c.Globals.Input, ctx, funcs)
		if err != nil {
			return err
		}
		c.Inputs[i] = input
	}

	c.Outputs = make([]*Output, len(c.OutputConfigs))
	for i, x := range c.OutputConfigs {

		output := &Output{
			PlanetResource: &PlanetResource{
				FilteredResource: &FilteredResource{
					Resource: &Resource{
						Uri: x.Uri,
					},
					Filter: x.Filter,
				},
				DropWays:      x.DropWays,
				DropRelations: x.DropRelations,
				DropVersion:   x.DropVersion,
				DropChangeset: x.DropChangeset,
				DropTimestamp: x.DropTimestamp,
				DropUserId:    x.DropUserId,
				DropUserName:  x.DropUserName,
				KeysToKeep:    x.KeysToKeep,
				KeysToDrop:    x.KeysToDrop,
			},
			WaysToNodes: x.WaysToNodes,
			Pretty:      x.Pretty,
		}

		err := output.Init(c.Globals.Output, ctx, funcs)
		if err != nil {
			return err
		}
		c.Outputs[i] = output
	}

	drop_nodes := true
	for _, o := range c.Outputs {
		if !o.DropNodes {
			drop_nodes = false
			break
		}
	}
	c.DropAllNodes = drop_nodes

	drop_ways := true
	for _, i := range c.Inputs {
		if !i.DropWays {
			drop_ways = false
			break
		}
	}
	if !drop_ways {
		for _, o := range c.Outputs {
			if !o.DropWays {
				drop_ways = false
				break
			}
		}
	}
	c.DropAllWays = drop_ways

	drop_relations := true
	for _, i := range c.Inputs {
		if !i.DropRelations {
			drop_relations = false
			break
		}
	}
	if !drop_relations {
		for _, o := range c.Outputs {
			if !o.DropRelations {
				drop_relations = false
				break
			}
		}
	}
	c.DropAllRelations = drop_relations

	ways_to_nodes := true
	for _, o := range c.Outputs {
		if !o.WaysToNodes {
			ways_to_nodes = false
			break
		}
	}
	c.ConvertAllWaysToNodes = ways_to_nodes

	drop_versions := true
	for _, o := range c.Outputs {
		if !o.DropVersion {
			drop_versions = false
			break
		}
	}
	c.DropAllVersions = drop_versions

	drop_timestamps := true
	for _, o := range c.Outputs {
		if !o.DropTimestamp {
			drop_timestamps = false
			break
		}
	}
	c.DropAllTimestamps = drop_timestamps

	drop_changesets := true
	for _, o := range c.Outputs {
		if !o.DropChangeset {
			drop_changesets = false
			break
		}
	}
	c.DropAllChangesets = drop_changesets

	drop_user_ids := true
	for _, o := range c.Outputs {
		if !o.DropUserId {
			drop_user_ids = false
			break
		}
	}
	c.DropAllUserIds = drop_user_ids

	drop_user_names := true
	for _, o := range c.Outputs {
		if !o.DropUserName {
			drop_user_names = false
			break
		}
	}
	c.DropAllUserNames = drop_user_names

	output_keys_keep := NewStringSet()
	for _, o := range c.Outputs {
		if len(o.KeysToKeep) == 0 {
			output_keys_keep = NewStringSet()
			break
		}
		output_keys_keep.Add(o.KeysToKeep...)
	}
	c.OutputKeysToKeep = output_keys_keep

	output_keys_drop := NewStringSet()
	for _, o := range c.Outputs {
		if len(o.KeysToDrop) == 0 {
			output_keys_drop = NewStringSet()
			break
		}
		output_keys_drop = output_keys_drop.Intersect(o.KeysToDrop)
	}
	c.OutputKeysToDrop = output_keys_drop

	for _, i := range c.Inputs {
		err := i.Init(c.Globals.Input, ctx, funcs)
		if err != nil {
			return err
		}

		if c.DropAllWays {
			i.DropWays = true
		}

		if c.DropAllRelations {
			i.DropRelations = true
		}

		if c.DropAllVersions {
			i.DropVersion = true
		}

		if c.DropAllChangesets {
			i.DropChangeset = true
		}

		if c.DropAllTimestamps {
			i.DropTimestamp = true
		}

		if c.DropAllUserIds {
			i.DropUserId = true
		}

		if c.DropAllUserNames {
			i.DropUserName = true
		}

		if len(i.KeysToKeep) == 0 {
			i.KeysToKeep = c.OutputKeysToKeep.Slice(true)
		} else {
			if c.OutputKeysToKeep.Len() > 0 {
				i.KeysToKeep = c.OutputKeysToKeep.Intersect(i.KeysToKeep).Slice(true)
			}
		}

		if c.OutputKeysToDrop.Len() > 0 {
			if len(i.KeysToDrop) == 0 {
				i.KeysToDrop = c.OutputKeysToDrop.Slice(true)
			} else {
				i.KeysToDrop = c.OutputKeysToDrop.Union(i.KeysToDrop).Slice(true)
			}
		}

	}

	return nil
}

func (c *Config) HasResourceType(t string) bool {
	has := false
	for _, i := range c.Inputs {
		if i.Type == t {
			has = true
			break
		}
	}
	for _, o := range c.Outputs {
		if o.Type == t {
			has = true
			break
		}
	}
	return has
}

func (c *Config) GetNameNodes() []string {
	nameNodes := make([]string, 0)

	for _, input := range c.Inputs {
		if input.IsType("hdfs") {
			nameNodes = append(nameNodes, input.NameNode)
		}
	}

	for _, output := range c.Outputs {
		if output.IsType("hdfs") {
			nameNodes = append(nameNodes, output.NameNode)
		}
	}

	return nameNodes
}

// HasDrop returns true if you'll be dropping anything during input
func (c *Config) HasDrop() bool {
	return c.DropAllWays || c.DropAllRelations || c.DropAllVersions || c.DropAllChangesets || c.DropAllTimestamps || c.DropAllUserIds || c.DropAllUserNames
}

func (c *Config) Validate() error {

	for _, input := range c.Inputs {
		if len(input.Uri) == 0 {
			return errors.New("Error: input_uri is missing.")
		}
	}

	for _, output := range c.Outputs {

		if output.WaysToNodes && output.DropWays {
			return errors.New("Error: cannot enable ways_to_nodes and drop_ways at the same time.")
		}

		if output.DropNodes && output.DropWays && output.DropRelations {
			return errors.New("Error: you cannot drop nodes, ways, and relations.  Output will be empty.")
		}

	}

	return nil
}

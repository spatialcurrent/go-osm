package osm

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

type PlanetResource struct {
	*FilteredResource
	DropNodes     bool     `hcl:"drop_nodes"`     // drop nodes
	DropWays      bool     `hcl:"drop_ways"`      // drop ways
	DropRelations bool     `hcl:"drop_relations"` // drop relations
	DropVersion   bool     `hcl:"drop_version"`   // drop version numbers
	DropChangeset bool     `hcl:"drop_changeset"` // drop changeset id
	DropTimestamp bool     `hcl:"drop_timestamp"` // drop last modified timestamp
	DropUserId    bool     `hcl:"drop_user_id"`   // drop the id of the user that last modified an element
	DropUserName  bool     `hcl:"drop_user_name"` // drop the name of the user that last modified an element
	KeysToKeep    []string `hcl:"keep_keys"`      // slice of keys to keep from read elements.  This is not a filter.
	KeysToDrop    []string `hcl:"drop_keys"`      // slice of keys to drop from read elements.  This is not a filter.
}

func (pr PlanetResource) HasDrop() bool {
	return pr.DropNodes || pr.DropWays || pr.DropRelations || pr.DropVersion || pr.DropChangeset || pr.DropTimestamp || pr.DropUserId || pr.DropUserName
}

func (pr *PlanetResource) Init(globals map[string]interface{}, ctx map[string]interface{}, funcs *dfl.FunctionMap) error {

	err := pr.FilteredResource.Init(globals, ctx, funcs)
	if err != nil {
		return err
	}

	if len(globals) > 0 {
		for k, v := range globals {
			switch k {
			case "drop_nodes":
				switch v.(type) {
				case bool:
					pr.DropNodes = v.(bool)
				}
			case "drop_ways":
				switch v.(type) {
				case bool:
					pr.DropWays = v.(bool)
				}
			case "drop_relations":
				switch v.(type) {
				case bool:
					pr.DropRelations = v.(bool)
				}
			case "drop_version":
				switch v.(type) {
				case bool:
					pr.DropVersion = v.(bool)
				}
			case "drop_changeset":
				switch v.(type) {
				case bool:
					pr.DropChangeset = v.(bool)
				}
			case "drop_timestamp":
				switch v.(type) {
				case bool:
					pr.DropTimestamp = v.(bool)
				}
			case "drop_user_id":
				switch v.(type) {
				case bool:
					pr.DropUserId = v.(bool)
				}
			case "drop_user_name":
				switch v.(type) {
				case bool:
					pr.DropUserName = v.(bool)
				}
			case "keep_keys":
				switch v.(type) {
				case []string:
					pr.KeysToKeep = v.([]string)
				}
			case "drop_keys":
				switch v.(type) {
				case []string:
					pr.KeysToDrop = v.([]string)
				}
			}
		}
	}

	return nil
}

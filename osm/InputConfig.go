package osm

type InputConfig struct {
	Uri           string   `hcl:"uri"`            // resource URI
	DropNodes     bool     `hcl:"drop_nodes"`     //drop nodes
	DropWays      bool     `hcl:"drop_ways"`      // drop ways
	DropRelations bool     `hcl:"drop_relations"` // drop relations
	DropVersion   bool     `hcl:"drop_version"`   // drop version numbers
	DropChangeset bool     `hcl:"drop_changeset"` // drop changeset id
	DropTimestamp bool     `hcl:"drop_timestamp"` // drop last modified timestamp
	DropUserId    bool     `hcl:"drop_user_id"`   // drop the id of the user that last modified an element
	DropUserName  bool     `hcl:"drop_user_name"` // drop the name of the user that last modified an element
	KeysToKeep    []string `hcl:"keep_keys"`      // slice of keys to keep from read elements.  This is not a filter.
	KeysToDrop    []string `hcl:"drop_keys"`      // slice of keys to drop from read elements.  This is not a filter.
	//WaysToNodes   bool     `hcl:"ways_to_nodes"`  // convert ways into nodes
	Filter *Filter `hcl:"filter"` // filter input
}

//func NewInputConfig(uri string, drop_nodes bool, drop_ways bool, drop_relations bool, ways_to_nodes bool, filter *Filter) InputConfig {
func NewInputConfig(uri string, drop_nodes bool, drop_ways bool, drop_relations bool, filter *Filter) InputConfig {
	return InputConfig{
		Uri:           uri,
		DropNodes:     drop_nodes,
		DropWays:      drop_ways,
		DropRelations: drop_relations,
		//WaysToNodes:   ways_to_nodes,
		Filter: filter,
	}
}

package osm

type Output struct {
	DropWays      bool
	DropRelations bool
	DropVersion   bool
	DropChangeset bool
	DropTimestamp bool
	DropUserId    bool
	DropUserName  bool
	KeysToKeep    []string
	KeysToDrop    []string
}

func (o Output) HasDrop() bool {
	return o.DropWays || o.DropRelations || o.DropVersion || o.DropChangeset || o.DropTimestamp || o.DropUserId || o.DropUserName
}

func (o Output) HasKeysToKeep() bool {
	return len(o.KeysToKeep) > 0
}

func (o Output) HasKeysToDrop() bool {
	return len(o.KeysToDrop) > 0
}

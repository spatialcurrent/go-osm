package osm

import (
//"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

func KeepNode(planet *Planet, fi *Filter, n *Node, dfl_cache *dfl.Cache) (bool, error) {

	if fi == nil {
		return true, nil
	}

	m := planet.GetTagsAsMap(n.GetTagsIndex())
	m["timestamp"] = n.Timestamp
	m["version"] = n.Version
	m["uid"] = n.UserId
	m["user"] = n.UserName

	if fi.HasKeysToKeep() {
		keep := false
		for _, k := range fi.KeysToKeep {
			if _, ok := m[k]; ok {
				keep = true
				break
			}
		}
		if !keep {
			return false, nil
		}
	}

	if fi.HasKeysToDrop() {
		keep := true
		for _, k := range fi.KeysToDrop {
			if _, ok := m[k]; ok {
				keep = false
				break
			}
		}
		if !keep {
			return false, nil
		}
	}

	if !fi.ContainsPoint(n.Longitude, n.Latitude) {
		return false, nil
	}

	if !fi.HasExpression() {
		return true, nil
	}

	return EvaluateExpression(m, fi.Expression, fi.Functions, fi.Attributes, dfl_cache)
}

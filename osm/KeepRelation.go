package osm

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

func KeepRelation(planet *Planet, fi *Filter, r *Relation, dfl_cache *dfl.Cache) (bool, error) {

	if fi == nil {
		return true, nil
	}

	m := planet.GetTagsAsMap(r.GetTagsIndex())
	m["timestamp"] = r.Timestamp
	m["version"] = r.Version
	m["uid"] = r.UserId
	m["user"] = r.UserName

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

	if !fi.HasExpression() {
		return true, nil
	}

	return EvaluateExpression(m, fi.Expression, fi.Functions, fi.Attributes, dfl_cache)
}

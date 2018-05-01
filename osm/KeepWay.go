package osm

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

func KeepWay(planet *Planet, fi FilterInput, w *Way, dfl_cache *dfl.Cache) (bool, error) {
	m := planet.GetTagsAsMap(w.GetTagsIndex())
	m["timestamp"] = w.Timestamp
	m["version"] = w.Version
	m["uid"] = w.UserId
	m["user"] = w.UserName

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

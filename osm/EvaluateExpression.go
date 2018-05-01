package osm

import (
	"strings"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

func EvaluateExpression(ctx map[string]interface{}, root dfl.Node, funcs dfl.FunctionMap, dfl_attributes []string, dfl_cache *dfl.Cache) (bool, error) {
	key := strings.Join(MapToSlice(ctx, dfl_attributes), "\n")
	if dfl_cache != nil && dfl_cache.Has(key) {
		return dfl_cache.Get(key), nil
	}
	result, err := root.Evaluate(ctx, funcs)
	if err != nil {
		return false, errors.Wrap(err, "Unknown error evaluating node.")
	}
	switch result.(type) {
	case bool:
		result_bool := result.(bool)
		if dfl_cache != nil {
			dfl_cache.Set(key, result_bool)
		}
		return result_bool, nil
	default:
		return false, errors.New("Error converting dfl output to bool")
	}
	return false, errors.New("Unknown error evaluating node.")
}

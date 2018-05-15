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

// EvaluateExpression evaluates a DFL expression against a dfl.Context ctx.
// If the parameter dfl_cache is not nil, then checks the cache with the input for previous result.
// If the cache does not contain the result, then save the result to the cache.
// Returns the boolean result of the expression, and an error if any.
func EvaluateExpression(ctx dfl.Context, root dfl.Node, funcs *dfl.FunctionMap, dfl_attributes []string, dfl_cache *dfl.Cache) (bool, error) {
	key := ""
	if dfl_cache != nil {
		key = strings.Join(MapToSlice(ctx, dfl_attributes), "\n")
		if dfl_cache.Has(key) {
			return dfl_cache.Get(key), nil
		}
	}
	result, err := root.Evaluate(ctx, *funcs)
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

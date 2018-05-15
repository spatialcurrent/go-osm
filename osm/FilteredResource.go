package osm

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

type FilteredResource struct {
	*Resource
	Filter *Filter `hcl:"filter"` // filter input
}

func (fr *FilteredResource) Init(globals map[string]interface{}, ctx map[string]interface{}, funcs *dfl.FunctionMap) error {

	err := fr.Resource.Init(ctx)
	if err != nil {
		return err
	}

	if fr.Filter == nil {
		fr.Filter = NewFilter([]string{}, []string{}, "", true, []float64{})
	}

	err = fr.Filter.Init(globals, funcs)
	if err != nil {
		return err
	}

	return nil

}

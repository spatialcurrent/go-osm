package osm

import (
	"strconv"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

type Filter struct {
	KeysToKeep     []string         `hcl:"keys_keep"`
	KeysToDrop     []string         `hcl:"keys_drop"`
	ExpressionText string           `hcl:"expression"`
	Expression     dfl.Node         `hcl:"-"`
	Functions      *dfl.FunctionMap `hcl:"-"`
	Attributes     []string         `hcl:"-"`
	UseCache       bool             `hcl:"use_cache"`
	BoundingBox    []float64        `hcl:"bbox"`
	MaxExtent      *Bounds          `hcl:"-"`
}

func (f *Filter) Init(globals map[string]interface{}, funcs *dfl.FunctionMap) error {

	if len(f.ExpressionText) > 0 {
		exp, err := dfl.Parse(f.ExpressionText)
		if err != nil {
			return errors.New("Error parsing DFL filter expression \"" + f.ExpressionText + "\"")
		}
		f.Expression = exp.Compile()
		f.Attributes = f.Expression.Attributes()
		f.Functions = funcs
	}

	if len(f.BoundingBox) > 0 {
		if len(f.BoundingBox) != 4 {
			return errors.New("Invalid number of bounding box values " + strconv.Itoa(len(f.BoundingBox)))
		}
		f.MaxExtent = NewBounds(f.BoundingBox[0], f.BoundingBox[1], f.BoundingBox[2], f.BoundingBox[3])
	}

	return nil
}

func (fi Filter) HasExpression() bool {
	return fi.Expression != nil
}

func (fi Filter) HasMaxExtent() bool {
	return fi.MaxExtent != nil
}

func (fi Filter) HasKeysToKeep() bool {
	return len(fi.KeysToKeep) > 0
}

func (fi Filter) HasKeysToDrop() bool {
	return len(fi.KeysToDrop) > 0
}

func (fi Filter) ContainsPoint(lon float64, lat float64) bool {
	if fi.MaxExtent != nil {
		return fi.MaxExtent.ContainsPoint(lon, lat)
	}
	return true
}

func NewFilter(keysToKeep []string, keysToDrop []string, exp string, useCache bool, bbox []float64) *Filter {
	fi := &Filter{
		KeysToKeep:     keysToKeep,
		KeysToDrop:     keysToDrop,
		ExpressionText: exp,
		UseCache:       useCache,
		BoundingBox:    bbox,
	}
	return fi
}

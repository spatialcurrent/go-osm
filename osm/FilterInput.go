package osm

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

type FilterInput struct {
	KeysToKeep []string
	KeysToDrop []string
	Expression dfl.Node
	Functions  dfl.FunctionMap
	Attributes []string
	UseCache   bool
	MaxExtent  *Bounds
}

func (fi FilterInput) HasExpression() bool {
	return fi.Expression != nil
}

func (fi FilterInput) HasMaxExtent() bool {
	return fi.MaxExtent != nil
}

func (fi FilterInput) HasKeysToKeep() bool {
	return len(fi.KeysToKeep) > 0
}

func (fi FilterInput) HasKeysToDrop() bool {
	return len(fi.KeysToDrop) > 0
}

func (fi FilterInput) ContainsPoint(lon float64, lat float64) bool {
	if fi.MaxExtent != nil {
		return fi.MaxExtent.ContainsPoint(lon, lat)
	}
	return true
}

func NewFilterInput(keysToKeep []string, keysToDrop []string, exp dfl.Node, funcs dfl.FunctionMap, useCache bool, bbox []float64) FilterInput {
	fi := FilterInput{
		KeysToKeep: keysToKeep,
		KeysToDrop: keysToDrop,
		Expression: exp,
		Functions:  funcs,
		UseCache:   useCache,
	}

	if fi.Expression != nil {
		fi.Attributes = fi.Expression.Attributes()
	}

	if len(bbox) == 4 {
		fi.MaxExtent = NewBounds(bbox[0], bbox[1], bbox[2], bbox[3])
	}

	return fi
}

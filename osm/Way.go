package osm

import (
	"strings"
	"time"
)

import (
	//"github.com/dhconnelly/rtreego"
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

type NodeReference struct {
	Reference int64 `xml:"ref,attr"`
}

type Way struct {
	Id             int64           `xml:"id,attr"`
	Version        int             `xml:"version,attr,omitempty"`
	Timestamp      *time.Time      `xml:"timestamp,attr,omitempty"`
	Changeset      int64           `xml:"changeset,attr,omitempty"`
	UserId         int64           `xml:"uid,attr,omitempty"`
	UserName       string          `xml:"user,attr,omitempty"`
	NodeReferences []NodeReference `xml:"nd"`
	Tags           []Tag           `xml:"tag"`
}

func (w Way) NumberOfNodes() int {
	return len(w.NodeReferences)
}

func (w Way) HasKey(key string) bool {
	for _, t := range w.Tags {
		if key == t.Key {
			return true
		}
	}
	return false
}

func (w Way) HasAnyKey(keys []string) bool {
	for _, k := range keys {
		if w.HasKey(k) {
			return true
		}
	}
	return false
}

func (w *Way) DropAttributes(output Output) {
	if output.DropVersion {
		w.DropVersion()
	}

	if output.DropTimestamp {
		w.DropTimestamp()
	}

	if output.DropChangeset {
		w.DropChangeset()
	}

	if output.DropUserId {
		w.DropUserId()
	}

	if output.DropUserName {
		w.DropUserName()
	}
}

func (w *Way) DropVersion() {
	w.Version = 0
}

func (w *Way) DropTimestamp() {
	w.Timestamp = nil
}

func (w *Way) DropChangeset() {
	w.Changeset = int64(0)
}

func (w *Way) DropUserId() {
	w.UserId = int64(0)
}

func (w *Way) DropUserName() {
	w.UserName = ""
}

func (w *Way) TagsAsMap() map[string]interface{} {
	m := map[string]interface{}{}
	for _, t := range w.Tags {
		m[t.Key] = t.Value
	}
	return m
}

func (w *Way) Evaluate(root dfl.Node, funcs dfl.FunctionMap, dfl_attributes []string, dfl_cache *dfl.Cache) (bool, error) {
	m := w.TagsAsMap()
	key := strings.Join(MapToSlice(m, dfl_attributes), "\n")
	if dfl_cache != nil && dfl_cache.Has(key) {
		return dfl_cache.Get(key), nil
	}
	result, err := root.Evaluate(m, funcs)
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

func (w *Way) Keep(fi FilterInput, dfl_cache *dfl.Cache) (bool, error) {
	if fi.HasKeysToKeep() {
		if !w.HasAnyKey(fi.KeysToKeep) {
			return false, nil
		}
	}
	if fi.HasKeysToDrop() {
		if w.HasAnyKey(fi.KeysToDrop) {
			return false, nil
		}
	}
	if !fi.HasExpression() {
		return true, nil
	}
	return w.Evaluate(fi.Expression, fi.Functions, fi.Attributes, dfl_cache)
}

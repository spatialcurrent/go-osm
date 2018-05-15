package osm

import (
//"strings"
//"time"
)

import (
//"github.com/dhconnelly/rtreego"
//"github.com/pkg/errors"
)

import (
//"github.com/spatialcurrent/go-dfl/dfl"
)

type NodeReference struct {
	Reference uint64 `xml:"ref,attr" parquet:"ref, type=UINT_64"`
}

type Way struct {
	TaggedElement
	NodeReferences []NodeReference `xml:"nd"`
}

func (w Way) NumberOfNodes() int {
	return len(w.NodeReferences)
}

func NewWay() *Way {
	return &Way{TaggedElement: TaggedElement{Element: Element{}}}
}

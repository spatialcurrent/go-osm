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
	Reference int64 `xml:"ref,attr"`
}

type Way struct {
	TaggedElement
	NodeReferences []NodeReference `xml:"nd"`
}

func (w Way) NumberOfNodes() int {
	return len(w.NodeReferences)
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

func NewWay() *Way {
	return &Way{TaggedElement: TaggedElement{Element: Element{}}}
}

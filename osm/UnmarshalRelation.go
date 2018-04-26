package osm

import (
	"encoding/xml"
)

import (
	"github.com/pkg/errors"
)

func UnmarshalRelation(decoder *xml.Decoder, e xml.StartElement, output Output) (Relation, error) {
	r := Relation{}
	err := decoder.DecodeElement(&r, &e)
	if err != nil {
		return r, errors.Wrap(err, "Error decoding relation.")
	}
	return r, nil
}

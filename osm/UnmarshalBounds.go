package osm

import (
	"encoding/xml"
)

import (
	"github.com/pkg/errors"
)

func UnmarshalBounds(decoder *xml.Decoder, e xml.StartElement) (Bounds, error) {
	b := Bounds{}
	err := decoder.DecodeElement(&b, &e)
	if err != nil {
		return b, errors.Wrap(err, "Error decoding bounds.")
	}
	return b, nil
}

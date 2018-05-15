package osm

import (
	"encoding/xml"
)

import (
	"github.com/pkg/errors"
)

// UnmarshalBounds unmarshals a bounds XML element.  For example:
//	<bounds minlon="-77.120100" minlat="38.791340" maxlon="-76.909060" maxlat="38.996030"></bounds>
func UnmarshalBounds(decoder *xml.Decoder, e xml.StartElement) (Bounds, error) {
	b := Bounds{}
	err := decoder.DecodeElement(&b, &e)
	if err != nil {
		return b, errors.Wrap(err, "Error decoding bounds.")
	}
	return b, nil
}

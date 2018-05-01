package osm

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"time"
)

import (
	"github.com/pkg/errors"
)

func MarshalWay(encoder *xml.Encoder, planet *Planet, output_config Output, w *Way) error {
	attrs := []xml.Attr{
		xml.Attr{Name: xml.Name{Space: "", Local: "id"}, Value: fmt.Sprint(w.Id)},
	}
	if !output_config.DropVersion {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "version"}, Value: strconv.Itoa(w.Version)})
	}
	if !output_config.DropTimestamp {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "timestamp"}, Value: w.Timestamp.Format(time.RFC3339)})
	}
	if !output_config.DropChangeset {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "changeset"}, Value: fmt.Sprint(w.Changeset)})
	}
	if !output_config.DropUserName {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "uid"}, Value: fmt.Sprint(w.UserId)})
	}
	if !output_config.DropUserName {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "user"}, Value: fmt.Sprint(w.UserName)})
	}
	token_way := xml.StartElement{
		Name: xml.Name{Space: "", Local: "way"},
		Attr: attrs,
	}
	err := encoder.EncodeToken(token_way)
	if err != nil {
		return errors.Wrap(err, "Error encoding way start element.")
	}
	for _, nr := range w.NodeReferences {
		token_nr := xml.StartElement{
			Name: xml.Name{Space: "", Local: "nd"},
			Attr: []xml.Attr{
				xml.Attr{Name: xml.Name{Space: "", Local: "ref"}, Value: fmt.Sprint(nr.Reference)},
			},
		}
		err = encoder.EncodeToken(token_nr)
		if err != nil {
			return errors.Wrap(err, "Error encoding node reference element.")
		}
		err = encoder.EncodeToken(token_nr.End())
		if err != nil {
			return errors.Wrap(err, "Error encoding node reference end element.")
		}
	}
	for _, tagIndex := range w.GetTagsIndex() {
		tag := planet.GetTag(tagIndex)
		token_tag := xml.StartElement{
			Name: xml.Name{Space: "", Local: "tag"},
			Attr: []xml.Attr{
				xml.Attr{Name: xml.Name{Space: "", Local: "k"}, Value: fmt.Sprint(tag.Key)},
				xml.Attr{Name: xml.Name{Space: "", Local: "v"}, Value: fmt.Sprint(tag.Value)},
			},
		}
		err = encoder.EncodeToken(token_tag)
		if err != nil {
			return errors.Wrap(err, "Error encoding tag element.")
		}
		err = encoder.EncodeToken(token_tag.End())
		if err != nil {
			return errors.Wrap(err, "Error encoding bounds end element.")
		}
	}
	err = encoder.EncodeToken(token_way.End())
	if err != nil {
		return errors.Wrap(err, "Error encoding way end element")
	}
	return nil
}

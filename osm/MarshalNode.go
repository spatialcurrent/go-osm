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

func MarshalNode(encoder *xml.Encoder, planet *Planet, output_config *Output, n *Node) error {
	attrs := []xml.Attr{
		xml.Attr{Name: xml.Name{Space: "", Local: "id"}, Value: fmt.Sprint(n.Id)},
		xml.Attr{Name: xml.Name{Space: "", Local: "lat"}, Value: strconv.FormatFloat(n.Latitude, 'f', 6, 64)},
		xml.Attr{Name: xml.Name{Space: "", Local: "lon"}, Value: strconv.FormatFloat(n.Longitude, 'f', 6, 64)},
	}
	if !output_config.DropVersion {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "version"}, Value: fmt.Sprint(n.Version)})
	}
	if !output_config.DropTimestamp {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "timestamp"}, Value: n.Timestamp.Format(time.RFC3339)})
	}
	if !output_config.DropChangeset {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "changeset"}, Value: fmt.Sprint(n.Changeset)})
	}
	if !output_config.DropUserId {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "uid"}, Value: fmt.Sprint(n.UserId)})
	}
	if !output_config.DropUserName {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "user"}, Value: fmt.Sprint(planet.UserNames[n.UserId])})
	}
	token_node := xml.StartElement{
		Name: xml.Name{Space: "", Local: "node"},
		Attr: attrs,
	}
	err := encoder.EncodeToken(token_node)
	if err != nil {
		return errors.Wrap(err, "Error encoding node start element.")
	}
	for _, tagIndex := range n.GetTagsIndex() {
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
	err = encoder.EncodeToken(token_node.End())
	if err != nil {
		return errors.Wrap(err, "Error encoding bounds end element.")
	}
	return nil
}

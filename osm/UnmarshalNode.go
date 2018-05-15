package osm

import (
	"encoding/xml"
	"strconv"
	"time"
)

import (
	"github.com/pkg/errors"
)

// UnmarshalNode unmarshals a node from an XML element.
// Returns the Node, the tags as a separate slice, and an error if any.  For example:
//	<node id="4580586204" lat="38.949900" lon="-77.080715" timestamp="2016-12-30T03:51:18Z">
//	  <tag k="name" v="District Taco"></tag>
//	  <tag k="amenity" v="restaurant"></tag>
//	  <tag k="addr:city" v="Washington"></tag>
//	  <tag k="addr:state" v="DC"></tag>
//	</node>
//
func UnmarshalNode(decoder *xml.Decoder, e xml.StartElement, input *Input) (*Node, uint64, string, []Tag, error) {
	n := NewNode()
	user_id := uint64(0)
	user_name := ""
	tags := make([]Tag, 0)

	for _, attr := range e.Attr {
		switch attr.Name.Local {
		case "id":
			id, err := strconv.ParseUint(attr.Value, 10, 64)
			if err != nil {
				return n, user_id, user_name, tags, errors.Wrap(err, "Error parsing node id")
			}
			n.Id = id
		case "version":
			if !input.DropVersion {
				version, err := strconv.ParseUint(attr.Value, 10, 16)
				if err != nil {
					return n, user_id, user_name, tags, errors.Wrap(err, "Error parsing node version")
				}
				n.Version = uint16(version)
			}
		case "changeset":
			if !input.DropChangeset {
				changeset, err := strconv.ParseUint(attr.Value, 10, 64)
				if err != nil {
					return n, user_id, user_name, tags, errors.Wrap(err, "Error parsing node changeset")
				}
				n.Changeset = changeset
			}
		case "timestamp":
			if !input.DropTimestamp {
				ts, err := time.Parse(time.RFC3339, attr.Value)
				if err != nil {
					return n, user_id, user_name, tags, errors.Wrap(err, "Error parsing node timestamp")
				}
				n.Timestamp = &ts
			}
		case "uid":
			if !input.DropUserId {
				uid, err := strconv.ParseUint(attr.Value, 10, 64)
				if err != nil {
					return n, user_id, user_name, tags, errors.Wrap(err, "Error parsing node uid")
				}
				//n.UserId = uid
				user_id = uid
			}
		case "user":
			if !input.DropUserName {
				//n.UserName = attr.Value
				user_name = attr.Value
			}
		case "lat":
			lat, err := strconv.ParseFloat(attr.Value, 64)
			if err != nil {
				return n, user_id, user_name, tags, errors.Wrap(err, "Error parsing node latitude")
			}
			n.Latitude = lat
		case "lon":
			lon, err := strconv.ParseFloat(attr.Value, 64)
			if err != nil {
				return n, user_id, user_name, tags, errors.Wrap(err, "Error parsing node longitude")
			}
			n.Longitude = lon
		}
	}

	tags = UnmarshalTags(decoder, input.KeysToKeep, input.KeysToDrop)

	return n, user_id, user_name, tags, nil
}

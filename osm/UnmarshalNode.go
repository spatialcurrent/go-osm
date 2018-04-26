package osm

import (
	"encoding/xml"
	"strconv"
	"time"
)

import (
	"github.com/pkg/errors"
)

func UnmarshalNode(decoder *xml.Decoder, e xml.StartElement, output Output) (Node, error) {
	n := Node{}

	for _, attr := range e.Attr {
		switch attr.Name.Local {
		case "id":
			id, err := strconv.ParseInt(attr.Value, 10, 64)
			if err != nil {
				return n, errors.Wrap(err, "Error parsing node id")
			}
			n.Id = id
		case "version":
			if !output.DropVersion {
				version, err := strconv.Atoi(attr.Value)
				if err != nil {
					return n, errors.Wrap(err, "Error parsing node version")
				}
				n.Version = version
			}
		case "changeset":
			if !output.DropChangeset {
				changeset, err := strconv.ParseInt(attr.Value, 10, 64)
				if err != nil {
					return n, errors.Wrap(err, "Error parsing node changeset")
				}
				n.Changeset = changeset
			}
		case "timestamp":
			if !output.DropTimestamp {
				ts, err := time.Parse(time.RFC3339, attr.Value)
				if err != nil {
					return n, errors.Wrap(err, "Error parsing node timestamp")
				}
				n.Timestamp = &ts
			}
		case "uid":
			if !output.DropUserId {
				uid, err := strconv.ParseInt(attr.Value, 10, 64)
				if err != nil {
					return n, errors.Wrap(err, "Error parsing node uid")
				}
				n.UserId = uid
			}
		case "user":
			if !output.DropUserName {
				n.UserName = attr.Value
			}
		case "lat":
			lat, err := strconv.ParseFloat(attr.Value, 64)
			if err != nil {
				return n, errors.Wrap(err, "Error parsing node latitude")
			}
			n.Latitude = lat
		case "lon":
			lon, err := strconv.ParseFloat(attr.Value, 64)
			if err != nil {
				return n, errors.Wrap(err, "Error parsing node longitude")
			}
			n.Longitude = lon
		}
	}

	n.Tags = UnmarshalTags(decoder, output.KeysToKeep, output.KeysToKeep)

	return n, nil
}

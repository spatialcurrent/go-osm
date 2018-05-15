package osm

import (
	"encoding/xml"
	"strconv"
	"time"
)

import (
	"github.com/pkg/errors"
)

// UnmarshalWay unmarshals a Way from an XML stream.
// Returns the way's tags as a separate slice, so they can be cached
func UnmarshalWay(decoder *xml.Decoder, e xml.StartElement, input *Input) (*Way, uint64, string, []Tag, error) {
	w := NewWay()
	user_id := uint64(0)
	user_name := ""
	tags := make([]Tag, 0)

	for _, attr := range e.Attr {
		switch attr.Name.Local {
		case "id":
			id, err := strconv.ParseUint(attr.Value, 10, 64)
			if err != nil {
				return w, user_id, user_name, tags, errors.Wrap(err, "Error parsing way id")
			}
			w.Id = id
		case "version":
			if !input.DropVersion {
				version, err := strconv.ParseUint(attr.Value, 10, 16)
				if err != nil {
					return w, user_id, user_name, tags, errors.Wrap(err, "Error parsing way version")
				}
				w.Version = uint16(version)
			}
		case "changeset":
			if !input.DropChangeset {
				changeset, err := strconv.ParseUint(attr.Value, 10, 64)
				if err != nil {
					return w, user_id, user_name, tags, errors.Wrap(err, "Error parsing way changeset")
				}
				w.Changeset = changeset
			}
		case "timestamp":
			if !input.DropTimestamp {
				ts, err := time.Parse(time.RFC3339, attr.Value)
				if err != nil {
					return w, user_id, user_name, tags, errors.Wrap(err, "Error parsing way timestamp")
				}
				w.Timestamp = &ts
			}
		case "uid":
			if !input.DropUserId {
				uid, err := strconv.ParseUint(attr.Value, 10, 64)
				if err != nil {
					return w, user_id, user_name, tags, errors.Wrap(err, "Error parsing way uid")
				}
				user_id = uid
			}
		case "user":
			if !input.DropUserName {
				user_name = attr.Value
			}
		}
	}

NodesAndTags:
	for {

		token, _ := decoder.Token()
		if token == nil {
			break
		}

		switch e := token.(type) {
		case xml.StartElement:
			switch e.Name.Local {
			case "tag":
				tag := Tag{}
				for _, attr := range e.Attr {
					switch attr.Name.Local {
					case "k":
						tag.Key = attr.Value
					case "v":
						tag.Value = attr.Value
					}
				}
				keep := true
				if len(input.KeysToKeep) > 0 {
					keep = false
					for _, k := range input.KeysToKeep {
						if tag.Key == k {
							keep = true
							break
						}
					}
				} else if len(input.KeysToDrop) > 0 {
					for _, k := range input.KeysToDrop {
						if tag.Key == k {
							keep = false
							break
						}
					}
				}
				if keep {
					tags = append(tags, tag)
				}
			case "nd":
				nr := NodeReference{}
				for _, attr := range e.Attr {
					switch attr.Name.Local {
					case "ref":
						ref, err := strconv.ParseUint(attr.Value, 10, 64)
						if err != nil {
							return w, user_id, user_name, tags, errors.Wrap(err, "Error parsing node reference for way.")
						}
						nr.Reference = ref
					}
				}
				w.NodeReferences = append(w.NodeReferences, nr)
			}
		case xml.EndElement:
			switch e.Name.Local {
			case "way":
				break NodesAndTags
			}
		}
	}

	return w, user_id, user_name, tags, nil
}

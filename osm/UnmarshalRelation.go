package osm

import (
	"encoding/xml"
	"strconv"
	"time"
)

import (
	"github.com/pkg/errors"
)

// UnmarshalRelation unmarhals a relation from an XML stream
// Returns the unmarshalled relation, and an error if any.
func UnmarshalRelation(decoder *xml.Decoder, e xml.StartElement, input *Input) (*Relation, uint64, string, []Tag, error) {
	r := NewRelation()
	user_id := uint64(0)
	user_name := ""
	tags := make([]Tag, 0)

	for _, attr := range e.Attr {
		switch attr.Name.Local {
		case "id":
			id, err := strconv.ParseUint(attr.Value, 10, 64)
			if err != nil {
				return r, user_id, user_name, tags, errors.Wrap(err, "Error parsing relation id")
			}
			r.Id = id
		case "version":
			if !input.DropVersion {
				version, err := strconv.ParseUint(attr.Value, 10, 16)
				if err != nil {
					return r, user_id, user_name, tags, errors.Wrap(err, "Error parsing relation version")
				}
				r.Version = uint16(version)
			}
		case "changeset":
			if !input.DropChangeset {
				changeset, err := strconv.ParseUint(attr.Value, 10, 64)
				if err != nil {
					return r, user_id, user_name, tags, errors.Wrap(err, "Error parsing relation changeset")
				}
				r.Changeset = changeset
			}
		case "timestamp":
			if !input.DropTimestamp {
				ts, err := time.Parse(time.RFC3339, attr.Value)
				if err != nil {
					return r, user_id, user_name, tags, errors.Wrap(err, "Error parsing relation timestamp")
				}
				r.Timestamp = &ts
			}
		case "uid":
			if !input.DropUserId {
				uid, err := strconv.ParseUint(attr.Value, 10, 64)
				if err != nil {
					return r, user_id, user_name, tags, errors.Wrap(err, "Error parsing relation uid")
				}
				user_id = uid
			}
		case "user":
			if !input.DropUserName {
				user_name = attr.Value
			}
		}
	}

MembersAndTags:
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
			case "member":
				rm := RelationMember{}
				for _, attr := range e.Attr {
					switch attr.Name.Local {
					case "type":
						rm.Type = attr.Value
					case "ref":
						ref, err := strconv.ParseUint(attr.Value, 10, 64)
						if err != nil {
							return r, user_id, user_name, tags, errors.Wrap(err, "Error parsing relation member reference")
						}
						rm.Reference = ref
					case "role":
						rm.Role = attr.Value
					}
				}
				r.Members = append(r.Members, rm)
			}
		case xml.EndElement:
			switch e.Name.Local {
			case "relation":
				break MembersAndTags
			}
		}
	}

	return r, user_id, user_name, tags, nil
}

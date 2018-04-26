package osm

import (
	"encoding/xml"
	"strconv"
	"time"
)

import (
	"github.com/pkg/errors"
)

func UnmarshalWay(decoder *xml.Decoder, e xml.StartElement, output Output) (Way, error) {
	w := Way{}

	for _, attr := range e.Attr {
		switch attr.Name.Local {
		case "id":
			id, err := strconv.ParseInt(attr.Value, 10, 64)
			if err != nil {
				return w, errors.Wrap(err, "Error parsing way id")
			}
			w.Id = id
		case "version":
			if !output.DropVersion {
				version, err := strconv.Atoi(attr.Value)
				if err != nil {
					return w, errors.Wrap(err, "Error parsing way version")
				}
				w.Version = version
			}
		case "changeset":
			if !output.DropChangeset {
				changeset, err := strconv.ParseInt(attr.Value, 10, 64)
				if err != nil {
					return w, errors.Wrap(err, "Error parsing way changeset")
				}
				w.Changeset = changeset
			}
		case "timestamp":
			if !output.DropTimestamp {
				ts, err := time.Parse(time.RFC3339, attr.Value)
				if err != nil {
					return w, errors.Wrap(err, "Error parsing way timestamp")
				}
				w.Timestamp = &ts
			}
		case "uid":
			if !output.DropUserId {
				uid, err := strconv.ParseInt(attr.Value, 10, 64)
				if err != nil {
					return w, errors.Wrap(err, "Error parsing way uid")
				}
				w.UserId = uid
			}
		case "user":
			if !output.DropUserName {
				w.UserName = attr.Value
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
				if len(output.KeysToKeep) > 0 {
					keep = false
					for _, k := range output.KeysToKeep {
						if tag.Key == k {
							keep = true
							break
						}
					}
				} else if len(output.KeysToDrop) > 0 {
					for _, k := range output.KeysToDrop {
						if tag.Key == k {
							keep = false
							break
						}
					}
				}
				if keep {
					w.Tags = append(w.Tags, tag)
				}
			case "nd":
				nr := NodeReference{}
				for _, attr := range e.Attr {
					switch attr.Name.Local {
					case "ref":
						ref, err := strconv.ParseInt(attr.Value, 10, 64)
						if err != nil {
							return w, errors.Wrap(err, "Error parsing node reference")
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

	return w, nil
}

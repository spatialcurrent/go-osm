package osm

import (
	"encoding/xml"
)

func UnmarshalTags(decoder *xml.Decoder, keep_keys []string, drop_keys []string) []Tag {
	tags := make([]Tag, 0)

Tags:
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
				if len(keep_keys) > 0 {
					keep = false
					for _, k := range keep_keys {
						if tag.Key == k {
							keep = true
							break
						}
					}
				} else if len(drop_keys) > 0 {
					for _, k := range drop_keys {
						if tag.Key == k {
							keep = false
							break
						}
					}
				}
				if keep {
					tags = append(tags, tag)
				}
			}
		case xml.EndElement:
			switch e.Name.Local {
			case "node":
				break Tags
			}
		}
	}

	return tags
}

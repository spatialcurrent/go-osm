package osm

type Tag struct {
	Key   string `xml:"k,attr"`
	Value string `xml:"v,attr"`
}

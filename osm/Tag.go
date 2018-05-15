package osm

// Tag is a key=value construct used in an OpenStreetMap for associating attributes with physical geospatial features or relationships among those features.
type Tag struct {
	Key   string `xml:"k,attr"` // the key
	Value string `xml:"v,attr"` // the value
}

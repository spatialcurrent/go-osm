package osm

func ParseBool(in string) bool {
	return in == "yes" || in == "true" || in == "y" || in == "1" || in == "t"
}

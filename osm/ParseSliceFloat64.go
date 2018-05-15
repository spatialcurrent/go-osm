package osm

import (
	"strconv"
	"strings"
)

func ParseSliceFloat64(in string) ([]float64, error) {
	out := make([]float64, 0)
	if len(in) > 0 {
		for _, s := range strings.Split(in, ",") {
			v, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return out, err
			}
			out = append(out, v)
		}
	}
	return out, nil
}

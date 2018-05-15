package osm

import (
	"bufio"
	"compress/gzip"
	"encoding/xml"
	"fmt"
	//"io"
	"os"
	"strconv"
	"strings"
	"time"
)

import (
	"github.com/pkg/errors"
)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

func MarshalPlanet(output *Output, config *Config, planet *Planet) error {

	var dfl_cache *dfl.Cache
	if output.Filter.HasExpression() && output.Filter.UseCache {
		dfl_cache = dfl.NewCache()
	}

	var output_file *os.File
	var writer *bufio.Writer
	//var input_file *os.File
	//var input_reader io.Reader

	if output.Uri == "stdout" {
		writer = bufio.NewWriter(os.Stdout)
	} else if output.Uri == "sterr" {
		writer = bufio.NewWriter(os.Stderr)
	} else if output.Scheme == "file" {
		if strings.HasSuffix(output.PathExpanded, ".osm") {
			f, err := os.OpenFile(output.PathExpanded, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return errors.Wrap(err, "error opening file to write osm file to disk")
			}
			output_file = f
			writer = bufio.NewWriter(f)
		} else if strings.HasSuffix(output.PathExpanded, ".osm.gz") {
			f, err := os.OpenFile(output.PathExpanded, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				return errors.Wrap(err, "error opening file to write osm file to disk")
			}
			output_file = f
			writer = bufio.NewWriter(gzip.NewWriter(f))
		} else {
			return errors.New("Invalid extension for output " + output.Uri)
		}
	} else {
		return errors.New("unknown output_uri " + output.Uri)
	}

	fmt.Fprint(writer, xml.Header)
	encoder := xml.NewEncoder(writer)
	if output.Pretty {
		encoder.Indent("", "    ")
	}
	attrs := make([]xml.Attr, 0)
	if !output.DropVersion {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "version"}, Value: planet.Version})
	}
	if !output.DropTimestamp {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "timestamp"}, Value: planet.Timestamp.Format(time.RFC3339)})
	}
	token_osm := xml.StartElement{
		Name: xml.Name{Space: "", Local: "osm"},
		Attr: attrs,
	}
	err := encoder.EncodeToken(token_osm)
	if err != nil {
		return errors.Wrap(err, "Error encoding osm start element.")
	}
	token := xml.StartElement{
		Name: xml.Name{Space: "", Local: "bounds"},
		Attr: []xml.Attr{
			xml.Attr{Name: xml.Name{Space: "", Local: "minlon"}, Value: strconv.FormatFloat(planet.Bounds.MinimumLongitude, 'f', 6, 64)},
			xml.Attr{Name: xml.Name{Space: "", Local: "minlat"}, Value: strconv.FormatFloat(planet.Bounds.MinimumLatitude, 'f', 6, 64)},
			xml.Attr{Name: xml.Name{Space: "", Local: "maxlon"}, Value: strconv.FormatFloat(planet.Bounds.MaximumLongitude, 'f', 6, 64)},
			xml.Attr{Name: xml.Name{Space: "", Local: "maxlat"}, Value: strconv.FormatFloat(planet.Bounds.MaximumLatitude, 'f', 6, 64)},
		},
	}
	err = encoder.EncodeToken(token)
	if err != nil {
		return errors.Wrap(err, "Error encoding bounds element.")
	}
	err = encoder.EncodeToken(token.End())
	if err != nil {
		return errors.Wrap(err, "Error encoding bounds end element.")
	}

	uid := planet.maxId
	set_way_nodes := NewUInt64Set()
	nodes := make([]*Node, 0)
	ways := make([]*Way, 0)
	if !output.DropWays {
		for _, w := range planet.Ways {
			keep, err := KeepWay(planet, output.Filter, w, dfl_cache)
			if err != nil {
				return err
			}
			if keep {
				if output.WaysToNodes {
					uid += 1
					n, err := planet.ConvertWayToNode(w, uid)
					if err != nil {
						//fmt.Println(err)
						//continue
						return errors.Wrap(err, "Error converting way to node.")
					}
					nodes = append(nodes, n)
				} else {
					ways = append(ways, w)
					for _, nr := range w.NodeReferences {
						set_way_nodes.Add(nr.Reference)
					}
				}
			}
		}
	}
	slice_way_nodes := set_way_nodes.Slice(true)

	for _, n := range planet.Nodes {
		keep := true
		if !slice_way_nodes.Contains(n.Id) {
			if output.DropNodes {
				keep = false
			} else {
				keep, err = KeepNode(planet, output.Filter, n, dfl_cache)
				if err != nil {
					return err
				}
			}
		}
		if keep {
			err := MarshalNode(encoder, planet, output, n)
			if err != nil {
				return errors.Wrap(err, "Error marshalling node")
			}
		}
	}

	for _, n := range nodes {
		err := MarshalNode(encoder, planet, output, n)
		if err != nil {
			return errors.Wrap(err, "Error marshalling node")
		}
	}

	for _, w := range ways {
		err := MarshalWay(encoder, planet, output, w)
		if err != nil {
			return errors.Wrap(err, "Error marshalling way")
		}
	}

	err = encoder.EncodeToken(token_osm.End())
	if err != nil {
		return errors.Wrap(err, "Error encoding osm end element.")
	}

	err = encoder.Flush()
	if err != nil {
		return errors.Wrap(err, "Error flushing buffered xml.")
	}

	writer.WriteString("\n")
	err = writer.Flush()
	if err != nil {
		return errors.Wrap(err, "Error flushing writer.")
	}

	if output_file != nil {
		err := output_file.Close()
		if err != nil {
			return errors.Wrap(err, "Error closing file writer for xml file.")
		}
	}

	return nil
}

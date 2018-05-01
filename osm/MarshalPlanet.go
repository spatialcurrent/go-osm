package osm

import (
	"bufio"
	"encoding/xml"
	"fmt"
	//"io"
	"os"
	"strconv"
	"time"
)

import (
	"github.com/pkg/errors"
)

func MarshalPlanet(output_uri string, output_scheme string, output_path string, output_config Output, planet *Planet) error {

	var writer *bufio.Writer
	//var input_file *os.File
	//var input_reader io.Reader

	if output_uri == "stdout" {
		writer = bufio.NewWriter(os.Stdout)
	} else if output_uri == "sterr" {
		writer = bufio.NewWriter(os.Stderr)
	} else if output_scheme == "file" {
		output_file, err := os.OpenFile(output_path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
		if err != nil {
			return errors.Wrap(err, "error opening file to write osm file to disk")
		}
		writer = bufio.NewWriter(output_file)
	} else {
		return errors.New("unknown output_uri " + output_uri)
	}

	fmt.Fprint(writer, xml.Header)
	encoder := xml.NewEncoder(writer)
	if output_config.Pretty {
		encoder.Indent("", "    ")
	}
	attrs := make([]xml.Attr, 0)
	if !output_config.DropVersion {
		attrs = append(attrs, xml.Attr{Name: xml.Name{Space: "", Local: "version"}, Value: planet.Version})
	}
	if !output_config.DropTimestamp {
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
	for _, n := range planet.Nodes {
		err := MarshalNode(encoder, planet, output_config, n)
		if err != nil {
			return errors.Wrap(err, "Error marshalling node")
		}
	}
	for _, w := range planet.Ways {
		err := MarshalWay(encoder, planet, output_config, w)
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

	return nil
}

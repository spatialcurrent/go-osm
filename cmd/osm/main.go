package main

import (
	"bufio"
	"compress/gzip"
	//"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

//import (
//	"github.com/golang/protobuf/proto"
//)

import (
	"github.com/mitchellh/go-homedir"
)

import (
	"github.com/spatialcurrent/go-composite-logger/compositelogger"
)

import (
	"github.com/spatialcurrent/go-osm/osm"
)

var GO_OSM_VERSION = "0.0.1"

func main() {

	start := time.Now()

	var input_uri string
	var output_uri string
	var include_keys_text string

	var ways_to_nodes bool
	var drop_relations bool
	var drop_version bool
	var drop_timestamp bool
	var drop_changeset bool
	var summarize bool

	var verbose bool
	var dry_run bool
	var version bool
	var help bool

	flag.StringVar(&input_uri, "input_uri", "", "Input uri")
	flag.StringVar(&output_uri, "output_uri", "", "Output uri")
	flag.StringVar(&include_keys_text, "include_keys", "", "Comma-separated list of tag keys to keep")

	flag.BoolVar(&ways_to_nodes, "ways_to_nodes", false, "Convert ways into nodes")
	flag.BoolVar(&drop_relations, "drop_relations", false, "Drop relations")
	flag.BoolVar(&drop_version, "drop_version", false, "Drop version")
	flag.BoolVar(&drop_timestamp, "drop_timestamp", false, "Drop timestamp")
	flag.BoolVar(&drop_changeset, "drop_changeset", false, "Drop changeset")
	flag.BoolVar(&summarize, "summarize", false, "Summarize data")

	flag.BoolVar(&verbose, "verbose", false, "Provide verbose output")
	flag.BoolVar(&dry_run, "dry_run", false, "Connect to destination, but don't import any data.")
	flag.BoolVar(&version, "version", false, "Version")
	flag.BoolVar(&help, "help", false, "Print help")

	flag.Parse()

	include_keys := strings.Split(include_keys_text, ",")

	if help {
		fmt.Println("Usage: osm -input_uri INPUT -output_uri OUTPUT [-verbose] [-dry_run] [-version] [-help]")
		flag.PrintDefaults()
		os.Exit(0)
	} else if len(os.Args) == 1 {
		fmt.Println("Error: Provided no arguments.")
		fmt.Println("Run \"osm --help\" for more information.")
		os.Exit(0)
	} else if flag.NArg() > 0 {
		fmt.Println("Error: Provided extra command line arguments:", strings.Join(flag.Args(), ", "))
		fmt.Println("Run \"osm --help\" for more information.")
		os.Exit(0)
	}

	if version {
		fmt.Println(GO_OSM_VERSION)
		os.Exit(0)
	}

	if len(include_keys) == 0 {
		fmt.Println("Error: Missing command line argument \"inlcude_keys\".")
		fmt.Println("Run \"osm --help\" for more information.")
		os.Exit(1)
	}

	if dry_run {
		os.Exit(1)
	}

	logger, err := compositelogger.NewDefaultLogger()
	if err != nil {
		fmt.Println("Error initializing composite logger.")
		fmt.Println(err)
		os.Exit(1)
	}

	input_bytes := make([]byte, 0)
	if input_uri == "stdin" {

		in, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("Error reading from stdin.")
			os.Exit(1)
		}
		input_bytes = []byte(strings.TrimSpace(string(in)))

	} else {

		input_file := ""
		if strings.HasPrefix(input_uri, "file://") {
			input_file = input_uri[7:]
		} else {
			input_file = input_uri
		}

		input_file_expanded, err := homedir.Expand(input_file)
		if err != nil {
			fmt.Println("Error expanding path")
			os.Exit(1)
		}

		in, err := ioutil.ReadFile(input_file_expanded)
		if err != nil {
			fmt.Println("Error reading from uri  " + input_uri + ".")
			fmt.Println(err)
			os.Exit(1)
		}

		if strings.HasSuffix(input_file_expanded, ".xml") || strings.HasSuffix(input_file_expanded, ".osm") {
			input_bytes = []byte(strings.TrimSpace(string(in)))
		} else {
			input_bytes = in
		}

	}

	planet := osm.Planet{}
	if strings.HasSuffix(input_uri, ".pbf") {
		fmt.Println("Protobuf not implemented yet.")
		os.Exit(1)
		//err = proto.Unmarshal(input_bytes, &planet)
		//if err != nil {
		//	fmt.Println("Error unmarhsalling input")
		//	fmt.Println(err)
		//	os.Exit(1)
		//}
	} else {
		err = xml.Unmarshal(input_bytes, &planet)
		if err != nil {
			fmt.Println("Error unmarhsalling input")
			fmt.Println(err)
			os.Exit(1)
		}
	}

	planet.Filter(include_keys)

	if ways_to_nodes {
		planet.ConvertWaysToNodes()
	}

	if drop_relations {
		planet.DropRelations()
	}

	if drop_version {
		planet.DropVersion()
	}

	if drop_timestamp {
		planet.DropTimestamp()
	}

	if drop_changeset {
		planet.DropChangeset()
	}

	if summarize {
		fmt.Println("Bounding Box:", planet.BoundingBox())
		fmt.Println("Number of Nodes:", len(planet.Nodes))
		fmt.Println("Number of Ways:", len(planet.Ways))
		fmt.Println("Number of Relations:", len(planet.Relations))
	}

	if len(output_uri) > 0 {

		//output_bytes, err := xml.MarshalIndent(&osm.PlanetFile{&planet}, "  ", "    ")
		output_bytes, err := xml.MarshalIndent(&planet, "  ", "    ")
		if err != nil {
			fmt.Println("Error marshalling output")
			fmt.Println(err)
			os.Exit(1)
		}
		output_text := xml.Header + string(output_bytes)

		if output_uri == "stdout" {
			fmt.Println(output_text)
		} else if output_uri == "stderr" {
			fmt.Fprintf(os.Stderr, output_text)
		} else {
			if verbose {
				fmt.Println("Writing to " + output_uri + ".")
			}

			output_path := ""
			if strings.HasPrefix(output_uri, "file://") {
				output_path = output_uri[7:]
			} else {
				output_path = output_uri
			}

			output_file, err := os.OpenFile(output_path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				os.Exit(1)
			}

			if strings.HasSuffix(output_path, ".gz") {
				gw := gzip.NewWriter(output_file)
				w := bufio.NewWriter(gw)
				w.WriteString(output_text)
				w.Flush()
			} else {
				w := bufio.NewWriter(output_file)
				w.WriteString(output_text)
				w.Flush()
			}

		}

	}

	elapsed := time.Since(start)
	logger.Info("Done in " + elapsed.String())

}

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

var XML_PRETTY_PREFIX = ""
var XML_PRETTY_INDENT = "    "

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
	var drop_uid bool
	var drop_user bool
	var drop_author bool

	var summarize bool
	var pretty bool

	var verbose bool
	var overwrite bool
	var dry_run bool
	var version bool
	var help bool

	flag.StringVar(&input_uri, "input_uri", "", "Input uri.  Supported file extensions: .osm, .osm.gz")
	flag.StringVar(&output_uri, "output_uri", "", "Output uri.  Supported file extensions: .osm, .osm.gz")
	flag.StringVar(&include_keys_text, "include_keys", "", "Comma-separated list of tag keys to keep")

	flag.BoolVar(&ways_to_nodes, "ways_to_nodes", false, "Convert ways into nodes for output")
	flag.BoolVar(&drop_relations, "drop_relations", false, "Drop relations from output")
	flag.BoolVar(&drop_version, "drop_version", false, "Drop version attribute from output")
	flag.BoolVar(&drop_timestamp, "drop_timestamp", false, "Drop timestamp attribute from output")
	flag.BoolVar(&drop_changeset, "drop_changeset", false, "Drop changeset attribute from output")

	flag.BoolVar(&drop_uid, "drop_uid", false, "Drop uid attribute from output")
	flag.BoolVar(&drop_user, "drop_user", false, "Drop user attribute from output")
	flag.BoolVar(&drop_author, "drop_author", false, "Drop author.  Synonymous to drop_uid and drop_user")

	flag.BoolVar(&summarize, "summarize", false, "Print data summary to stdout (bounding box, number of nodes, number of ways, and number of relations)")
	flag.BoolVar(&pretty, "pretty", false, "Pretty output.  Adds indents.")

	flag.BoolVar(&verbose, "verbose", false, "Provide verbose output")
	flag.BoolVar(&overwrite, "overwrite", false, "Overwrite output file.")
	flag.BoolVar(&dry_run, "dry_run", false, "Test user input but do not execute.")
	flag.BoolVar(&version, "version", false, "Prints version to stdout")
	flag.BoolVar(&help, "help", false, "Print help")

	flag.Parse()

	include_keys := make([]string, 0)
	if len(include_keys_text) > 0 {
		include_keys = strings.Split(include_keys_text, ",")
	}

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

	output_path := ""
	output_path_expanded := ""
	output_exists := false
	if len(output_uri) > 0 && output_uri != "stdout" && output_uri != "stdin" {

		if strings.HasPrefix(output_uri, "file://") {
			output_path = output_uri[7:]
		} else {
			output_path = output_uri
		}

		p, err := homedir.Expand(output_path)
		if err != nil {
			fmt.Println("Error expanding output path")
			os.Exit(1)
		}
		output_path_expanded = p

		if _, err := os.Stat(output_path_expanded); os.IsNotExist(err) {
			output_exists = false
		} else {
			output_exists = true
		}

		if output_exists {
			if !overwrite {
				fmt.Println("Output file already exists at " + output_uri + ".")
				fmt.Println("If you'd like to overwrite this file, then set the overwrite command line flag.")
				fmt.Println("Run \"osm --help\" for more information.")
				os.Exit(1)
			} else if verbose {
				fmt.Println("File already exists at " + output_uri + ".")
			}
		}

	}

	if drop_author {
		drop_uid = true
		drop_user = true
	}

	if dry_run {
		os.Exit(0)
	}

	if len(output_uri) > 0 && output_uri != "stdout" && output_uri != "stdin" && output_exists && overwrite {
		err := os.Remove(output_path_expanded)
		if err != nil {
			fmt.Println("Error deleting existing file at output location " + output_uri + ".")
			fmt.Println(err)
			os.Exit(1)
		}
		if verbose {
			fmt.Println("Deleted existing file at " + output_uri + ".")
		}
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

		input_path := ""
		if strings.HasPrefix(input_uri, "file://") {
			input_path = input_uri[7:]
		} else {
			input_path = input_uri
		}

		input_path_expanded, err := homedir.Expand(input_path)
		if err != nil {
			fmt.Println("Error expanding path")
			os.Exit(1)
		}

		if strings.HasSuffix(input_path_expanded, ".osm.gz") || strings.HasSuffix(input_path_expanded, ".xml.gz") {

			input_file, err := os.Open(input_path_expanded)
			if err != nil {
				fmt.Println("Error opening input file at " + input_uri + ".")
				fmt.Println(err)
				os.Exit(1)
			}
			defer input_file.Close()

			gr, err := gzip.NewReader(input_file)
			if err != nil {
				fmt.Println("Error creating gzip reader for file at " + input_uri + ".")
				fmt.Println(err)
				os.Exit(1)
			}
			defer gr.Close()

			in, err := ioutil.ReadAll(gr)
			if err != nil {
				fmt.Println("Error reading from gzip file at " + input_uri + ".")
				fmt.Println(err)
				os.Exit(1)
			}
			input_bytes = []byte(strings.TrimSpace(string(in)))

		} else if strings.HasSuffix(input_path_expanded, ".osm") || strings.HasSuffix(input_path_expanded, ".xml") {

			in, err := ioutil.ReadFile(input_path_expanded)
			if err != nil {
				fmt.Println("Error reading from uri  " + input_uri + ".")
				fmt.Println(err)
				os.Exit(1)
			}
			input_bytes = []byte(strings.TrimSpace(string(in)))

		} else if strings.HasSuffix(input_path_expanded, ".osm.pbf") || strings.HasSuffix(input_path_expanded, ".xml.pbf") {
			fmt.Println("The OSM PBF format is not supported yet.")
			os.Exit(1)
		} else if strings.HasSuffix(input_path_expanded, ".o5m") {
			fmt.Println("The o5m format is not supported yet.")
			os.Exit(1)
		} else {
			fmt.Println("Unknown file extension for input at " + input_uri + ".")
			os.Exit(1)
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

	planet.DropAttributes(drop_version, drop_timestamp, drop_changeset, drop_uid, drop_user)

	if summarize {
		fmt.Println("Bounding Box:", planet.BoundingBox())
		fmt.Println("Number of Nodes:", len(planet.Nodes))
		fmt.Println("Number of Ways:", len(planet.Ways))
		fmt.Println("Number of Relations:", len(planet.Relations))
	}

	if len(output_uri) > 0 {

		output_bytes := make([]byte, 0)
		if pretty {
			output_bytes, err = xml.MarshalIndent(&planet, XML_PRETTY_PREFIX, XML_PRETTY_INDENT)
		} else {
			output_bytes, err = xml.Marshal(&planet)
		}

		if err != nil {
			fmt.Println("Error marshalling output")
			fmt.Println(err)
			os.Exit(1)
		}
		//output_text := xml.Header + string(output_bytes)

		if output_uri == "stdout" {
			fmt.Println(xml.Header)
			fmt.Println(string(output_bytes))
		} else if output_uri == "stderr" {
			fmt.Fprintf(os.Stderr, xml.Header)
			fmt.Fprintf(os.Stderr, string(output_bytes))
		} else {
			if verbose {
				fmt.Println("Writing to " + output_uri + ".")
			}

			output_file, err := os.OpenFile(output_path_expanded, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
			if err != nil {
				os.Exit(1)
			}

			if strings.HasSuffix(output_path_expanded, ".gz") {
				gw := gzip.NewWriter(output_file)
				w := bufio.NewWriter(gw)
				_, err := w.WriteString(xml.Header)
				if err != nil {
					fmt.Println("Error writing XML Header to gzip file at " + output_uri + ".")
					os.Exit(1)
				}
				_, err = w.Write(output_bytes)
				if err != nil {
					fmt.Println("Error writing string to gzip file at " + output_uri + ".")
					os.Exit(1)
				}
				_, err = w.WriteString("\n")
				if err != nil {
					fmt.Println("Error writing last newline to gzip file at " + output_uri + ".")
					os.Exit(1)
				}
				err = w.Flush()
				if err != nil {
					fmt.Println("Error flushing output to bufio writer at " + output_uri + ".")
					os.Exit(1)
				}
				err = gw.Flush()
				if err != nil {
					fmt.Println("Error flushing output to gzip writer at " + output_uri + ".")
					os.Exit(1)
				}

				err = gw.Close()
				if err != nil {
					fmt.Println("Error closing gzip writer")
					os.Exit(1)
				}
				err = output_file.Close()
				if err != nil {
					fmt.Println("Error closing file writer")
					os.Exit(1)
				}

			} else {
				defer output_file.Close()
				w := bufio.NewWriter(output_file)
				_, err := w.WriteString(xml.Header)
				if err != nil {
					fmt.Println("Error writing XML Header to file at " + output_uri + ".")
					os.Exit(1)
				}
				_, err = w.Write(output_bytes)
				if err != nil {
					fmt.Println("Error writing string to file at " + output_uri + ".")
					os.Exit(1)
				}
				_, err = w.WriteString("\n")
				if err != nil {
					fmt.Println("Error writing last newline to file at " + output_uri + ".")
					os.Exit(1)
				}
				w.Flush()
				if err != nil {
					fmt.Println("Error flushing output to bufio writer at " + output_uri + ".")
					os.Exit(1)
				}
			}

		}

	}

	elapsed := time.Since(start)
	logger.Info("Done in " + elapsed.String())

}

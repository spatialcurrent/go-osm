package main

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io"
	//"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"
)

import (
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	//"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

import (
	"github.com/spatialcurrent/go-composite-logger/compositelogger"
	"github.com/spatialcurrent/go-dfl/dfl"
)

import (
	"github.com/spatialcurrent/go-osm/osm"
	"github.com/spatialcurrent/go-osm/s3util"
	"github.com/spatialcurrent/go-osm/xmlutil"
)

var GO_OSM_VERSION = "0.0.2"

var SUPPORTED_SCHEMES = []string{
	"file",
	"http",
	"https",
	"s3",
}

var XML_PRETTY_PREFIX = ""
var XML_PRETTY_INDENT = "    "

func parse_uri(uri string, schemes []string) (string, string) {
	for _, scheme := range schemes {
		if strings.HasPrefix(strings.ToLower(uri), scheme+"://") {
			return scheme, uri[len(scheme+"://"):]
		}
	}
	return "file", uri
}

func connect_to_aws(aws_access_key_id string, aws_secret_access_key string, aws_region string) *session.Session {
	aws_session := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Credentials: credentials.NewStaticCredentials(aws_access_key_id, aws_secret_access_key, ""),
			MaxRetries:  aws.Int(3),
			Region:      aws.String(aws_region),
		},
	}))
	return aws_session
}

func dfl_build_funcs() dfl.FunctionMap {
	funcs := dfl.FunctionMap{}

	funcs["len"] = func(ctx dfl.Context, args []string) (interface{}, error) {
		if len(args) != 1 {
			return 0, errors.New("Invalid number of arguments to len.")
		}
		return len(args[0]), nil
	}

	return funcs
}

func parse_slice_string(in string) []string {
	out := make([]string, 0)
	if len(in) > 0 {
		out = strings.Split(in, ",")
	}
	return out
}

func parse_slice_float64(in string) ([]float64, error) {
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

func parse_bool(in string) bool {
	return in == "yes" || in == "true" || in == "y" || in == "1" || in == "t"
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	start := time.Now()

	var aws_default_region string
	var aws_access_key_id string
	var aws_secret_access_key string

	var input_uri string

	var ini_uri string

	var filter_keys_keep_text string
	var filter_keys_drop_text string

	var filter_dfl_use_cache bool
	var filter_dfl_exp_text string

	var ways_to_nodes bool

	var bbox_text string

	// ---------------------------------------------------------
	// Output flags
	var output_uri string
	var drop_text string
	var drop_ways bool
	var drop_relations bool
	var drop_version bool
	var drop_timestamp bool
	var drop_changeset bool
	var drop_uid bool
	var drop_user bool
	var drop_author bool
	var output_keys_keep_text string
	var output_keys_drop_text string
	// ---------------------------------------------------------

	var summarize bool
	var summarize_keys_text string

	var pretty bool
	var stream bool
	var async bool

	var read_buffer_size int

	var profile bool
	var verbose bool
	var overwrite bool
	var dry_run bool
	var version bool
	var help bool

	flag.StringVar(&aws_default_region, "aws_default_region", "", "Defaults to value of environment variable AWS_DEFAULT_REGION.")
	flag.StringVar(&aws_access_key_id, "aws_access_key_id", "", "Defaults to value of environment variable AWS_ACCESS_KEY_ID")
	flag.StringVar(&aws_secret_access_key, "aws_secret_access_key", "", "Defaults to value of environment variable AWS_SECRET_ACCESS_KEY.")

	if len(aws_default_region) == 0 {
		aws_default_region = os.Getenv("AWS_DEFAULT_REGION")
	}
	if len(aws_access_key_id) == 0 {
		aws_access_key_id = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	if len(aws_secret_access_key) == 0 {
		aws_secret_access_key = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}

	// Input Flags
	flag.StringVar(&input_uri, "input_uri", "", "Input uri.  \"stdin\" or uri to input file.")
	flag.StringVar(&ini_uri, "ini_uri", "", "Uri to ini file.")

	// Filter Flags
	flag.StringVar(&filter_keys_keep_text, "filter_keys_keep", "", "Only keep nodes or ways that have a key in the provided comma-separated list of keys")
	flag.StringVar(&filter_keys_drop_text, "filter_keys_drop", "", "Drop nodes or ways that have a key in the provided comma-separated list of keys")

	flag.BoolVar(&filter_dfl_use_cache, "filter_dfl_cache", false, "Use cache for DFL results.   Use wisely.  Can increase performance.")
	flag.StringVar(&filter_dfl_exp_text, "filter_dfl_exp", "", "DFL filter expression")

	flag.StringVar(&bbox_text, "bbox", "", "Filter by bounding box (minx,miny,maxx,maxy)")

	flag.BoolVar(&ways_to_nodes, "ways_to_nodes", false, "Convert ways into nodes for output")

	flag.StringVar(&drop_text, "drop", "", "Convenience flag.  A comma-separated list of features or attributes to drop: ways, relations, version, timestamp, changeset, uid, user, author")

	flag.BoolVar(&drop_ways, "drop_ways", false, "Drop ways from output")
	flag.BoolVar(&drop_relations, "drop_relations", false, "Drop relations from output")

	flag.BoolVar(&drop_version, "drop_version", false, "Drop version attribute from output")
	flag.BoolVar(&drop_timestamp, "drop_timestamp", false, "Drop timestamp attribute from output")
	flag.BoolVar(&drop_changeset, "drop_changeset", false, "Drop changeset attribute from output")
	flag.BoolVar(&drop_uid, "drop_uid", false, "Drop uid attribute from output")
	flag.BoolVar(&drop_user, "drop_user", false, "Drop user attribute from output")
	flag.BoolVar(&drop_author, "drop_author", false, "Drop author.  Synonymous to drop_uid and drop_user")

	// Output Flags
	flag.StringVar(&output_uri, "output_uri", "", "Output uri. \"stdout\", \"stderr\", or uri to output file.")
	flag.StringVar(&output_keys_keep_text, "output_keys_keep", "", "Comma-separated list of tag keys to keep in output.  Drop all other keys.")
	flag.StringVar(&output_keys_drop_text, "output_keys_drop", "", "Comma-separated list of keys to drop in output.  Keep everything else.")

	flag.BoolVar(&summarize, "summarize", false, "Print data summary to stdout (bounding box, number of nodes, number of ways, and number of relations)")
	flag.StringVar(&summarize_keys_text, "summarize_keys", "", "Comma-separated list of keys to summarize")
	flag.BoolVar(&pretty, "pretty", false, "Pretty output.  Adds indents.")

	flag.BoolVar(&stream, "stream", false, "Stream input.")
	flag.BoolVar(&async, "async", false, "Process input using async functions.")

	flag.IntVar(&read_buffer_size, "read_buffer_size", 4096, "Size of buffer when reading files from disk")

	flag.BoolVar(&profile, "profile", false, "Profile performance")
	flag.BoolVar(&verbose, "verbose", false, "Provide verbose output")
	flag.BoolVar(&overwrite, "overwrite", false, "Overwrite output file.")
	flag.BoolVar(&dry_run, "dry_run", false, "Test user input but do not execute.")
	flag.BoolVar(&version, "version", false, "Prints version to stdout")
	flag.BoolVar(&help, "help", false, "Print help")

	flag.Parse()

	drop := parse_slice_string(drop_text)
	filter_keys_keep := parse_slice_string(filter_keys_keep_text)
	filter_keys_drop := parse_slice_string(filter_keys_drop_text)

	drop_ways = drop_ways || stringSliceContains(drop, "ways")
	drop_relations = drop_relations || stringSliceContains(drop, "relations")
	drop_timestamp = drop_timestamp || stringSliceContains(drop, "timestamp")
	drop_changeset = drop_changeset || stringSliceContains(drop, "changeset")
	drop_version = drop_version || stringSliceContains(drop, "version")
	drop_author = drop_author || stringSliceContains(drop, "author")
	drop_uid = drop_uid || stringSliceContains(drop, "uid")
	drop_user = drop_user || stringSliceContains(drop, "user")

	if len(filter_keys_keep) > 0 && len(filter_keys_drop) > 0 {
		fmt.Println("-filter_keys_keep (" + filter_keys_keep_text + ") and -filter_keys_drop (" + filter_keys_drop_text + ") are mutually exclusive")
		os.Exit(1)
	}

	bbox, err := parse_slice_float64(bbox_text)
	if err != nil {
		fmt.Println("Invalid bounding box " + bbox_text)
		os.Exit(1)
	}

	if len(bbox) != 0 && len(bbox) != 4 {
		fmt.Println("Invalid length of bounding box " + bbox_text)
		os.Exit(1)
	}

	if drop_author {
		drop_uid = true
		drop_user = true
	}

	outputConfig := osm.Output{
		DropWays:      drop_ways,
		DropRelations: drop_relations,
		DropVersion:   drop_version,
		DropChangeset: drop_changeset,
		DropTimestamp: drop_timestamp,
		DropUserId:    drop_uid,
		DropUserName:  drop_user,
		KeysToKeep:    []string{},
		KeysToDrop:    []string{},
		Pretty:        pretty,
	}

	if len(ini_uri) > 0 {
		ini_scheme, ini_path := parse_uri(ini_uri, []string{"file"})
		if ini_scheme == "file" {
			ini_path_expanded, err := homedir.Expand(ini_path)
			if err != nil {
				fmt.Println("Error expanding ini path " + ini_uri)
				os.Exit(1)
			}
			cfg, err := ini.Load(ini_path_expanded)
			if err != nil {
				fmt.Printf("Fail to read file: %v", err)
				os.Exit(1)
			}
			outputConfig.DropVersion = !parse_bool(cfg.Section("points").Key("osm_version").String())
			outputConfig.DropChangeset = !parse_bool(cfg.Section("points").Key("osm_changeset").String())
			outputConfig.DropTimestamp = !parse_bool(cfg.Section("points").Key("osm_timestamp").String())
			outputConfig.DropUserId = !parse_bool(cfg.Section("points").Key("osm_id").String())
			outputConfig.DropUserName = !parse_bool(cfg.Section("points").Key("osm_user").String())
			outputConfig.KeysToKeep = parse_slice_string(cfg.Section("points").Key("attributes").String())
		}
	}

	// Parse Output Flags
	if len(output_keys_keep_text) > 0 {
		outputConfig.KeysToKeep = parse_slice_string(output_keys_keep_text)
	}
	if len(output_keys_drop_text) > 0 {
		outputConfig.KeysToDrop = parse_slice_string(output_keys_drop_text)
	}

	if len(outputConfig.KeysToKeep) > 0 && len(outputConfig.KeysToDrop) > 0 {
		fmt.Println("-output_keys_keep (" + output_keys_keep_text + ") and -output_keys_drop (" + output_keys_drop_text + ") are mutually exclusive")
		os.Exit(1)
	}

	// Parse Summarize Flags
	summarize_keys := parse_slice_string(summarize_keys_text)

	if help {
		fmt.Println("Usage: osm -input_uri INPUT -output_uri OUTPUT [-verbose] [-dry_run] [-version] [-help]")
		fmt.Println("Supported Schemes: " + strings.Join(SUPPORTED_SCHEMES, ", "))
		fmt.Println("Supported Input File Extensions: .osm, .osm.gz, .osm.bz2")
		fmt.Println("Supported Output File Extensions: .osm, .osm.gz, .geojson, .geojson.gz")
		fmt.Println("Options:")
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

	if len(input_uri) == 0 {
		fmt.Println("Error: input_uri or version option is required.")
		fmt.Println("Run \"osm --help\" for more information.")
		os.Exit(1)
	}

	if ways_to_nodes && drop_ways {
		fmt.Println("Error: cannot enable ways_to_nodes and drop_ways at the same time.")
		os.Exit(1)
	}

	output_scheme := "" // stdin, stdout, stderr, file, http, https, s3
	output_path := ""

	if len(output_uri) > 0 {
		if output_uri == "stdout" {
			output_scheme = "stdout"
		} else if output_uri == "stderr" {
			output_scheme = "stderr"
		} else {
			output_scheme, output_path = parse_uri(output_uri, SUPPORTED_SCHEMES)
		}
	}

	output_path_expanded := ""
	output_exists := false

	var aws_session *session.Session
	var s3_client *s3.S3
	output_s3_bucket := ""
	output_s3_key := ""

	if output_scheme == "file" {

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

	} else if output_scheme == "s3" {

		aws_session = connect_to_aws(aws_access_key_id, aws_secret_access_key, aws_default_region)
		s3_client = s3.New(aws_session)
		b, k, err := s3util.ParsePath(output_path)
		if err != nil {
			fmt.Println("Error parsing AWS S3 path")
			fmt.Println(err)
			os.Exit(1)
		}
		output_s3_bucket = b
		output_s3_key = k
		output_exists = s3util.ObjectExists(s3_client, output_s3_bucket, output_s3_key)

	} else {
		output_path_expanded = output_path
	}

	if output_exists {
		if !overwrite {
			fmt.Println("Output file already exists at output location " + output_uri + ".")
			fmt.Println("If you'd like to overwrite this file, then set the overwrite command line flag.")
			fmt.Println("Run \"osm --help\" for more information.")
			os.Exit(1)
		} else if verbose {
			fmt.Println("File already exists at output location " + output_uri + ".")
		}
	}

	if dry_run {
		os.Exit(0)
	}

	if output_exists && overwrite {
		if output_scheme == "file" {
			err := os.Remove(output_path_expanded)
			if err != nil {
				fmt.Println("Error deleting existing file at output location " + output_uri + ".")
				fmt.Println(err)
				os.Exit(1)
			}
			if verbose {
				fmt.Println("Deleted existing file at output location " + output_uri + ".")
			}
		} else if output_scheme == "s3" {
			err := s3util.DeleteObject(s3_client, output_s3_bucket, output_s3_key)
			if err != nil {
				fmt.Println("Error deleting existing object on AWS S3 at output location " + output_uri + ".")
				fmt.Println(err)
				os.Exit(1)
			}
			if verbose {
				fmt.Println("Deleted existing object on AWS S3 at output location " + output_uri + ".")
			}
		}
	}

	if !output_exists && output_scheme == "s3" {
		if !s3util.BucketExists(s3_client, output_s3_bucket) {
			err := s3util.CreateBucket(s3_client, aws_default_region, output_s3_bucket)
			if err != nil {
				fmt.Println("Error creating AWS S3 bucket.")
				os.Exit(1)
			}
		}
	}

	logger, err := compositelogger.NewDefaultLogger()
	if err != nil {
		fmt.Println("Error initializing composite logger.")
		fmt.Println(err)
		os.Exit(1)
	}

	var filter_dfl_exp dfl.Node
	if len(filter_dfl_exp_text) > 0 {
		filter_dfl_exp, err = dfl.Parse(filter_dfl_exp_text)
		if err != nil {
			fmt.Println("Error parsing DFL filter expression", filter_dfl_exp_text)
			fmt.Println(err)
			os.Exit(1)
		}
		filter_dfl_exp = filter_dfl_exp.Compile()
	}

	funcs := dfl_build_funcs()

	start_read := time.Now()

	var input_file *os.File
	var input_reader io.Reader
	if input_uri == "stdin" {
		input_reader = bufio.NewReader(os.Stdin)
	} else {

		input_scheme, input_path := parse_uri(input_uri, SUPPORTED_SCHEMES)

		if input_scheme == "file" {

			input_path_expanded, err := homedir.Expand(input_path)
			if err != nil {
				fmt.Println("Error expanding path")
				os.Exit(1)
			}

			if strings.HasSuffix(input_path_expanded, ".osm.gz") {

				f, err := os.Open(input_path_expanded)
				if err != nil {
					fmt.Println("Error opening input file at " + input_uri + ".")
					fmt.Println(err)
					os.Exit(1)
				}
				input_file = f

				gr, err := gzip.NewReader(input_file)
				if err != nil {
					fmt.Println("Error creating gzip reader for file at " + input_uri + ".")
					fmt.Println(err)
					os.Exit(1)
				}
				input_reader = gr

			} else if strings.HasSuffix(input_path_expanded, ".osm.bz2") {

				f, err := os.Open(input_path_expanded)
				if err != nil {
					fmt.Println("Error opening input file at " + input_uri + ".")
					fmt.Println(err)
					os.Exit(1)
				}
				input_file = f
				input_reader = bzip2.NewReader(bufio.NewReaderSize(input_file, read_buffer_size))

			} else if strings.HasSuffix(input_path_expanded, ".osm") {

				f, err := os.Open(input_path_expanded)
				if err != nil {
					fmt.Println("Error opening input file at " + input_uri + ".")
					fmt.Println(err)
					os.Exit(1)
				}
				input_file = f
				input_reader = bufio.NewReaderSize(input_file, read_buffer_size)

			} else if strings.HasSuffix(input_path_expanded, ".osm.pbf") {
				fmt.Println("The OSM PBF format is not supported yet.")
				os.Exit(1)
			} else if strings.HasSuffix(input_path_expanded, ".o5m") {
				fmt.Println("The o5m format is not supported yet.")
				os.Exit(1)
			} else {
				fmt.Println("Unknown file extension for input at " + input_uri + ".")
				os.Exit(1)
			}

		} else if input_scheme == "s3" {

			if s3_client == nil {
				if aws_session == nil {
					aws_session = connect_to_aws(aws_access_key_id, aws_secret_access_key, aws_default_region)
				}
				s3_client = s3.New(aws_session)
			}

			input_s3_bucket, input_s3_key, err := s3util.ParsePath(input_path)
			if err != nil {
				fmt.Println("Error parsing AWS S3 path")
				fmt.Println(err)
				os.Exit(1)
			}

			if strings.HasSuffix(input_s3_key, ".osm.gz") || strings.HasSuffix(input_s3_key, ".osm.bz2") || strings.HasSuffix(input_s3_key, ".osm") {

				in, err := s3util.GetObject(s3_client, input_s3_bucket, input_s3_key)
				if err != nil {
					fmt.Println("Error reading from AWS S3 uri " + input_uri + ".")
					fmt.Println(err)
					os.Exit(1)
				}
				input_reader = bytes.NewReader(in)

			} else if strings.HasSuffix(input_s3_key, ".osm.pbf") {
				fmt.Println("The OSM PBF format is not supported yet.")
				os.Exit(1)
			} else if strings.HasSuffix(input_s3_key, ".o5m") {
				fmt.Println("The o5m format is not supported yet.")
				os.Exit(1)
			} else {
				fmt.Println("Unknown file extension for input at " + input_uri + ".")
				os.Exit(1)
			}

		}

	}

	if profile {
		logger.Info("Opened file " + input_uri + " in " + time.Since(start_read).String())
	}

	planet := osm.NewPlanet()
	if strings.HasSuffix(input_uri, ".pbf") {
		fmt.Println("Protobuf not implemented yet.")
		os.Exit(1)
	} else {

		if verbose {
			logger.Info("Unmarshalling planet file")
		}

		fi := osm.NewFilterInput(filter_keys_keep, filter_keys_drop, filter_dfl_exp, funcs, filter_dfl_use_cache, bbox)

		start_unmarshal := time.Now()
		err = osm.UnmarshalPlanet(
			planet,
			input_reader,
			stream,
			outputConfig,
			fi,
			ways_to_nodes)
		if err != nil {
			logger.Warn(errors.Wrap(err, "Error unmarhsalling input"))
			os.Exit(1)
		}
		if input_file != nil {
			input_file.Close()
		}

		if profile {
			logger.Info("Unmarshalled in " + time.Since(start_unmarshal).String())
		}

	}

	if summarize {
		summary := planet.Summarize(summarize_keys, async)
		summary.Print()
	}

	if len(output_uri) > 0 {

		if output_uri == "stdout" || output_uri == "stderr" || strings.HasSuffix(output_path, ".osm") {
			start_marshal := time.Now()
			err := osm.MarshalPlanet(output_uri, output_scheme, output_path, outputConfig, planet)
			if err != nil {
				logger.Warn(errors.Wrap(err, "Error marshalling planet."))
				os.Exit(1)
			}
			if profile {
				logger.Info("Marshalled in " + time.Since(start_marshal).String())
			}
		} else if strings.HasSuffix(output_path, ".osm.gz") {
			output_bytes := make([]byte, 0)
			if pretty {
				output_bytes, err = xml.MarshalIndent(planet, XML_PRETTY_PREFIX, XML_PRETTY_INDENT)
			} else {
				output_bytes, err = xml.Marshal(planet)
			}

			if err != nil {
				logger.Warn(errors.Wrap(err, "Error marshalling output"))
				os.Exit(1)
			}

			if verbose {
				logger.Info("Writing to " + output_uri + ".")
			}

			err := xmlutil.WriteBytes(output_uri, output_scheme, output_path_expanded, output_bytes, s3_client, output_s3_bucket, output_s3_key)
			if err != nil {
				logger.Warn(errors.Wrap(err, "Error writing xml to output"))
				os.Exit(1)
			}
		} else if strings.HasSuffix(output_path, ".geojson.gz") || strings.HasSuffix(output_path, ".geojson") {

			output_bytes, err := json.Marshal(planet.FeatureCollection())
			if err != nil {
				logger.Warn(errors.Wrap(err, "Could not marshal feature collection as response"))
				os.Exit(1)
			}

			if output_uri == "stdout" {
				fmt.Println(string(output_bytes))
			} else if output_uri == "stderr" {
				fmt.Fprintf(os.Stderr, string(output_bytes))
			} else if output_scheme == "s3" {

			} else if output_scheme == "file" {

				output_file, err := os.OpenFile(output_path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
				if err != nil {
					logger.Warn(errors.Wrap(err, "error opening file to write GeoJSON to disk"))
					os.Exit(1)
				}
				w := bufio.NewWriter(output_file)

				_, err = w.Write(output_bytes)
				if err != nil {
					logger.Warn(errors.Wrap(err, "Error writing string to GeoJSON file"))
					os.Exit(1)
				}

				_, err = w.WriteString("\n")
				if err != nil {
					logger.Warn(errors.Wrap(err, "Error writing last newline to GeoJSON file"))
					os.Exit(1)
				}

				w.Flush()
				if err != nil {
					logger.Warn(errors.Wrap(err, "Error flushing output to bufio writer for GeoJSON file"))
					os.Exit(1)
				}

				err = output_file.Close()
				if err != nil {
					logger.Warn(errors.Wrap(err, "Error closing file writer for GeoJSON file."))
					os.Exit(1)
				}
			}

		}
	}

	elapsed := time.Since(start)
	logger.Info("Done in " + elapsed.String())

}

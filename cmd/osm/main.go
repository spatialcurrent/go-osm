package main

import (
	"bufio"
	//"bytes"
	//"compress/bzip2"
	//"compress/gzip"
	"encoding/json"
	//"encoding/xml"
	"flag"
	"fmt"
	//"io"
	//"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

import (
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	//"gopkg.in/ini.v1"
	"github.com/colinmarc/hdfs"
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
	//"github.com/spatialcurrent/go-osm/xmlutil"
)

var GO_OSM_VERSION = "0.0.3"

var XML_PRETTY_PREFIX = ""
var XML_PRETTY_INDENT = "    "

var GDAL_INI_KEYS = []string{"osm_version", "osm_changeset", "osm_timestamp", "osm_id", "osm_user", "osm_attributes"}

type Message struct {
	Message string
	Fields  map[string]interface{}
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

func dfl_build_funcs() *dfl.FunctionMap {
	funcs := dfl.FunctionMap{}

	funcs["len"] = func(ctx dfl.Context, args []string) (interface{}, error) {
		if len(args) != 1 {
			return 0, errors.New("Invalid number of arguments to len.")
		}
		return len(args[0]), nil
	}

	return &funcs
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	start := time.Now()

	var aws_default_region string
	var aws_access_key_id string
	var aws_secret_access_key string

	var config_uri string

	var input_uri_text string

	var gdal_ini_uri string
	var gdal_ini_section string

	var filter_keys_keep_text string
	var filter_keys_drop_text string

	var filter_dfl_use_cache bool
	var filter_dfl_exp_text string

	var ways_to_nodes bool

	var bbox_text string

	// ---------------------------------------------------------
	// Output flags
	var output_uri_text string
	var drop_text string
	var drop_nodes bool
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

	// Config Flags
	flag.StringVar(&config_uri, "config_uri", "", "Config uri.  Uri to config file.  If given, ignores most command line flags.  Defaults to value of environment variable GO_OSM_CONFIG_URI.")

	// Input Flags
	flag.StringVar(&input_uri_text, "input_uri", "", "A single or colon-separated list of input uris.  Supports wildcards.  \"stdin\" or uri to input file.")
	flag.StringVar(&gdal_ini_uri, "gdal_ini_uri", "", "Uri to GDAL ini file for convience.  See http://www.gdal.org/drv_osm.html.")
	flag.StringVar(&gdal_ini_section, "gdal_ini_section", "points", "Section to parse in GDAL in file.  See http://www.gdal.org/drv_osm.html.")

	// Filter Flags
	flag.StringVar(&filter_keys_keep_text, "filter_keys_keep", "", "Only keep nodes or ways that have a key in the provided comma-separated list of keys")
	flag.StringVar(&filter_keys_drop_text, "filter_keys_drop", "", "Drop nodes or ways that have a key in the provided comma-separated list of keys")

	flag.BoolVar(&filter_dfl_use_cache, "filter_dfl_cache", false, "Use cache for DFL results.   Use wisely.  Can increase performance.")
	flag.StringVar(&filter_dfl_exp_text, "filter_dfl_exp", "", "DFL filter expression")

	flag.StringVar(&bbox_text, "bbox", "", "Filter by bounding box (minx,miny,maxx,maxy)")

	flag.BoolVar(&ways_to_nodes, "ways_to_nodes", false, "Convert ways into nodes for output")

	flag.StringVar(&drop_text, "drop", "", "Convenience flag.  A comma-separated list of features or attributes to drop: ways, relations, version, timestamp, changeset, uid, user, author")

	flag.BoolVar(&drop_nodes, "drop_nodes", false, "Drop nodes from output")
	flag.BoolVar(&drop_ways, "drop_ways", false, "Drop ways from output")
	flag.BoolVar(&drop_relations, "drop_relations", false, "Drop relations from output")

	flag.BoolVar(&drop_version, "drop_version", false, "Drop version attribute from output")
	flag.BoolVar(&drop_timestamp, "drop_timestamp", false, "Drop timestamp attribute from output")
	flag.BoolVar(&drop_changeset, "drop_changeset", false, "Drop changeset attribute from output")
	flag.BoolVar(&drop_uid, "drop_uid", false, "Drop uid attribute from output")
	flag.BoolVar(&drop_user, "drop_user", false, "Drop user attribute from output")
	flag.BoolVar(&drop_author, "drop_author", false, "Drop author.  Synonymous to drop_uid and drop_user")

	// Output Flags
	flag.StringVar(&output_uri_text, "output_uri", "", "A single or colon-separated list of uutput uris. \"stdout\", \"stderr\", or uri to output file.")
	flag.StringVar(&output_keys_keep_text, "output_keys_keep", "", "Comma-separated list of tag keys to keep in output.  Drop all other keys.")
	flag.StringVar(&output_keys_drop_text, "output_keys_drop", "", "Comma-separated list of keys to drop in output.  Keep everything else.")

	flag.BoolVar(&summarize, "summarize", false, "Print data summary to stdout (bounding box, number of nodes, number of ways, and number of relations)")
	flag.StringVar(&summarize_keys_text, "summarize_keys", "", "Comma-separated list of keys to summarize")
	flag.BoolVar(&pretty, "pretty", false, "Pretty output.  Adds indents.")

	flag.IntVar(&read_buffer_size, "read_buffer_size", 4096, "Size of buffer when reading files from disk")

	flag.BoolVar(&profile, "profile", false, "Profile performance")
	flag.BoolVar(&verbose, "verbose", false, "Provide verbose output")
	flag.BoolVar(&overwrite, "overwrite", false, "Overwrite output file.")
	flag.BoolVar(&dry_run, "dry_run", false, "Test user input but do not execute.")
	flag.BoolVar(&version, "version", false, "Prints version to stdout")
	flag.BoolVar(&help, "help", false, "Print help")

	flag.Parse()

	if help {
		fmt.Println("Usage: osm -input_uri INPUT[:INPUT_2][:INPUT_3] -output_uri OUTPUT [-verbose] [-dry_run] [-version] [-help] [A=1] [B=2]")
		fmt.Println("Supported Schemes: " + strings.Join(osm.SUPPORTED_SCHEMES, ", "))
		fmt.Println("Supported Input File Extensions: .osm, .osm.gz, .osm.bz2")
		fmt.Println("Supported Output File Extensions: .osm, .osm.gz, .geojson, .geojson.gz")
		fmt.Println("Options:")
		flag.PrintDefaults()
		os.Exit(0)
	} else if version {
		fmt.Println(GO_OSM_VERSION)
		os.Exit(0)
	} else if len(os.Args) == 1 {
		fmt.Println("Error: Provided no arguments.")
		fmt.Println("Run \"osm --help\" for more information.")
		os.Exit(0)
	}

	ctx := map[string]interface{}{}
	for _, a := range flag.Args() {
		if !strings.Contains(a, "=") {
			fmt.Println("Context attribute \"" + a + "\" does not contain \"=\".")
			os.Exit(1)
		}
		parts := strings.SplitN(a, "=", 2)
		ctx[parts[0]] = dfl.TryConvertString(parts[1])
	}

	if len(config_uri) == 0 {
		config_uri = os.Getenv("GO_OSM_CONFIG_URI")
	}

	if len(input_uri_text) == 0 {
		input_uri_text = os.Getenv("GO_OSM_INPUT_URI")
	}

	funcs := dfl_build_funcs()

	filter_keys_keep := osm.ParseSliceString(filter_keys_keep_text)
	filter_keys_drop := osm.ParseSliceString(filter_keys_drop_text)

	if len(filter_keys_keep) > 0 && len(filter_keys_drop) > 0 {
		fmt.Println("-filter_keys_keep (" + filter_keys_keep_text + ") and -filter_keys_drop (" + filter_keys_drop_text + ") are mutually exclusive")
		os.Exit(1)
	}

	drop := osm.ParseSliceString(drop_text)
	drop_nodes = drop_nodes || stringSliceContains(drop, "nodes")
	drop_ways = drop_ways || stringSliceContains(drop, "ways")
	drop_relations = drop_relations || stringSliceContains(drop, "relations")
	drop_timestamp = drop_timestamp || stringSliceContains(drop, "timestamp")
	drop_changeset = drop_changeset || stringSliceContains(drop, "changeset")
	drop_version = drop_version || stringSliceContains(drop, "version")
	drop_author = drop_author || stringSliceContains(drop, "author")
	drop_uid = drop_uid || stringSliceContains(drop, "uid")
	drop_user = drop_user || stringSliceContains(drop, "user")

	if drop_author {
		drop_uid = true
		drop_user = true
	}

	if drop_uid && !drop_user {
		fmt.Println("You cannot drop the user id but keep the user name.")
		os.Exit(1)
	}

	var config *osm.Config
	if len(config_uri) > 0 {

		c, err := osm.LoadConfig(config_uri)
		if err != nil {
			fmt.Println("Error loading config.")
			fmt.Println(err)
			os.Exit(1)
		}

		if len(input_uri_text) > 0 {
			input_configs := make([]osm.InputConfig, 0)
			input_uris := strings.Split(input_uri_text, ":")
			for _, input_uri := range input_uris {
				scheme, input_path_glob := osm.SplitUri(input_uri, osm.SUPPORTED_SCHEMES)
				if scheme == "file" && strings.Contains(input_path_glob, "*") {
					input_path_expanded, err := homedir.Expand(input_path_glob)
					if err != nil {
						fmt.Println(errors.Wrap(err, "Error expanding input file path"))
						os.Exit(1)
					}
					input_paths, err := filepath.Glob(input_path_expanded)
					if err != nil {
						fmt.Println(errors.Wrap(err, "Error globing input_uri "+input_uri))
						os.Exit(1)
					}
					for _, input_path := range input_paths {
						input_filter := osm.NewFilter(filter_keys_keep, filter_keys_drop, "", true, []float64{})
						input_config := osm.NewInputConfig(input_path, drop_nodes, drop_ways, drop_relations, input_filter)
						input_configs = append(input_configs, input_config)
					}
				} else {
					input_filter := osm.NewFilter(filter_keys_keep, filter_keys_drop, "", true, []float64{})
					input_config := osm.NewInputConfig(input_uri, drop_nodes, drop_ways, drop_relations, input_filter)
					input_configs = append(input_configs, input_config)
				}
			}
			c.InputConfigs = input_configs
		}

		config = c

	} else {

		bbox, err := osm.ParseSliceFloat64(bbox_text)
		if err != nil {
			fmt.Println("Invalid bounding box " + bbox_text)
			os.Exit(1)
		}

		if len(bbox) != 0 && len(bbox) != 4 {
			fmt.Println("Invalid length of bounding box " + bbox_text)
			os.Exit(1)
		}

		input_filter := osm.NewFilter(filter_keys_keep, filter_keys_drop, filter_dfl_exp_text, filter_dfl_use_cache, bbox)

		input_configs := make([]osm.InputConfig, 0)
		if len(input_uri_text) > 0 {
			input_uris := strings.Split(input_uri_text, ":")
			for _, input_uri := range input_uris {
				scheme, input_path_glob := osm.SplitUri(input_uri, osm.SUPPORTED_SCHEMES)
				if scheme == "file" && strings.Contains(input_path_glob, "*") {
					input_path_expanded, err := homedir.Expand(input_path_glob)
					if err != nil {
						fmt.Println(errors.Wrap(err, "Error expanding input file path"))
						os.Exit(1)
					}
					input_paths, err := filepath.Glob(input_path_expanded)
					if err != nil {
						fmt.Println(errors.Wrap(err, "Error globing input_uri "+input_uri))
						os.Exit(1)
					}
					for _, input_path := range input_paths {
						input_filter := osm.NewFilter(filter_keys_keep, filter_keys_drop, "", true, []float64{})
						input_config := osm.NewInputConfig(input_path, drop_nodes, drop_ways, drop_relations, input_filter)
						input_configs = append(input_configs, input_config)
					}
				} else {
					input_config := osm.NewInputConfig(input_uri, drop_nodes, drop_ways, drop_relations, input_filter)
					input_configs = append(input_configs, input_config)
				}
			}
		}

		output_configs := make([]osm.OutputConfig, 0)
		if len(output_uri_text) > 0 {
			output_uris := strings.Split(output_uri_text, ":")
			for _, output_uri := range output_uris {
				output_configs = append(output_configs, osm.NewOutputConfig(
					output_uri,
					input_filter,
					drop_ways,
					drop_nodes,
					drop_relations,
					drop_version,
					drop_changeset,
					drop_timestamp,
					drop_uid,
					drop_user,
					ways_to_nodes,
					pretty,
				))
			}
		}

		if len(gdal_ini_uri) > 0 {
			gdal_ini, err := osm.LoadIniSection(gdal_ini_uri, gdal_ini_section, GDAL_INI_KEYS)
			if err != nil {
				fmt.Println(gdal_ini)
				os.Exit(1)
			}
			for _, outputConfig := range output_configs {
				outputConfig.DropVersion = !osm.ParseBool(gdal_ini["osm_version"])
				outputConfig.DropChangeset = !osm.ParseBool(gdal_ini["osm_changeset"])
				outputConfig.DropTimestamp = !osm.ParseBool(gdal_ini["osm_timestamp"])
				outputConfig.DropUserId = !osm.ParseBool(gdal_ini["osm_uid"])
				outputConfig.DropUserName = !osm.ParseBool(gdal_ini["osm_user"])
				outputConfig.KeysToKeep = osm.ParseSliceString(gdal_ini["attributes"])
			}
		}

		for _, outputConfig := range output_configs {
			// Parse Output Flags
			if len(output_keys_keep_text) > 0 {
				outputConfig.KeysToKeep = osm.ParseSliceString(output_keys_keep_text)
			}
			if len(output_keys_drop_text) > 0 {
				outputConfig.KeysToDrop = osm.ParseSliceString(output_keys_drop_text)
			}

			if len(outputConfig.KeysToKeep) > 0 && len(outputConfig.KeysToDrop) > 0 {
				fmt.Println("-output_keys_keep (" + output_keys_keep_text + ") and -output_keys_drop (" + output_keys_drop_text + ") are mutually exclusive")
				os.Exit(1)
			}
		}

		config = &osm.Config{
			InputConfigs:  input_configs,
			OutputConfigs: output_configs,
		}

	}

	logger, err := compositelogger.NewDefaultLogger()
	if err != nil {
		fmt.Println("Error initializing composite logger.")
		fmt.Println(err)
		os.Exit(1)
	}

	// Initialize config, including input & output resources
	if verbose {
		logger.Info("Initializing...")
	}

	err = config.Init(ctx, funcs)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if verbose {

		logger.InfoWithFields("Config", map[string]interface{}{
			"Inputs":           len(config.Inputs),
			"Outputs":          len(config.Outputs),
			"AllWaysToNodes":   config.ConvertAllWaysToNodes,
			"DropAllNodes":     config.DropAllNodes,
			"DropAllWays":      config.DropAllWays,
			"DropAllRelations": config.DropAllRelations,
		})

		for i, input := range config.Inputs {
			fields := map[string]interface{}{"uri": input.Uri}
			if input.Filter != nil {
				fields["filter_keys_keep"] = strings.Join(input.Filter.KeysToKeep, ",")
			}
			logger.InfoWithFields("Input "+strconv.Itoa(i), fields)
		}
		for i, output := range config.Outputs {
			logger.InfoWithFields("Output "+strconv.Itoa(i), map[string]interface{}{"uri": output.Uri})
		}
	}

	// Parse Summarize Flags
	summarize_keys := osm.ParseSliceString(summarize_keys_text)

	err = config.Validate()
	if err != nil {
		fmt.Println(err)
		fmt.Println("Run \"osm --help\" for more information.")
		os.Exit(1)
	}

	var aws_session *session.Session
	var s3_client *s3.S3

	hdfs_clients := map[string]*hdfs.Client{}

	if config.HasResourceType("s3") {
		aws_session = connect_to_aws(aws_access_key_id, aws_secret_access_key, aws_default_region)
		s3_client = s3.New(aws_session)
	}

	if config.HasResourceType("hdfs") {
		for _, nameNode := range config.GetNameNodes() {
			hdfs_client, err := hdfs.New(nameNode)
			if err != nil {
				logger.Warn(errors.Wrap(err, "Could not connect to HDFS name node with domain "+nameNode+"."))
				os.Exit(1)
			}
			hdfs_clients[nameNode] = hdfs_client
		}
	}

	for _, input := range config.Inputs {

		switch input.GetType() {
		case "file":
			input.Exists = input.FileExists()
		case "hdfs":
			_, err := hdfs_clients[input.NameNode].Stat(input.Path)
			input.Exists = !os.IsNotExist(err)
		case "s3":
			input.Exists = s3util.ObjectExists(s3_client, input.Bucket, input.Key)
		}

		if !input.Exists {
			fmt.Println("Input at uri " + input.Uri + " does not exist.")
		}
	}

	for _, output := range config.Outputs {

		switch output.GetType() {
		case "file":
			output.Exists = output.FileExists()
		case "hdfs":
			_, err := hdfs_clients[output.NameNode].Stat(output.Path)
			output.Exists = !os.IsNotExist(err)
		case "s3":
			output.Exists = s3util.ObjectExists(s3_client, output.Bucket, output.Key)
		}

		if (!output.IsType("stream")) && output.Exists {
			if !overwrite {
				fmt.Println("Output file already exists at output location " + output.Uri + ".")
				fmt.Println("If you'd like to overwrite this file, then set the overwrite command line flag.")
				fmt.Println("Run \"osm --help\" for more information.")
				os.Exit(1)
			} else if verbose {
				fmt.Println("File already exists at output location " + output.Uri + ".")
			}
		}

	}

	if dry_run {
		os.Exit(0)
	}

	if overwrite {
		for _, output := range config.Outputs {
			if (!output.IsType("stream")) && output.Exists {
				switch output.GetType() {
				case "file":
					err := os.Remove(output.PathExpanded)
					if err != nil {
						fmt.Println("Error deleting existing file at output location " + output.Uri + ".")
						fmt.Println(err)
						os.Exit(1)
					}
					if verbose {
						fmt.Println("Deleted existing file at output location " + output.Uri + ".")
					}
				case "hdfs":
					err := hdfs_clients[output.NameNode].Remove(output.PathExpanded)
					if err != nil {
						fmt.Println(errors.Wrap(err, "Error deleting file on HDFS at uri "+output.Uri))
						os.Exit(1)
					}
				case "s3":
					err := s3util.DeleteObject(s3_client, output.Bucket, output.Key)
					if err != nil {
						fmt.Println(errors.Wrap(err, "Error deleting existing object on AWS S3 at output location "+output.Uri+"."))
						os.Exit(1)
					}
					if verbose {
						fmt.Println("Deleted existing object on AWS S3 at output location " + output.Uri + ".")
					}
				}
			}
		}
	}

	for _, output := range config.Outputs {
		if output.IsType("file") && !output.Exists {
			basepath := filepath.Dir(output.PathExpanded)
			if _, err := os.Stat(basepath); os.IsNotExist(err) {
				if verbose {
					logger.InfoWithFields("Creating parent directory for output.", map[string]interface{}{"path": basepath})
				}
				err := os.MkdirAll(basepath, os.ModePerm)
				if err != nil {
					fmt.Println(errors.Wrap(err, "Error creating parent directory for "+output.Uri))
					os.Exit(1)
				}
			}
		} else if output.IsType("hdfs") && !output.Exists {
			basepath := filepath.Dir(output.PathExpanded)
			if _, err := hdfs_clients[output.NameNode].Stat(output.PathExpanded); os.IsNotExist(err) {
				if verbose {
					logger.InfoWithFields("Creating parent directory for output.", map[string]interface{}{"path": basepath})
				}
				err := hdfs_clients[output.NameNode].MkdirAll(basepath, os.ModePerm)
				if err != nil {
					fmt.Println(errors.Wrap(err, "Error creating parent directory for "+output.Uri))
					os.Exit(1)
				}
			}
		} else if output.IsType("s3") && !output.Exists {
			if !s3util.BucketExists(s3_client, output.Bucket) {
				err := s3util.CreateBucket(s3_client, aws_default_region, output.Bucket)
				if err != nil {
					fmt.Println("Error creating AWS S3 bucket.")
					os.Exit(1)
				}
			}
		}
	}

	planet := osm.NewPlanet()

	err = planet.Init()
	if err != nil {
		fmt.Println(errors.Wrap(err, "Error initializing planet"))
		os.Exit(1)
	}

	for i, input := range config.Inputs {

		start_read := time.Now()

		err := input.Open(read_buffer_size, s3_client, hdfs_clients)
		if err != nil {
			fmt.Println(errors.Wrap(err, "Error opening input file at "+input.Uri))
			os.Exit(1)
		}

		if profile {
			logger.InfoWithFields("Opened input "+strconv.Itoa(i), map[string]interface{}{"uri": input.Uri, "duration": time.Since(start_read).String()})
		}

		if verbose {
			logger.InfoWithFields("Importing data from planet file", map[string]interface{}{
				"uri":            input.Uri,
				"drop_nodes":     input.DropNodes,
				"drop_ways":      input.DropWays,
				"drop_relations": input.DropRelations,
			})
		}

		start_unmarshal := time.Now()

		err = osm.UnmarshalPlanet(
			planet,
			input,
			logger)
		if err != nil {
			logger.Warn(errors.Wrap(err, "Error importing data from planet file at "+input.Uri))
			os.Exit(1)
		}

		err = input.Close()
		if err != nil {
			fmt.Println("Error closing input at uri " + input.Uri)
			os.Exit(1)
		}

		if profile {
			logger.InfoWithFields("Finished importing data from planet file", map[string]interface{}{"uri": input.Uri, "duration": time.Since(start_unmarshal).String()})
		}
	}

	if summarize {
		start_summarize := time.Now()
		summary := planet.Summarize(summarize_keys)
		summary.Print()
		if profile {
			logger.InfoWithFields("Finished summary ", map[string]interface{}{"duration": time.Since(start_summarize).String()})
		}
	}

	start_output := time.Now()
	ch := make(chan interface{})
	go func(ch chan interface{}) {
		for msg := range ch {
			switch msg.(type) {
			case error:
				logger.Warn(err)
			case Message:
				logger.InfoWithFields(msg.(Message).Message, msg.(Message).Fields)
			default:
				logger.Info(msg)
			}
		}
	}(ch)

	var wg sync.WaitGroup
	for i, o := range config.Outputs {
		wg.Add(1)
		go func(wg *sync.WaitGroup, planet *osm.Planet, output_id int, output *osm.Output, ch chan<- interface{}, verbose bool) {

			start_marshal := time.Now()

			if output.Uri == "stdout" || output.Uri == "stderr" || strings.HasSuffix(output.Path, ".osm") || strings.HasSuffix(output.Path, ".osm.gz") {
				err := osm.MarshalPlanet(output, config, planet)
				if err != nil {
					ch <- errors.Wrap(err, "Output "+strconv.Itoa(output_id)+" | Error marshalling to "+output.Uri)
					wg.Done()
					return
				}
			} else if strings.HasSuffix(output.Path, ".geojson.gz") || strings.HasSuffix(output.Path, ".geojson") {

				output_fc, err := planet.FeatureCollection(output)
				if err != nil {
					ch <- errors.Wrap(err, "Could not create feature collection from planet")
					wg.Done()
					return
				}

				output_bytes, err := json.Marshal(output_fc)
				if err != nil {
					ch <- errors.Wrap(err, "Could not marshal feature collection as response")
					wg.Done()
					return
				}

				if output.Uri == "stdout" {
					fmt.Println(string(output_bytes))
				} else if output.Uri == "stderr" {
					fmt.Fprintf(os.Stderr, string(output_bytes))
				} else if output.Scheme == "s3" {

				} else if output.Scheme == "file" {

					output_file, err := os.OpenFile(output.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
					if err != nil {
						ch <- errors.Wrap(err, "error opening file to write GeoJSON to disk")
						wg.Done()
						return
					}
					w := bufio.NewWriter(output_file)

					_, err = w.Write(output_bytes)
					if err != nil {
						ch <- errors.Wrap(err, "Error writing string to GeoJSON file")
						wg.Done()
						return
					}

					_, err = w.WriteString("\n")
					if err != nil {
						ch <- errors.Wrap(err, "Error writing last newline to GeoJSON file")
						wg.Done()
						return
					}

					w.Flush()
					if err != nil {
						ch <- errors.Wrap(err, "Error flushing output to bufio writer for GeoJSON file")
						wg.Done()
						return
					}

					err = output_file.Close()
					if err != nil {
						ch <- errors.Wrap(err, "Error closing file writer for GeoJSON file.")
						wg.Done()
						return
					}
				}

			}

			if profile {
				ch <- Message{Message: "Writing complete", Fields: map[string]interface{}{"uri": output.Uri, "duration": time.Since(start_marshal).String()}}
			}

			wg.Done()

		}(&wg, planet, i, o, ch, verbose)
	}

	wg.Wait()
	close(ch)

	if profile {
		logger.Info("Writing to all output finished in " + time.Since(start_output).String())
	}

	elapsed := time.Since(start)
	logger.Info("Done in " + elapsed.String())

}

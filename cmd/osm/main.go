package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	//"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

//import (
//	"github.com/golang/protobuf/proto"
//)

import (
	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
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
)

var GO_OSM_VERSION = "0.0.1"

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

func parse_path_s3(path string) (string, string, error) {
	if !strings.Contains(path, "/") {
		return "", "", errors.New("AWS S3 path does not include bucket.")
	}
	parts := strings.Split(path, "/")
	return parts[0], strings.Join(parts[1:], "/"), nil
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

func s3_bucket_exists(s3_client *s3.S3, bucket string) bool {

	input := &s3.HeadBucketInput{
		Bucket: aws.String(bucket),
	}

	_, err := s3_client.HeadBucket(input)
	if err != nil {
		return false
	}

	return true
}

func s3_object_exists(s3_client *s3.S3, bucket string, key string) bool {

	input := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s3_client.HeadObject(input)
	if err != nil {
		return false
	}

	return true
}

func s3_delete_object(s3_client *s3.S3, bucket string, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	_, err := s3_client.DeleteObject(input)
	if err != nil {
		return err
	}

	return nil
}

func s3_get_object(s3_client *s3.S3, bucket string, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	result, err := s3_client.GetObject(input)
	if err != nil {
		return make([]byte, 0), err
	}

	if strings.HasSuffix(key, ".gz") {

		gr, err := gzip.NewReader(result.Body)
		if err != nil {
			fmt.Println("Error creating gzip reader for AWS S3 object at s3://" + bucket + "/" + key + ".")
			fmt.Println(err)
			os.Exit(1)
		}
		defer gr.Close()

		obj, err := ioutil.ReadAll(gr)
		if err != nil {
			return make([]byte, 0), err
		}
		return obj, nil
	}

	obj, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return make([]byte, 0), err
	}

	return obj, nil

}

func s3_create_bucket(s3_client *s3.S3, region string, bucket string) error {
	input := &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(region),
		},
	}

	result, err := s3_client.CreateBucket(input)
	if err != nil {
		return err
	}

	fmt.Println("Create Bucket Result:", result)

	return nil

}

func s3_put_object(s3_client *s3.S3, bucket string, key string, data []byte) error {

	input := &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	if strings.HasSuffix(key, ".gz") {

		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		gw.Write(data)
		gw.Close()
		body := buf.Bytes()
		input.Body = bytes.NewReader(body)
		input.ContentLength = aws.Int64(int64(len(body)))
		input.ContentType = aws.String(http.DetectContentType(body))

	} else {

		input.Body = bytes.NewReader(data)
		input.ContentLength = aws.Int64(int64(len(data)))
		input.ContentType = aws.String(http.DetectContentType(data))

	}

	result, err := s3_client.PutObject(input)
	if err != nil {
		return err
	}

	//fmt.Println("Put Object Result:", result)

	return nil

}

func dfl_build_funcs() map[string]func(map[string]interface{}, []string) (interface{}, error) {
	funcs := map[string]func(map[string]interface{}, []string) (interface{}, error){}

	funcs["len"] = func(ctx map[string]interface{}, args []string) (interface{}, error) {
		if len(args) != 1 {
			return 0, errors.New("Invalid number of arguments to len.")
		}
		return len(args[0]), nil
	}

	return funcs
}

func main() {

	start := time.Now()

	var aws_default_region string
	var aws_access_key_id string
	var aws_secret_access_key string

	var input_uri string
	var output_uri string

	var include_keys_text string

	var dfl_filter_text string

	var ways_to_nodes bool

	var drop_relations bool
	var drop_version bool
	var drop_timestamp bool
	var drop_changeset bool
	var drop_uid bool
	var drop_user bool
	var drop_author bool

	var summarize bool
	var summarize_keys_text string

	var pretty bool

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

	flag.StringVar(&input_uri, "input_uri", "", "Input uri.  \"stdin\" or uri to input file.")
	flag.StringVar(&output_uri, "output_uri", "", "Output uri. \"stdout\", \"stderr\", or uri to output file.")

	flag.StringVar(&include_keys_text, "include_keys", "", "Comma-separated list of tag keys to keep")
	flag.StringVar(&dfl_filter_text, "dfl", "", "DFL filter")

	flag.BoolVar(&ways_to_nodes, "ways_to_nodes", false, "Convert ways into nodes for output")

	flag.BoolVar(&drop_relations, "drop_relations", false, "Drop relations from output")
	flag.BoolVar(&drop_version, "drop_version", false, "Drop version attribute from output")
	flag.BoolVar(&drop_timestamp, "drop_timestamp", false, "Drop timestamp attribute from output")
	flag.BoolVar(&drop_changeset, "drop_changeset", false, "Drop changeset attribute from output")

	flag.BoolVar(&drop_uid, "drop_uid", false, "Drop uid attribute from output")
	flag.BoolVar(&drop_user, "drop_user", false, "Drop user attribute from output")
	flag.BoolVar(&drop_author, "drop_author", false, "Drop author.  Synonymous to drop_uid and drop_user")

	flag.BoolVar(&summarize, "summarize", false, "Print data summary to stdout (bounding box, number of nodes, number of ways, and number of relations)")
	flag.StringVar(&summarize_keys_text, "summarize_keys", "", "Comma-separated list of keys to summarize")
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

	summarize_keys := make([]string, 0)
	if len(summarize_keys_text) > 0 {
		summarize_keys = strings.Split(summarize_keys_text, ",")
	}

	if help {
		fmt.Println("Usage: osm -input_uri INPUT -output_uri OUTPUT [-verbose] [-dry_run] [-version] [-help]")
		fmt.Println("Supported Schemes: " + strings.Join(SUPPORTED_SCHEMES, ", "))
		fmt.Println("Supported File Extensions: .osm, .osm.gz")
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
		b, k, err := parse_path_s3(output_path)
		if err != nil {
			fmt.Println("Error parsing AWS S3 path")
			fmt.Println(err)
			os.Exit(1)
		}
		output_s3_bucket = b
		output_s3_key = k
		output_exists = s3_object_exists(s3_client, output_s3_bucket, output_s3_key)

	} else {
		output_path_expanded = output_path
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

	if drop_author {
		drop_uid = true
		drop_user = true
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
			err := s3_delete_object(s3_client, output_s3_bucket, output_s3_key)
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
		if !s3_bucket_exists(s3_client, output_s3_bucket) {
			err := s3_create_bucket(s3_client, aws_default_region, output_s3_bucket)
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

	input_bytes := make([]byte, 0)
	if input_uri == "stdin" {

		in, err := ioutil.ReadAll(os.Stdin)
		if err != nil {
			fmt.Println("Error reading from stdin.")
			os.Exit(1)
		}
		input_bytes = []byte(strings.TrimSpace(string(in)))

	} else {

		input_scheme, input_path := parse_uri(input_uri, SUPPORTED_SCHEMES)

		if input_scheme == "file" {

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

		} else if input_scheme == "s3" {

			if s3_client == nil {
				if aws_session == nil {
					aws_session = connect_to_aws(aws_access_key_id, aws_secret_access_key, aws_default_region)
				}
				s3_client = s3.New(aws_session)
			}

			input_s3_bucket, input_s3_key, err := parse_path_s3(input_path)
			if err != nil {
				fmt.Println("Error parsing AWS S3 path")
				fmt.Println(err)
				os.Exit(1)
			}

			if strings.HasSuffix(input_s3_key, ".osm.gz") || strings.HasSuffix(input_s3_key, ".xml.gz") || strings.HasSuffix(input_s3_key, ".osm") || strings.HasSuffix(input_s3_key, ".xml") {

				in, err := s3_get_object(s3_client, input_s3_bucket, input_s3_key)
				if err != nil {
					fmt.Println("Error reading from AWS S3 uri " + input_uri + ".")
					fmt.Println(err)
					os.Exit(1)
				}
				input_bytes = in

			} else if strings.HasSuffix(input_s3_key, ".osm.pbf") || strings.HasSuffix(input_s3_key, ".xml.pbf") {
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

	var root dfl.Node
	if len(dfl_filter_text) > 0 {
		root, err = dfl.Parse(dfl_filter_text)
		if err != nil {
			fmt.Println("Error parsing DFL filter text", dfl_filter_text)
			fmt.Println(err)
			os.Exit(1)
		}
	}

	funcs := dfl_build_funcs()

	planet.Filter(include_keys, root, funcs)

	if ways_to_nodes {
		planet.ConvertWaysToNodes()
	}

	if drop_relations {
		planet.DropRelations()
	}

	planet.DropAttributes(drop_version, drop_timestamp, drop_changeset, drop_uid, drop_user)

	if summarize {
		summary := planet.Summarize(summarize_keys)
		summary.Print()
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
		} else if output_scheme == "file" {

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

		} else if output_scheme == "s3" {
			err := s3_put_object(
				s3_client,
				output_s3_bucket,
				output_s3_key,
				append(append([]byte(xml.Header), output_bytes...), []byte("\n")...))
			if err != nil {
				fmt.Println("Error uploading object to AWS S3 at output location " + output_uri + ".")
				fmt.Println(err)
				os.Exit(1)
			}
			if verbose {
				fmt.Println("Deleted uploading object to AWS S3 at output location " + output_uri + ".")
			}
		}

	}

	elapsed := time.Since(start)
	logger.Info("Done in " + elapsed.String())

}

# go-osm

# Description

**go-osm** is a tool for manipulating OSM planet files.  **go-osm** supports the Dynamic Filter Language (DFL).

# Building

```
cd scripts
bash build.sh
```

To build for windows run:

```
GOOS=windows GOARCH=amd64 go build github.com/spatialcurrent/go-osm/cmd/osm
```

# Usage

```
Usage: osm -input_uri INPUT -output_uri OUTPUT [-verbose] [-dry_run] [-version] [-help]
Supported Schemes: file, http, https, s3
Supported File Extensions: .osm, .osm.gz
Options:
  -aws_access_key_id string
    	Defaults to value of environment variable AWS_ACCESS_KEY_ID
  -aws_default_region string
    	Defaults to value of environment variable AWS_DEFAULT_REGION.
  -aws_secret_access_key string
    	Defaults to value of environment variable AWS_SECRET_ACCESS_KEY.
  -dfl string
    	DFL filter
  -drop_author
    	Drop author.  Synonymous to drop_uid and drop_user
  -drop_changeset
    	Drop changeset attribute from output
  -drop_relations
    	Drop relations from output
  -drop_timestamp
    	Drop timestamp attribute from output
  -drop_uid
    	Drop uid attribute from output
  -drop_user
    	Drop user attribute from output
  -drop_version
    	Drop version attribute from output
  -dry_run
    	Test user input but do not execute.
  -help
    	Print help
  -include_keys string
    	Comma-separated list of tag keys to keep
  -input_uri string
    	Input uri.  "stdin" or uri to input file.
  -output_uri string
    	Output uri. "stdout", "stderr", or uri to output file.
  -overwrite
    	Overwrite output file.
  -pretty
    	Pretty output.  Adds indents.
  -summarize
    	Print data summary to stdout (bounding box, number of nodes, number of ways, and number of relations)
  -summarize_keys string
    	Comma-separated list of keys to summarize
  -verbose
    	Provide verbose output
  -version
    	Prints version to stdout
  -ways_to_nodes
    	Convert ways into nodes for output
```

# Examples

Filter Washington, DC .osm.pbf planet file to only features that include a certain tag.

```shell
osmconvert district-of-columbia-latest.osm.pbf | osm \
-input_uri stdin \
-output_uri district-of-columbia-latest-filtered-nodes-cleaned.osm \
-include_keys amenity,aeroway,craft,leisure,shop,station,tourism \
-ways_to_nodes \
-drop_version \
-drop_timestamp \
-drop_changeset \
-drop_relations \
-verbose
```

Summarize OSM planet file in S3 folder

```shell
AWS_DEFAULT_REGION=us-east-1 osm -input_uri s3://<YOUR BUCKET>/data/district-of-columbia-latest.osm.gz -summarize
Bounding Box: -77.120100,38.791340,-76.909060,38.996030
Number of Nodes: 1701544
Number of Ways: 206181
Number of Relations: 3198
```

Breweries in Washington, DC

```
AWS_DEFAULT_REGION=us-east-1 ./osm -input_uri s3://<YOUR BUCKET>/data/district-of-columbia-latest.osm.gz -summarize -ways_to_nodes -dfl '@craft like brewery' -drop_relations -output_uri breweries.osm
Total Number of Nodes: 5
Total Number of Ways: 0
Total Number of Relations: 0
```

Breweries & Distilleries in Washington, DC as GeoJson

```
./osm -input_uri district-of-columbia-latest.osm.bz2 -summarize -pretty -verbose -drop_relations -drop_timestamp -drop_changeset -drop_version -ways_to_nodes -include_keys craft -dfl_use_cache -dfl '(@craft like brewery) or (@craft like distillery)'  -output_uri breweries_and_distilleries.geojson -drop_tags 'dcgis:gis_id' -stream -overwrite -ways_to_nodes
```

# Contributing

[Spatial Current, Inc.](https://spatialcurrent.io) is currently accepting pull requests for this repository.  We'd love to have your contributions!  Please see [Contributing.md](https://github.com/spatialcurrent/go-osm/blob/master/CONTRIBUTING.md) for how to get started.

# License

This work is distributed under the **MIT License**.  See **LICENSE** file.

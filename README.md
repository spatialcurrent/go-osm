# go-osm

# Description

**go-osm** is a tool for manipulating OSM planet files.

# Building

```
cd scripts
bash build.sh
```

# Usage

```
Usage: osm -input_uri INPUT -output_uri OUTPUT [-verbose] [-dry_run] [-version] [-help]
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
    	Input uri.  Supported file extensions: .osm, .osm.gz
  -output_uri string
    	Output uri.  Supported file extensions: .osm, .osm.gz
  -overwrite
    	Overwrite output file.
  -pretty
    	Pretty output.  Adds indents.
  -summarize
    	Print data summary to stdout (bounding box, number of nodes, number of ways, and number of relations)
  -verbose
    	Provide verbose output
  -version
    	Prints version to stdout
  -ways_to_nodes
    	Convert ways into nodes for output
```

# Contributing

[Spatial Current, Inc.](https://spatialcurrent.io) is currently accepting pull requests for this repository.  We'd love to have your contributions!  Please see [Contributing.md](https://github.com/spatialcurrent/go-osm/blob/master/CONTRIBUTING.md) for how to get started.

# License

This work is distributed under the **MIT License**.  See **LICENSE** file.

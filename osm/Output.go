package osm

import (
//"fmt"
//"os"
)

import (
//"github.com/aws/aws-sdk-go/service/s3"
//"github.com/mitchellh/go-homedir"
//"github.com/pkg/errors"
//"gopkg.in/ini.v1"
)

//import (
//	"github.com/spatialcurrent/go-osm/s3util"
//)

import (
	"github.com/spatialcurrent/go-dfl/dfl"
)

// Output is a struct for holding all the configuration describing an output destination
type Output struct {
	*PlanetResource
	WaysToNodes bool `hcl:"ways_to_nodes"` // convert ways into nodes
	Pretty      bool `hcl:"pretty"`        // write pretty output (newlines and tabs for .osm XML)
}

func (o *Output) Init(globals map[string]interface{}, ctx map[string]interface{}, funcs *dfl.FunctionMap) error {

	err := o.PlanetResource.Init(globals, ctx, funcs)
	if err != nil {
		return err
	}

	if len(globals) > 0 {

		if v, ok := globals["ways_to_nodes"]; ok {
			switch v.(type) {
			case bool:
				o.WaysToNodes = v.(bool)
			}
		}

		if v, ok := globals["pretty"]; ok {
			switch v.(type) {
			case bool:
				o.Pretty = v.(bool)
			}
		}
	}

	return nil
}

// HasDrop returns true if any property of elements will be dropped in the output.
func (o Output) HasDrop() bool {
	return o.DropWays || o.DropRelations || o.DropVersion || o.DropChangeset || o.DropTimestamp || o.DropUserId || o.DropUserName
}

// HasKeysToKeep returns true if there are keys to keep in the output, otherwise false.
func (o Output) HasKeysToKeep() bool {
	return len(o.KeysToKeep) > 0
}

// HasKeysToDrop returns true if there are keys to drop in the output, otherwise false.
func (o Output) HasKeysToDrop() bool {
	return len(o.KeysToDrop) > 0
}

/*func (o *Output) LoadGDALIniConfig(gdal_ini_uri string) error {
	gdal_ini_scheme, gdal_ini_path := SplitUri(gdal_ini_uri, []string{"file"})
	if gdal_ini_scheme == "file" {
		gdal_ini_path_expanded, err := homedir.Expand(gdal_ini_path)
		if err != nil {
			fmt.Println("Error expanding ini path " + gdal_ini_uri)
			os.Exit(1)
		}
		cfg, err := ini.Load(gdal_ini_path_expanded)
		if err != nil {
			return errors.Wrap(err, "Fail to read file")
		}
		o.DropVersion = !ParseBool(cfg.Section(gdal_ini_section).Key("osm_version").String())
		o.DropChangeset = !ParseBool(cfg.Section(gdal_ini_section).Key("osm_changeset").String())
		o.DropTimestamp = !ParseBool(cfg.Section(gdal_ini_section).Key("osm_timestamp").String())
		o.DropUserId = !ParseBool(cfg.Section(gdal_ini_section).Key("osm_id").String())
		o.DropUserName = !ParseBool(cfg.Section(gdal_ini_section).Key("osm_user").String())
		o.KeysToKeep = ParseSliceString(cfg.Section(gdal_ini_section).Key("attributes").String())
	}
}
*/

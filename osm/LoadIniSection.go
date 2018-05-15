package osm

import (
	"github.com/pkg/errors"
	"gopkg.in/ini.v1"
)

func LoadIniSection(uri string, section string, keys []string) (map[string]string, error) {
	m := make(map[string]string, len(keys))

	scheme, path := SplitUri(uri, []string{"file"})
	if scheme != "file" {
		return m, errors.New("Unsupported scheme for ini uri " + uri)
	}

	cfg, err := ini.Load(path)
	if err != nil {
		return m, errors.Wrap(err, "Fail to read file")
	}

	for _, k := range keys {
		m[k] = cfg.Section(section).Key(k).String()
	}

	return m, nil
}

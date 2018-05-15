package osm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"text/template"
)

import (
	"github.com/spatialcurrent/go-graph/graph"
)

func RenderTemplate(template_text string, ctx map[string]interface{}) (string, error) {
	templateFunctions := template.FuncMap{
		"lower": func(value string) string {
			return strings.ToLower(value)
		},
		"upper": func(value string) string {
			return strings.ToUpper(value)
		},
		"replace": func(old string, new string, value string) string {
			return strings.Replace(value, old, new, -1)
		},
		"float64": func(value interface{}) string {
			switch value.(type) {
			case string:
				if len(value.(string)) == 0 {
					return "0.0"
				}
				f, err := strconv.ParseFloat(value.(string), 64)
				if err != nil {
					return "0.0"
				}
				return fmt.Sprintf("%f", f)
			case int:
				return fmt.Sprintf("%f", float64(value.(int)))
			case int64:
				return fmt.Sprintf("%f", float64(value.(int64)))
			case float64:
				return fmt.Sprintf("%f", value.(float64))
			}
			return "0.0"
		},
		"json": func(value interface{}) string {
			out_bytes, err := json.Marshal(value)
			if err != nil {
				return ""
			}
			return string(out_bytes)
		},
		"map": func(value interface{}) map[string]interface{} {
			switch value.(type) {
			case []interface{}:
				out := map[string]interface{}{}
				for _, v := range value.([]interface{}) {
					v_msi := v.(map[string]interface{})
					out[v_msi["name"].(string)] = v_msi["value"]
				}
				return out
			case []map[string]interface{}:
				out := map[string]interface{}{}
				for _, v := range value.([]map[string]interface{}) {
					out[fmt.Sprintf("%v", v["name"])] = v["value"]
				}
				return out
			case map[interface{}]interface{}:
				return graph.StringifyMapKeys(value).(map[string]interface{})
			case map[string]interface{}:
				return value.(map[string]interface{})
			}
			return map[string]interface{}{}
		},
	}
	tmpl, err := template.New("tmpl").Funcs(templateFunctions).Parse(template_text)
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(buf, "tmpl", ctx)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

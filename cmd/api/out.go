package api

import (
	"fmt"
	"os"
	"strconv"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/luscis/openlan/pkg/libol"
)

func OutJson(data interface{}) error {
	if out, err := libol.Marshal(data, true); err == nil {
		fmt.Println(string(out))
	} else {
		return err
	}
	return nil
}

func OutYaml(data interface{}) error {
	if out, err := yaml.Marshal(data); err == nil {
		fmt.Println(string(out))
	} else {
		return err
	}
	return nil
}

func OutTable(data interface{}, tmpl string) error {
	funcMap := template.FuncMap{
		"ps": func(space int, args ...interface{}) string {
			format := "%" + strconv.Itoa(space) + "s"
			if space < 0 {
				format = "%-" + strconv.Itoa(space) + "s"
			}
			return fmt.Sprintf(format, args...)
		},
		"pi": func(space int, args ...interface{}) string {
			format := "%" + strconv.Itoa(space) + "d"
			if space < 0 {
				format = "%-" + strconv.Itoa(space) + "d"
			}
			return fmt.Sprintf(format, args...)
		},
		"pu": func(space int, args ...interface{}) string {
			format := "%" + strconv.Itoa(space) + "u"
			if space < 0 {
				format = "%-" + strconv.Itoa(space) + "u"
			}
			return fmt.Sprintf(format, args...)
		},
		"pt": func(value int64) string {
			return libol.PrettyTime(value)
		},
		"p2": func(space int, format, key1, key2 string) string {
			value := fmt.Sprintf(format, key1, key2)
			format = "%" + strconv.Itoa(space) + "s"
			if space < 0 {
				format = "%-" + strconv.Itoa(space) + "s"
			}
			return fmt.Sprintf(format, value)
		},
		"ut": func(value int64) string {
			return libol.UnixTime(value)
		},
	}
	if tmpl, err := template.New("main").Funcs(funcMap).Parse(tmpl); err != nil {
		return err
	} else {
		if err := tmpl.Execute(os.Stdout, data); err != nil {
			return err
		}
	}
	return nil
}

func Out(data interface{}, format string, tmpl string) error {
	libol.Debug("Out %s %s", format, tmpl)
	switch format {
	case "json":
		return OutJson(data)
	case "yaml":
		return OutYaml(data)
	default:
		return OutTable(data, tmpl)
	}
}

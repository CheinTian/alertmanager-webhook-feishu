package tmpl

import (
	"embed"
	"errors"
	"fmt"
	amtemplate "github.com/prometheus/alertmanager/template"
	"github.com/sirupsen/logrus"
	"net/url"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

//go:embed templates/*
var fs embed.FS
var embedTemplates map[string]*template.Template
var customTemplates map[string]*template.Template
var funcMap template.FuncMap

func init() {
	// func
	funcMap = template.FuncMap{
		"date": func(dt time.Time, zone string) string {
			loc, err := time.LoadLocation(zone)
			if err != nil {
				logrus.Error(err)
				return err.Error()
			}
			dt = dt.In(loc)
			return dt.Format("2006-01-02 15:04:05")
		},
		"isNonZeroDate": func(dt time.Time) bool {
			return !(dt == time.Time{})
		},
		"in": func(m map[string]string, key string) bool {
			_, ok := m[key]
			return ok
		},
		"toUpper": strings.ToUpper,
		"toLink": func(s string) string {
			return fmt.Sprintf("[%s](%s)", s, s)
		},
		"displayKV": func(k, v string) string {
			_, err := url.ParseRequestURI(v)
			if err != nil {
				return fmt.Sprintf("%s:%s", k, v)
			}
			return fmt.Sprintf("[%s](%s)", k, v)
		},
		"contains": strings.Contains,
		"default": func(a, b string) string {
			if a == "" {
				return b
			}
			return a
		},
		"dict": func(values ...interface{}) (map[string]interface{}, error) {
			if len(values)%2 != 0 {
				return nil, errors.New("invalid dict call")
			}
			dict := make(map[string]interface{}, len(values)/2)
			for i := 0; i < len(values); i += 2 {
				key, ok := values[i].(string)
				if !ok {
					return nil, errors.New("dict keys must be strings")
				}
				dict[key] = values[i+1]
			}
			return dict, nil
		},
		"filterByStatus": func(alerts []amtemplate.Alert, status string) []amtemplate.Alert {
			// 预分配切片，提高性能
			filteredAlerts := make([]amtemplate.Alert, 0, len(alerts))

			for _, alert := range alerts {
				// 直接比较 Status 字段
				if alert.Status == status {
					filteredAlerts = append(filteredAlerts, alert)
				}
			}

			return filteredAlerts
		},
		"removeEmptyLines": func(s string) string {
			return strings.ReplaceAll(s, "\n\n", "\n")
		},
	}

	// embed
	dir, err := fs.ReadDir("templates")
	if err != nil {
		panic(err)
	}

	embedTemplates = make(map[string]*template.Template)
	for _, entry := range dir {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		if !strings.HasSuffix(filename, ".tmpl") {
			continue
		}

		t, err := template.New(filename).Funcs(funcMap).ParseFS(fs, "templates/"+filename)
		if err != nil {
			panic(err)
		}

		embedTemplates[t.Name()] = t
	}

	// custom
	customTemplates = make(map[string]*template.Template)
}

func GetEmbedTemplate(filename string) (*template.Template, error) {
	if t, ok := embedTemplates[filename]; ok {
		return t, nil
	}

	return nil, errors.New("template not found")
}

func GetCustomTemplate(filename string) (*template.Template, error) {
	if t, ok := customTemplates[filename]; ok {
		return t, nil
	}

	t, err := template.New(filepath.Base(filename)).Funcs(funcMap).ParseFiles(filename)
	if err != nil {
		return nil, err
	}
	customTemplates[filename] = t

	return t, nil
}

package dataview

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/tidwall/gjson"
)

// Definition contains the view definition of when to use a specific view
type Definition struct {
	FileName    string   `json:"-"`
	Name        string   `json:"name,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	Fragments   []string `json:"fragments,omitempty"`
	Type        string   `json:"type,omitempty"`
	ContentType string   `json:"contentType,omitempty"`
	Self        string   `json:"self,omitempty"`
	Columns     []string `json:"columns,omitempty"`
}

// DefinitionCollection collection of view definitions
type DefinitionCollection struct {
	Definitions []Definition `json:"definitions,omitempty"`
}

// DataView data view containing pre-definied views
type DataView struct {
	Paths       []string
	Extension   string
	Pattern     string
	Definitions []Definition
	Logger      *logger.Logger
}

// NewDataView creates a new data view which selected a view based in json data
func NewDataView(pattern string, extension string, log *logger.Logger, paths ...string) (*DataView, error) {
	if log == nil {
		log = logger.NewDummyLogger("dataview")
	}
	view := &DataView{
		Paths:     paths,
		Pattern:   pattern,
		Extension: extension,
		Logger:    log,
	}
	return view, nil
}

func (v *DataView) LoadDefinitions() error {
	definitions := make([]Definition, 0)
	v.Logger.Debugf("Looking for definitions in: %v", v.Paths)
	for _, path := range v.Paths {
		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if !d.IsDir() {
				if !strings.EqualFold(filepath.Ext(d.Name()), v.Extension) {
					return nil
				}
				if m, err := regexp.MatchString(v.Pattern, filepath.Base(d.Name())); err == nil && !m {
					return nil
				}

				contents, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}

				v.Logger.Debugf("Found view definition: %s", d.Name())
				viewDefinition := &DefinitionCollection{}
				if err := json.Unmarshal(contents, &viewDefinition); err != nil {
					v.Logger.Warnf("Could not load view definitions. %s", err)
					return err
				}
				definitions = append(definitions, viewDefinition.Definitions...)
			}
			return nil
		})
		if err != nil {
			v.Logger.Warnf("View discovery has errors. %s", err)
			return err
		}
		v.Logger.Debugf("Loaded definitions: %d", len(definitions))

	}
	// sort by priority
	sort.Slice(definitions, func(i, j int) bool {
		return definitions[i].Priority < definitions[j].Priority
	})
	v.Definitions = definitions
	return nil
}

func (v *DataView) GetView(data *gjson.Result, contentType ...string) ([]string, error) {
	err := v.LoadDefinitions()
	if err != nil {
		return nil, err
	}
	if data.IsArray() {
		if len(data.Array()) == 0 {
			return nil, nil
		}
		data = &data.Array()[0]
	}

	var matchingDefinition *Definition
	for _, definition := range v.Definitions {
		isMatch := true

		for _, fragment := range definition.Fragments {
			if result := data.Get(fragment); !result.Exists() {
				// v.Logger.Debugf("Data did not contain fragment. view=%s, fragment=%s, input=%s", definition.FileName, fragment, data.Raw)
				isMatch = false
				break
			}
		}
		if definition.Type != "" {
			if v := data.Get("type"); v.Exists() {
				if match, err := regexp.MatchString("(?i)"+definition.Type, v.Str); err == nil && !match {
					isMatch = false
				}
			} else {
				isMatch = false
			}
		}

		if len(contentType) > 0 {
			if match, err := regexp.MatchString("(?i)"+definition.ContentType, contentType[0]); err == nil && !match {
				isMatch = false
			}
		}

		if definition.Self != "" {
			if v := data.Get("self"); v.Exists() {
				if match, err := regexp.MatchString("(?i)"+definition.Self, v.Str); err == nil && !match {
					isMatch = false
				}
			} else {
				isMatch = false
			}
		}

		if isMatch {
			matchingDefinition = &definition
			break
		}
	}
	if matchingDefinition != nil {
		v.Logger.Debugf("Found matching view: name=%s", matchingDefinition.Name)
		return matchingDefinition.Columns, nil
	}
	v.Logger.Debug("No matching view found")
	return nil, nil
}

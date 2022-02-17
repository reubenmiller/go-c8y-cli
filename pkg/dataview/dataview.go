package dataview

import (
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/reubenmiller/go-c8y-cli/pkg/logger"
	"github.com/reubenmiller/go-c8y-cli/pkg/matcher"
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
	mu          sync.RWMutex
	Paths       []string
	Extension   string
	Pattern     string
	Definitions []Definition
	Logger      *logger.Logger
	ActiveView  *Definition
}

// NewDataView creates a new data view which selected a view based in json data
func NewDataView(pattern string, extension string, log *logger.Logger, paths ...string) (*DataView, error) {
	if log == nil {
		log = logger.NewDummyLogger("dataview")
	}
	view := &DataView{
		mu:        sync.RWMutex{},
		Paths:     paths,
		Pattern:   pattern,
		Extension: extension,
		Logger:    log,
	}
	return view, nil
}

// LoadDefinitions load the view definitions
func (v *DataView) LoadDefinitions() error {

	if len(v.GetDefinitions()) > 0 {
		v.Logger.Debugf("Views already loaded")
		return nil
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	definitions := make([]Definition, 0)
	v.Logger.Debugf("Looking for definitions in: %v", v.Paths)
	for _, path := range v.Paths {
		v.Logger.Debugf("Current view path: %s", path)

		if stat, err := os.Stat(path); err != nil || !(err == nil && stat.IsDir()) {
			if err == nil && stat != nil && !stat.IsDir() {
				v.Logger.Debugf("Skipping view path because it is not a folder. path=%s", path)
			} else {
				v.Logger.Debugf("Skipping view path because it does not exist. path=%s, error=%s", path, err)
			}
			continue
		}
		err := filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				// do not block walking folder
				v.Logger.Warnf("Failed to walk path: %s, err=%s. file will be ignored", path, err)
				return nil
			}
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
				for i := range viewDefinition.Definitions {
					viewDefinition.Definitions[i].FileName = d.Name()
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

// GetViewByName get view by name. Accepts wildcard name
func (v *DataView) GetViewByName(pattern string) ([]string, error) {
	err := v.LoadDefinitions()
	if err != nil {
		return nil, err
	}

	var matchingDefinition *Definition

	for _, definition := range v.GetDefinitions() {

		if match, _ := matcher.MatchWithWildcards(definition.Name, pattern); match {
			matchingDefinition = &definition
			break
		}
	}

	if matchingDefinition == nil {
		return nil, nil
	}

	return matchingDefinition.Columns, nil
}

// GetViews get a list of view names
func (v *DataView) GetViews(pattern string) ([]Definition, error) {
	err := v.LoadDefinitions()
	if err != nil {
		return nil, err
	}

	matches := []Definition{}

	for _, definition := range v.GetDefinitions() {
		if match, _ := matcher.MatchWithWildcards(definition.Name, pattern); match {
			matches = append(matches, definition)
		}
	}

	return matches, nil
}

func (v *DataView) GetDefinitions() []Definition {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.Definitions
}

// GetActiveView get the active view
func (v *DataView) GetActiveView() *Definition {
	v.mu.RLock()
	defer v.mu.RUnlock()
	return v.ActiveView
}

// ClearActiveView clear the active view
func (v *DataView) ClearActiveView() {
	v.mu.Lock()
	defer v.mu.Unlock()
	v.ActiveView = nil
}

func (v *DataView) GetView(data *gjson.Result, contentType ...string) ([]string, error) {
	if view := v.GetActiveView(); view != nil {
		v.Logger.Debugf("Using already active view")
		return view.Columns, nil
	}

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
	definitions := v.GetDefinitions()
	v.mu.Lock()
	defer v.mu.Unlock()

	var matchingDefinition *Definition
	for _, definition := range definitions {
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
		v.ActiveView = matchingDefinition
		return matchingDefinition.Columns, nil
	}
	v.Logger.Debug("No matching view found")
	return nil, nil
}

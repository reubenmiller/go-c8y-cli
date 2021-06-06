package models

import "strings"

type Specification struct {
	Information SpecificationInformation `yaml:"information"`
	Endpoints   []EndPoint               `yaml:"endpoints"`
}

type SpecificationInformation struct {
	Name            string `yaml:"name"`
	Description     string `yaml:"description"`
	DescriptionLong string `yaml:"descriptionLong"`
	Link            string `yaml:"link"`
}

type EndPoint struct {
	Name               string       `yaml:"name"`
	Method             string       `yaml:"method"`
	Accept             string       `yaml:"accept,omitempty"`
	CollectionType     string       `yaml:"collectionType,omitempty"`
	CollectionProperty string       `yaml:"collectionProperty,omitempty"`
	Path               string       `yaml:"path"`
	Examples           Examples     `yaml:"examples"`
	Alias              Aliases      `yaml:"alias"`
	Skip               *bool        `yaml:"skip,omitempty"`
	QueryParameters    []Parameter  `yaml:"queryParameters,omitempty"`
	PathParameters     []Parameter  `yaml:"pathParameters,omitempty"`
	HeaderParameters   []Parameter  `yaml:"headerParameters,omitempty"`
	Body               []Parameter  `yaml:"body,omitempty"`
	BodyContent        *BodyContent `yaml:"bodyContent,omitempty"`
}

func (p *EndPoint) GetAllParameters() []Parameter {
	parameters := make([]Parameter, 0)
	if len(p.QueryParameters) > 0 {
		parameters = append(parameters, p.QueryParameters...)
	}
	if len(p.PathParameters) > 0 {
		parameters = append(parameters, p.PathParameters...)
	}
	if len(p.HeaderParameters) > 0 {
		parameters = append(parameters, p.HeaderParameters...)
	}
	if len(p.Body) > 0 {
		parameters = append(parameters, p.Body...)
	}
	return parameters
}

type Aliases struct {
	Go         string `yaml:"go"`
	PowerShell string `yaml:"powershell"`
}

type Examples struct {
	Powershell []Example `yaml:"powershell"`
	Go         []Example `yaml:"go"`
}

type Example struct {
	Description  string           `yaml:"description,omitempty"`
	Command      string           `yaml:"command"`
	AssertStdout *OutputAssertion `yaml:"assertStdOut,omitempty"`
	AssertStderr *OutputAssertion `yaml:"assertStdErr,omitempty"`
	BeforeEach   []string         `yaml:"beforeEach,omitempty"`
	AfterEach    []string         `yaml:"afterEach,omitempty"`
	SkipTest     bool             `yaml:"skipTest,omitempty"`
}

type BodyContent struct {
	Type string `yaml:"type,omitempty"`
}

type Parameter struct {
	Name            string   `yaml:"name,omitempty"`
	Type            string   `yaml:"type,omitempty"`
	Value           string   `yaml:"value,omitempty"`
	Property        string   `yaml:"property,omitempty"`
	Pipeline        *bool    `yaml:"pipeline,omitempty"`
	PipelineAliases []string `yaml:"pipelineAliases,omitempty"`
	Required        *bool    `yaml:"required,omitempty"`
	Description     string   `yaml:"description,omitempty"`
	Default         string   `yaml:"default,omitempty"`
	Position        *int     `yaml:"position,omitempty"`
	ValidationSet   []string `yaml:"validationSet,omitempty"`
	Skip            *bool    `yaml:"skip,omitempty"`
}

func (p *Parameter) GetTargetProperty() string {
	if p.Property != "" {
		return p.Property
	}
	return p.Name
}

func (p *EndPoint) IsCollection() bool {
	return strings.EqualFold(p.Method, "GET") &&
		(p.CollectionProperty != "" || strings.Contains(strings.ToLower(p.Accept), "collection"))
}

func (p *EndPoint) SupportsTemplates() bool {
	return strings.EqualFold(p.Method, "PUT") || strings.EqualFold(p.Method, "POST")
}

func (p *EndPoint) IsBodyFormData() bool {
	return p.BodyContent != nil && p.BodyContent.Type == "formdata"
}

func (p *Parameter) IsTypeDateTime() bool {
	return strings.EqualFold(p.Type, "datetime")
}

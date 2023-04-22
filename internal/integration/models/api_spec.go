package models

import "strings"

type Specification struct {
	Group    Group     `yaml:"group"`
	Commands []Command `yaml:"commands"`
}

type Group struct {
	Name            string `yaml:"name"`
	Description     string `yaml:"description"`
	DescriptionLong string `yaml:"descriptionLong"`
	Link            string `yaml:"link"`
	Skip            bool   `yaml:"skip"`
}

type BodyTemplate struct {
	Type      string `yaml:"type"`
	ApplyLast string `yaml:"applyLast"`
	Template  string `yaml:"template"`
}

type Command struct {
	Name               string         `yaml:"name"`
	Description        string         `yaml:"description"`
	DescriptionLong    string         `yaml:"descriptionLong"`
	Deprecated         string         `yaml:"deprecated"`
	DeprecatedAt       string         `yaml:"deprecatedAt"`
	Method             string         `yaml:"method"`
	SemanticMethod     string         `yaml:"semanticMethod"`
	Accept             string         `yaml:"accept,omitempty"`
	CollectionType     string         `yaml:"collectionType,omitempty"`
	CollectionProperty string         `yaml:"collectionProperty,omitempty"`
	Path               string         `yaml:"path"`
	Examples           Examples       `yaml:"examples"`
	ExampleList        []Example      `yaml:"exampleList"`
	Alias              Aliases        `yaml:"alias"`
	Hidden             *bool          `yaml:"hidden,omitempty"`
	Skip               *bool          `yaml:"skip,omitempty"`
	QueryParameters    []Parameter    `yaml:"queryParameters,omitempty"`
	PathParameters     []Parameter    `yaml:"pathParameters,omitempty"`
	HeaderParameters   []Parameter    `yaml:"headerParameters,omitempty"`
	Body               []Parameter    `yaml:"body,omitempty"`
	BodyContent        *BodyContent   `yaml:"bodyContent,omitempty"`
	BodyTemplates      []BodyTemplate `yaml:"bodyTemplates,omitempty"`
	BodyRequiredKeys   []string       `yaml:"bodyRequiredKeys,omitempty"`
}

func (c *Command) SupportsProcessingMode() bool {
	return c.Method == "DELETE" || c.Method == "PUT" || c.Method == "POST"
}
func (c *Command) IsHidden() bool {
	return c.Hidden != nil && *c.Hidden
}

func (c *Command) ShouldIgnore() bool {
	return c.Skip != nil && *c.Skip
}

func (c *Command) GetDescriptionLong() string {
	var sb strings.Builder

	if c.Description != "" {
		sb.WriteString(c.Description)
	}
	if c.DescriptionLong != "" {
		sb.WriteString("\n\n")
		sb.WriteString(c.DescriptionLong)
	}
	return sb.String()

}

func (c *Command) IsDeprecated() bool {
	return c.Deprecated != ""
}

func (c *Command) GetExamples() string {
	var sb strings.Builder
	for _, ex := range c.ExampleList {
		sb.WriteString("  $ " + strings.TrimSpace(ex.Command) + "\n")
		sb.WriteString("  " + strings.TrimSpace(ex.Description) + "\n\n")
	}
	return strings.TrimRight(sb.String(), "\n")
}

func (c *Command) GetMethod() string {
	if c.SemanticMethod != "" {
		return c.SemanticMethod
	}
	return c.Method
}

func (c *Command) GetAllParameters() []Parameter {
	parameters := make([]Parameter, 0)
	if len(c.QueryParameters) > 0 {
		for _, param := range c.QueryParameters {
			if len(param.Children) > 0 {
				parameters = append(parameters, param.Children...)
			} else {
				param.TargetType = ParamQueryParameter
				parameters = append(parameters, param)
			}
		}
	}
	if len(c.PathParameters) > 0 {
		for _, p := range c.PathParameters {
			p.TargetType = ParamPath
		}
		parameters = append(parameters, c.PathParameters...)
	}
	if len(c.HeaderParameters) > 0 {
		for _, p := range c.HeaderParameters {
			p.TargetType = ParamHeader
		}
		parameters = append(parameters, c.HeaderParameters...)
	}
	if len(c.Body) > 0 {
		for _, p := range c.Body {
			p.TargetType = ParamBody
		}
		parameters = append(parameters, c.Body...)
	}
	return parameters
}

func (c *Command) GetQueryParameters() []Parameter {
	parameters := make([]Parameter, 0)
	if len(c.QueryParameters) > 0 {
		for _, param := range c.QueryParameters {
			if len(param.Children) > 0 {
				parameters = append(parameters, param.Children...)
			} else {
				parameters = append(parameters, param)
			}
		}
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
	Name            string      `yaml:"name,omitempty"`
	ShortName       string      `yaml:"shortname,omitempty"`
	Type            string      `yaml:"type,omitempty"`
	Value           string      `yaml:"value,omitempty"`
	Format          string      `yaml:"format,omitempty"`
	Property        string      `yaml:"property,omitempty"`
	Hidden          *bool       `yaml:"hidden,omitempty"`
	Pipeline        *bool       `yaml:"pipeline,omitempty"`
	PipelineAliases []string    `yaml:"pipelineAliases,omitempty"`
	Required        *bool       `yaml:"required,omitempty"`
	Description     string      `yaml:"description,omitempty"`
	Default         string      `yaml:"default,omitempty"`
	Position        *int        `yaml:"position,omitempty"`
	ValidationSet   []string    `yaml:"validationSet,omitempty"`
	Skip            *bool       `yaml:"skip,omitempty"`
	Children        []Parameter `yaml:"children,omitempty"`
	DependsOn       []string    `yaml:"dependsOn,omitempty"`

	TargetType TargetType `yaml:"-"`
}

type TargetType int

const (
	ParamHeader TargetType = iota
	ParamBody
	ParamPath
	ParamQueryParameter
)

func (p *Parameter) IsRequired() bool {
	return p.Required != nil && *p.Required
}

func (p *Parameter) AcceptsPipeline() bool {
	return p.Pipeline != nil && *p.Pipeline
}

func (p *Parameter) IsHidden() bool {
	return p.Hidden != nil && *p.Hidden
}

func (p *Parameter) GetDescription() string {
	var sb strings.Builder
	sb.WriteString(p.Description)
	if p.Required != nil && *p.Required {
		sb.WriteString(" (required)")
	}
	if p.Pipeline != nil && *p.Pipeline {
		sb.WriteString(" (accepts pipeline)")
	}
	return sb.String()
}

func (p *Parameter) GetTargetProperty() string {
	if p.Property != "" {
		return p.Property
	}
	return p.Name
}

func (p *Command) IsCollection() bool {
	return strings.EqualFold(p.Method, "GET") &&
		(p.CollectionProperty != "" || strings.Contains(strings.ToLower(p.Accept), "collection"))
}

func (p *Command) SupportsTemplates() bool {
	return strings.EqualFold(p.Method, "PUT") || strings.EqualFold(p.Method, "POST")
}

func (p *Command) IsBodyFormData() bool {
	return p.BodyContent != nil && p.BodyContent.Type == "formdata"
}

func (p *Command) GetBodyContentType() string {
	if p.BodyContent == nil {
		return ""
	}
	return p.BodyContent.Type
}

func (p *Parameter) IsTypeDateTime() bool {
	return strings.EqualFold(p.Type, "datetime")
}

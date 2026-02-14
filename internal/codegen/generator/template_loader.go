package generator

import (
	"bytes"
	"embed"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
	"unicode"

	"github.com/gofhir/models/internal/codegen/analyzer"
)

// Kind constants for type categorization.
const (
	kindResource = "resource"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS

// TemplateData holds common data for templates.
type TemplateData struct {
	PackageName string
	Version     string
	FileType    string
}

// TypesTemplateData holds data for datatypes and resources templates.
type TypesTemplateData struct {
	TemplateData
	Types []*analyzer.AnalyzedType
}

// RegistryTemplateData holds data for registry template.
type RegistryTemplateData struct {
	TemplateData
	ResourceNames []string
}

// CodeSystemsTemplateData holds data for codesystems template.
type CodeSystemsTemplateData struct {
	TemplateData
	ValueSets []ValueSetData
}

// ValueSetData holds processed value set data for templates.
type ValueSetData struct {
	Name     string
	TypeName string
	Title    string
	Codes    []CodeData
}

// CodeData holds processed code data for templates.
type CodeData struct {
	Code      string
	Display   string
	ConstName string
}

// BuildersTemplateData holds data for builders template.
type BuildersTemplateData struct {
	TemplateData
	Resources []ResourceBuilderData
}

// ResourceBuilderData holds data for a single resource builder.
type ResourceBuilderData struct {
	Name       string
	LowerName  string
	Properties []PropertyBuilderData
}

// PropertyBuilderData holds processed property data for builder templates.
type PropertyBuilderData struct {
	Name        string
	GoType      string
	IsArray     bool
	IsPointer   bool
	IsChoice    bool
	ElementType string // For arrays: the element type (e.g., "HumanName" from "[]HumanName")
	BaseType    string // For pointers: the base type (e.g., "string" from "*string")
}

// BackbonesTemplateData holds data for backbones template.
type BackbonesTemplateData struct {
	TemplateData
	Backbones []*analyzer.AnalyzedType
}

// loadTemplate loads a template by name from embedded files.
func loadTemplate(name string) (*template.Template, error) {
	content, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	return tmpl, nil
}

// executeTemplate executes a template and returns formatted Go code.
func executeTemplate(tmpl *template.Template, data interface{}) ([]byte, error) {
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return buf.Bytes(), fmt.Errorf("failed to format code: %w (unformatted content available)", err)
	}

	return formatted, nil
}

// writeTemplateFile executes a template and writes to file.
func writeTemplateFile(outputPath, templateName string, data interface{}) error {
	tmpl, err := loadTemplate(templateName)
	if err != nil {
		return err
	}

	content, err := executeTemplate(tmpl, data)
	if err != nil {
		// Write unformatted content for debugging
		unformattedPath := outputPath + ".unformatted"
		if writeErr := os.WriteFile(unformattedPath, content, 0o600); writeErr != nil {
			return fmt.Errorf("%w (also failed to write debug file: %v)", err, writeErr)
		}
		return fmt.Errorf("%w (saved to %s)", err, unformattedPath)
	}

	return os.WriteFile(outputPath, content, 0o600)
}

// generateRegistryFromTemplate generates registry.go using template.
func (c *CodeGen) generateRegistryFromTemplate() error {
	var resourceNames []string
	for _, t := range c.types {
		if t.Kind == kindResource {
			resourceNames = append(resourceNames, t.Name)
		}
	}

	sort.Strings(resourceNames)

	data := RegistryTemplateData{
		TemplateData: TemplateData{
			PackageName: c.config.PackageName,
			Version:     strings.ToUpper(c.config.Version),
			FileType:    "registry",
		},
		ResourceNames: resourceNames,
	}

	path := filepath.Join(c.config.OutputDir, "registry.go")
	return writeTemplateFile(path, "registry.go.tmpl", data)
}

// generateInterfacesFromTemplate generates interfaces.go using template.
func (c *CodeGen) generateInterfacesFromTemplate() error {
	data := TemplateData{
		PackageName: c.config.PackageName,
		Version:     strings.ToUpper(c.config.Version),
		FileType:    "interfaces",
	}

	path := filepath.Join(c.config.OutputDir, "interfaces.go")
	return writeTemplateFile(path, "interfaces.go.tmpl", data)
}

// generateCodeSystemsFromTemplate generates codesystems.go using template.
func (c *CodeGen) generateCodeSystemsFromTemplate() error {
	if c.analyzer == nil || len(c.analyzer.UsedBindings) == 0 {
		return nil
	}

	// Collect and sort used value sets
	valueSetURLs := make([]string, 0, len(c.analyzer.UsedBindings))
	for url := range c.analyzer.UsedBindings {
		valueSetURLs = append(valueSetURLs, url)
	}
	sort.Strings(valueSetURLs)

	// Track generated type names to avoid duplicates
	generatedTypes := make(map[string]bool)
	valueSets := make([]ValueSetData, 0, len(valueSetURLs))

	for _, url := range valueSetURLs {
		vs := c.valueSets.Get(url)
		if vs == nil {
			continue
		}

		typeName := sanitizeTypeName(vs.Name)
		if generatedTypes[typeName] {
			continue
		}
		generatedTypes[typeName] = true

		vsData := ValueSetData{
			Name:     vs.Name,
			TypeName: typeName,
			Title:    vs.Title,
			Codes:    make([]CodeData, 0, len(vs.Codes)),
		}

		for _, code := range vs.Codes {
			vsData.Codes = append(vsData.Codes, CodeData{
				Code:      code.Code,
				Display:   code.Display,
				ConstName: toPascalCaseCode(code.Code),
			})
		}

		valueSets = append(valueSets, vsData)
	}

	data := CodeSystemsTemplateData{
		TemplateData: TemplateData{
			PackageName: c.config.PackageName,
			Version:     strings.ToUpper(c.config.Version),
			FileType:    "codesystems",
		},
		ValueSets: valueSets,
	}

	path := filepath.Join(c.config.OutputDir, "codesystems.go")
	return writeTemplateFile(path, "codesystems.go.tmpl", data)
}

// toLowerFirstChar converts the first character to lowercase.
func toLowerFirstChar(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// SummaryTemplateData holds data for summary template.
type SummaryTemplateData struct {
	TemplateData
	Resources []ResourceSummaryData
}

// ResourceSummaryData holds summary field data for a resource.
type ResourceSummaryData struct {
	Name          string
	SummaryFields []string
}

// generateSummaryFromTemplate generates summary.go using template.
func (c *CodeGen) generateSummaryFromTemplate() error {
	resources := make([]ResourceSummaryData, 0)

	for _, t := range c.types {
		if t.Kind != kindResource {
			continue
		}

		summaryFields := make([]string, 0)
		for _, prop := range t.Properties {
			if prop.IsSummary {
				summaryFields = append(summaryFields, prop.JSONName)
			}
		}

		// Only include resources that have summary fields
		if len(summaryFields) > 0 {
			sort.Strings(summaryFields)
			resources = append(resources, ResourceSummaryData{
				Name:          t.Name,
				SummaryFields: summaryFields,
			})
		}
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	data := SummaryTemplateData{
		TemplateData: TemplateData{
			PackageName: c.config.PackageName,
			Version:     strings.ToUpper(c.config.Version),
			FileType:    "summary",
		},
		Resources: resources,
	}

	path := filepath.Join(c.config.OutputDir, "summary.go")
	return writeTemplateFile(path, "summary.go.tmpl", data)
}

// ============================================================================
// NEW: Separate File Generation Functions
// ============================================================================

// generateDatatypesSeparately generates one file per datatype.
func (c *CodeGen) generateDatatypesSeparately() error {
	for _, t := range c.types {
		// Include datatype, primitive, and backbone types (like Dosage, Timing)
		// Backbone types that appear here are top-level datatypes with nested backbones
		if t.Kind != "datatype" && t.Kind != "primitive" && t.Kind != "backbone" {
			continue
		}

		// Skip Element and BackboneElement (they go in datatype_base.go)
		if t.Name == "Element" || t.Name == "BackboneElement" {
			continue
		}

		data := TypesTemplateData{
			TemplateData: TemplateData{
				PackageName: c.config.PackageName,
				Version:     strings.ToUpper(c.config.Version),
				FileType:    "datatypes",
			},
			Types: []*analyzer.AnalyzedType{t},
		}

		// Naming convention: datatype_<lowercase_name>.go
		filename := fmt.Sprintf("datatype_%s.go", strings.ToLower(t.Name))
		path := filepath.Join(c.config.OutputDir, filename)

		if err := writeTemplateFile(path, "datatypes.go.tmpl", data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}
	}

	// Generate datatype_base.go for Element and BackboneElement
	baseTypes := make([]*analyzer.AnalyzedType, 0)
	for _, t := range c.types {
		if t.Name == "Element" || t.Name == "BackboneElement" {
			baseTypes = append(baseTypes, t)
		}
	}

	if len(baseTypes) > 0 {
		data := TypesTemplateData{
			TemplateData: TemplateData{
				PackageName: c.config.PackageName,
				Version:     strings.ToUpper(c.config.Version),
				FileType:    "datatypes",
			},
			Types: baseTypes,
		}

		path := filepath.Join(c.config.OutputDir, "datatype_base.go")
		if err := writeTemplateFile(path, "datatypes.go.tmpl", data); err != nil {
			return fmt.Errorf("failed to generate datatype_base.go: %w", err)
		}
	}

	return nil
}

// generateResourcesSeparately generates one file per resource.
func (c *CodeGen) generateResourcesSeparately() error {
	for _, t := range c.types {
		if t.Kind != kindResource {
			continue
		}

		data := TypesTemplateData{
			TemplateData: TemplateData{
				PackageName: c.config.PackageName,
				Version:     strings.ToUpper(c.config.Version),
				FileType:    "resources",
			},
			Types: []*analyzer.AnalyzedType{t},
		}

		// Naming convention: resource_<lowercase_name>.go
		filename := fmt.Sprintf("resource_%s.go", strings.ToLower(t.Name))
		path := filepath.Join(c.config.OutputDir, filename)

		if err := writeTemplateFile(path, "resources.go.tmpl", data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}
	}

	return nil
}

// generateBackbonesSeparately generates backbone files grouped by parent resource.
func (c *CodeGen) generateBackbonesSeparately() error {
	// Group backbones by parent resource
	backbonesByParent := make(map[string][]*analyzer.AnalyzedType)

	for _, t := range c.types {
		if len(t.BackboneTypes) == 0 {
			continue
		}

		// The parent name is the resource/datatype name
		parentName := t.Name
		backbonesByParent[parentName] = append(backbonesByParent[parentName], t.BackboneTypes...)
	}

	// Generate one file per parent
	for parentName, backbones := range backbonesByParent {
		sort.Slice(backbones, func(i, j int) bool {
			return backbones[i].Name < backbones[j].Name
		})

		data := BackbonesTemplateData{
			TemplateData: TemplateData{
				PackageName: c.config.PackageName,
				Version:     strings.ToUpper(c.config.Version),
				FileType:    "backbones",
			},
			Backbones: backbones,
		}

		// Naming convention: backbone_<lowercase_parent>.go
		filename := fmt.Sprintf("backbone_%s.go", strings.ToLower(parentName))
		path := filepath.Join(c.config.OutputDir, filename)

		if err := writeTemplateFile(path, "backbones.go.tmpl", data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}
	}

	return nil
}

// buildResourceBuilderData converts an AnalyzedType to ResourceBuilderData.
func buildResourceBuilderData(t *analyzer.AnalyzedType) ResourceBuilderData {
	resource := ResourceBuilderData{
		Name:       t.Name,
		LowerName:  toLowerFirstChar(t.Name),
		Properties: make([]PropertyBuilderData, 0, len(t.Properties)),
	}

	for _, prop := range t.Properties {
		propData := PropertyBuilderData{
			Name:      prop.Name,
			GoType:    prop.GoType,
			IsArray:   prop.IsArray,
			IsPointer: prop.IsPointer,
			IsChoice:  prop.IsChoice,
		}

		if prop.IsArray {
			propData.ElementType = strings.TrimPrefix(prop.GoType, "[]")
		}
		if prop.IsPointer {
			propData.BaseType = strings.TrimPrefix(prop.GoType, "*")
		}

		resource.Properties = append(resource.Properties, propData)
	}

	return resource
}

// generateBuildersSeparately generates one fluent builder file per resource.
func (c *CodeGen) generateBuildersSeparately() error {
	for _, t := range c.types {
		if t.Kind != kindResource {
			continue
		}

		data := BuildersTemplateData{
			TemplateData: TemplateData{
				PackageName: c.config.PackageName,
				Version:     strings.ToUpper(c.config.Version),
				FileType:    "builders",
			},
			Resources: []ResourceBuilderData{buildResourceBuilderData(t)},
		}

		filename := fmt.Sprintf("builder_%s.go", strings.ToLower(t.Name))
		path := filepath.Join(c.config.OutputDir, filename)

		if err := writeTemplateFile(path, "fluent_builders.go.tmpl", data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}
	}

	return nil
}

// generateOptionsSeparately generates one functional options file per resource.
func (c *CodeGen) generateOptionsSeparately() error {
	for _, t := range c.types {
		if t.Kind != kindResource {
			continue
		}

		data := BuildersTemplateData{
			TemplateData: TemplateData{
				PackageName: c.config.PackageName,
				Version:     strings.ToUpper(c.config.Version),
				FileType:    "options",
			},
			Resources: []ResourceBuilderData{buildResourceBuilderData(t)},
		}

		filename := fmt.Sprintf("options_%s.go", strings.ToLower(t.Name))
		path := filepath.Join(c.config.OutputDir, filename)

		if err := writeTemplateFile(path, "functional_options.go.tmpl", data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}
	}

	return nil
}

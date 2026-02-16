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

// ResourceConsolidatedData holds data for the consolidated resource template
// (struct + backbones + JSON + XML + builder + options in a single file).
type ResourceConsolidatedData struct {
	TemplateData
	Resource  *analyzer.AnalyzedType
	Backbones []*analyzer.AnalyzedType
	Builder   ResourceBuilderData
}

// DatatypesConsolidatedData holds data for the consolidated datatypes template
// (all datatype structs + backbone structs + XML marshal/unmarshal in a single file).
type DatatypesConsolidatedData struct {
	TemplateData
	Types     []*analyzer.AnalyzedType
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

// ============================================================================
// XML Serialization Generation
// ============================================================================

// xmlTemplateFuncMap returns template functions used by XML templates.
func xmlTemplateFuncMap() template.FuncMap {
	return template.FuncMap{
		// isExtField detects _field extension companion properties (e.g., "_birthDate").
		"isExtField": func(prop analyzer.AnalyzedProperty) bool {
			return strings.HasPrefix(prop.JSONName, "_")
		},

		// xmlPrimitiveFunc maps a Go type to the XML primitive encoding function name.
		"xmlPrimitiveFunc": func(goType string) string {
			baseType := strings.TrimPrefix(goType, "*")
			switch baseType {
			case "string":
				return "xmlEncodePrimitiveString"
			case "bool":
				return "xmlEncodePrimitiveBool"
			case "int":
				return "xmlEncodePrimitiveInt"
			case "int64":
				return "xmlEncodePrimitiveInt64"
			case "uint32":
				return "xmlEncodePrimitiveUint32"
			case "float64":
				return "xmlEncodePrimitiveFloat64"
			default:
				// Custom code type (e.g., *AdministrativeGender, *NarrativeStatus)
				return "xmlEncodePrimitiveCode"
			}
		},

		// extFieldRef returns the extension companion field reference (e.g., "r.BirthDateExt")
		// or "nil" if no extension companion exists.
		"extFieldRef": func(receiver string, prop analyzer.AnalyzedProperty) string {
			if !prop.HasExtension || prop.IsChoice {
				return "nil"
			}
			return fmt.Sprintf("%s.%sExt", receiver, prop.Name)
		},

		// hasIdField checks whether a type has an "id" property.
		"hasIdField": func(t *analyzer.AnalyzedType) bool {
			for _, prop := range t.Properties {
				if prop.JSONName == "id" {
					return true
				}
			}
			return false
		},

		// xmlPrimitiveArrayFunc maps a Go array type to the XML primitive array encoding function name.
		"xmlPrimitiveArrayFunc": func(goType string) string {
			elemType := strings.TrimPrefix(goType, "[]")
			switch elemType {
			case "string":
				return "xmlEncodePrimitiveStringArray"
			case "bool":
				return "xmlEncodePrimitiveBoolArray"
			case "int":
				return "xmlEncodePrimitiveIntArray"
			case "int64":
				return "xmlEncodePrimitiveInt64Array"
			case "uint32":
				return "xmlEncodePrimitiveUint32Array"
			case "float64":
				return "xmlEncodePrimitiveFloat64Array"
			default:
				// Custom code type array (e.g., []ReferenceHandlingPolicy)
				return "xmlEncodePrimitiveCodeArray"
			}
		},

		// xmlPrimitiveDecodeFunc maps a Go type to the XML primitive decode function name.
		"xmlPrimitiveDecodeFunc": func(goType string) string {
			baseType := strings.TrimPrefix(goType, "*")
			switch baseType {
			case "string":
				return "xmlDecodePrimitiveString"
			case "bool":
				return "xmlDecodePrimitiveBool"
			case "int":
				return "xmlDecodePrimitiveInt"
			case "int64":
				return "xmlDecodePrimitiveInt64"
			case "uint32":
				return "xmlDecodePrimitiveUint32"
			case "float64":
				return "xmlDecodePrimitiveFloat64"
			default:
				// Custom code type (e.g., *AdministrativeGender)
				return "xmlDecodePrimitiveCode[" + baseType + "]"
			}
		},

		// elemType extracts the element type from a slice type: "[]Foo" -> "Foo"
		"elemType": func(goType string) string {
			return strings.TrimPrefix(goType, "[]")
		},

		// derefType extracts the base type from a pointer type: "*Foo" -> "Foo"
		"derefType": func(goType string) string {
			return strings.TrimPrefix(goType, "*")
		},

		// hasResourceField checks whether any property of a type has GoType "Resource".
		"hasResourceField": func(t *analyzer.AnalyzedType) bool {
			for _, prop := range t.Properties {
				if prop.GoType == "Resource" {
					return true
				}
			}
			return false
		},

		// resourceFieldName returns the Go field name of the first Resource-typed property.
		"resourceFieldName": func(t *analyzer.AnalyzedType) string {
			for _, prop := range t.Properties {
				if prop.GoType == "Resource" {
					return prop.Name
				}
			}
			return ""
		},
	}
}

// loadTemplateWithFuncs loads a template with custom functions.
func loadTemplateWithFuncs(name string, funcMap template.FuncMap) (*template.Template, error) {
	content, err := templatesFS.ReadFile("templates/" + name)
	if err != nil {
		return nil, fmt.Errorf("failed to read template %s: %w", name, err)
	}

	tmpl, err := template.New(name).Funcs(funcMap).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse template %s: %w", name, err)
	}

	return tmpl, nil
}

// writeXMLTemplateFile executes an XML template with FuncMap and writes to file.
func writeXMLTemplateFile(outputPath, templateName string, data interface{}) error {
	tmpl, err := loadTemplateWithFuncs(templateName, xmlTemplateFuncMap())
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

// generateXMLHelpers generates xml_helpers.go from template.
func (c *CodeGen) generateXMLHelpers() error {
	data := TemplateData{
		PackageName: c.config.PackageName,
		Version:     strings.ToUpper(c.config.Version),
		FileType:    "xml_helpers",
	}

	path := filepath.Join(c.config.OutputDir, "xml_helpers.go")
	return writeTemplateFile(path, "xml_helpers.go.tmpl", data)
}

// ============================================================================
// Consolidated File Generation
// ============================================================================

// generateResourcesConsolidated generates one file per resource containing:
// struct + backbones + JSON marshal/unmarshal + XML marshal/unmarshal.
func (c *CodeGen) generateResourcesConsolidated() error {
	for _, t := range c.types {
		if t.Kind != kindResource {
			continue
		}

		var backbones []*analyzer.AnalyzedType
		if len(t.BackboneTypes) > 0 {
			backbones = make([]*analyzer.AnalyzedType, len(t.BackboneTypes))
			copy(backbones, t.BackboneTypes)
			sort.Slice(backbones, func(i, j int) bool {
				return backbones[i].Name < backbones[j].Name
			})
		}

		data := ResourceConsolidatedData{
			TemplateData: TemplateData{
				PackageName: c.config.PackageName,
				Version:     strings.ToUpper(c.config.Version),
				FileType:    "resource_consolidated",
			},
			Resource:  t,
			Backbones: backbones,
			Builder:   buildResourceBuilderData(t),
		}

		filename := fmt.Sprintf("resource_%s.go", strings.ToLower(t.Name))
		path := filepath.Join(c.config.OutputDir, filename)

		if err := writeXMLTemplateFile(path, "resource_consolidated.go.tmpl", data); err != nil {
			return fmt.Errorf("failed to generate %s: %w", filename, err)
		}
	}

	return nil
}

// generateDatatypesConsolidated generates a single file with all datatypes,
// their backbone elements, and all XML marshal/unmarshal methods.
func (c *CodeGen) generateDatatypesConsolidated() error {
	var allTypes []*analyzer.AnalyzedType
	var allBackbones []*analyzer.AnalyzedType

	// Collect base types first (Element, BackboneElement)
	for _, t := range c.types {
		if t.Name == "Element" || t.Name == "BackboneElement" {
			allTypes = append(allTypes, t)
		}
	}

	// Collect all other datatypes and their backbones
	for _, t := range c.types {
		if t.Kind != "datatype" && t.Kind != "primitive" && t.Kind != "backbone" {
			continue
		}
		if t.Name == "Element" || t.Name == "BackboneElement" {
			continue
		}

		allTypes = append(allTypes, t)

		if len(t.BackboneTypes) > 0 {
			allBackbones = append(allBackbones, t.BackboneTypes...)
		}
	}

	sort.Slice(allBackbones, func(i, j int) bool {
		return allBackbones[i].Name < allBackbones[j].Name
	})

	data := DatatypesConsolidatedData{
		TemplateData: TemplateData{
			PackageName: c.config.PackageName,
			Version:     strings.ToUpper(c.config.Version),
			FileType:    "datatypes_consolidated",
		},
		Types:     allTypes,
		Backbones: allBackbones,
	}

	path := filepath.Join(c.config.OutputDir, "datatypes.go")
	return writeXMLTemplateFile(path, "datatypes_consolidated.go.tmpl", data)
}

// Package generator provides code generation utilities for FHIR types.
package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

// TerminologyCodegen generates embedded ValueSet code from FHIR specifications.
type TerminologyCodegen struct {
	codeSystems map[string]*codeSystemDef
	valueSets   map[string]*valueSetDef
}

type codeSystemDef struct {
	URL     string
	Name    string
	Content string
	Codes   []string
}

type valueSetDef struct {
	URL   string
	Name  string
	Codes []string // Just the code strings, not full CodeInfo
}

// NewTerminologyCodegen creates a new terminology code generator.
func NewTerminologyCodegen() *TerminologyCodegen {
	return &TerminologyCodegen{
		codeSystems: make(map[string]*codeSystemDef),
		valueSets:   make(map[string]*valueSetDef),
	}
}

// LoadFromFile loads ValueSets and CodeSystems from a FHIR bundle file.
func (g *TerminologyCodegen) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return g.LoadFromBundle(data)
}

// LoadFromBundle loads ValueSets and CodeSystems from a FHIR Bundle JSON.
func (g *TerminologyCodegen) LoadFromBundle(data []byte) error {
	var bundle struct {
		ResourceType string `json:"resourceType"`
		Entry        []struct {
			Resource json.RawMessage `json:"resource"`
		} `json:"entry"`
	}

	if err := json.Unmarshal(data, &bundle); err != nil {
		return fmt.Errorf("failed to parse bundle: %w", err)
	}

	// First pass: load all CodeSystems
	for _, entry := range bundle.Entry {
		if entry.Resource == nil {
			continue
		}

		var base struct {
			ResourceType string `json:"resourceType"`
		}
		if err := json.Unmarshal(entry.Resource, &base); err != nil {
			continue
		}

		if base.ResourceType == "CodeSystem" {
			g.loadCodeSystem(entry.Resource)
		}
	}

	// Second pass: load ValueSets and resolve CodeSystem references
	for _, entry := range bundle.Entry {
		if entry.Resource == nil {
			continue
		}

		var base struct {
			ResourceType string `json:"resourceType"`
		}
		if err := json.Unmarshal(entry.Resource, &base); err != nil {
			continue
		}

		if base.ResourceType == "ValueSet" {
			g.loadValueSet(entry.Resource)
		}
	}

	return nil
}

type codeSystemResource struct {
	URL     string              `json:"url"`
	Name    string              `json:"name"`
	Content string              `json:"content"`
	Concept []codeSystemConcept `json:"concept,omitempty"`
}

type codeSystemConcept struct {
	Code    string              `json:"code"`
	Concept []codeSystemConcept `json:"concept,omitempty"`
}

func (g *TerminologyCodegen) loadCodeSystem(data []byte) {
	var cs codeSystemResource
	if err := json.Unmarshal(data, &cs); err != nil || cs.URL == "" {
		return
	}

	// Only load CodeSystems with actual content
	if cs.Content != "complete" && cs.Content != "fragment" {
		return
	}

	codes := flattenCSConcepts(cs.Concept)
	if len(codes) > 0 {
		g.codeSystems[cs.URL] = &codeSystemDef{
			URL:     cs.URL,
			Name:    cs.Name,
			Content: cs.Content,
			Codes:   codes,
		}
	}
}

func flattenCSConcepts(concepts []codeSystemConcept) []string {
	codes := make([]string, 0, len(concepts))
	for _, c := range concepts {
		codes = append(codes, c.Code)
		if len(c.Concept) > 0 {
			codes = append(codes, flattenCSConcepts(c.Concept)...)
		}
	}
	return codes
}

type valueSetResource struct {
	URL       string             `json:"url"`
	Name      string             `json:"name"`
	Compose   *valueSetCompose   `json:"compose,omitempty"`
	Expansion *valueSetExpansion `json:"expansion,omitempty"`
}

type valueSetCompose struct {
	Include []valueSetInclude `json:"include,omitempty"`
}

type valueSetInclude struct {
	System  string            `json:"system,omitempty"`
	Concept []valueSetConcept `json:"concept,omitempty"`
}

type valueSetConcept struct {
	Code string `json:"code"`
}

type valueSetExpansion struct {
	Contains []expansionContains `json:"contains,omitempty"`
}

type expansionContains struct {
	Code string `json:"code,omitempty"`
}

func (g *TerminologyCodegen) loadValueSet(data []byte) {
	var vs valueSetResource
	if err := json.Unmarshal(data, &vs); err != nil || vs.URL == "" {
		return
	}

	var codes []string

	// First try expansion
	if vs.Expansion != nil && len(vs.Expansion.Contains) > 0 {
		for _, c := range vs.Expansion.Contains {
			if c.Code != "" {
				codes = append(codes, c.Code)
			}
		}
	} else if vs.Compose != nil {
		// Otherwise expand from compose
		for _, include := range vs.Compose.Include {
			if len(include.Concept) > 0 {
				// Explicit concepts
				for _, c := range include.Concept {
					codes = append(codes, c.Code)
				}
			} else if cs, ok := g.codeSystems[include.System]; ok {
				// All codes from CodeSystem
				codes = append(codes, cs.Codes...)
			}
		}
	}

	if len(codes) > 0 {
		// Remove duplicates
		codes = uniqueStrings(codes)
		g.valueSets[vs.URL] = &valueSetDef{
			URL:   vs.URL,
			Name:  vs.Name,
			Codes: codes,
		}
	}
}

func uniqueStrings(s []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(s))
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// Stats returns statistics about loaded terminology.
func (g *TerminologyCodegen) Stats() (codeSystems, valueSets, totalCodes int) {
	codeSystems = len(g.codeSystems)
	valueSets = len(g.valueSets)
	for _, vs := range g.valueSets {
		totalCodes += len(vs.Codes)
	}
	return
}

// GenerateEmbeddedValueSets generates Go code with embedded ValueSets.
// If requiredOnly is true, only generates ValueSets that match the provided URLs.
func (g *TerminologyCodegen) GenerateEmbeddedValueSets(packageName, fhirVersion string, requiredURLs []string) ([]byte, error) {
	// Filter ValueSets if required URLs provided
	var valueSetsToGenerate []*valueSetDef
	if len(requiredURLs) > 0 {
		urlSet := make(map[string]bool)
		for _, url := range requiredURLs {
			urlSet[normalizeURL(url)] = true
		}
		for url, vs := range g.valueSets {
			if urlSet[normalizeURL(url)] {
				valueSetsToGenerate = append(valueSetsToGenerate, vs)
			}
		}
	} else {
		for _, vs := range g.valueSets {
			valueSetsToGenerate = append(valueSetsToGenerate, vs)
		}
	}

	// Sort for deterministic output
	sort.Slice(valueSetsToGenerate, func(i, j int) bool {
		return valueSetsToGenerate[i].URL < valueSetsToGenerate[j].URL
	})

	// Generate code
	data := struct {
		Package       string
		FHIRVersion   string
		VersionSuffix string
		ValueSets     []*valueSetDef
		TotalCodes    int
	}{
		Package:       packageName,
		FHIRVersion:   fhirVersion,
		VersionSuffix: versionToSuffix(fhirVersion),
		ValueSets:     valueSetsToGenerate,
	}

	for _, vs := range valueSetsToGenerate {
		data.TotalCodes += len(vs.Codes)
	}

	tmpl, err := template.New("terminology").Parse(terminologyTemplate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.Bytes(), nil
}

func normalizeURL(url string) string {
	if idx := strings.Index(url, "|"); idx != -1 {
		return url[:idx]
	}
	return url
}

func versionToSuffix(fhirVersion string) string {
	switch fhirVersion {
	case "4.0.1":
		return "R4"
	case "4.3.0":
		return "R4B"
	case "5.0.0":
		return "R5"
	default:
		return strings.ReplaceAll(fhirVersion, ".", "_")
	}
}

// WriteToFile generates and writes embedded ValueSets to a file.
func (g *TerminologyCodegen) WriteToFile(outputPath, packageName, fhirVersion string, requiredURLs []string) error {
	code, err := g.GenerateEmbeddedValueSets(packageName, fhirVersion, requiredURLs)
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	//nolint:gosec // Generated code file should be readable
	return os.WriteFile(outputPath, code, 0o644)
}

const terminologyTemplate = `// Code generated by gofhir-codegen. DO NOT EDIT.
// FHIR Version: {{.FHIRVersion}}
// ValueSets: {{len .ValueSets}}
// Total Codes: {{.TotalCodes}}

package {{.Package}}

// embeddedValueSets{{.VersionSuffix}} contains pre-loaded ValueSets for FHIR {{.FHIRVersion}}.
// This eliminates the need to load valuesets.json at runtime.
var embeddedValueSets{{.VersionSuffix}} = map[string]map[string]bool{
{{- range .ValueSets}}
	// {{.Name}}
	"{{.URL}}": {
		{{- range .Codes}}
		"{{.}}": true,
		{{- end}}
	},
{{- end}}
}

func init() {
	registerEmbeddedValueSets("{{.FHIRVersion}}", embeddedValueSets{{.VersionSuffix}})
}
`

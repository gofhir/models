// Package generator implements FHIR to Go code generation.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/gofhir/models/internal/codegen/analyzer"
	"github.com/gofhir/models/internal/codegen/parser"
)

// Config holds code generation configuration.
type Config struct {
	// SpecsDir is the directory containing FHIR specifications
	SpecsDir string
	// OutputDir is the directory to write generated code
	OutputDir string
	// PackageName is the Go package name for generated code
	PackageName string
	// Version is the FHIR version (r4, r4b, r5)
	Version string
}

// CodeGen generates Go code from FHIR specifications.
type CodeGen struct {
	config      Config
	analyzer    *analyzer.Analyzer
	types       []*analyzer.AnalyzedType
	valueSets   *parser.ValueSetRegistry
	rawSDs      []*parser.StructureDefinition // All SDs before filtering, used for hierarchy
	fhirVersion string                        // FHIR release of the loaded specs, e.g. "4.0.1"
}

// New creates a new CodeGen instance.
func New(config Config) *CodeGen {
	return &CodeGen{
		config:    config,
		types:     make([]*analyzer.AnalyzedType, 0),
		valueSets: parser.NewValueSetRegistry(),
	}
}

// LoadTypes loads and analyzes all StructureDefinitions from the specs directory.
func (c *CodeGen) LoadTypes() error {
	specsDir := filepath.Join(c.config.SpecsDir, c.config.Version)

	// Load ValueSets first (needed for binding resolution).
	//
	// This is fatal on purpose. Without valuesets.json every required binding
	// silently degrades from a generated enum to *string: ~205 exported types
	// vanish, the package still compiles, and `go build ./...` reports nothing.
	// A missing or unparseable file must stop generation, not warn.
	valueSetsFile := filepath.Join(specsDir, "valuesets.json")
	data, err := os.ReadFile(valueSetsFile)
	if err != nil {
		return fmt.Errorf("failed to read required value sets from %s: %w", valueSetsFile, err)
	}
	if err = c.valueSets.LoadFromBundle(data); err != nil {
		return fmt.Errorf("failed to load value sets from %s: %w", valueSetsFile, err)
	}
	if n := c.valueSets.Count(); n == 0 {
		return fmt.Errorf("no value sets parsed from %s: refusing to generate code with no enums", valueSetsFile)
	}

	// Collect all StructureDefinitions from both bundles
	var allSDs []*parser.StructureDefinition

	// Load datatypes from profiles-types.json
	typesFile := filepath.Join(specsDir, "profiles-types.json")
	typeSDs, err := c.loadStructureDefinitions(typesFile)
	if err != nil {
		return fmt.Errorf("failed to load types: %w", err)
	}
	allSDs = append(allSDs, typeSDs...)

	// Load resources from profiles-resources.json
	resourcesFile := filepath.Join(specsDir, "profiles-resources.json")
	resourceSDs, err := c.loadStructureDefinitions(resourcesFile)
	if err != nil {
		return fmt.Errorf("failed to load resources: %w", err)
	}
	allSDs = append(allSDs, resourceSDs...)

	// Load all SDs without filtering, needed for the complete type hierarchy
	// (includes abstract types like DomainResource and Resource).
	rawTypeSDs, err := c.loadAllStructureDefinitions(typesFile)
	if err != nil {
		return fmt.Errorf("failed to load raw types: %w", err)
	}
	rawResourceSDs, err := c.loadAllStructureDefinitions(resourcesFile)
	if err != nil {
		return fmt.Errorf("failed to load raw resources: %w", err)
	}
	c.rawSDs = append(c.rawSDs, rawTypeSDs...)
	c.rawSDs = append(c.rawSDs, rawResourceSDs...)

	// Derive the FHIR release from the specs themselves, so generated code never
	// carries a hand-written version string.
	fhirVersion, err := detectFHIRVersion(c.rawSDs)
	if err != nil {
		return fmt.Errorf("failed to determine FHIR version from %s: %w", specsDir, err)
	}
	c.fhirVersion = fhirVersion

	// Create ONE analyzer with ALL definitions and value sets
	c.analyzer = analyzer.NewAnalyzer(allSDs, c.valueSets)

	// Analyze each StructureDefinition
	for _, sd := range allSDs {
		analyzed, err := c.analyzer.Analyze(sd)
		if err != nil {
			// Skip types that fail analysis (e.g., incomplete definitions)
			continue
		}
		c.types = append(c.types, analyzed)
	}

	return nil
}

// loadAllStructureDefinitions loads all StructureDefinitions from a Bundle file
// without any filtering. Used to build a complete type hierarchy that includes
// abstract types like DomainResource and Resource.
func (c *CodeGen) loadAllStructureDefinitions(path string) ([]*parser.StructureDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	bundle, err := parser.ParseBundle(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bundle: %w", err)
	}

	sds, err := parser.ExtractStructureDefinitions(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to extract definitions: %w", err)
	}

	return sds, nil
}

// loadStructureDefinitions loads and filters StructureDefinitions from a Bundle file.
func (c *CodeGen) loadStructureDefinitions(path string) ([]*parser.StructureDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	bundle, err := parser.ParseBundle(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse bundle: %w", err)
	}

	sds, err := parser.ExtractStructureDefinitions(bundle)
	if err != nil {
		return nil, fmt.Errorf("failed to extract definitions: %w", err)
	}

	// Filter StructureDefinitions:
	// - Keep abstract base types: Element, BackboneElement
	// - Skip primitive types (they map to Go builtins)
	// - Skip other abstract types
	var filtered []*parser.StructureDefinition
	for _, sd := range sds {
		// Keep base types even if abstract
		if sd.Name == "Element" || sd.Name == "BackboneElement" {
			filtered = append(filtered, sd)
			continue
		}

		// Skip primitive types - they map to Go builtins
		if sd.Kind == parser.KindPrimitiveType {
			continue
		}

		// Filter out other abstract types
		if !sd.Abstract {
			filtered = append(filtered, sd)
		}
	}

	return filtered, nil
}

// detectFHIRVersion derives the FHIR release (e.g. "4.0.1") from the
// StructureDefinitions being generated from, so that generated code never
// carries a hand-written version string.
//
// Definitions without a fhirVersion are ignored. Disagreement is fatal: a model
// that cannot state its version unambiguously must not be emitted.
func detectFHIRVersion(sds []*parser.StructureDefinition) (string, error) {
	// Track one example definition per distinct value to make conflicts diagnosable.
	examples := make(map[string]string)
	for _, sd := range sds {
		if sd.FHIRVersion == "" {
			continue
		}
		if _, seen := examples[sd.FHIRVersion]; !seen {
			examples[sd.FHIRVersion] = sd.Name
		}
	}

	switch len(examples) {
	case 1:
		for version := range examples {
			return version, nil
		}
	case 0:
		return "", fmt.Errorf("no StructureDefinition declares a fhirVersion (checked %d definitions)", len(sds))
	}

	versions := make([]string, 0, len(examples))
	for version := range examples {
		versions = append(versions, version)
	}
	sort.Strings(versions)

	conflicts := make([]string, 0, len(versions))
	for _, version := range versions {
		conflicts = append(conflicts, fmt.Sprintf("%s (e.g. %s)", version, examples[version]))
	}
	return "", fmt.Errorf("StructureDefinitions declare conflicting fhirVersion values: %s", strings.Join(conflicts, ", "))
}

// Generate writes all generated code to the output directory.
func (c *CodeGen) Generate() error {
	if err := os.MkdirAll(c.config.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate decimal.go (FHIR decimal type with precision preservation)
	if err := c.generateDecimalType(); err != nil {
		return fmt.Errorf("failed to generate decimal type: %w", err)
	}

	// Generate interfaces.go (shared interfaces, small file)
	if err := c.generateInterfacesFromTemplate(); err != nil {
		return fmt.Errorf("failed to generate interfaces: %w", err)
	}

	// Generate integer64.go (R5 only: the FHIR integer64 primitive)
	if err := c.generateInteger64FromTemplate(); err != nil {
		return fmt.Errorf("failed to generate integer64: %w", err)
	}

	// Generate helpers.go (generic pointer/value helpers)
	if err := c.generateHelpersFromTemplate(); err != nil {
		return fmt.Errorf("failed to generate helpers: %w", err)
	}

	// Generate doc.go (package documentation)
	if err := c.generateDocFromTemplate(); err != nil {
		return fmt.Errorf("failed to generate doc: %w", err)
	}

	// Generate marshal.go (JSON marshaling entry points)
	if err := c.generateMarshalFromTemplate(); err != nil {
		return fmt.Errorf("failed to generate marshal: %w", err)
	}

	// Generate registry.go (resource factories and unmarshal functions)
	if err := c.generateRegistryFromTemplate(); err != nil {
		return fmt.Errorf("failed to generate registry: %w", err)
	}

	// Generate codesystems.go (types used by datatypes and resources)
	if err := c.generateCodeSystemsFromTemplate(); err != nil {
		return fmt.Errorf("failed to generate codesystems: %w", err)
	}

	// Generate summary.go (summary fields per resource type)
	if err := c.generateSummaryFromTemplate(); err != nil {
		return fmt.Errorf("failed to generate summary: %w", err)
	}

	// Generate fhirpath_model.go (runtime metadata for FHIRPath evaluation)
	if err := c.generateFHIRPathModel(); err != nil {
		return fmt.Errorf("failed to generate fhirpath model: %w", err)
	}

	// Generate consolidated datatypes (all structs + backbones + XML in one file)
	if err := c.generateDatatypesConsolidated(); err != nil {
		return fmt.Errorf("failed to generate datatypes: %w", err)
	}

	// Generate consolidated resources (one file per resource: struct + backbones + JSON + XML + builder + options)
	if err := c.generateResourcesConsolidated(); err != nil {
		return fmt.Errorf("failed to generate resources: %w", err)
	}

	// Generate XML helpers (shared encoding/decoding utilities)
	if err := c.generateXMLHelpers(); err != nil {
		return fmt.Errorf("failed to generate XML helpers: %w", err)
	}

	return nil
}

// toPascalCaseCode converts a code value to PascalCase for use as a constant name.
func toPascalCaseCode(code string) string {
	// Handle special symbol codes first
	symbolMap := map[string]string{
		"<":  "LessThan",
		"<=": "LessOrEqual",
		">":  "GreaterThan",
		">=": "GreaterOrEqual",
		"=":  "Equal",
		"!=": "NotEqual",
		"+":  "Plus",
		"-":  "Minus",
		"*":  "Asterisk",
		"/":  "Slash",
		"#":  "Hash",
		"&":  "Ampersand",
		"|":  "Pipe",
	}

	if replacement, ok := symbolMap[code]; ok {
		return replacement
	}

	// Handle common patterns
	code = strings.ReplaceAll(code, "-", " ")
	code = strings.ReplaceAll(code, "_", " ")
	code = strings.ReplaceAll(code, ".", " ")
	code = strings.ReplaceAll(code, "/", " ")
	code = strings.ReplaceAll(code, "(", " ")
	code = strings.ReplaceAll(code, ")", " ")

	words := strings.Fields(code)
	for i, word := range words {
		if word != "" {
			words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
		}
	}
	return strings.Join(words, "")
}

// Package generator implements FHIR to Go code generation.
package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/gofhir/fhir/internal/codegen/analyzer"
	"github.com/gofhir/fhir/internal/codegen/parser"
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
	config       Config
	analyzer     *analyzer.Analyzer
	types        []*analyzer.AnalyzedType
	valueSets    *parser.ValueSetRegistry
	usedBindings map[string]bool // Track which bindings are actually used
}

// New creates a new CodeGen instance.
func New(config Config) *CodeGen {
	return &CodeGen{
		config:       config,
		types:        make([]*analyzer.AnalyzedType, 0),
		valueSets:    parser.NewValueSetRegistry(),
		usedBindings: make(map[string]bool),
	}
}

// LoadTypes loads and analyzes all StructureDefinitions from the specs directory.
func (c *CodeGen) LoadTypes() error {
	specsDir := filepath.Join(c.config.SpecsDir, c.config.Version)

	// Load ValueSets first (needed for binding resolution)
	valueSetsFile := filepath.Join(specsDir, "valuesets.json")
	if data, err := os.ReadFile(valueSetsFile); err == nil {
		if err := c.valueSets.LoadFromBundle(data); err != nil {
			// Non-fatal: continue without value sets
			fmt.Printf("Warning: failed to load value sets: %v\n", err)
		}
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

// Generate writes all generated code to the output directory.
func (c *CodeGen) Generate() error {
	if err := os.MkdirAll(c.config.OutputDir, 0o755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Generate interfaces.go (shared interfaces, small file)
	if err := c.generateInterfacesFromTemplate(); err != nil {
		return fmt.Errorf("failed to generate interfaces: %w", err)
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

	// NEW: Generate separate files for datatypes (one file per datatype)
	if err := c.generateDatatypesSeparately(); err != nil {
		return fmt.Errorf("failed to generate datatypes: %w", err)
	}

	// NEW: Generate separate files for resources (one file per resource)
	if err := c.generateResourcesSeparately(); err != nil {
		return fmt.Errorf("failed to generate resources: %w", err)
	}

	// NEW: Generate separate backbone files (grouped by parent resource)
	if err := c.generateBackbonesSeparately(); err != nil {
		return fmt.Errorf("failed to generate backbones: %w", err)
	}

	// NEW: Generate separate builder files (one per resource)
	if err := c.generateBuildersSeparately(); err != nil {
		return fmt.Errorf("failed to generate builders: %w", err)
	}

	// NEW: Generate separate option files (one per resource)
	if err := c.generateOptionsSeparately(); err != nil {
		return fmt.Errorf("failed to generate options: %w", err)
	}

	return nil
}

// sanitizeTypeName converts a ValueSet name to a valid Go type name.
func sanitizeTypeName(name string) string {
	// Remove/replace invalid characters
	name = strings.ReplaceAll(name, " ", "")
	name = strings.ReplaceAll(name, "-", "")
	name = strings.ReplaceAll(name, "_", "")
	name = strings.ReplaceAll(name, ".", "")
	name = strings.ReplaceAll(name, "(", "")
	name = strings.ReplaceAll(name, ")", "")
	name = strings.ReplaceAll(name, "/", "")

	// Ensure first character is uppercase
	if name != "" {
		runes := []rune(name)
		r, _ := utf8.DecodeRuneInString(strings.ToUpper(string(runes[0])))
		runes[0] = r
		name = string(runes)
	}

	return name
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

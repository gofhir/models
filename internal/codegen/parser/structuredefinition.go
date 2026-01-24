// Package parser provides parsing of FHIR StructureDefinition JSON files.
package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// StructureDefinition represents a FHIR StructureDefinition resource.
// This is a simplified version containing only the fields needed for code generation.
type StructureDefinition struct {
	ResourceType   string        `json:"resourceType"`
	ID             string        `json:"id"`
	URL            string        `json:"url"`
	Version        string        `json:"version"`
	Name           string        `json:"name"`
	Title          string        `json:"title"`
	Status         string        `json:"status"`
	Kind           string        `json:"kind"` // primitive-type, complex-type, resource, logical
	Abstract       bool          `json:"abstract"`
	Type           string        `json:"type"`
	BaseDefinition string        `json:"baseDefinition"`
	Derivation     string        `json:"derivation"` // specialization, constraint
	Snapshot       *Snapshot     `json:"snapshot"`
	Differential   *Differential `json:"differential"`
}

// Snapshot contains the complete list of elements for this definition.
type Snapshot struct {
	Element []ElementDefinition `json:"element"`
}

// Differential contains only the differences from the base definition.
type Differential struct {
	Element []ElementDefinition `json:"element"`
}

// ElementDefinition defines a single element within a StructureDefinition.
type ElementDefinition struct {
	ID               string           `json:"id"`
	Path             string           `json:"path"`
	SliceName        string           `json:"sliceName,omitempty"`
	Short            string           `json:"short"`
	Definition       string           `json:"definition"`
	Comment          string           `json:"comment,omitempty"`
	Min              int              `json:"min"`
	Max              string           `json:"max"` // "0", "1", "*", or a number
	Base             *Base            `json:"base,omitempty"`
	Type             []TypeRef        `json:"type,omitempty"`
	ContentReference string           `json:"contentReference,omitempty"`
	Binding          *Binding         `json:"binding,omitempty"`
	Constraint       []Constraint     `json:"constraint,omitempty"`
	MustSupport      bool             `json:"mustSupport,omitempty"`
	IsModifier       bool             `json:"isModifier,omitempty"`
	IsSummary        bool             `json:"isSummary,omitempty"`
	Fixed            json.RawMessage  `json:"fixed,omitempty"`   // fixed[x]
	Pattern          json.RawMessage  `json:"pattern,omitempty"` // pattern[x]
	Example          []Example        `json:"example,omitempty"`
	MinValue         json.RawMessage  `json:"minValue,omitempty"`
	MaxValue         json.RawMessage  `json:"maxValue,omitempty"`
	MaxLength        *int             `json:"maxLength,omitempty"`
	Condition        []string         `json:"condition,omitempty"`
	Mapping          []ElementMapping `json:"mapping,omitempty"`
}

// Base contains information about the base element this is derived from.
type Base struct {
	Path string `json:"path"`
	Min  int    `json:"min"`
	Max  string `json:"max"`
}

// TypeRef defines a type that an element can have.
type TypeRef struct {
	Code          string   `json:"code"`
	Profile       []string `json:"profile,omitempty"`
	TargetProfile []string `json:"targetProfile,omitempty"`
	Aggregation   []string `json:"aggregation,omitempty"`
	Versioning    string   `json:"versioning,omitempty"`
}

// Binding describes the binding to a value set.
type Binding struct {
	Strength    string `json:"strength"` // required, extensible, preferred, example
	Description string `json:"description,omitempty"`
	ValueSet    string `json:"valueSet,omitempty"`
}

// Constraint represents a FHIRPath constraint on an element.
type Constraint struct {
	Key          string `json:"key"`
	Requirements string `json:"requirements,omitempty"`
	Severity     string `json:"severity"` // error, warning
	Human        string `json:"human"`
	Expression   string `json:"expression,omitempty"`
	XPath        string `json:"xpath,omitempty"`
	Source       string `json:"source,omitempty"`
}

// Example provides an example value for an element.
type Example struct {
	Label string          `json:"label"`
	Value json.RawMessage `json:"value"`
}

// ElementMapping maps an element to another specification.
type ElementMapping struct {
	Identity string `json:"identity"`
	Language string `json:"language,omitempty"`
	Map      string `json:"map"`
	Comment  string `json:"comment,omitempty"`
}

// Bundle represents a FHIR Bundle containing multiple resources.
type Bundle struct {
	ResourceType string        `json:"resourceType"`
	ID           string        `json:"id,omitempty"`
	Type         string        `json:"type"`
	Entry        []BundleEntry `json:"entry,omitempty"`
}

// BundleEntry is a single entry in a Bundle.
type BundleEntry struct {
	FullURL  string          `json:"fullUrl,omitempty"`
	Resource json.RawMessage `json:"resource,omitempty"`
}

// ParseStructureDefinition parses a single StructureDefinition from JSON data.
func ParseStructureDefinition(data []byte) (*StructureDefinition, error) {
	var sd StructureDefinition
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, fmt.Errorf("failed to parse StructureDefinition: %w", err)
	}
	if sd.ResourceType != ResourceTypeStructureDefinition {
		return nil, fmt.Errorf("expected resourceType '%s', got '%s'", ResourceTypeStructureDefinition, sd.ResourceType)
	}
	return &sd, nil
}

// ParseStructureDefinitionFile parses a StructureDefinition from a file.
func ParseStructureDefinitionFile(path string) (*StructureDefinition, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", path, err)
	}
	return ParseStructureDefinition(data)
}

// ParseBundle parses a Bundle from JSON data.
func ParseBundle(data []byte) (*Bundle, error) {
	var b Bundle
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("failed to parse Bundle: %w", err)
	}
	if b.ResourceType != ResourceTypeBundle {
		return nil, fmt.Errorf("expected resourceType '%s', got '%s'", ResourceTypeBundle, b.ResourceType)
	}
	return &b, nil
}

// ExtractStructureDefinitions extracts all StructureDefinitions from a Bundle.
func ExtractStructureDefinitions(bundle *Bundle) ([]*StructureDefinition, error) {
	results := make([]*StructureDefinition, 0, len(bundle.Entry))

	for i, entry := range bundle.Entry {
		if len(entry.Resource) == 0 {
			continue
		}

		// Check if this is a StructureDefinition
		var peek struct {
			ResourceType string `json:"resourceType"`
		}
		if err := json.Unmarshal(entry.Resource, &peek); err != nil {
			continue
		}

		if peek.ResourceType != ResourceTypeStructureDefinition {
			continue
		}

		sd, err := ParseStructureDefinition(entry.Resource)
		if err != nil {
			return nil, fmt.Errorf("failed to parse entry %d: %w", i, err)
		}
		results = append(results, sd)
	}

	return results, nil
}

// LoadStructureDefinitionsFromDir loads all StructureDefinitions from a directory.
// It looks for both individual JSON files and Bundle files.
func LoadStructureDefinitionsFromDir(dir string) ([]*StructureDefinition, error) {
	var results []*StructureDefinition
	seen := make(map[string]bool)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read %s: %w", path, err)
		}

		// Check the resource type
		var peek struct {
			ResourceType string `json:"resourceType"`
		}
		if err := json.Unmarshal(data, &peek); err != nil {
			// Skip invalid JSON files
			return nil
		}

		switch peek.ResourceType {
		case ResourceTypeStructureDefinition:
			sd, err := ParseStructureDefinition(data)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}
			if !seen[sd.URL] {
				seen[sd.URL] = true
				results = append(results, sd)
			}

		case ResourceTypeBundle:
			bundle, err := ParseBundle(data)
			if err != nil {
				return fmt.Errorf("failed to parse bundle %s: %w", path, err)
			}
			sds, err := ExtractStructureDefinitions(bundle)
			if err != nil {
				return fmt.Errorf("failed to extract from bundle %s: %w", path, err)
			}
			for _, sd := range sds {
				if !seen[sd.URL] {
					seen[sd.URL] = true
					results = append(results, sd)
				}
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return results, nil
}

// FilterByKind filters StructureDefinitions by their Kind.
func FilterByKind(sds []*StructureDefinition, kinds ...string) []*StructureDefinition {
	kindSet := make(map[string]bool)
	for _, k := range kinds {
		kindSet[k] = true
	}

	var results []*StructureDefinition
	for _, sd := range sds {
		if kindSet[sd.Kind] {
			results = append(results, sd)
		}
	}
	return results
}

// FilterNonAbstract filters out abstract StructureDefinitions.
func FilterNonAbstract(sds []*StructureDefinition) []*StructureDefinition {
	var results []*StructureDefinition
	for _, sd := range sds {
		if !sd.Abstract {
			results = append(results, sd)
		}
	}
	return results
}

// Resource type constants.
const (
	ResourceTypeStructureDefinition = "StructureDefinition"
	ResourceTypeBundle              = "Bundle"
)

// Kind constants for StructureDefinition.
const (
	KindPrimitiveType = "primitive-type"
	KindComplexType   = "complex-type"
	KindResource      = "resource"
	KindLogical       = "logical"
)

// IsPrimitive returns true if this is a primitive type definition.
func (sd *StructureDefinition) IsPrimitive() bool {
	return sd.Kind == KindPrimitiveType
}

// IsComplexType returns true if this is a complex type (datatype) definition.
func (sd *StructureDefinition) IsComplexType() bool {
	return sd.Kind == KindComplexType
}

// IsResource returns true if this is a resource definition.
func (sd *StructureDefinition) IsResource() bool {
	return sd.Kind == KindResource
}

// GetElements returns the elements from Snapshot, or Differential if Snapshot is nil.
func (sd *StructureDefinition) GetElements() []ElementDefinition {
	if sd.Snapshot != nil && len(sd.Snapshot.Element) > 0 {
		return sd.Snapshot.Element
	}
	if sd.Differential != nil {
		return sd.Differential.Element
	}
	return nil
}

// IsChoiceType returns true if this element is a choice type (ends with [x]).
func (ed *ElementDefinition) IsChoiceType() bool {
	return strings.HasSuffix(ed.Path, "[x]")
}

// GetBaseName returns the element name without the [x] suffix for choice types.
func (ed *ElementDefinition) GetBaseName() string {
	path := ed.Path
	if idx := strings.LastIndex(path, "."); idx >= 0 {
		path = path[idx+1:]
	}
	return strings.TrimSuffix(path, "[x]")
}

// IsRequired returns true if this element has min >= 1.
func (ed *ElementDefinition) IsRequired() bool {
	return ed.Min >= 1
}

// IsArray returns true if this element can have multiple values.
func (ed *ElementDefinition) IsArray() bool {
	return ed.Max == "*" || (ed.Max != "" && ed.Max != "0" && ed.Max != "1")
}

// IsBackboneElement returns true if this element defines a backbone element.
func (ed *ElementDefinition) IsBackboneElement() bool {
	for _, t := range ed.Type {
		if t.Code == "BackboneElement" || t.Code == "Element" {
			return true
		}
	}
	return false
}

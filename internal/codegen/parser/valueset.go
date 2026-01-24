// Package parser provides FHIR specification parsing utilities.
package parser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ValueSet represents a FHIR ValueSet resource.
type ValueSet struct {
	ResourceType string           `json:"resourceType"`
	ID           string           `json:"id"`
	URL          string           `json:"url"`
	Name         string           `json:"name"`
	Title        string           `json:"title"`
	Status       string           `json:"status"`
	Compose      *ValueSetCompose `json:"compose,omitempty"`
}

// ValueSetCompose defines the content of the value set.
type ValueSetCompose struct {
	Include []ValueSetInclude `json:"include,omitempty"`
}

// ValueSetInclude specifies which codes are included.
type ValueSetInclude struct {
	System  string            `json:"system,omitempty"`
	Concept []ValueSetConcept `json:"concept,omitempty"`
}

// ValueSetConcept represents a code in the value set.
type ValueSetConcept struct {
	Code    string `json:"code"`
	Display string `json:"display,omitempty"`
}

// CodeSystem represents a FHIR CodeSystem resource.
type CodeSystem struct {
	ResourceType string              `json:"resourceType"`
	ID           string              `json:"id"`
	URL          string              `json:"url"`
	Name         string              `json:"name"`
	Title        string              `json:"title"`
	Status       string              `json:"status"`
	Content      string              `json:"content"`
	Concept      []CodeSystemConcept `json:"concept,omitempty"`
}

// CodeSystemConcept represents a concept in a code system.
type CodeSystemConcept struct {
	Code       string              `json:"code"`
	Display    string              `json:"display,omitempty"`
	Definition string              `json:"definition,omitempty"`
	Concept    []CodeSystemConcept `json:"concept,omitempty"` // Nested concepts
}

// ParsedValueSet represents a processed value set ready for code generation.
type ParsedValueSet struct {
	URL   string // Canonical URL
	Name  string // Name for Go type
	Title string // Human-readable title
	Codes []ParsedCode
}

// ParsedCode represents a single code value.
type ParsedCode struct {
	Code    string // The actual code value
	Display string // Human-readable display
}

// ValueSetRegistry holds parsed value sets indexed by URL.
type ValueSetRegistry struct {
	valueSets   map[string]*ParsedValueSet
	codeSystems map[string]*CodeSystem
}

// NewValueSetRegistry creates a new registry.
func NewValueSetRegistry() *ValueSetRegistry {
	return &ValueSetRegistry{
		valueSets:   make(map[string]*ParsedValueSet),
		codeSystems: make(map[string]*CodeSystem),
	}
}

// LoadFromBundle loads ValueSets and CodeSystems from a bundle.
func (r *ValueSetRegistry) LoadFromBundle(data []byte) error {
	var bundle Bundle
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
			var cs CodeSystem
			if err := json.Unmarshal(entry.Resource, &cs); err != nil {
				continue
			}
			r.codeSystems[cs.URL] = &cs
		}
	}

	// Second pass: load ValueSets and resolve references
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
			var vs ValueSet
			if err := json.Unmarshal(entry.Resource, &vs); err != nil {
				continue
			}

			parsed := r.parseValueSet(&vs)
			if parsed != nil && len(parsed.Codes) > 0 {
				r.valueSets[vs.URL] = parsed
			}
		}
	}

	return nil
}

// parseValueSet converts a ValueSet to a ParsedValueSet.
func (r *ValueSetRegistry) parseValueSet(vs *ValueSet) *ParsedValueSet {
	parsed := &ParsedValueSet{
		URL:   vs.URL,
		Name:  vs.Name,
		Title: vs.Title,
	}

	if vs.Compose == nil {
		return parsed
	}

	for _, include := range vs.Compose.Include {
		// If concepts are explicitly listed
		if len(include.Concept) > 0 {
			for _, c := range include.Concept {
				parsed.Codes = append(parsed.Codes, ParsedCode(c))
			}
			continue
		}

		// Otherwise, try to resolve from CodeSystem
		if cs, ok := r.codeSystems[include.System]; ok {
			codes := r.flattenConcepts(cs.Concept)
			parsed.Codes = append(parsed.Codes, codes...)
		}
	}

	return parsed
}

// flattenConcepts recursively flattens nested concepts.
func (r *ValueSetRegistry) flattenConcepts(concepts []CodeSystemConcept) []ParsedCode {
	codes := make([]ParsedCode, 0, len(concepts))
	for _, c := range concepts {
		codes = append(codes, ParsedCode{
			Code:    c.Code,
			Display: c.Display,
		})
		// Recursively add nested concepts
		if len(c.Concept) > 0 {
			codes = append(codes, r.flattenConcepts(c.Concept)...)
		}
	}
	return codes
}

// Get returns a parsed value set by URL (handles versioned URLs).
func (r *ValueSetRegistry) Get(url string) *ParsedValueSet {
	// Try exact match first
	if vs, ok := r.valueSets[url]; ok {
		return vs
	}

	// Try without version suffix (e.g., "http://....|4.0.1" -> "http://....")
	if idx := strings.Index(url, "|"); idx != -1 {
		baseURL := url[:idx]
		if vs, ok := r.valueSets[baseURL]; ok {
			return vs
		}
	}

	return nil
}

// Count returns the number of loaded value sets.
func (r *ValueSetRegistry) Count() int {
	return len(r.valueSets)
}

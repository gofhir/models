// Package parser provides FHIR specification parsing utilities.
package parser

import (
	"encoding/json"
	"fmt"
	"sort"
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
//
// ValueSet is the composition case: an include may name other ValueSets instead
// of listing codes, and the result is their contents. Ignoring it resolved a
// composed ValueSet to whichever part happened to be written inline —
// version-independent-all-resource-types came out as 41 obsolete type names
// with no Patient among them, because its other half is a reference.
type ValueSetInclude struct {
	System   string            `json:"system,omitempty"`
	Concept  []ValueSetConcept `json:"concept,omitempty"`
	ValueSet []string          `json:"valueSet,omitempty"`
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
	// System is the CodeSystem the code belongs to. A code on its own is not a
	// coding — "male" means nothing without saying which vocabulary it is from —
	// so this is what lets a generated enum produce a usable Coding.
	System string
}

// ValueSetRegistry holds parsed value sets indexed by URL.
type ValueSetRegistry struct {
	valueSets   map[string]*ParsedValueSet
	codeSystems map[string]*CodeSystem
	// raw keeps the unparsed ValueSets so an include that names another one can
	// be followed. Bundle order does not put a target before the ValueSet that
	// references it, so resolution cannot happen while loading.
	raw map[string]*ValueSet
}

// NewValueSetRegistry creates a new registry.
func NewValueSetRegistry() *ValueSetRegistry {
	return &ValueSetRegistry{
		valueSets:   make(map[string]*ParsedValueSet),
		codeSystems: make(map[string]*CodeSystem),
		raw:         make(map[string]*ValueSet),
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
			r.raw[vs.URL] = &vs
		}
	}

	// Third pass: resolve. An include naming another ValueSet needs every raw
	// ValueSet already in hand, which the second pass cannot promise.
	for url, vs := range r.raw {
		parsed := r.parseValueSet(vs)
		if parsed != nil && len(parsed.Codes) > 0 {
			r.valueSets[url] = parsed
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
	// One code becomes one Go constant, and the name comes from the code alone.
	// Composition makes duplicates ordinary — a set that includes another one may
	// restate a code the other already has — and two constants of the same name
	// do not compile. First occurrence wins.
	r.collectCodes(vs, parsed, make(map[string]bool), make(map[string]bool))
	return parsed
}

// collectCodes appends vs's codes to parsed, following includes that name other
// ValueSets. visited stops a cycle: nothing in the published specs has one, but a
// composed ValueSet is a graph and a self-reference would otherwise not return.
func (r *ValueSetRegistry) collectCodes(vs *ValueSet, parsed *ParsedValueSet, visited, seen map[string]bool) {
	if vs == nil || vs.Compose == nil {
		return
	}
	if visited[vs.URL] {
		return
	}
	visited[vs.URL] = true

	add := func(codes []ParsedCode) {
		for _, c := range codes {
			if c.Code == "" || seen[c.Code] {
				continue
			}
			seen[c.Code] = true
			parsed.Codes = append(parsed.Codes, c)
		}
	}

	for _, include := range vs.Compose.Include {
		// Codes listed inline.
		if len(include.Concept) > 0 {
			for _, c := range include.Concept {
				add([]ParsedCode{{Code: c.Code, Display: c.Display, System: include.System}})
			}
		} else if cs, ok := r.codeSystems[include.System]; ok {
			// The whole CodeSystem.
			add(r.flattenConcepts(cs.Concept, cs.URL))
		}

		// And whatever other ValueSets this one is built from.
		for _, ref := range include.ValueSet {
			if target := r.rawValueSet(ref); target != nil {
				r.collectCodes(target, parsed, visited, seen)
			}
		}
	}
}

// rawValueSet looks up an unparsed ValueSet, tolerating a version suffix the way
// Get does: a reference may be written as "…/ValueSet/all-types|4.0.1".
func (r *ValueSetRegistry) rawValueSet(url string) *ValueSet {
	if vs, ok := r.raw[url]; ok {
		return vs
	}
	if i := strings.Index(url, "|"); i > 0 {
		if vs, ok := r.raw[url[:i]]; ok {
			return vs
		}
	}
	return nil
}

// flattenConcepts recursively flattens nested concepts.
func (r *ValueSetRegistry) flattenConcepts(concepts []CodeSystemConcept, system string) []ParsedCode {
	codes := make([]ParsedCode, 0, len(concepts))
	for _, c := range concepts {
		codes = append(codes, ParsedCode{
			Code:    c.Code,
			Display: c.Display,
			System:  system,
		})
		// Recursively add nested concepts
		if len(c.Concept) > 0 {
			codes = append(codes, r.flattenConcepts(c.Concept, system)...)
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

// All returns every parsed ValueSet, ordered by canonical URL so callers that
// derive names or detect collisions get the same answer on every run.
func (r *ValueSetRegistry) All() []*ParsedValueSet {
	urls := make([]string, 0, len(r.valueSets))
	for url := range r.valueSets {
		urls = append(urls, url)
	}
	sort.Strings(urls)

	out := make([]*ParsedValueSet, 0, len(urls))
	for _, url := range urls {
		out = append(out, r.valueSets[url])
	}
	return out
}

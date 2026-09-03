// Package analyzer analyzes FHIR StructureDefinitions and determines Go types.
package analyzer

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/gofhir/models/internal/codegen/parser"
)

// Kind constants for type categorization.
const (
	kindDatatype = "datatype"
	kindBackbone = "backbone"
)

// maxEnumCodes caps how large a ValueSet may be before its binding falls back to
// a plain string. resource-types, all-types, mimetypes and currencies are the ones
// this excludes: emitting hundreds of constants for them costs more than it buys.
//
// Shared with the collision index, so both agree on which ValueSets could ever
// become a Go type.
const maxEnumCodes = 100

// Analyzer processes StructureDefinitions and produces analyzed types for code generation.
type Analyzer struct {
	definitions  map[string]*parser.StructureDefinition
	valueSets    *parser.ValueSetRegistry
	UsedBindings map[string]bool // Track which bindings are used (exported for generator)
	// valueSetNameClaims maps a sanitized Go type name to the canonical URLs of
	// the ValueSets that would claim it, so a genuine collision is
	// distinguishable from a name that is simply in use.
	valueSetNameClaims map[string][]string
	// bindingNames maps a canonical ValueSet URL to the name HL7 gives that
	// binding, when every element binding to it agrees on one usable name. See
	// collectBindingNames for why agreement is required.
	bindingNames map[string]string
}

// NewAnalyzer creates a new Analyzer with the given StructureDefinitions and ValueSets.
func NewAnalyzer(definitions []*parser.StructureDefinition, valueSets *parser.ValueSetRegistry) *Analyzer {
	defMap := make(map[string]*parser.StructureDefinition)
	for _, sd := range definitions {
		defMap[sd.URL] = sd
		defMap[sd.Name] = sd
		defMap[sd.Type] = sd
	}
	a := &Analyzer{
		definitions:  defMap,
		valueSets:    valueSets,
		UsedBindings: make(map[string]bool),
		bindingNames: collectBindingNames(definitions),
	}
	// Claims are computed from the names that will actually be emitted, so a
	// collision is judged on the final name rather than on the ValueSet title
	// that a binding name may have replaced.
	a.dropLessSpecificBindingNames(valueSets)
	a.dropBindingNamesClaimedByTypes(definitions, valueSets)
	a.dropCollidingBindingNames(valueSets)
	a.valueSetNameClaims = a.buildValueSetNameClaims(valueSets)
	return a
}

// misspelledBindingNames lists binding names that are typos in the published
// specification, and are therefore not used.
//
// This is deliberately a list of individual upstream defects rather than a rule:
// each one is a specific string HL7 got wrong, verifiable against the element it
// annotates, and there is no general test that would catch it.
var misspelledBindingNames = map[string]string{
	// ConceptMap.additionalAttribute.type in R5 — "Map" appears twice, the
	// second time in lower case. The title-derived ConceptMapAttributeType is
	// what the binding name was evidently meant to say.
	"ConceptMapmapAttributeType": "ConceptMap.additionalAttribute.type (R5)",
}

// dropBindingNamesClaimedByTypes removes binding names that a resource or
// datatype already answers to.
//
// The enums share a package with the 445 resource types, and nothing in the
// specification stops a binding being named after a resource: in R4B and R5 the
// subscription-status ValueSet is bound as SubscriptionStatus, which is also a
// resource. Taking the binding name there produces two declarations of one
// identifier and the package stops compiling.
//
// Only the ValueSet gives way. A resource name is fixed by the specification and
// is what callers write, so it is not available to be renamed.
func (a *Analyzer) dropBindingNamesClaimedByTypes(definitions []*parser.StructureDefinition, valueSets *parser.ValueSetRegistry) {
	if valueSets == nil || len(a.bindingNames) == 0 {
		return
	}
	taken := make(map[string]bool, len(definitions))
	for _, sd := range definitions {
		if sd == nil {
			continue
		}
		taken[sanitizeTypeName(sd.Name)] = true
		taken[sanitizeTypeName(sd.Type)] = true
	}
	for url, bound := range a.bindingNames {
		if taken[bound] {
			delete(a.bindingNames, url)
		}
	}
}

// dropLessSpecificBindingNames removes binding names that lose context.
//
// A binding name is usually the more precise of the two — MedicationStatusCodes
// against MedicationStatus. Sometimes it is the reverse: the artifact-assessment
// bindings are named Disposition, InformationType and WorkflowStatus, and
// verificationresult-status is named plainly Status. Those read as a general
// vocabulary rather than one belonging to a resource, and an exported Status in a
// package holding hundreds of enums is worse than the name it would replace.
//
// The test for it is structural rather than a judgement call: the binding name is
// a suffix of the title-derived name, so it says strictly less about what the
// ValueSet is. Where that holds the longer name wins.
func (a *Analyzer) dropLessSpecificBindingNames(valueSets *parser.ValueSetRegistry) {
	if valueSets == nil || len(a.bindingNames) == 0 {
		return
	}
	for _, vs := range valueSets.All() {
		url := canonicalValueSetURL(vs.URL)
		bound, ok := a.bindingNames[url]
		if !ok {
			continue
		}
		if title := sanitizeTypeName(vs.Name); title != bound && strings.HasSuffix(title, bound) {
			delete(a.bindingNames, url)
		}
	}
}

// dropCollidingBindingNames removes binding names that would introduce a new
// collision.
//
// A binding name is an improvement, not an obligation: it turns
// MedicationStatusCodes into MedicationStatus. But it changes which names are in
// play, and can make two ValueSets that were distinct under their titles land on
// the same one — R5 annotates Specimen.combined, whose codes are grouped and
// pooled, with bindingName="PublicationStatus", a name belonging to an unrelated
// ValueSet. Where that happens the binding name is dropped and both keep their
// title-derived names, which were already unique.
//
// Falling back is better than adding a manual override for each case: the
// override list should hold genuine upstream naming clashes, not ones this
// change introduced.
func (a *Analyzer) dropCollidingBindingNames(valueSets *parser.ValueSetRegistry) {
	if valueSets == nil || len(a.bindingNames) == 0 {
		return
	}

	// Every name that will be in play, and who claims it.
	claimants := make(map[string][]string)
	for _, vs := range valueSets.All() {
		if len(vs.Codes) == 0 || len(vs.Codes) > maxEnumCodes {
			continue
		}
		url := canonicalValueSetURL(vs.URL)
		name := sanitizeTypeName(vs.Name)
		if bound, ok := a.bindingNames[url]; ok {
			name = bound
		}
		claimants[name] = append(claimants[name], url)
	}

	for name, urls := range claimants {
		if len(urls) < 2 {
			continue
		}
		// Only give up the binding name where it is what caused the clash.
		for _, url := range urls {
			if a.bindingNames[url] == name {
				delete(a.bindingNames, url)
			}
		}
	}
}

// collectBindingNames indexes the name HL7 gives each required binding, via the
// elementdefinition-bindingName extension.
//
// The ValueSet's own title is a poor source for a Go type name: it produces
// MedicationStatusCodes where the specification calls the binding
// MedicationStatus, and Currencies where it says CurrencyCode. bindingName is
// what HL7 intends the binding to be called.
//
// It is only used when every element binding to a given ValueSet agrees on one
// name. 16 ValueSets in R4 are bound from several places under different names —
// request-priority is CommunicationPriority, TaskPriority, ServiceRequestPriority
// and two more — and a ValueSet becomes exactly one Go type, so there is no
// answer to pick. Those keep the title-derived name.
//
// Names that are not already Go identifiers are skipped too: the specification
// carries a few like "appointment-type" and "LOINC LL379-9 answerlist", and
// sanitizing them would land back on the name we already had.
func collectBindingNames(definitions []*parser.StructureDefinition) map[string]string {
	claimed := make(map[string]map[string]bool)
	for _, sd := range definitions {
		if sd == nil || sd.Snapshot == nil {
			continue
		}
		for i := range sd.Snapshot.Element {
			elem := &sd.Snapshot.Element[i]
			if elem.Binding == nil || elem.Binding.Strength != "required" {
				continue
			}
			name := elem.Binding.Name()
			if name == "" || !isGoIdentifier(name) {
				continue
			}
			if _, misspelled := misspelledBindingNames[name]; misspelled {
				continue
			}
			url := canonicalValueSetURL(elem.Binding.ValueSet)
			if url == "" {
				continue
			}
			if claimed[url] == nil {
				claimed[url] = make(map[string]bool)
			}
			claimed[url][name] = true
		}
	}

	names := make(map[string]string, len(claimed))
	for url, set := range claimed {
		if len(set) != 1 {
			continue // ambiguous: several bindings, several names
		}
		for name := range set {
			names[url] = strings.ToUpper(name[:1]) + name[1:]
		}
	}
	return names
}

// isGoIdentifier reports whether s can be used as a Go type name unchanged.
func isGoIdentifier(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}

// AnalyzedType represents a fully analyzed type ready for code generation.
type AnalyzedType struct {
	Name           string             // Go type name (PascalCase)
	FHIRName       string             // Original FHIR name
	Kind           string             // primitive, datatype, resource, backbone
	Description    string             // Documentation
	URL            string             // Canonical URL
	IsAbstract     bool               // Whether this is an abstract type
	Properties     []AnalyzedProperty // Fields of this type
	Constraints    []AnalyzedConstraint
	BackboneTypes  []*AnalyzedType // Nested backbone element types for this resource
	ParentResource string          // For backbone types: name of the parent resource
}

// AnalyzedProperty represents a single property of a type.
type AnalyzedProperty struct {
	Name           string   // Go field name (PascalCase)
	JSONName       string   // JSON field name (camelCase)
	GoType         string   // Complete Go type (e.g., "*string", "[]Coding")
	Description    string   // Documentation
	IsPointer      bool     // Whether this field is a pointer
	IsArray        bool     // Whether this field is an array
	IsRequired     bool     // Whether min >= 1
	IsPrimitive    bool     // Whether the base type is a primitive
	IsChoice       bool     // Whether this is a choice type field
	ChoiceTypes    []string // For choice types, the list of allowed types
	ChoiceBaseName string   // For choice types, the base element name (e.g., "value" for "value[x]")
	FHIRType       string   // Original FHIR type code
	Binding        *AnalyzedBinding
	HasExtension   bool     // Whether this primitive needs a _field for extensions
	IsBackbone     bool     // Whether this is a backbone element reference
	BackboneType   string   // For backbone: the specific backbone type name (e.g., "PatientContact")
	IsSummary      bool     // Whether this field is marked as isSummary in FHIR spec
	TargetTypes    []string // For Reference/canonical types: allowed target resource type names
	ContentRef     string   // For contentReference properties: the target FHIR path (e.g., "Questionnaire.item")
}

// AnalyzedBinding represents a value set binding.
type AnalyzedBinding struct {
	Strength string // required, extensible, preferred, example
	ValueSet string // ValueSet URL
}

// AnalyzedConstraint represents a FHIRPath constraint.
type AnalyzedConstraint struct {
	Key        string
	Severity   string
	Human      string
	Expression string
}

// Analyze processes a StructureDefinition and returns an AnalyzedType.
func (a *Analyzer) Analyze(sd *parser.StructureDefinition) (*AnalyzedType, error) {
	if sd == nil {
		return nil, fmt.Errorf("StructureDefinition is nil")
	}

	kind := a.determineKind(sd)

	analyzed := &AnalyzedType{
		Name:        sd.Name,
		FHIRName:    sd.Name,
		Kind:        kind,
		Description: sd.Title,
		URL:         sd.URL,
		IsAbstract:  sd.Abstract,
	}

	elements := sd.GetElements()
	if len(elements) == 0 {
		return analyzed, nil
	}

	// For resources, datatypes, and backbone types, extract nested backbone elements
	if kind == "resource" || kind == kindDatatype || kind == kindBackbone {
		backbones := a.extractBackboneElements(sd)
		analyzed.BackboneTypes = backbones
	}

	// Skip the root element (first element is always the type itself)
	for i := 1; i < len(elements); i++ {
		elem := elements[i]

		// Skip slices for now
		if elem.SliceName != "" {
			continue
		}

		// Skip nested elements (backbone children) - they'll be handled separately
		if a.isNestedElement(elem.Path, sd.Type) {
			continue
		}

		props, err := a.analyzeElement(&elem, sd.Type, sd.Name)
		if err != nil {
			return nil, fmt.Errorf("failed to analyze element %s: %w", elem.Path, err)
		}
		analyzed.Properties = append(analyzed.Properties, props...)
	}

	// Extract constraints from the root element
	if len(elements) > 0 {
		for _, c := range elements[0].Constraint {
			analyzed.Constraints = append(analyzed.Constraints, AnalyzedConstraint{
				Key:        c.Key,
				Severity:   c.Severity,
				Human:      c.Human,
				Expression: c.Expression,
			})
		}
	}

	return analyzed, nil
}

// extractBackboneElements extracts all backbone element types from a resource.
func (a *Analyzer) extractBackboneElements(sd *parser.StructureDefinition) []*AnalyzedType {
	elements := sd.GetElements()
	backboneMap := make(map[string]*AnalyzedType)

	// First pass: identify all backbone element paths
	for _, elem := range elements {
		if elem.IsBackboneElement() {
			// Get the backbone type name: ResourceName + FieldName(s)
			// e.g., Patient.contact -> PatientContact
			// e.g., Bundle.entry.search -> BundleEntrySearch
			backboneName := a.getBackboneTypeName(elem.Path)
			backboneMap[elem.Path] = &AnalyzedType{
				Name:           backboneName,
				FHIRName:       elem.Path,
				Kind:           "backbone",
				Description:    elem.Short,
				ParentResource: sd.Name,
				Properties:     []AnalyzedProperty{},
			}
		}
	}

	// Second pass: assign properties to backbone types
	for _, elem := range elements {
		if elem.SliceName != "" {
			continue
		}

		// Find the parent backbone for this element
		parentPath := a.findParentBackbonePath(elem.Path, backboneMap)
		if parentPath == "" {
			continue
		}

		// Only process direct children of the backbone
		suffix := strings.TrimPrefix(elem.Path, parentPath+".")
		if suffix == elem.Path || strings.Contains(suffix, ".") {
			continue
		}

		// Get the backbone type
		backbone := backboneMap[parentPath]
		if backbone == nil {
			continue
		}

		// Create the property
		fieldName := suffix
		switch {
		case elem.IsChoiceType():
			//nolint:errcheck // Choice type analysis errors are non-fatal; skip on error
			props, _ := a.analyzeChoiceType(&elem, strings.TrimSuffix(fieldName, "[x]"))
			backbone.Properties = append(backbone.Properties, props...)
		case elem.ContentReference != "":
			// Content reference - resolve to the referenced type
			goType, isBackboneRef, backboneTypeName := a.resolveContentReference(elem.ContentReference, elem.IsArray())
			prop := AnalyzedProperty{
				Name:         toGoFieldName(fieldName),
				JSONName:     toLowerFirst(fieldName),
				GoType:       goType,
				Description:  elem.Short,
				IsPointer:    !elem.IsArray(),
				IsArray:      elem.IsArray(),
				IsRequired:   elem.IsRequired(),
				IsPrimitive:  false,
				FHIRType:     "ContentReference",
				IsBackbone:   isBackboneRef,
				BackboneType: backboneTypeName,
				ContentRef:   strings.TrimPrefix(elem.ContentReference, "#"),
			}
			backbone.Properties = append(backbone.Properties, prop)
		case elem.IsBackboneElement():
			// Nested backbone element - use specific type name
			backboneTypeName := a.getBackboneTypeName(elem.Path)
			isArray := elem.IsArray()
			var goType string
			if isArray {
				goType = "[]" + backboneTypeName
			} else {
				goType = "*" + backboneTypeName
			}

			prop := AnalyzedProperty{
				Name:         toGoFieldName(fieldName),
				JSONName:     toLowerFirst(fieldName),
				GoType:       goType,
				Description:  elem.Short,
				IsPointer:    !isArray,
				IsArray:      isArray,
				IsRequired:   elem.IsRequired(),
				IsPrimitive:  false,
				FHIRType:     "BackboneElement",
				IsBackbone:   true,
				BackboneType: backboneTypeName,
			}
			backbone.Properties = append(backbone.Properties, prop)
		case len(elem.Type) > 0:
			prop := a.createProperty(&elem, fieldName, elem.Type[0])
			backbone.Properties = append(backbone.Properties, prop)
		}
	}

	// Convert map to slice
	backbones := make([]*AnalyzedType, 0, len(backboneMap))
	for _, bb := range backboneMap {
		backbones = append(backbones, bb)
	}

	return backbones
}

// getBackboneTypeName generates a Go type name for a backbone element path.
func (a *Analyzer) getBackboneTypeName(path string) string {
	// Split the path and PascalCase each part
	// e.g., "Patient.contact" -> "PatientContact"
	// e.g., "Bundle.entry.search" -> "BundleEntrySearch"
	parts := strings.Split(path, ".")
	result := ""
	for _, part := range parts {
		result += toPascalCase(part)
	}
	return result
}

// findParentBackbonePath finds the immediate parent backbone path for an element.
func (a *Analyzer) findParentBackbonePath(elemPath string, backboneMap map[string]*AnalyzedType) string {
	// Find the longest matching backbone path
	longestMatch := ""
	for bbPath := range backboneMap {
		if strings.HasPrefix(elemPath, bbPath+".") && len(bbPath) > len(longestMatch) {
			longestMatch = bbPath
		}
	}
	return longestMatch
}

// determineKind determines the kind of type (primitive, datatype, resource, backbone).
func (a *Analyzer) determineKind(sd *parser.StructureDefinition) string {
	switch sd.Kind {
	case parser.KindPrimitiveType:
		return "primitive"
	case parser.KindResource:
		return "resource"
	case parser.KindComplexType:
		// Check if it's a backbone element
		if strings.Contains(sd.BaseDefinition, "BackboneElement") {
			return kindBackbone
		}
		return kindDatatype
	default:
		return kindDatatype
	}
}

// isNestedElement checks if an element path indicates a nested (backbone) element.
func (a *Analyzer) isNestedElement(path, rootType string) bool {
	// Remove the root type prefix
	suffix := strings.TrimPrefix(path, rootType+".")
	if suffix == path {
		return false
	}
	// If there's still a dot, it's nested
	return strings.Contains(suffix, ".")
}

// analyzeElement analyzes a single element and returns properties.
// May return multiple properties for choice types.
func (a *Analyzer) analyzeElement(elem *parser.ElementDefinition, rootType, _ string) ([]AnalyzedProperty, error) {
	// Get the field name from the path
	fieldName := a.extractFieldName(elem.Path, rootType)
	if fieldName == "" {
		return nil, nil
	}

	// Handle choice types (value[x], effective[x], etc.)
	if elem.IsChoiceType() {
		return a.analyzeChoiceType(elem, fieldName)
	}

	// Handle content references
	if elem.ContentReference != "" {
		return a.analyzeContentReference(elem, fieldName)
	}

	// Handle backbone elements - use specific type instead of generic BackboneElement
	if elem.IsBackboneElement() {
		backboneTypeName := a.getBackboneTypeName(elem.Path)
		isArray := elem.IsArray()
		var goType string
		if isArray {
			goType = "[]" + backboneTypeName
		} else {
			goType = "*" + backboneTypeName
		}

		prop := AnalyzedProperty{
			Name:         toGoFieldName(fieldName),
			JSONName:     toLowerFirst(fieldName),
			GoType:       goType,
			Description:  elem.Short,
			IsPointer:    !isArray,
			IsArray:      isArray,
			IsRequired:   elem.IsRequired(),
			IsPrimitive:  false,
			FHIRType:     "BackboneElement",
			IsBackbone:   true,
			BackboneType: backboneTypeName,
		}
		return []AnalyzedProperty{prop}, nil
	}

	// Regular element
	if len(elem.Type) == 0 {
		return nil, nil
	}

	prop := a.createProperty(elem, fieldName, elem.Type[0])
	return []AnalyzedProperty{prop}, nil
}

// analyzeChoiceType handles choice type elements like value[x].
func (a *Analyzer) analyzeChoiceType(elem *parser.ElementDefinition, baseName string) ([]AnalyzedProperty, error) {
	props := make([]AnalyzedProperty, 0, len(elem.Type)*2) // *2 for extension fields
	choiceTypes := make([]string, 0, len(elem.Type))

	for _, typeRef := range elem.Type {
		choiceTypes = append(choiceTypes, typeRef.Code)
	}

	// Generate a property for each possible type
	for _, typeRef := range elem.Type {
		typeName := typeRef.Code
		// Field name: PascalCase(baseName) + PascalCase(typeName)
		// e.g., "deceased" + "Boolean" = "DeceasedBoolean"
		fieldName := toPascalCase(baseName) + toPascalCase(typeName)

		// Interfaces should not have pointers, even in choice types
		isInterface := (typeName == "Resource" || typeName == "DomainResource")
		usePointer := !isInterface // true for most types, false for interfaces

		prop := AnalyzedProperty{
			Name:           fieldName,
			JSONName:       toLowerFirst(baseName) + toPascalCase(typeName),
			GoType:         a.resolveGoType(typeName, usePointer, false),
			Description:    elem.Short,
			IsPointer:      usePointer,
			IsArray:        false,
			IsRequired:     false,
			IsPrimitive:    IsPrimitiveType(typeName),
			IsChoice:       true,
			ChoiceTypes:    choiceTypes,
			ChoiceBaseName: baseName,
			FHIRType:       typeName,
			HasExtension:   IsPrimitiveType(typeName),
		}

		if elem.Binding != nil {
			prop.Binding = &AnalyzedBinding{
				Strength: elem.Binding.Strength,
				ValueSet: elem.Binding.ValueSet,
			}
		}

		props = append(props, prop)

		// Add extension field for primitives
		if prop.HasExtension {
			extProp := AnalyzedProperty{
				Name:        fieldName + "Ext",
				JSONName:    "_" + toLowerFirst(baseName) + toPascalCase(typeName),
				GoType:      "*Element",
				Description: fmt.Sprintf("Extension for %s", fieldName),
				IsPointer:   true,
				IsArray:     false,
				IsPrimitive: false,
				FHIRType:    "Element",
			}
			props = append(props, extProp)
		}
	}

	return props, nil
}

// analyzeContentReference handles content references.
// Content references point to another element's definition within the same or different resource.
// Format: "#ResourceType.path.to.element" (e.g., "#TestScript.setup.action.operation")
func (a *Analyzer) analyzeContentReference(elem *parser.ElementDefinition, fieldName string) ([]AnalyzedProperty, error) {
	// Resolve the content reference to get the actual Go type
	goType, isBackbone, backboneTypeName := a.resolveContentReference(elem.ContentReference, elem.IsArray())

	prop := AnalyzedProperty{
		Name:         toGoFieldName(fieldName),
		JSONName:     toLowerFirst(fieldName),
		GoType:       goType,
		Description:  elem.Short,
		IsPointer:    !elem.IsArray(),
		IsArray:      elem.IsArray(),
		IsRequired:   elem.IsRequired(),
		FHIRType:     "ContentReference",
		IsBackbone:   isBackbone,
		BackboneType: backboneTypeName,
		ContentRef:   strings.TrimPrefix(elem.ContentReference, "#"),
	}
	return []AnalyzedProperty{prop}, nil
}

// resolveContentReference parses a contentReference URL and returns the Go type.
func (a *Analyzer) resolveContentReference(ref string, isArray bool) (goType string, isBackbone bool, backboneTypeName string) {
	// Remove the leading "#" from the reference
	if !strings.HasPrefix(ref, "#") {
		// Invalid format, return interface{} as fallback
		if isArray {
			return "[]interface{}", false, ""
		}
		return "*interface{}", false, ""
	}

	refPath := strings.TrimPrefix(ref, "#")

	// The refPath is the full element path like "TestScript.setup.action.operation"
	// We need to find if this is a BackboneElement and generate the appropriate type name

	// Extract the resource type from the path (first segment)
	parts := strings.Split(refPath, ".")
	if len(parts) < 2 {
		if isArray {
			return "[]interface{}", false, ""
		}
		return "*interface{}", false, ""
	}

	resourceType := parts[0]

	// Try to find the StructureDefinition for this resource
	sd := a.definitions[resourceType]
	if sd == nil {
		// Resource not found, try to generate a reasonable type name anyway
		// This handles cases where the referenced resource might not be loaded
		backboneTypeName = a.getBackboneTypeName(refPath)
		if isArray {
			return "[]" + backboneTypeName, true, backboneTypeName
		}
		return "*" + backboneTypeName, true, backboneTypeName
	}

	// Find the referenced element in the StructureDefinition
	var targetElem *parser.ElementDefinition
	for i := range sd.Snapshot.Element {
		if sd.Snapshot.Element[i].Path == refPath {
			targetElem = &sd.Snapshot.Element[i]
			break
		}
	}

	if targetElem == nil {
		// Element not found, generate backbone type name from path
		backboneTypeName = a.getBackboneTypeName(refPath)
		if isArray {
			return "[]" + backboneTypeName, true, backboneTypeName
		}
		return "*" + backboneTypeName, true, backboneTypeName
	}

	// Check if the target element is a BackboneElement
	if targetElem.IsBackboneElement() {
		backboneTypeName = a.getBackboneTypeName(refPath)
		if isArray {
			return "[]" + backboneTypeName, true, backboneTypeName
		}
		return "*" + backboneTypeName, true, backboneTypeName
	}

	// If it has types, use the first type
	if len(targetElem.Type) > 0 {
		typeName := targetElem.Type[0].Code
		goType = a.resolveGoType(typeName, !isArray, isArray)
		return goType, false, ""
	}

	// Fallback: assume it's a backbone element based on the path
	backboneTypeName = a.getBackboneTypeName(refPath)
	if isArray {
		return "[]" + backboneTypeName, true, backboneTypeName
	}
	return "*" + backboneTypeName, true, backboneTypeName
}

// createProperty creates an AnalyzedProperty from an element and type reference.
func (a *Analyzer) createProperty(elem *parser.ElementDefinition, fieldName string, typeRef parser.TypeRef) AnalyzedProperty {
	typeName := typeRef.Code
	isArray := elem.IsArray()
	isPrimitive := IsPrimitiveType(typeName)

	// Determine if pointer is needed.
	//
	// Interfaces never need one — they are already references — and a nil slice
	// already means absent, so arrays do not either. Everything else does,
	// including required complex types.
	//
	// Requiring a complex type used to mean a value field, on the reasoning that
	// the type should express the obligation. It does not: Go has no way to force
	// a field to be set, so r4.Observation{} compiles and marshals as
	// {"code":{}} — an empty element that violates ele-1 and that no validator
	// accepts. The obligation is a matter for a FHIR validator; what the type has
	// to be able to say is "absent", and only a pointer says that.
	isInterface := (typeName == "Resource" || typeName == "DomainResource")
	isPointer := !isArray && !isInterface

	// Check for required binding with code type - use custom type
	goType := a.resolveGoTypeWithBinding(typeName, isPointer, isArray, elem.Binding)

	prop := AnalyzedProperty{
		Name:         toGoFieldName(fieldName),
		JSONName:     toLowerFirst(fieldName),
		GoType:       goType,
		Description:  elem.Short,
		IsPointer:    isPointer,
		IsArray:      isArray,
		IsRequired:   elem.IsRequired(),
		IsPrimitive:  isPrimitive,
		IsChoice:     false,
		FHIRType:     typeName,
		HasExtension: isPrimitive,
		IsSummary:    elem.IsSummary,
	}

	if (typeRef.Code == "Reference" || typeRef.Code == "canonical") && len(typeRef.TargetProfile) > 0 {
		targets := make([]string, 0, len(typeRef.TargetProfile))
		for _, url := range typeRef.TargetProfile {
			parts := strings.Split(url, "/")
			targets = append(targets, parts[len(parts)-1])
		}
		prop.TargetTypes = targets
	}

	if elem.Binding != nil {
		prop.Binding = &AnalyzedBinding{
			Strength: elem.Binding.Strength,
			ValueSet: elem.Binding.ValueSet,
		}
	}

	return prop
}

// resolveGoTypeWithBinding resolves Go type, using custom types for required bindings.
func (a *Analyzer) resolveGoTypeWithBinding(fhirType string, isPointer, isArray bool, binding *parser.Binding) string {
	// Only apply custom types for code fields with required binding
	if fhirType == "code" && binding != nil && binding.Strength == "required" {
		if vs := a.getValueSetForBinding(binding.ValueSet); vs != nil {
			// Track that this binding is used
			a.UsedBindings[binding.ValueSet] = true

			// Must match what the code-system template emits for this ValueSet.
			customType := a.ValueSetTypeName(vs.URL, vs.Name)
			if isArray {
				// Same reasoning as resolveGoType: a repeating code is a
				// repeating primitive, and needs to represent an absent slot.
				return "[]*" + customType
			}
			if isPointer {
				return "*" + customType
			}
			return customType
		}
	}

	return a.resolveGoType(fhirType, isPointer, isArray)
}

// valueSetCollisionOverrides resolves ValueSets whose FHIR `name` collides with
// another's, keyed by the colliding Go type name and then by canonical URL.
//
// FHIR names are not unique. In R4, medication-status and
// medication-statement-status are both named "Medication Status Codes", which both
// sanitize to MedicationStatusCodes. The generator used to emit whichever came
// first and silently drop the other, while still pointing the dropped ValueSet's
// fields at the surviving type — so Medication.status offered MedicationStatement's
// codes and had no constant for `inactive`, its only other legal value.
//
// Overrides are keyed on the collision, not on the URL alone, and are applied only
// when that collision is actually present. The R4 clash was fixed upstream: R4B
// renamed one side to "MedicationStatement Status Codes" and R5 to
// "MedicationStatementStatusCodes". A URL-keyed override would rename
// MedicationStatusCodes out of r4b and r5 as well, deleting an exported type from
// two packages that never had the problem.
//
// The surviving name is left alone and only the shadowed ValueSet is renamed, so
// existing constants keep working.
//
// A collision with no entry here fails generation rather than resolving by bundle
// order; see generateCodeSystemsFromTemplate.
//
// Binding names now reach the same answer without help: medication-status is bound
// as MedicationStatus, so the two ValueSets no longer land on one name and the
// entry below never fires. Emptying the map produces byte-identical output in all
// three versions. It is kept because it costs nothing and is the only thing
// standing between a future upstream collision and a failed build — the binding
// name that currently resolves this one is upstream data, and can change.
var valueSetCollisionOverrides = map[string]map[string]string{
	"MedicationStatusCodes": {
		// Bound by Medication.status: active | inactive | entered-in-error.
		"http://hl7.org/fhir/ValueSet/medication-status": "MedicationStatus",
	},
}

// ValueSetTypeName returns the Go type name for a ValueSet.
//
// Both the analyzer (choosing a field's type) and the code-system template
// (emitting the type) must agree on this, or fields end up bound to types that
// were never generated.
func (a *Analyzer) ValueSetTypeName(url, name string) string {
	base := a.baseTypeName(url, name)

	// Only a real collision justifies renaming.
	if len(a.valueSetNameClaims[base]) < 2 {
		return base
	}
	if override, ok := valueSetCollisionOverrides[base][canonicalValueSetURL(url)]; ok {
		return override
	}
	// Unresolved: keep the base name so the template reports the collision with
	// both URLs rather than this returning something arbitrary.
	return base
}

// buildValueSetNameClaims indexes which ValueSets lay claim to each sanitized Go
// type name, so a collision can be recognized as such.
//
// Only ValueSets that could actually become an enum are counted — the same filter
// getValueSetForBinding applies — because two unused ValueSets sharing a name is
// not a problem anyone can observe.
func (a *Analyzer) buildValueSetNameClaims(valueSets *parser.ValueSetRegistry) map[string][]string {
	claims := make(map[string][]string)
	if valueSets == nil {
		return claims
	}
	for _, vs := range valueSets.All() {
		if len(vs.Codes) == 0 || len(vs.Codes) > maxEnumCodes {
			continue
		}
		base := a.baseTypeName(vs.URL, vs.Name)
		claims[base] = append(claims[base], canonicalValueSetURL(vs.URL))
	}
	return claims
}

// baseTypeName is the name a ValueSet takes before collision handling: HL7's
// binding name when there is an unambiguous one, and the sanitized ValueSet
// title otherwise.
func (a *Analyzer) baseTypeName(url, name string) string {
	if bound, ok := a.bindingNames[canonicalValueSetURL(url)]; ok {
		return bound
	}
	return sanitizeTypeName(name)
}

// canonicalValueSetURL strips the |version suffix FHIR bindings carry, so
// "…/medication-status|4.0.1" and "…/medication-status" are the same key.
func canonicalValueSetURL(url string) string {
	if i := strings.IndexByte(url, '|'); i >= 0 {
		return url[:i]
	}
	return url
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
		runes[0] = unicode.ToUpper(runes[0])
		name = string(runes)
	}

	return name
}

// getValueSetForBinding retrieves and validates a ValueSet for use as a Go type.
func (a *Analyzer) getValueSetForBinding(url string) *parser.ParsedValueSet {
	if a.valueSets == nil {
		return nil
	}

	vs := a.valueSets.Get(url)
	if vs == nil || len(vs.Codes) == 0 {
		return nil
	}

	// Skip very large value sets (like all-types, mimetypes)
	if len(vs.Codes) > maxEnumCodes {
		return nil
	}

	return vs
}

// resolveGoType converts a FHIR type to a Go type string.
func (a *Analyzer) resolveGoType(fhirType string, isPointer, isArray bool) string {
	goType := FHIRToGoType(fhirType)

	if isArray {
		// A repeating primitive is []*T, not []T.
		//
		// FHIR aligns a primitive array with its _field extension array by
		// position, and uses null to mark a slot with no value:
		//
		//	"event":  [null]
		//	"_event": [{"extension": [{"url": ".../cqf-expression", ...}]}]
		//
		// which says the value is not a literal — it is computed by the
		// expression in the extension. []T cannot hold that: the null becomes
		// the zero value, so an absent date turns into an empty string and the
		// extension is left describing nothing.
		//
		// Complex-typed arrays keep []T; null is not permitted there.
		if isFHIRPrimitive(fhirType) {
			return "[]*" + goType
		}
		return "[]" + goType
	}
	if isPointer {
		return "*" + goType
	}
	return goType
}

// isFHIRPrimitive reports whether a FHIR type code names a primitive, and
// therefore participates in _field alignment.
//
// Decided on the FHIR type rather than the resolved Go type, so a code bound to a
// generated enum still counts: `code` is a primitive whichever Go type it lands
// on, and it carries extensions the same way.
func isFHIRPrimitive(fhirType string) bool {
	_, ok := PrimitiveTypeMap[fhirType]
	return ok
}

// extractFieldName extracts the field name from an element path.
func (a *Analyzer) extractFieldName(path, rootType string) string {
	suffix := strings.TrimPrefix(path, rootType+".")
	if suffix == path || suffix == "" {
		return ""
	}
	// Remove [x] suffix
	suffix = strings.TrimSuffix(suffix, "[x]")
	return suffix
}

// toGoFieldName converts a FHIR field name to a Go field name.
func toGoFieldName(name string) string {
	// Handle special cases
	switch name {
	case "class":
		return "Class"
	case "import":
		return "Import"
	case "type":
		return "Type"
	case "package":
		return "Package"
	case "interface":
		return "Interface"
	}

	// Convert to PascalCase
	return toPascalCase(name)
}

// toPascalCase converts a string to PascalCase.
func toPascalCase(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}

// toLowerFirst returns the string with the first character lowercased.
func toLowerFirst(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// TypeHierarchy returns a map from FHIR type name to its immediate parent type name,
// extracted from each StructureDefinition's BaseDefinition URL.
// e.g. "Patient" → "DomainResource", "Age" → "Quantity", "code" → "string".
func (a *Analyzer) TypeHierarchy() map[string]string {
	result := make(map[string]string)
	seen := make(map[string]bool)
	for _, sd := range a.definitions {
		if seen[sd.Name] {
			continue
		}
		seen[sd.Name] = true
		if sd.BaseDefinition == "" {
			continue
		}
		parts := strings.Split(sd.BaseDefinition, "/")
		parent := parts[len(parts)-1]
		if parent != "" && parent != sd.Name {
			result[sd.Name] = parent
		}
	}
	return result
}

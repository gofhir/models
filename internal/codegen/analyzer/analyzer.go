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

// Analyzer processes StructureDefinitions and produces analyzed types for code generation.
type Analyzer struct {
	definitions  map[string]*parser.StructureDefinition
	valueSets    *parser.ValueSetRegistry
	UsedBindings map[string]bool // Track which bindings are used (exported for generator)
}

// NewAnalyzer creates a new Analyzer with the given StructureDefinitions and ValueSets.
func NewAnalyzer(definitions []*parser.StructureDefinition, valueSets *parser.ValueSetRegistry) *Analyzer {
	defMap := make(map[string]*parser.StructureDefinition)
	for _, sd := range definitions {
		defMap[sd.URL] = sd
		defMap[sd.Name] = sd
		defMap[sd.Type] = sd
	}
	return &Analyzer{
		definitions:  defMap,
		valueSets:    valueSets,
		UsedBindings: make(map[string]bool),
	}
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
	Name         string   // Go field name (PascalCase)
	JSONName     string   // JSON field name (camelCase)
	GoType       string   // Complete Go type (e.g., "*string", "[]Coding")
	Description  string   // Documentation
	IsPointer    bool     // Whether this field is a pointer
	IsArray      bool     // Whether this field is an array
	IsRequired   bool     // Whether min >= 1
	IsPrimitive  bool     // Whether the base type is a primitive
	IsChoice     bool     // Whether this is a choice type field
	ChoiceTypes  []string // For choice types, the list of allowed types
	FHIRType     string   // Original FHIR type code
	Binding      *AnalyzedBinding
	HasExtension bool   // Whether this primitive needs a _field for extensions
	IsBackbone   bool   // Whether this is a backbone element reference
	BackboneType string // For backbone: the specific backbone type name (e.g., "PatientContact")
	IsSummary    bool   // Whether this field is marked as isSummary in FHIR spec
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
			Name:         fieldName,
			JSONName:     toLowerFirst(baseName) + toPascalCase(typeName),
			GoType:       a.resolveGoType(typeName, usePointer, false),
			Description:  elem.Short,
			IsPointer:    usePointer,
			IsArray:      false,
			IsRequired:   false,
			IsPrimitive:  IsPrimitiveType(typeName),
			IsChoice:     true,
			ChoiceTypes:  choiceTypes,
			FHIRType:     typeName,
			HasExtension: IsPrimitiveType(typeName),
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

	// Determine if pointer is needed
	// - Interfaces NEVER need pointers (they're already references)
	// - Arrays don't need pointer (nil slice is fine)
	// - Required primitives could be non-pointer, but we use pointer for JSON omitempty
	// - Complex types are always pointers when optional
	isInterface := (typeName == "Resource" || typeName == "DomainResource")
	isPointer := false
	if !isArray && !isInterface {
		isPointer = (elem.Min == 0 || isPrimitive)
	}

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

			// Sanitize the type name to match what generator produces
			customType := sanitizeTypeName(vs.Name)
			if isArray {
				return "[]" + customType
			}
			if isPointer {
				return "*" + customType
			}
			return customType
		}
	}

	return a.resolveGoType(fhirType, isPointer, isArray)
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
	if len(vs.Codes) > 100 {
		return nil
	}

	return vs
}

// resolveGoType converts a FHIR type to a Go type string.
func (a *Analyzer) resolveGoType(fhirType string, isPointer, isArray bool) string {
	goType := FHIRToGoType(fhirType)

	if isArray {
		return "[]" + goType
	}
	if isPointer {
		return "*" + goType
	}
	return goType
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

package analyzer

// PrimitiveTypeMap maps FHIR primitive types to Go types.
var PrimitiveTypeMap = map[string]string{
	// Boolean
	"boolean": "bool",

	// Numeric types
	"integer":     "int",
	"integer64":   "int64",
	"decimal":     "float64",
	"unsignedInt": "uint32",
	"positiveInt": "uint32",

	// String types (all represented as string in Go)
	"string":       "string",
	"uri":          "string",
	"url":          "string",
	"canonical":    "string",
	"base64Binary": "string",
	"code":         "string",
	"oid":          "string",
	"id":           "string",
	"markdown":     "string",
	"uuid":         "string",
	"xhtml":        "string",

	// Date/Time types (stored as string, parsed when needed)
	"instant":  "string",
	"date":     "string",
	"dateTime": "string",
	"time":     "string",
}

// ComplexTypeMap maps FHIR complex types to Go type names.
// These are datatypes that will be generated as structs.
var ComplexTypeMap = map[string]string{
	// Base types
	"Element":         "Element",
	"BackboneElement": "BackboneElement",
	"Resource":        "Resource",
	"DomainResource":  "DomainResource",

	// General-purpose datatypes
	"Address":             "Address",
	"Age":                 "Age",
	"Annotation":          "Annotation",
	"Attachment":          "Attachment",
	"CodeableConcept":     "CodeableConcept",
	"CodeableReference":   "CodeableReference",
	"Coding":              "Coding",
	"ContactDetail":       "ContactDetail",
	"ContactPoint":        "ContactPoint",
	"Contributor":         "Contributor",
	"Count":               "Count",
	"DataRequirement":     "DataRequirement",
	"Distance":            "Distance",
	"Dosage":              "Dosage",
	"Duration":            "Duration",
	"Expression":          "Expression",
	"Extension":           "Extension",
	"HumanName":           "HumanName",
	"Identifier":          "Identifier",
	"Meta":                "Meta",
	"Money":               "Money",
	"MoneyQuantity":       "MoneyQuantity",
	"Narrative":           "Narrative",
	"ParameterDefinition": "ParameterDefinition",
	"Period":              "Period",
	"Population":          "Population",
	"ProdCharacteristic":  "ProdCharacteristic",
	"ProductShelfLife":    "ProductShelfLife",
	"Quantity":            "Quantity",
	"Range":               "Range",
	"Ratio":               "Ratio",
	"RatioRange":          "RatioRange",
	"Reference":           "Reference",
	"RelatedArtifact":     "RelatedArtifact",
	"SampledData":         "SampledData",
	"Signature":           "Signature",
	"SimpleQuantity":      "SimpleQuantity",
	"Timing":              "Timing",
	"TriggerDefinition":   "TriggerDefinition",
	"UsageContext":        "UsageContext",

	// Special types
	"Availability":          "Availability",
	"ExtendedContactDetail": "ExtendedContactDetail",
	"VirtualServiceDetail":  "VirtualServiceDetail",
	"MarketingStatus":       "MarketingStatus",
}

// SpecialResourceTypes are resources that have special handling.
var SpecialResourceTypes = map[string]string{
	"Resource":       "Resource",
	"DomainResource": "DomainResource",
	"Bundle":         "Bundle",
	"Parameters":     "Parameters",
}

// FHIRPathSystemTypes maps FHIRPath system type URLs to Go types.
// These are used in StructureDefinitions for primitive element IDs.
var FHIRPathSystemTypes = map[string]string{
	"http://hl7.org/fhirpath/System.String":   "string",
	"http://hl7.org/fhirpath/System.Boolean":  "bool",
	"http://hl7.org/fhirpath/System.Integer":  "int",
	"http://hl7.org/fhirpath/System.Decimal":  "float64",
	"http://hl7.org/fhirpath/System.Date":     "string",
	"http://hl7.org/fhirpath/System.DateTime": "string",
	"http://hl7.org/fhirpath/System.Time":     "string",
}

// FHIRToGoType converts a FHIR type name to a Go type name.
func FHIRToGoType(fhirType string) string {
	// Check FHIRPath system types (URLs)
	if goType, ok := FHIRPathSystemTypes[fhirType]; ok {
		return goType
	}

	// Check primitives first
	if goType, ok := PrimitiveTypeMap[fhirType]; ok {
		return goType
	}

	// Check known complex types
	if goType, ok := ComplexTypeMap[fhirType]; ok {
		return goType
	}

	// Check special resource types
	if goType, ok := SpecialResourceTypes[fhirType]; ok {
		return goType
	}

	// Default: use the FHIR type name as-is (it's a resource or custom type)
	return fhirType
}

// IsPrimitiveType returns true if the FHIR type is a primitive type.
func IsPrimitiveType(fhirType string) bool {
	_, ok := PrimitiveTypeMap[fhirType]
	return ok
}

// IsComplexType returns true if the FHIR type is a complex type (datatype).
func IsComplexType(fhirType string) bool {
	_, ok := ComplexTypeMap[fhirType]
	return ok
}

// GoTypeRequiresPointer returns true if the Go type should be a pointer when optional.
func GoTypeRequiresPointer(goType string, isRequired bool) bool {
	// Arrays never need pointer (nil slice is fine for optional)
	// All primitives need pointer for JSON omitempty to work correctly
	// Complex types are always pointers when optional

	// For primitives, we always use pointers to distinguish "not set" from "zero value"
	switch goType {
	case "bool", "int", "int64", "uint32", "float64", "string":
		return true
	default:
		return !isRequired
	}
}

// PrimitiveTypesNeedingExtension returns the list of primitive types that support extensions.
// In FHIR, all primitives can have extensions via the _field pattern.
func PrimitiveTypesNeedingExtension() []string {
	types := make([]string, 0, len(PrimitiveTypeMap))
	for t := range PrimitiveTypeMap {
		types = append(types, t)
	}
	return types
}

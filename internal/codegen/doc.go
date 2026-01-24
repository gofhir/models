// Package codegen provides code generation from FHIR StructureDefinitions.
//
// This package is internal and not part of the public API.
// It contains:
//   - parser: Parses FHIR StructureDefinition JSON files
//   - analyzer: Analyzes elements and determines Go types
//   - generator: Generates Go code from templates
//   - templates: Go text/template files for code generation
package codegen

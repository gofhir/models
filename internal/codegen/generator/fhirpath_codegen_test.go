package generator

import (
	"slices"
	"strings"
	"testing"

	"github.com/gofhir/models/internal/codegen/parser"
)

// sd is a StructureDefinition with a snapshot of the given element ids, each
// its own path unless it names a slice.
func sd(name, kind, derivation, typ string, ids ...string) *parser.StructureDefinition {
	def := &parser.StructureDefinition{Name: name, Kind: kind, Derivation: derivation, Type: typ, Snapshot: &parser.Snapshot{}}
	for _, id := range ids {
		path, _, _ := strings.Cut(id, ":")
		def.Snapshot.Element = append(def.Snapshot.Element, parser.ElementDefinition{ID: id, Path: path})
	}
	return def
}

// Only a definition of a type contributes its snapshot: a constraint written
// under the same paths cannot repeat or reorder a type's children, a slice is
// not a child, and a logical model is no type an instance has. A root, which
// R4 writes with no derivation, is a definition; a type that constrains
// another has its base's children.
func TestChildElementsAreEachTypesDefinitionsOwn(t *testing.T) {
	c := &CodeGen{rawSDs: []*parser.StructureDefinition{
		sd("Element", "complex-type", "", "Element", "Element", "Element.id", "Element.extension"),
		sd("Quantity", "complex-type", "specialization", "Quantity", "Quantity", "Quantity.id", "Quantity.value", "Quantity.unit"),
		sd("SimpleQuantity", "complex-type", "constraint", "Quantity", "Quantity", "Quantity.id", "Quantity.unit", "Quantity.value"),
		sd("Observation", "resource", "specialization", "Observation",
			"Observation", "Observation.id", "Observation.status", "Observation.component",
			"Observation.component.code", "Observation.component.value[x]", "Observation.value[x]"),
		sd("vitalsigns", "resource", "constraint", "Observation",
			"Observation", "Observation.value[x]", "Observation.status", "Observation.component:systolic", "Observation.extra"),
		sd("Shape", "logical", "specialization", "Shape", "Shape", "Shape.side"),
	}}

	children := c.buildChildElements()
	for path, want := range map[string][]string{
		"Element":               {"id", "extension"},
		"Quantity":              {"id", "value", "unit"},
		"SimpleQuantity":        {"id", "value", "unit"},
		"Observation":           {"id", "status", "component", "value[x]"},
		"Observation.component": {"code", "value[x]"},
		"Shape":                 nil,
	} {
		if got := children[path]; !slices.Equal(got, want) {
			t.Errorf("children[%q] = %v, want %v", path, got, want)
		}
	}
}

// typed gives the element at path in def one type, with the FHIR type
// structuredefinition-fhir-type names for it when fhirType is not "".
func typed(def *parser.StructureDefinition, path, code, fhirType string) *parser.StructureDefinition {
	for i := range def.Snapshot.Element {
		e := &def.Snapshot.Element[i]
		if e.Path != path {
			continue
		}
		ref := parser.TypeRef{Code: code}
		if fhirType != "" {
			ref.Extension = []parser.TypeExtension{{
				URL:      "http://hl7.org/fhir/StructureDefinition/structuredefinition-fhir-type",
				ValueURL: fhirType,
			}}
		}
		e.Type = []parser.TypeRef{ref}
	}
	return def
}

// An element whose code is a System type is listed under the FHIR type its
// extension names, each path its own: a backbone element's id is not
// BackboneElement's. A primitive's value, which names the primitive, is left
// out, and so are a constraint's snapshot, a slice and an element without
// the extension.
func TestFHIRTypesNameEachSystemTypedElementsFHIRType(t *testing.T) {
	const system = "http://hl7.org/fhirpath/System.String"
	backbone := typed(sd("BackboneElement", "complex-type", "specialization", "BackboneElement",
		"BackboneElement", "BackboneElement.id"), "BackboneElement.id", system, "id")
	patient := sd("Patient", "resource", "specialization", "Patient",
		"Patient", "Patient.id", "Patient.contact", "Patient.contact.id", "Patient.gender")
	typed(patient, "Patient.id", system, "id")
	typed(patient, "Patient.contact.id", system, "string")
	typed(patient, "Patient.gender", "code", "")
	str := typed(sd("string", "primitive-type", "specialization", "string",
		"string", "string.id", "string.value"), "string.id", system, "string")
	typed(str, "string.value", system, "string")
	extension := typed(sd("Extension", "complex-type", "specialization", "Extension",
		"Extension", "Extension.url", "Extension.url:slice"), "Extension.url", system, "uri")
	typed(extension, "Extension.url:slice", system, "url")
	profile := typed(sd("MyPatient", "resource", "constraint", "Patient",
		"Patient", "Patient.id"), "Patient.id", system, "string")

	c := &CodeGen{rawSDs: []*parser.StructureDefinition{profile, backbone, patient, str, extension}}
	got := c.buildFHIRTypes()
	want := map[string]string{
		"BackboneElement.id": "id",
		"Patient.id":         "id",
		"Patient.contact.id": "string",
		"string.id":          "string",
		"Extension.url":      "uri",
	}
	if len(got) != len(want) {
		t.Errorf("buildFHIRTypes() = %v, want %v", got, want)
	}
	for path, fhirType := range want {
		if got[path] != fhirType {
			t.Errorf("buildFHIRTypes()[%q] = %q, want %q", path, got[path], fhirType)
		}
	}
}

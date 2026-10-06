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

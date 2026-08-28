package r4b

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFHIRPathModel_NotNil(t *testing.T) {
	require.NotNil(t, FHIRPathModel())
}

func TestFHIRPathModel_FHIRVersion(t *testing.T) {
	// Must match StructureDefinition.fhirVersion in specs/r4b, not the package name.
	assert.Equal(t, "4.3.0", FHIRPathModel().FHIRVersion())
}

func TestFHIRPathModel_HasType(t *testing.T) {
	m := FHIRPathModel()

	for _, name := range []string{"Patient", "Observation", "Bundle",
		"HumanName", "Quantity", "CodeableConcept",
		"string", "code", "dateTime", "boolean",
	} {
		assert.True(t, m.HasType(name), "expected %q to resolve", name)
	}

	// Root and abstract types resolve even though ParentType returns "" for them.
	assert.True(t, m.HasType("Element"))
	assert.True(t, m.HasType("Resource"))
	assert.True(t, m.HasType("DomainResource"))
	assert.Empty(t, m.ParentType("Element"))

	// Base is R5-only: Element and Resource are the roots in R4B.
	assert.False(t, m.HasType("Base"))

	// Constraint-derived definitions resolve, and must agree with ParentType.
	assert.True(t, m.HasType("SimpleQuantity"))
	assert.Equal(t, "Quantity", m.ParentType("SimpleQuantity"))

	// Unresolvable specifiers, case sensitivity, and System types.
	assert.False(t, m.HasType("string1"))
	assert.False(t, m.HasType("NonExistentType"))
	assert.False(t, m.HasType(""))
	assert.False(t, m.HasType("patient"))
	assert.False(t, m.HasType("String"))
	assert.False(t, m.HasType("Integer"))
}

func TestFHIRPathModel_HasTypeCoversHierarchy(t *testing.T) {
	m := FHIRPathModel()

	// IsSubtype must never walk into a type that HasType rejects.
	for child, parent := range m.type2Parent {
		assert.True(t, m.HasType(child), "type2Parent key %q is not declared", child)
		assert.True(t, m.HasType(parent), "type2Parent value %q is not declared", parent)
	}

	for name := range m.resources {
		assert.True(t, m.HasType(name), "resource %q is not declared", name)
	}
}

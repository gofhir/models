package r4

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFHIRPathModel_NotNil(t *testing.T) {
	m := FHIRPathModel()
	require.NotNil(t, m)
}

func TestFHIRPathModel_FHIRVersion(t *testing.T) {
	// Must match StructureDefinition.fhirVersion in specs/r4, not the package name.
	assert.Equal(t, "4.0.1", FHIRPathModel().FHIRVersion())
}

func TestFHIRPathModel_ChoiceTypes(t *testing.T) {
	m := FHIRPathModel()

	// Observation.value[x] has 11 permitted types
	types := m.ChoiceTypes("Observation.value")
	require.NotNil(t, types)
	assert.Contains(t, types, "Quantity")
	assert.Contains(t, types, "string")
	assert.Contains(t, types, "boolean")
	assert.Contains(t, types, "Period")

	// Patient.deceased[x]
	deceased := m.ChoiceTypes("Patient.deceased")
	require.NotNil(t, deceased)
	assert.Contains(t, deceased, "boolean")
	assert.Contains(t, deceased, "dateTime")

	// Non-choice path returns nil
	assert.Nil(t, m.ChoiceTypes("Patient.name"))
	assert.Nil(t, m.ChoiceTypes("Patient.nonexistent"))
}

func TestFHIRPathModel_TypeOf(t *testing.T) {
	m := FHIRPathModel()

	assert.Equal(t, "HumanName", m.TypeOf("Patient.name"))
	assert.Equal(t, "code", m.TypeOf("Patient.gender"))
	assert.Equal(t, "boolean", m.TypeOf("Patient.active"))
	assert.Equal(t, "Reference", m.TypeOf("Observation.subject"))
	assert.Equal(t, "code", m.TypeOf("Observation.status"))
	assert.Equal(t, "CodeableConcept", m.TypeOf("Observation.code"))

	// Backbone element path
	assert.Equal(t, "BackboneElement", m.TypeOf("Patient.contact"))

	// Nested backbone property
	assert.Equal(t, "HumanName", m.TypeOf("Patient.contact.name"))

	// Choice type variant paths
	assert.Equal(t, "Quantity", m.TypeOf("Observation.valueQuantity"))
	assert.Equal(t, "string", m.TypeOf("Observation.valueString"))

	// Extension companion fields are excluded
	assert.Empty(t, m.TypeOf("Patient._birthDate"))

	// Unknown path returns empty string
	assert.Empty(t, m.TypeOf("Patient.nonexistent"))
}

func TestFHIRPathModel_ReferenceTargets(t *testing.T) {
	m := FHIRPathModel()

	// Observation.subject allows Patient, Group, Device, Location
	targets := m.ReferenceTargets("Observation.subject")
	require.NotNil(t, targets)
	assert.Contains(t, targets, "Patient")
	assert.Contains(t, targets, "Group")
	assert.Contains(t, targets, "Device")
	assert.Contains(t, targets, "Location")

	// Patient.managingOrganization → only Organization
	orgTargets := m.ReferenceTargets("Patient.managingOrganization")
	require.NotNil(t, orgTargets)
	assert.Contains(t, orgTargets, "Organization")

	// Patient.generalPractitioner → Practitioner, Organization, PractitionerRole
	gpTargets := m.ReferenceTargets("Patient.generalPractitioner")
	require.NotNil(t, gpTargets)
	assert.Contains(t, gpTargets, "Practitioner")

	// Non-reference field returns nil
	assert.Nil(t, m.ReferenceTargets("Patient.name"))
	assert.Nil(t, m.ReferenceTargets("Observation.status"))
}

func TestFHIRPathModel_ParentType(t *testing.T) {
	m := FHIRPathModel()

	assert.Equal(t, "DomainResource", m.ParentType("Patient"))
	assert.Equal(t, "DomainResource", m.ParentType("Observation"))
	assert.Equal(t, "Resource", m.ParentType("Bundle"))
	assert.Equal(t, "Resource", m.ParentType("DomainResource"))
	assert.Equal(t, "Quantity", m.ParentType("Age"))
	assert.Equal(t, "Quantity", m.ParentType("Duration"))
	assert.Equal(t, "Element", m.ParentType("BackboneElement"))

	// Unknown type returns empty string
	assert.Empty(t, m.ParentType("NonExistentType"))
}

func TestFHIRPathModel_IsSubtype(t *testing.T) {
	m := FHIRPathModel()

	// Direct parent
	assert.True(t, m.IsSubtype("Patient", "DomainResource"))
	assert.True(t, m.IsSubtype("Age", "Quantity"))

	// Transitive: Patient → DomainResource → Resource
	assert.True(t, m.IsSubtype("Patient", "Resource"))

	// Bundle → Resource (skips DomainResource)
	assert.True(t, m.IsSubtype("Bundle", "Resource"))

	// Reflexive
	assert.True(t, m.IsSubtype("Patient", "Patient"))

	// False cases
	assert.False(t, m.IsSubtype("Patient", "Observation"))
	assert.False(t, m.IsSubtype("Patient", "HumanName"))
	assert.False(t, m.IsSubtype("HumanName", "Resource"))
}

func TestFHIRPathModel_ResolvePath(t *testing.T) {
	m := FHIRPathModel()

	// Recursive structure: Questionnaire.item.item reuses Questionnaire.item definition
	assert.Equal(t, "Questionnaire.item", m.ResolvePath("Questionnaire.item.item"))

	// Self-defined paths are returned unchanged
	assert.Equal(t, "Patient.name", m.ResolvePath("Patient.name"))
	assert.Equal(t, "Observation.value", m.ResolvePath("Observation.value"))
}

func TestFHIRPathModel_IsResource(t *testing.T) {
	m := FHIRPathModel()

	assert.True(t, m.IsResource("Patient"))
	assert.True(t, m.IsResource("Observation"))
	assert.True(t, m.IsResource("Bundle"))
	assert.True(t, m.IsResource("Parameters"))

	assert.False(t, m.IsResource("HumanName"))
	assert.False(t, m.IsResource("CodeableConcept"))
	assert.False(t, m.IsResource("BackboneElement"))
	assert.False(t, m.IsResource("NonExistentType"))
}

func TestFHIRPathModel_HasType(t *testing.T) {
	m := FHIRPathModel()

	// Resources, data types and primitives are all type specifiers a FHIRPath
	// expression may legitimately name.
	for _, name := range []string{"Patient", "Observation", "Bundle",
		"HumanName", "Quantity", "CodeableConcept",
		"string", "code", "dateTime", "boolean",
	} {
		assert.True(t, m.HasType(name), "expected %q to resolve", name)
	}

	// Root and abstract types resolve too, even though ParentType returns ""
	// for them — which is why HasType cannot be derived from ParentType alone.
	assert.True(t, m.HasType("Element"))
	assert.True(t, m.HasType("Resource"))
	assert.True(t, m.HasType("DomainResource"))
	assert.True(t, m.HasType("BackboneElement"))
	assert.Empty(t, m.ParentType("Element"))
	assert.Empty(t, m.ParentType("Resource"))

	// Base is R5-only: Element and Resource are the roots in R4.
	assert.False(t, m.HasType("Base"))

	// Constraint-derived definitions resolve, and must agree with ParentType:
	// the model cannot report a parent for a type it claims not to declare.
	assert.True(t, m.HasType("SimpleQuantity"))
	assert.Equal(t, "Quantity", m.ParentType("SimpleQuantity"))

	// Unresolvable specifiers. The FHIRPath conformance suite requires
	// Patient.gender.as(string1) to be an execution error rather than empty.
	assert.False(t, m.HasType("string1"))
	assert.False(t, m.HasType("strng"))
	assert.False(t, m.HasType("NonExistentType"))
	assert.False(t, m.HasType(""))

	// Case-sensitive: FHIR names resources in PascalCase and primitives in
	// lower camelCase, and treats them as distinct types.
	assert.False(t, m.HasType("patient"))
	assert.False(t, m.HasType("String"))

	// System types belong to the FHIRPath language, not to this model.
	assert.False(t, m.HasType("Integer"))
	assert.False(t, m.HasType("Decimal"))
}

// typeRegistry mirrors the optional interface gofhir/fhirpath probes for. The
// engine detects it by assertion, so a signature change here would silently
// switch type-specifier validation back off instead of breaking the build.
type typeRegistry interface {
	HasType(typeName string) bool
}

func TestFHIRPathModel_SatisfiesTypeRegistry(t *testing.T) {
	var _ typeRegistry = FHIRPathModel()

	r, ok := any(FHIRPathModel()).(typeRegistry)
	require.True(t, ok)
	assert.True(t, r.HasType("Patient"))
	assert.False(t, r.HasType("string1"))
}

func TestFHIRPathModel_HasTypeCoversHierarchy(t *testing.T) {
	m := FHIRPathModel()

	// Every name the hierarchy mentions must be a type the model declares,
	// on both sides. Otherwise IsSubtype could walk into a type that HasType
	// rejects, and the two would contradict each other.
	for child, parent := range m.type2Parent {
		assert.True(t, m.HasType(child), "type2Parent key %q is not declared", child)
		assert.True(t, m.HasType(parent), "type2Parent value %q is not declared", parent)
	}

	// Same for resources, which is a subset of the declared types.
	for name := range m.resources {
		assert.True(t, m.HasType(name), "resource %q is not declared", name)
	}
}

func TestFHIRPathModel_DeterministicOutput(t *testing.T) {
	// Both calls return the same pointer — the singleton is deterministic.
	m1 := FHIRPathModel()
	m2 := FHIRPathModel()
	assert.Same(t, m1, m2)
}

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

func TestFHIRPathModel_DeterministicOutput(t *testing.T) {
	// Both calls return the same pointer — the singleton is deterministic.
	m1 := FHIRPathModel()
	m2 := FHIRPathModel()
	assert.Same(t, m1, m2)
}

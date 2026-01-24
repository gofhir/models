package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Sample StructureDefinition JSON for testing
var samplePatientSD = []byte(`{
	"resourceType": "StructureDefinition",
	"id": "Patient",
	"url": "http://hl7.org/fhir/StructureDefinition/Patient",
	"version": "4.0.1",
	"name": "Patient",
	"title": "Patient",
	"status": "active",
	"kind": "resource",
	"abstract": false,
	"type": "Patient",
	"baseDefinition": "http://hl7.org/fhir/StructureDefinition/DomainResource",
	"derivation": "specialization",
	"snapshot": {
		"element": [
			{
				"id": "Patient",
				"path": "Patient",
				"short": "Information about an individual",
				"min": 0,
				"max": "*"
			},
			{
				"id": "Patient.id",
				"path": "Patient.id",
				"short": "Logical id of this artifact",
				"min": 0,
				"max": "1",
				"type": [{"code": "id"}]
			},
			{
				"id": "Patient.active",
				"path": "Patient.active",
				"short": "Whether this patient record is active",
				"min": 0,
				"max": "1",
				"type": [{"code": "boolean"}]
			},
			{
				"id": "Patient.name",
				"path": "Patient.name",
				"short": "A name associated with the patient",
				"min": 0,
				"max": "*",
				"type": [{"code": "HumanName"}]
			},
			{
				"id": "Patient.deceased[x]",
				"path": "Patient.deceased[x]",
				"short": "Indicates if the patient is deceased",
				"min": 0,
				"max": "1",
				"type": [
					{"code": "boolean"},
					{"code": "dateTime"}
				]
			},
			{
				"id": "Patient.birthDate",
				"path": "Patient.birthDate",
				"short": "The date of birth for the patient",
				"min": 0,
				"max": "1",
				"type": [{"code": "date"}],
				"constraint": [
					{
						"key": "pat-1",
						"severity": "error",
						"human": "Birth date must be in the past",
						"expression": "birthDate <= today()"
					}
				]
			}
		]
	}
}`)

var sampleCodingSD = []byte(`{
	"resourceType": "StructureDefinition",
	"id": "Coding",
	"url": "http://hl7.org/fhir/StructureDefinition/Coding",
	"name": "Coding",
	"title": "Coding",
	"status": "active",
	"kind": "complex-type",
	"abstract": false,
	"type": "Coding",
	"baseDefinition": "http://hl7.org/fhir/StructureDefinition/Element",
	"snapshot": {
		"element": [
			{
				"id": "Coding",
				"path": "Coding",
				"short": "A reference to a code",
				"min": 0,
				"max": "*"
			},
			{
				"id": "Coding.system",
				"path": "Coding.system",
				"short": "Identity of the terminology system",
				"min": 0,
				"max": "1",
				"type": [{"code": "uri"}]
			},
			{
				"id": "Coding.code",
				"path": "Coding.code",
				"short": "Symbol in syntax defined by the system",
				"min": 0,
				"max": "1",
				"type": [{"code": "code"}],
				"binding": {
					"strength": "example",
					"valueSet": "http://hl7.org/fhir/ValueSet/example"
				}
			}
		]
	}
}`)

var sampleBundle = []byte(`{
	"resourceType": "Bundle",
	"id": "test-bundle",
	"type": "collection",
	"entry": [
		{
			"fullUrl": "http://hl7.org/fhir/StructureDefinition/Patient",
			"resource": ` + string(samplePatientSD) + `
		},
		{
			"fullUrl": "http://hl7.org/fhir/StructureDefinition/Coding",
			"resource": ` + string(sampleCodingSD) + `
		}
	]
}`)

func TestParseStructureDefinition(t *testing.T) {
	t.Run("valid patient", func(t *testing.T) {
		sd, err := ParseStructureDefinition(samplePatientSD)
		require.NoError(t, err)
		require.NotNil(t, sd)

		assert.Equal(t, "StructureDefinition", sd.ResourceType)
		assert.Equal(t, "Patient", sd.ID)
		assert.Equal(t, "Patient", sd.Name)
		assert.Equal(t, "resource", sd.Kind)
		assert.False(t, sd.Abstract)
		assert.Equal(t, "http://hl7.org/fhir/StructureDefinition/Patient", sd.URL)
	})

	t.Run("valid coding", func(t *testing.T) {
		sd, err := ParseStructureDefinition(sampleCodingSD)
		require.NoError(t, err)
		require.NotNil(t, sd)

		assert.Equal(t, "Coding", sd.Name)
		assert.Equal(t, "complex-type", sd.Kind)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := ParseStructureDefinition([]byte("not json"))
		assert.Error(t, err)
	})

	t.Run("wrong resource type", func(t *testing.T) {
		_, err := ParseStructureDefinition([]byte(`{"resourceType": "Patient"}`))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "expected resourceType")
	})
}

func TestParseBundle(t *testing.T) {
	t.Run("valid bundle", func(t *testing.T) {
		bundle, err := ParseBundle(sampleBundle)
		require.NoError(t, err)
		require.NotNil(t, bundle)

		assert.Equal(t, "Bundle", bundle.ResourceType)
		assert.Equal(t, "test-bundle", bundle.ID)
		assert.Len(t, bundle.Entry, 2)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := ParseBundle([]byte("not json"))
		assert.Error(t, err)
	})
}

func TestExtractStructureDefinitions(t *testing.T) {
	bundle, err := ParseBundle(sampleBundle)
	require.NoError(t, err)

	sds, err := ExtractStructureDefinitions(bundle)
	require.NoError(t, err)
	require.Len(t, sds, 2)

	names := make(map[string]bool)
	for _, sd := range sds {
		names[sd.Name] = true
	}
	assert.True(t, names["Patient"])
	assert.True(t, names["Coding"])
}

func TestStructureDefinitionMethods(t *testing.T) {
	sd, err := ParseStructureDefinition(samplePatientSD)
	require.NoError(t, err)

	t.Run("IsResource", func(t *testing.T) {
		assert.True(t, sd.IsResource())
		assert.False(t, sd.IsComplexType())
		assert.False(t, sd.IsPrimitive())
	})

	t.Run("GetElements", func(t *testing.T) {
		elements := sd.GetElements()
		assert.NotEmpty(t, elements)
		assert.Equal(t, "Patient", elements[0].Path)
	})
}

func TestElementDefinitionMethods(t *testing.T) {
	sd, err := ParseStructureDefinition(samplePatientSD)
	require.NoError(t, err)
	elements := sd.GetElements()

	t.Run("IsChoiceType", func(t *testing.T) {
		// Find deceased[x]
		var deceasedElem *ElementDefinition
		for i := range elements {
			if elements[i].Path == "Patient.deceased[x]" {
				deceasedElem = &elements[i]
				break
			}
		}
		require.NotNil(t, deceasedElem)
		assert.True(t, deceasedElem.IsChoiceType())
		assert.Equal(t, "deceased", deceasedElem.GetBaseName())
	})

	t.Run("IsArray", func(t *testing.T) {
		// Find name (max = "*")
		var nameElem *ElementDefinition
		for i := range elements {
			if elements[i].Path == "Patient.name" {
				nameElem = &elements[i]
				break
			}
		}
		require.NotNil(t, nameElem)
		assert.True(t, nameElem.IsArray())
		assert.False(t, nameElem.IsRequired())
	})

	t.Run("IsRequired", func(t *testing.T) {
		// All patient elements have min=0
		for _, elem := range elements {
			assert.False(t, elem.IsRequired())
		}
	})
}

func TestFilterByKind(t *testing.T) {
	patientSD, err := ParseStructureDefinition(samplePatientSD)
	require.NoError(t, err)
	codingSD, err := ParseStructureDefinition(sampleCodingSD)
	require.NoError(t, err)

	sds := []*StructureDefinition{patientSD, codingSD}

	t.Run("filter resources", func(t *testing.T) {
		filtered := FilterByKind(sds, KindResource)
		require.Len(t, filtered, 1)
		assert.Equal(t, "Patient", filtered[0].Name)
	})

	t.Run("filter complex types", func(t *testing.T) {
		filtered := FilterByKind(sds, KindComplexType)
		require.Len(t, filtered, 1)
		assert.Equal(t, "Coding", filtered[0].Name)
	})

	t.Run("filter multiple kinds", func(t *testing.T) {
		filtered := FilterByKind(sds, KindResource, KindComplexType)
		require.Len(t, filtered, 2)
	})
}

func TestFilterNonAbstract(t *testing.T) {
	sd1 := &StructureDefinition{Name: "Concrete", Abstract: false}
	sd2 := &StructureDefinition{Name: "Abstract", Abstract: true}

	filtered := FilterNonAbstract([]*StructureDefinition{sd1, sd2})
	require.Len(t, filtered, 1)
	assert.Equal(t, "Concrete", filtered[0].Name)
}

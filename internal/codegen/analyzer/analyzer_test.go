package analyzer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofhir/fhir/internal/codegen/parser"
)

var samplePatientSD = []byte(`{
	"resourceType": "StructureDefinition",
	"id": "Patient",
	"url": "http://hl7.org/fhir/StructureDefinition/Patient",
	"name": "Patient",
	"title": "Patient Resource",
	"status": "active",
	"kind": "resource",
	"abstract": false,
	"type": "Patient",
	"baseDefinition": "http://hl7.org/fhir/StructureDefinition/DomainResource",
	"snapshot": {
		"element": [
			{
				"id": "Patient",
				"path": "Patient",
				"short": "Information about an individual",
				"min": 0,
				"max": "*",
				"constraint": [
					{
						"key": "pat-1",
						"severity": "error",
						"human": "Contact should have a name or organization",
						"expression": "contact.exists() implies (contact.all(name.exists() or organization.exists()))"
					}
				]
			},
			{
				"id": "Patient.id",
				"path": "Patient.id",
				"short": "Logical id",
				"min": 0,
				"max": "1",
				"type": [{"code": "id"}]
			},
			{
				"id": "Patient.active",
				"path": "Patient.active",
				"short": "Whether record is active",
				"min": 0,
				"max": "1",
				"type": [{"code": "boolean"}]
			},
			{
				"id": "Patient.name",
				"path": "Patient.name",
				"short": "A name for the patient",
				"min": 0,
				"max": "*",
				"type": [{"code": "HumanName"}]
			},
			{
				"id": "Patient.birthDate",
				"path": "Patient.birthDate",
				"short": "Date of birth",
				"min": 0,
				"max": "1",
				"type": [{"code": "date"}]
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
				"id": "Patient.gender",
				"path": "Patient.gender",
				"short": "male | female | other | unknown",
				"min": 0,
				"max": "1",
				"type": [{"code": "code"}],
				"binding": {
					"strength": "required",
					"valueSet": "http://hl7.org/fhir/ValueSet/administrative-gender"
				}
			}
		]
	}
}`)

func TestAnalyzer_Analyze(t *testing.T) {
	sd, err := parser.ParseStructureDefinition(samplePatientSD)
	require.NoError(t, err)

	analyzer := NewAnalyzer([]*parser.StructureDefinition{sd}, nil)

	t.Run("analyze patient", func(t *testing.T) {
		result, err := analyzer.Analyze(sd)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, "Patient", result.Name)
		assert.Equal(t, "resource", result.Kind)
		assert.Equal(t, "Patient Resource", result.Description)
		assert.False(t, result.IsAbstract)

		// Should have constraints
		require.Len(t, result.Constraints, 1)
		assert.Equal(t, "pat-1", result.Constraints[0].Key)
	})

	t.Run("properties exist", func(t *testing.T) {
		result, err := analyzer.Analyze(sd)
		require.NoError(t, err)

		propMap := make(map[string]AnalyzedProperty)
		for _, p := range result.Properties {
			propMap[p.Name] = p
		}

		// Check id field
		idProp, ok := propMap["Id"]
		require.True(t, ok, "should have Id property")
		assert.Equal(t, "id", idProp.JSONName)
		assert.Equal(t, "*string", idProp.GoType)
		assert.True(t, idProp.IsPrimitive)
		assert.True(t, idProp.HasExtension)

		// Check active field
		activeProp, ok := propMap["Active"]
		require.True(t, ok, "should have Active property")
		assert.Equal(t, "active", activeProp.JSONName)
		assert.Equal(t, "*bool", activeProp.GoType)
		assert.True(t, activeProp.IsPrimitive)

		// Check name field (array)
		nameProp, ok := propMap["Name"]
		require.True(t, ok, "should have Name property")
		assert.Equal(t, "name", nameProp.JSONName)
		assert.Equal(t, "[]HumanName", nameProp.GoType)
		assert.True(t, nameProp.IsArray)
		assert.False(t, nameProp.IsPrimitive)

		// Check gender field with binding
		genderProp, ok := propMap["Gender"]
		require.True(t, ok, "should have Gender property")
		require.NotNil(t, genderProp.Binding)
		assert.Equal(t, "required", genderProp.Binding.Strength)
	})

	t.Run("choice type expansion", func(t *testing.T) {
		result, err := analyzer.Analyze(sd)
		require.NoError(t, err)

		propMap := make(map[string]AnalyzedProperty)
		for _, p := range result.Properties {
			propMap[p.Name] = p
		}

		// Choice type should be expanded to multiple fields
		deceasedBool, ok := propMap["DeceasedBoolean"]
		require.True(t, ok, "should have DeceasedBoolean property")
		assert.Equal(t, "deceasedBoolean", deceasedBool.JSONName)
		assert.Equal(t, "*bool", deceasedBool.GoType)
		assert.True(t, deceasedBool.IsChoice)
		assert.Contains(t, deceasedBool.ChoiceTypes, "boolean")
		assert.Contains(t, deceasedBool.ChoiceTypes, "dateTime")

		deceasedDateTime, ok := propMap["DeceasedDateTime"]
		require.True(t, ok, "should have DeceasedDateTime property")
		assert.Equal(t, "deceasedDateTime", deceasedDateTime.JSONName)
		assert.Equal(t, "*string", deceasedDateTime.GoType)
		assert.True(t, deceasedDateTime.IsChoice)

		// Should have extension fields for primitive choice types
		_, hasExtBool := propMap["DeceasedBooleanExt"]
		assert.True(t, hasExtBool, "should have DeceasedBooleanExt for primitive")

		_, hasExtDateTime := propMap["DeceasedDateTimeExt"]
		assert.True(t, hasExtDateTime, "should have DeceasedDateTimeExt for primitive")
	})

	t.Run("primitive extension fields", func(t *testing.T) {
		result, err := analyzer.Analyze(sd)
		require.NoError(t, err)

		propMap := make(map[string]AnalyzedProperty)
		for _, p := range result.Properties {
			propMap[p.Name] = p
		}

		// Primitives should have HasExtension = true
		idProp := propMap["Id"]
		assert.True(t, idProp.HasExtension)

		// Complex types should not
		nameProp := propMap["Name"]
		assert.False(t, nameProp.HasExtension)
	})
}

func TestAnalyzer_NilInput(t *testing.T) {
	analyzer := NewAnalyzer(nil, nil)
	result, err := analyzer.Analyze(nil)
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestFHIRToGoType(t *testing.T) {
	tests := []struct {
		fhirType string
		expected string
	}{
		// Primitives
		{"boolean", "bool"},
		{"integer", "int"},
		{"integer64", "int64"},
		{"decimal", "float64"},
		{"string", "string"},
		{"uri", "string"},
		{"date", "string"},
		{"dateTime", "string"},
		{"code", "string"},
		{"id", "string"},
		{"unsignedInt", "uint32"},
		{"positiveInt", "uint32"},

		// Complex types
		{"Coding", "Coding"},
		{"CodeableConcept", "CodeableConcept"},
		{"Reference", "Reference"},
		{"HumanName", "HumanName"},
		{"Extension", "Extension"},

		// Resources (pass through)
		{"Patient", "Patient"},
		{"Observation", "Observation"},
	}

	for _, tt := range tests {
		t.Run(tt.fhirType, func(t *testing.T) {
			result := FHIRToGoType(tt.fhirType)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsPrimitiveType(t *testing.T) {
	primitives := []string{"boolean", "integer", "string", "decimal", "date", "dateTime", "code", "uri"}
	for _, p := range primitives {
		assert.True(t, IsPrimitiveType(p), "%s should be primitive", p)
	}

	nonPrimitives := []string{"Coding", "Patient", "HumanName", "Reference"}
	for _, np := range nonPrimitives {
		assert.False(t, IsPrimitiveType(np), "%s should not be primitive", np)
	}
}

func TestHelperFunctions(t *testing.T) {
	t.Run("toPascalCase", func(t *testing.T) {
		assert.Equal(t, "Name", toPascalCase("name"))
		assert.Equal(t, "BirthDate", toPascalCase("birthDate"))
		assert.Equal(t, "", toPascalCase(""))
	})

	t.Run("toLowerFirst", func(t *testing.T) {
		assert.Equal(t, "name", toLowerFirst("Name"))
		assert.Equal(t, "birthDate", toLowerFirst("BirthDate"))
		assert.Equal(t, "", toLowerFirst(""))
	})

	t.Run("toGoFieldName reserved words", func(t *testing.T) {
		assert.Equal(t, "Class", toGoFieldName("class"))
		assert.Equal(t, "Type", toGoFieldName("type"))
		assert.Equal(t, "Import", toGoFieldName("import"))
	})
}

// Sample TestScript-like SD with contentReference for testing
var sampleTestScriptSD = []byte(`{
	"resourceType": "StructureDefinition",
	"id": "TestScript",
	"url": "http://hl7.org/fhir/StructureDefinition/TestScript",
	"name": "TestScript",
	"title": "TestScript Resource",
	"status": "active",
	"kind": "resource",
	"abstract": false,
	"type": "TestScript",
	"baseDefinition": "http://hl7.org/fhir/StructureDefinition/DomainResource",
	"snapshot": {
		"element": [
			{
				"id": "TestScript",
				"path": "TestScript",
				"short": "Describes a set of tests",
				"min": 0,
				"max": "*"
			},
			{
				"id": "TestScript.id",
				"path": "TestScript.id",
				"short": "Logical id",
				"min": 0,
				"max": "1",
				"type": [{"code": "id"}]
			},
			{
				"id": "TestScript.setup",
				"path": "TestScript.setup",
				"short": "A series of required setup operations",
				"min": 0,
				"max": "1",
				"type": [{"code": "BackboneElement"}]
			},
			{
				"id": "TestScript.setup.action",
				"path": "TestScript.setup.action",
				"short": "A setup operation or assert",
				"min": 1,
				"max": "*",
				"type": [{"code": "BackboneElement"}]
			},
			{
				"id": "TestScript.setup.action.operation",
				"path": "TestScript.setup.action.operation",
				"short": "The setup operation to perform",
				"min": 0,
				"max": "1",
				"type": [{"code": "BackboneElement"}]
			},
			{
				"id": "TestScript.setup.action.operation.type",
				"path": "TestScript.setup.action.operation.type",
				"short": "The operation code type",
				"min": 0,
				"max": "1",
				"type": [{"code": "Coding"}]
			},
			{
				"id": "TestScript.setup.action.assert",
				"path": "TestScript.setup.action.assert",
				"short": "The assertion to perform",
				"min": 0,
				"max": "1",
				"type": [{"code": "BackboneElement"}]
			},
			{
				"id": "TestScript.setup.action.assert.label",
				"path": "TestScript.setup.action.assert.label",
				"short": "Tracking/logging assertion label",
				"min": 0,
				"max": "1",
				"type": [{"code": "string"}]
			},
			{
				"id": "TestScript.test",
				"path": "TestScript.test",
				"short": "A test in this script",
				"min": 0,
				"max": "*",
				"type": [{"code": "BackboneElement"}]
			},
			{
				"id": "TestScript.test.action",
				"path": "TestScript.test.action",
				"short": "A test operation or assert",
				"min": 1,
				"max": "*",
				"type": [{"code": "BackboneElement"}]
			},
			{
				"id": "TestScript.test.action.operation",
				"path": "TestScript.test.action.operation",
				"short": "The setup operation to perform",
				"min": 0,
				"max": "1",
				"contentReference": "#TestScript.setup.action.operation"
			},
			{
				"id": "TestScript.test.action.assert",
				"path": "TestScript.test.action.assert",
				"short": "The setup assertion to perform",
				"min": 0,
				"max": "1",
				"contentReference": "#TestScript.setup.action.assert"
			},
			{
				"id": "TestScript.teardown",
				"path": "TestScript.teardown",
				"short": "A series of required clean up steps",
				"min": 0,
				"max": "1",
				"type": [{"code": "BackboneElement"}]
			},
			{
				"id": "TestScript.teardown.action",
				"path": "TestScript.teardown.action",
				"short": "One or more teardown operations",
				"min": 1,
				"max": "*",
				"type": [{"code": "BackboneElement"}]
			},
			{
				"id": "TestScript.teardown.action.operation",
				"path": "TestScript.teardown.action.operation",
				"short": "The teardown operation to perform",
				"min": 1,
				"max": "1",
				"contentReference": "#TestScript.setup.action.operation"
			}
		]
	}
}`)

func TestAnalyzer_ResolveContentReference(t *testing.T) {
	sd, err := parser.ParseStructureDefinition(sampleTestScriptSD)
	require.NoError(t, err)

	analyzer := NewAnalyzer([]*parser.StructureDefinition{sd}, nil)

	t.Run("invalid reference without hash", func(t *testing.T) {
		goType, isBackbone, backboneName := analyzer.resolveContentReference("InvalidRef", false)
		assert.Equal(t, "*interface{}", goType)
		assert.False(t, isBackbone)
		assert.Empty(t, backboneName)
	})

	t.Run("invalid reference without hash array", func(t *testing.T) {
		goType, isBackbone, backboneName := analyzer.resolveContentReference("InvalidRef", true)
		assert.Equal(t, "[]interface{}", goType)
		assert.False(t, isBackbone)
		assert.Empty(t, backboneName)
	})

	t.Run("short path reference", func(t *testing.T) {
		goType, isBackbone, backboneName := analyzer.resolveContentReference("#TestScript", false)
		assert.Equal(t, "*interface{}", goType)
		assert.False(t, isBackbone)
		assert.Empty(t, backboneName)
	})

	t.Run("valid backbone reference", func(t *testing.T) {
		goType, isBackbone, backboneName := analyzer.resolveContentReference("#TestScript.setup.action.operation", false)
		assert.Equal(t, "*TestScriptSetupActionOperation", goType)
		assert.True(t, isBackbone)
		assert.Equal(t, "TestScriptSetupActionOperation", backboneName)
	})

	t.Run("valid backbone reference array", func(t *testing.T) {
		goType, isBackbone, backboneName := analyzer.resolveContentReference("#TestScript.setup.action.operation", true)
		assert.Equal(t, "[]TestScriptSetupActionOperation", goType)
		assert.True(t, isBackbone)
		assert.Equal(t, "TestScriptSetupActionOperation", backboneName)
	})

	t.Run("reference to unknown resource", func(t *testing.T) {
		goType, isBackbone, backboneName := analyzer.resolveContentReference("#UnknownResource.some.path", false)
		// Should still generate a reasonable type name
		assert.Equal(t, "*UnknownResourceSomePath", goType)
		assert.True(t, isBackbone)
		assert.Equal(t, "UnknownResourceSomePath", backboneName)
	})
}

func TestAnalyzer_ContentReferenceInBackboneElements(t *testing.T) {
	sd, err := parser.ParseStructureDefinition(sampleTestScriptSD)
	require.NoError(t, err)

	analyzer := NewAnalyzer([]*parser.StructureDefinition{sd}, nil)

	t.Run("backbone elements with contentReference", func(t *testing.T) {
		result, err := analyzer.Analyze(sd)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Find the backbone types
		backboneMap := make(map[string]*AnalyzedType)
		for _, bb := range result.BackboneTypes {
			backboneMap[bb.Name] = bb
		}

		// Check TestScriptTestAction has Operation and Assert from contentReference
		testAction, ok := backboneMap["TestScriptTestAction"]
		require.True(t, ok, "should have TestScriptTestAction backbone")

		propMap := make(map[string]AnalyzedProperty)
		for _, p := range testAction.Properties {
			propMap[p.Name] = p
		}

		// Operation should reference TestScriptSetupActionOperation
		opProp, ok := propMap["Operation"]
		require.True(t, ok, "TestScriptTestAction should have Operation property")
		assert.Equal(t, "*TestScriptSetupActionOperation", opProp.GoType)
		assert.Equal(t, "operation", opProp.JSONName)
		assert.True(t, opProp.IsBackbone)
		assert.Equal(t, "TestScriptSetupActionOperation", opProp.BackboneType)

		// Assert should reference TestScriptSetupActionAssert
		assertProp, ok := propMap["Assert"]
		require.True(t, ok, "TestScriptTestAction should have Assert property")
		assert.Equal(t, "*TestScriptSetupActionAssert", assertProp.GoType)
		assert.Equal(t, "assert", assertProp.JSONName)
		assert.True(t, assertProp.IsBackbone)

		// Check TestScriptTeardownAction has Operation from contentReference
		teardownAction, ok := backboneMap["TestScriptTeardownAction"]
		require.True(t, ok, "should have TestScriptTeardownAction backbone")

		teardownPropMap := make(map[string]AnalyzedProperty)
		for _, p := range teardownAction.Properties {
			teardownPropMap[p.Name] = p
		}

		teardownOpProp, ok := teardownPropMap["Operation"]
		require.True(t, ok, "TestScriptTeardownAction should have Operation property")
		assert.Equal(t, "*TestScriptSetupActionOperation", teardownOpProp.GoType)
		assert.True(t, teardownOpProp.IsRequired) // min: 1 in spec
	})
}

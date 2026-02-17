package r4b

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr[T any](v T) *T { return &v }

func TestPatient_MarshalXML_Basic(t *testing.T) {
	gender := AdministrativeGenderMale

	patient := Patient{
		Id:        ptr("123"),
		Active:    ptr(true),
		Gender:    &gender,
		BirthDate: ptr("1974-12-25"),
	}

	data, err := MarshalResourceXML(&patient)
	require.NoError(t, err)

	xml := string(data)

	// Has XML declaration
	assert.True(t, strings.HasPrefix(xml, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>"))

	// Root element with FHIR namespace
	assert.Contains(t, xml, `<Patient xmlns="http://hl7.org/fhir">`)

	// Primitives as self-closing elements with value attribute (FHIR convention)
	assert.Contains(t, xml, `<id value="123"/>`)
	assert.Contains(t, xml, `<active value="true"/>`)
	assert.Contains(t, xml, `<gender value="male"/>`)
	assert.Contains(t, xml, `<birthDate value="1974-12-25"/>`)

	// Closing tag
	assert.Contains(t, xml, `</Patient>`)
}

func TestPatient_MarshalXML_OmitsEmpty(t *testing.T) {
	patient := Patient{
		Id: ptr("empty-test"),
	}

	data, err := MarshalResourceXML(&patient)
	require.NoError(t, err)

	xml := string(data)

	// Should NOT contain fields that are nil
	assert.NotContains(t, xml, "active")
	assert.NotContains(t, xml, "gender")
	assert.NotContains(t, xml, "birthDate")
	assert.NotContains(t, xml, "meta")
}

func TestPatient_MarshalXML_Narrative(t *testing.T) {
	status := NarrativeStatusGenerated
	patient := Patient{
		Id: ptr("narrative-test"),
		Text: &Narrative{
			Status: &status,
			Div:    ptr(`<div xmlns="http://www.w3.org/1999/xhtml"><p>Test patient</p></div>`),
		},
	}

	data, err := MarshalResourceXML(&patient)
	require.NoError(t, err)

	xml := string(data)

	// Raw XHTML injected verbatim
	assert.Contains(t, xml, `<div xmlns="http://www.w3.org/1999/xhtml"><p>Test patient</p></div>`)
	// Narrative status as self-closing element
	assert.Contains(t, xml, `<status value="generated"/>`)
}

func TestPatient_MarshalXML_PrimitiveExtensions(t *testing.T) {
	extURL := "http://example.org/ext"
	extValue := "some-value"

	patient := Patient{
		Id:        ptr("ext-test"),
		BirthDate: ptr("1974-12-25"),
		BirthDateExt: &Element{
			Extension: []Extension{
				{
					Url:         extURL,
					ValueString: &extValue,
				},
			},
		},
	}

	data, err := MarshalResourceXML(&patient)
	require.NoError(t, err)

	xml := string(data)

	// BirthDate should have value attribute AND extension child (NOT self-closing)
	assert.Contains(t, xml, `<birthDate value="1974-12-25">`)
	assert.Contains(t, xml, `<extension url="http://example.org/ext">`)
	assert.Contains(t, xml, `<valueString value="some-value"/>`)
	assert.Contains(t, xml, `</birthDate>`)
}

func TestPatient_MarshalXML_ChoiceType(t *testing.T) {
	patient := Patient{
		Id:              ptr("choice-test"),
		DeceasedBoolean: ptr(false),
	}

	data, err := MarshalResourceXML(&patient)
	require.NoError(t, err)

	xml := string(data)

	assert.Contains(t, xml, `<deceasedBoolean value="false"/>`)
}

func TestPatient_MarshalXML_RepeatingElements(t *testing.T) {
	patient := Patient{
		Id: ptr("repeat-test"),
		Name: []HumanName{
			{
				Family: ptr("Smith"),
				Given:  []string{"John", "Robert"},
			},
		},
	}

	data, err := MarshalResourceXML(&patient)
	require.NoError(t, err)

	xml := string(data)

	// Each given is a separate self-closing element
	assert.Contains(t, xml, `<given value="John"/>`)
	assert.Contains(t, xml, `<given value="Robert"/>`)
	assert.Contains(t, xml, `<family value="Smith"/>`)
}

func TestPatient_MarshalXML_ContainedResource(t *testing.T) {
	org := &Organization{
		Id:   ptr("org1"),
		Name: ptr("Test Org"),
	}

	patient := Patient{
		Id:        ptr("contained-test"),
		Contained: []Resource{org},
		ManagingOrganization: &Reference{
			Reference: ptr("#org1"),
		},
	}

	data, err := MarshalResourceXML(&patient)
	require.NoError(t, err)

	xml := string(data)

	// Contained wraps the resource in <contained><ResourceType>...</ResourceType></contained>
	assert.Contains(t, xml, `<contained>`)
	assert.Contains(t, xml, `<Organization>`)
	assert.Contains(t, xml, `<id value="org1"/>`)
	assert.Contains(t, xml, `<name value="Test Org"/>`)
	assert.Contains(t, xml, `</Organization>`)
	assert.Contains(t, xml, `</contained>`)
}

func TestExtension_MarshalXML_UrlAttribute(t *testing.T) {
	ext := Extension{
		Url:         "http://example.org/my-ext",
		ValueString: ptr("hello"),
	}

	data, err := MarshalResourceXMLIndent(&Patient{
		Id:        ptr("ext-url-test"),
		Extension: []Extension{ext},
	}, "", "  ")
	require.NoError(t, err)

	xml := string(data)

	// Extension url is an XML attribute, has child so NOT self-closing
	assert.Contains(t, xml, `<extension url="http://example.org/my-ext">`)
	// valueString is self-closing
	assert.Contains(t, xml, `<valueString value="hello"/>`)
}

func TestCoding_MarshalXML_IdAttribute(t *testing.T) {
	coding := Coding{
		Id:     ptr("c1"),
		System: ptr("http://loinc.org"),
		Code:   ptr("12345"),
	}

	// Marshal as part of a CodeableConcept inside a resource
	obs := Observation{
		Id: ptr("obs-coding-test"),
		Code: CodeableConcept{
			Coding: []Coding{coding},
		},
	}

	data, err := MarshalResourceXML(&obs)
	require.NoError(t, err)

	xml := string(data)

	// Coding id is an XML attribute, has children so NOT self-closing
	assert.Contains(t, xml, `<coding id="c1">`)
	// Primitives inside coding are self-closing
	assert.Contains(t, xml, `<system value="http://loinc.org"/>`)
	assert.Contains(t, xml, `<code value="12345"/>`)
}

func TestMarshalResourceXML_Declaration(t *testing.T) {
	patient := Patient{Id: ptr("decl-test")}

	data, err := MarshalResourceXML(&patient)
	require.NoError(t, err)

	xml := string(data)

	assert.True(t, strings.HasPrefix(xml, `<?xml version="1.0" encoding="UTF-8"?>`))
}

func TestMarshalResourceXMLIndent(t *testing.T) {
	patient := Patient{
		Id:     ptr("indent-test"),
		Active: ptr(true),
	}

	data, err := MarshalResourceXMLIndent(&patient, "", "  ")
	require.NoError(t, err)

	xml := string(data)

	// Should have indentation with self-closing tags
	assert.Contains(t, xml, "\n  <id")
	assert.Contains(t, xml, "\n  <active")
}

func TestBundle_MarshalXML_EntryResource(t *testing.T) {
	bundleType := BundleTypeSearchset

	patient := &Patient{
		Id: ptr("p1"),
	}

	bundle := Bundle{
		Id:   ptr("bundle-test"),
		Type: &bundleType,
		Entry: []BundleEntry{
			{
				Resource: patient,
			},
		},
	}

	data, err := MarshalResourceXML(&bundle)
	require.NoError(t, err)

	xml := string(data)

	assert.Contains(t, xml, `<Bundle xmlns="http://hl7.org/fhir">`)
	assert.Contains(t, xml, `<type value="searchset"/>`)
	assert.Contains(t, xml, `<entry>`)
	// Entry resource is encoded with its type name (no namespace)
	assert.Contains(t, xml, `<Patient>`)
	assert.Contains(t, xml, `<id value="p1"/>`)
	assert.Contains(t, xml, `</Patient>`)
	assert.Contains(t, xml, `</entry>`)
}

func TestObservation_MarshalXML_ComplexStructure(t *testing.T) {
	status := ObservationStatusFinal

	obs := Observation{
		Id:     ptr("obs-1"),
		Status: &status,
		Code: CodeableConcept{
			Coding: []Coding{
				{
					System:  ptr("http://loinc.org"),
					Code:    ptr("8480-6"),
					Display: ptr("Systolic blood pressure"),
				},
			},
		},
		ValueQuantity: &Quantity{
			Value:  NewDecimalFromFloat64(120.0),
			Unit:   ptr("mmHg"),
			System: ptr("http://unitsofmeasure.org"),
			Code:   ptr("mm[Hg]"),
		},
	}

	data, err := MarshalResourceXML(&obs)
	require.NoError(t, err)

	xml := string(data)

	assert.Contains(t, xml, `<Observation xmlns="http://hl7.org/fhir">`)
	assert.Contains(t, xml, `<status value="final"/>`)
	assert.Contains(t, xml, `<system value="http://loinc.org"/>`)
	assert.Contains(t, xml, `<code value="8480-6"/>`)
	assert.Contains(t, xml, `<display value="Systolic blood pressure"/>`)
	assert.Contains(t, xml, `<value value="120"/>`)
	assert.Contains(t, xml, `<unit value="mmHg"/>`)
}

func TestPatient_MarshalXML_ElementOrder(t *testing.T) {
	// FHIR requires element order to match StructureDefinition
	gender := AdministrativeGenderFemale
	patient := Patient{
		Id:        ptr("order-test"),
		Active:    ptr(true),
		Gender:    &gender,
		BirthDate: ptr("1990-01-01"),
		Name: []HumanName{
			{Family: ptr("Test")},
		},
	}

	data, err := MarshalResourceXML(&patient)
	require.NoError(t, err)

	xml := string(data)

	// id must come before active, active before name, name before gender, gender before birthDate
	idPos := strings.Index(xml, "<id ")
	activePos := strings.Index(xml, "<active ")
	namePos := strings.Index(xml, "<name>")
	genderPos := strings.Index(xml, "<gender ")
	birthDatePos := strings.Index(xml, "<birthDate ")

	assert.Less(t, idPos, activePos, "id should come before active")
	assert.Less(t, activePos, namePos, "active should come before name")
	assert.Less(t, namePos, genderPos, "name should come before gender")
	assert.Less(t, genderPos, birthDatePos, "gender should come before birthDate")
}

func TestMarshalXML_SelfClosingTags(t *testing.T) {
	// Verify FHIR convention: empty elements use self-closing tags
	patient := Patient{
		Id:        ptr("self-close-test"),
		Active:    ptr(true),
		BirthDate: ptr("2000-01-01"),
	}

	data, err := MarshalResourceXML(&patient)
	require.NoError(t, err)

	xml := string(data)

	// Self-closing form (FHIR convention)
	assert.Contains(t, xml, `<id value="self-close-test"/>`)
	assert.Contains(t, xml, `<active value="true"/>`)
	assert.Contains(t, xml, `<birthDate value="2000-01-01"/>`)

	// Should NOT contain explicit close tags for empty primitives
	assert.NotContains(t, xml, `</id>`)
	assert.NotContains(t, xml, `</active>`)
	assert.NotContains(t, xml, `</birthDate>`)
}

package r4b

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatient_UnmarshalXML_Basic(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Patient xmlns="http://hl7.org/fhir"><id value="123"/><active value="true"/><gender value="male"/><birthDate value="1974-12-25"/></Patient>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	patient, ok := resource.(*Patient)
	require.True(t, ok)
	assert.Equal(t, "123", *patient.Id)
	assert.Equal(t, true, *patient.Active)
	assert.Equal(t, AdministrativeGenderMale, *patient.Gender)
	assert.Equal(t, "1974-12-25", *patient.BirthDate)
}

func TestPatient_UnmarshalXML_OmitsEmpty(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Patient xmlns="http://hl7.org/fhir"><id value="empty-test"/></Patient>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	patient := resource.(*Patient)
	assert.Equal(t, "empty-test", *patient.Id)
	assert.Nil(t, patient.Active)
	assert.Nil(t, patient.Gender)
	assert.Nil(t, patient.BirthDate)
	assert.Nil(t, patient.Meta)
}

func TestPatient_UnmarshalXML_Narrative(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Patient xmlns="http://hl7.org/fhir"><id value="narrative-test"/><text><status value="generated"/><div xmlns="http://www.w3.org/1999/xhtml"><p>Test patient</p></div></text></Patient>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	patient := resource.(*Patient)
	require.NotNil(t, patient.Text)
	assert.Equal(t, NarrativeStatusGenerated, *patient.Text.Status)
	require.NotNil(t, patient.Text.Div)
	assert.Contains(t, *patient.Text.Div, `<div xmlns="http://www.w3.org/1999/xhtml">`)
	assert.Contains(t, *patient.Text.Div, `<p>Test patient</p>`)
}

func TestPatient_UnmarshalXML_PrimitiveExtensions(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Patient xmlns="http://hl7.org/fhir"><id value="ext-test"/><birthDate value="1974-12-25"><extension url="http://example.org/ext"><valueString value="some-value"/></extension></birthDate></Patient>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	patient := resource.(*Patient)
	assert.Equal(t, "1974-12-25", *patient.BirthDate)
	require.NotNil(t, patient.BirthDateExt)
	require.Len(t, patient.BirthDateExt.Extension, 1)
	assert.Equal(t, "http://example.org/ext", patient.BirthDateExt.Extension[0].Url)
	assert.Equal(t, "some-value", *patient.BirthDateExt.Extension[0].ValueString)
}

func TestPatient_UnmarshalXML_ChoiceType(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Patient xmlns="http://hl7.org/fhir"><id value="choice-test"/><deceasedBoolean value="false"/></Patient>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	patient := resource.(*Patient)
	require.NotNil(t, patient.DeceasedBoolean)
	assert.Equal(t, false, *patient.DeceasedBoolean)
}

func TestPatient_UnmarshalXML_RepeatingElements(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Patient xmlns="http://hl7.org/fhir"><id value="repeat-test"/><name><family value="Smith"/><given value="John"/><given value="Robert"/></name></Patient>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	patient := resource.(*Patient)
	require.Len(t, patient.Name, 1)
	assert.Equal(t, "Smith", *patient.Name[0].Family)
	require.Len(t, patient.Name[0].Given, 2)
	assert.Equal(t, "John", patient.Name[0].Given[0])
	assert.Equal(t, "Robert", patient.Name[0].Given[1])
}

func TestPatient_UnmarshalXML_ContainedResource(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Patient xmlns="http://hl7.org/fhir"><id value="contained-test"/><contained><Organization><id value="org1"/><name value="Test Org"/></Organization></contained><managingOrganization><reference value="#org1"/></managingOrganization></Patient>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	patient := resource.(*Patient)
	require.Len(t, patient.Contained, 1)
	org, ok := patient.Contained[0].(*Organization)
	require.True(t, ok)
	assert.Equal(t, "org1", *org.Id)
	assert.Equal(t, "Test Org", *org.Name)
	require.NotNil(t, patient.ManagingOrganization)
	assert.Equal(t, "#org1", *patient.ManagingOrganization.Reference)
}

func TestExtension_UnmarshalXML_UrlAttribute(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Patient xmlns="http://hl7.org/fhir"><id value="ext-url-test"/><extension url="http://example.org/my-ext"><valueString value="hello"/></extension></Patient>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	patient := resource.(*Patient)
	require.Len(t, patient.Extension, 1)
	assert.Equal(t, "http://example.org/my-ext", patient.Extension[0].Url)
	assert.Equal(t, "hello", *patient.Extension[0].ValueString)
}

func TestCoding_UnmarshalXML_IdAttribute(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Observation xmlns="http://hl7.org/fhir"><id value="obs-coding-test"/><code><coding id="c1"><system value="http://loinc.org"/><code value="12345"/></coding></code></Observation>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	obs := resource.(*Observation)
	require.Len(t, obs.Code.Coding, 1)
	assert.Equal(t, "c1", *obs.Code.Coding[0].Id)
	assert.Equal(t, "http://loinc.org", *obs.Code.Coding[0].System)
	assert.Equal(t, "12345", *obs.Code.Coding[0].Code)
}

func TestBundle_UnmarshalXML_EntryResource(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Bundle xmlns="http://hl7.org/fhir"><id value="bundle-test"/><type value="searchset"/><entry><Patient><id value="p1"/></Patient></entry></Bundle>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	bundle := resource.(*Bundle)
	assert.Equal(t, "bundle-test", *bundle.Id)
	assert.Equal(t, BundleTypeSearchset, *bundle.Type)
	require.Len(t, bundle.Entry, 1)
	require.NotNil(t, bundle.Entry[0].Resource)
	patient, ok := bundle.Entry[0].Resource.(*Patient)
	require.True(t, ok)
	assert.Equal(t, "p1", *patient.Id)
}

func TestObservation_UnmarshalXML_ComplexStructure(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><Observation xmlns="http://hl7.org/fhir"><id value="obs-1"/><status value="final"/><code><coding><system value="http://loinc.org"/><code value="8480-6"/><display value="Systolic blood pressure"/></coding></code><valueQuantity><value value="120"/><unit value="mmHg"/><system value="http://unitsofmeasure.org"/><code value="mm[Hg]"/></valueQuantity></Observation>`)

	resource, err := UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	obs := resource.(*Observation)
	assert.Equal(t, "obs-1", *obs.Id)
	assert.Equal(t, ObservationStatusFinal, *obs.Status)
	require.Len(t, obs.Code.Coding, 1)
	assert.Equal(t, "http://loinc.org", *obs.Code.Coding[0].System)
	assert.Equal(t, "8480-6", *obs.Code.Coding[0].Code)
	assert.Equal(t, "Systolic blood pressure", *obs.Code.Coding[0].Display)
	require.NotNil(t, obs.ValueQuantity)
	assert.Equal(t, "120", obs.ValueQuantity.Value.String())
	assert.Equal(t, "mmHg", *obs.ValueQuantity.Unit)
}

func TestPatient_XML_Roundtrip(t *testing.T) {
	gender := AdministrativeGenderFemale
	original := &Patient{
		Id:        ptr("roundtrip-test"),
		Active:    ptr(true),
		Gender:    &gender,
		BirthDate: ptr("1990-01-01"),
		Name: []HumanName{
			{
				Family: ptr("Test"),
				Given:  []string{"Jane", "Marie"},
			},
		},
	}

	// Marshal to XML
	data, err := MarshalResourceXML(original)
	require.NoError(t, err)

	// Unmarshal back
	resource, err := UnmarshalResourceXML(data)
	require.NoError(t, err)

	patient := resource.(*Patient)
	assert.Equal(t, *original.Id, *patient.Id)
	assert.Equal(t, *original.Active, *patient.Active)
	assert.Equal(t, *original.Gender, *patient.Gender)
	assert.Equal(t, *original.BirthDate, *patient.BirthDate)
	require.Len(t, patient.Name, 1)
	assert.Equal(t, *original.Name[0].Family, *patient.Name[0].Family)
	assert.Equal(t, original.Name[0].Given, patient.Name[0].Given)
}

func TestBundle_XML_Roundtrip(t *testing.T) {
	bundleType := BundleTypeSearchset
	original := &Bundle{
		Id:   ptr("bundle-roundtrip"),
		Type: &bundleType,
		Entry: []BundleEntry{
			{
				Resource: &Patient{Id: ptr("p1")},
			},
			{
				Resource: &Observation{
					Id: ptr("o1"),
					Code: CodeableConcept{
						Coding: []Coding{
							{System: ptr("http://loinc.org"), Code: ptr("1234")},
						},
					},
				},
			},
		},
	}

	data, err := MarshalResourceXML(original)
	require.NoError(t, err)

	resource, err := UnmarshalResourceXML(data)
	require.NoError(t, err)

	bundle := resource.(*Bundle)
	assert.Equal(t, *original.Id, *bundle.Id)
	assert.Equal(t, *original.Type, *bundle.Type)
	require.Len(t, bundle.Entry, 2)

	p, ok := bundle.Entry[0].Resource.(*Patient)
	require.True(t, ok)
	assert.Equal(t, "p1", *p.Id)

	o, ok := bundle.Entry[1].Resource.(*Observation)
	require.True(t, ok)
	assert.Equal(t, "o1", *o.Id)
	require.Len(t, o.Code.Coding, 1)
	assert.Equal(t, "1234", *o.Code.Coding[0].Code)
}

func TestUnmarshalResourceXML_UnknownType(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?><UnknownResource xmlns="http://hl7.org/fhir"><id value="test"/></UnknownResource>`)

	_, err := UnmarshalResourceXML(xmlData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "UnknownResource")
}

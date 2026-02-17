package r4_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofhir/models/r4"
)

func TestPatientBuilder(t *testing.T) {
	t.Run("build patient with fluent API", func(t *testing.T) {
		family := "Garcia"
		use := r4.NameUseOfficial

		patient := r4.NewPatientBuilder().
			SetId("patient-789").
			SetActive(true).
			SetGender(r4.AdministrativeGenderFemale).
			SetBirthDate("1985-06-20").
			AddName(r4.HumanName{
				Use:    &use,
				Family: &family,
				Given:  []string{"Maria"},
			}).
			Build()

		require.NotNil(t, patient)
		assert.Equal(t, "patient-789", *patient.Id)
		assert.True(t, *patient.Active)
		assert.Equal(t, r4.AdministrativeGenderFemale, *patient.Gender)
		assert.Equal(t, "1985-06-20", *patient.BirthDate)
		require.Len(t, patient.Name, 1)
		assert.Equal(t, "Garcia", *patient.Name[0].Family)
	})

	t.Run("add multiple elements", func(t *testing.T) {
		system := "http://hospital.example.org/mrn"
		value1 := "MRN-001"
		value2 := "MRN-002"

		patient := r4.NewPatientBuilder().
			SetId("patient-multi").
			AddIdentifier(r4.Identifier{System: &system, Value: &value1}).
			AddIdentifier(r4.Identifier{System: &system, Value: &value2}).
			Build()

		require.NotNil(t, patient)
		require.Len(t, patient.Identifier, 2)
		assert.Equal(t, "MRN-001", *patient.Identifier[0].Value)
		assert.Equal(t, "MRN-002", *patient.Identifier[1].Value)
	})

	t.Run("JSON round trip", func(t *testing.T) {
		family := "Johnson"
		city := "Boston"
		use := r4.AddressUseHome

		original := r4.NewPatientBuilder().
			SetId("pt-json").
			SetActive(true).
			SetGender(r4.AdministrativeGenderMale).
			AddName(r4.HumanName{Family: &family, Given: []string{"Robert"}}).
			AddAddress(r4.Address{Use: &use, City: &city}).
			Build()

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded r4.Patient
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *original.Id, *decoded.Id)
		assert.Equal(t, *original.Active, *decoded.Active)
		assert.Equal(t, *original.Gender, *decoded.Gender)
		require.Len(t, decoded.Name, 1)
		assert.Equal(t, *original.Name[0].Family, *decoded.Name[0].Family)
	})

	t.Run("empty builder", func(t *testing.T) {
		patient := r4.NewPatientBuilder().Build()

		require.NotNil(t, patient)
		assert.Nil(t, patient.Id)
		assert.Nil(t, patient.Active)
		assert.Empty(t, patient.Name)
	})
}

func TestObservationBuilder(t *testing.T) {
	t.Run("build observation with fluent API", func(t *testing.T) {
		codeSystem := "http://loinc.org"
		codeCode := "8480-6"
		codeDisplay := "Systolic blood pressure"
		value := r4.NewDecimalFromFloat64(120.0)
		unit := "mmHg"
		unitSystem := "http://unitsofmeasure.org"
		unitCode := "mm[Hg]"

		obs := r4.NewObservationBuilder().
			SetId("obs-bp").
			SetStatus(r4.ObservationStatusFinal).
			SetCode(r4.CodeableConcept{
				Coding: []r4.Coding{
					{System: &codeSystem, Code: &codeCode, Display: &codeDisplay},
				},
			}).
			SetValueQuantity(r4.Quantity{
				Value:  value,
				Unit:   &unit,
				System: &unitSystem,
				Code:   &unitCode,
			}).
			SetEffectiveDateTime("2024-06-15T14:30:00Z").
			Build()

		require.NotNil(t, obs)
		assert.Equal(t, "obs-bp", *obs.Id)
		assert.Equal(t, r4.ObservationStatusFinal, *obs.Status)
		require.NotNil(t, obs.ValueQuantity)
		assert.Equal(t, "120", obs.ValueQuantity.Value.String())
		assert.Equal(t, "mmHg", *obs.ValueQuantity.Unit)
		assert.Equal(t, "2024-06-15T14:30:00Z", *obs.EffectiveDateTime)
	})

	t.Run("add categories and performers", func(t *testing.T) {
		catSystem := "http://terminology.hl7.org/CodeSystem/observation-category"
		catCode := "vital-signs"
		ref := "Practitioner/123"

		obs := r4.NewObservationBuilder().
			SetId("obs-cat").
			SetStatus(r4.ObservationStatusFinal).
			SetCode(r4.CodeableConcept{}).
			AddCategory(r4.CodeableConcept{
				Coding: []r4.Coding{{System: &catSystem, Code: &catCode}},
			}).
			AddPerformer(r4.Reference{Reference: &ref}).
			Build()

		require.NotNil(t, obs)
		require.Len(t, obs.Category, 1)
		require.Len(t, obs.Performer, 1)
		assert.Equal(t, "vital-signs", *obs.Category[0].Coding[0].Code)
		assert.Equal(t, "Practitioner/123", *obs.Performer[0].Reference)
	})
}

func TestPractitionerBuilder(t *testing.T) {
	t.Run("build practitioner", func(t *testing.T) {
		family := "Wilson"
		use := r4.NameUseOfficial

		practitioner := r4.NewPractitionerBuilder().
			SetId("prac-001").
			SetActive(true).
			SetGender(r4.AdministrativeGenderFemale).
			AddName(r4.HumanName{
				Use:    &use,
				Family: &family,
				Given:  []string{"Sarah"},
				Prefix: []string{"Dr."},
			}).
			Build()

		require.NotNil(t, practitioner)
		assert.Equal(t, "prac-001", *practitioner.Id)
		assert.True(t, *practitioner.Active)
		assert.Equal(t, r4.AdministrativeGenderFemale, *practitioner.Gender)
		require.Len(t, practitioner.Name, 1)
		assert.Equal(t, "Wilson", *practitioner.Name[0].Family)
	})
}

func TestOrganizationBuilder(t *testing.T) {
	t.Run("build organization", func(t *testing.T) {
		org := r4.NewOrganizationBuilder().
			SetId("org-001").
			SetActive(true).
			SetName("General Hospital").
			Build()

		require.NotNil(t, org)
		assert.Equal(t, "org-001", *org.Id)
		assert.True(t, *org.Active)
		assert.Equal(t, "General Hospital", *org.Name)
	})
}

func TestConditionBuilder(t *testing.T) {
	t.Run("build condition", func(t *testing.T) {
		clinicalSystem := "http://terminology.hl7.org/CodeSystem/condition-clinical"
		clinicalCode := "active"
		ref := "Patient/123"

		condition := r4.NewConditionBuilder().
			SetId("cond-001").
			SetClinicalStatus(r4.CodeableConcept{
				Coding: []r4.Coding{{System: &clinicalSystem, Code: &clinicalCode}},
			}).
			SetSubject(r4.Reference{Reference: &ref}).
			SetOnsetDateTime("2024-01-15").
			Build()

		require.NotNil(t, condition)
		assert.Equal(t, "cond-001", *condition.Id)
		require.NotNil(t, condition.ClinicalStatus)
		assert.Equal(t, "active", *condition.ClinicalStatus.Coding[0].Code)
		assert.Equal(t, "Patient/123", *condition.Subject.Reference)
		assert.Equal(t, "2024-01-15", *condition.OnsetDateTime)
	})
}

func TestBundleBuilder(t *testing.T) {
	t.Run("build transaction bundle", func(t *testing.T) {
		bundle := r4.NewBundleBuilder().
			SetId("bundle-001").
			SetType(r4.BundleTypeTransaction).
			Build()

		require.NotNil(t, bundle)
		assert.Equal(t, "bundle-001", *bundle.Id)
		assert.Equal(t, r4.BundleTypeTransaction, *bundle.Type)
	})
}

func TestMixedBuilderPatterns(t *testing.T) {
	t.Run("functional options and builder produce same result", func(t *testing.T) {
		family := "Test"

		// Using functional options
		patient1 := r4.NewPatient(
			r4.WithPatientId("test-001"),
			r4.WithPatientActive(true),
			r4.WithPatientGender(r4.AdministrativeGenderMale),
			r4.WithPatientName(r4.HumanName{Family: &family}),
		)

		// Using fluent builder
		patient2 := r4.NewPatientBuilder().
			SetId("test-001").
			SetActive(true).
			SetGender(r4.AdministrativeGenderMale).
			AddName(r4.HumanName{Family: &family}).
			Build()

		// Both should produce equivalent results
		assert.Equal(t, *patient1.Id, *patient2.Id)
		assert.Equal(t, *patient1.Active, *patient2.Active)
		assert.Equal(t, *patient1.Gender, *patient2.Gender)
		require.Len(t, patient1.Name, 1)
		require.Len(t, patient2.Name, 1)
		assert.Equal(t, *patient1.Name[0].Family, *patient2.Name[0].Family)
	})
}

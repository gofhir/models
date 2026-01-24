package r5_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/robertoaraneda/gofhir/r5"
)

func TestPatientFunctionalOptions(t *testing.T) {
	t.Run("create patient with options", func(t *testing.T) {
		patient := r5.NewPatient(
			r5.WithPatientId("patient-123"),
			r5.WithPatientActive(true),
			r5.WithPatientGender(r5.AdministrativeGenderMale),
			r5.WithPatientBirthDate("1990-01-15"),
		)

		require.NotNil(t, patient)
		assert.Equal(t, "patient-123", *patient.Id)
		assert.True(t, *patient.Active)
		assert.Equal(t, r5.AdministrativeGenderMale, *patient.Gender)
		assert.Equal(t, "1990-01-15", *patient.BirthDate)
	})

	t.Run("add multiple names", func(t *testing.T) {
		family := "Smith"
		use := r5.NameUseOfficial

		patient := r5.NewPatient(
			r5.WithPatientId("patient-456"),
			r5.WithPatientName(r5.HumanName{
				Use:    &use,
				Family: &family,
				Given:  []string{"John"},
			}),
			r5.WithPatientName(r5.HumanName{
				Family: &family,
				Given:  []string{"Johnny"},
			}),
		)

		require.NotNil(t, patient)
		require.Len(t, patient.Name, 2)
		assert.Equal(t, "Smith", *patient.Name[0].Family)
		assert.Equal(t, r5.NameUseOfficial, *patient.Name[0].Use)
	})

	t.Run("add identifiers", func(t *testing.T) {
		system := "http://hospital.example.org/mrn"
		value := "12345"

		patient := r5.NewPatient(
			r5.WithPatientIdentifier(r5.Identifier{
				System: &system,
				Value:  &value,
			}),
		)

		require.NotNil(t, patient)
		require.Len(t, patient.Identifier, 1)
		assert.Equal(t, "http://hospital.example.org/mrn", *patient.Identifier[0].System)
		assert.Equal(t, "12345", *patient.Identifier[0].Value)
	})

	t.Run("empty patient", func(t *testing.T) {
		patient := r5.NewPatient()

		require.NotNil(t, patient)
		assert.Nil(t, patient.Id)
		assert.Nil(t, patient.Active)
		assert.Empty(t, patient.Name)
	})
}

func TestObservationFunctionalOptions(t *testing.T) {
	t.Run("create observation with options", func(t *testing.T) {
		codeSystem := "http://loinc.org"
		codeCode := "8867-4"
		codeDisplay := "Heart rate"

		obs := r5.NewObservation(
			r5.WithObservationId("obs-123"),
			r5.WithObservationStatus(r5.ObservationStatusFinal),
			r5.WithObservationCode(r5.CodeableConcept{
				Coding: []r5.Coding{
					{System: &codeSystem, Code: &codeCode, Display: &codeDisplay},
				},
			}),
			r5.WithObservationEffectiveDateTime("2024-01-15T10:30:00Z"),
		)

		require.NotNil(t, obs)
		assert.Equal(t, "obs-123", *obs.Id)
		assert.Equal(t, r5.ObservationStatusFinal, *obs.Status)
		assert.Equal(t, "2024-01-15T10:30:00Z", *obs.EffectiveDateTime)
		require.Len(t, obs.Code.Coding, 1)
		assert.Equal(t, "http://loinc.org", *obs.Code.Coding[0].System)
	})

	t.Run("observation with value quantity", func(t *testing.T) {
		value := 72.0
		unit := "bpm"
		system := "http://unitsofmeasure.org"
		code := "/min"

		obs := r5.NewObservation(
			r5.WithObservationId("obs-456"),
			r5.WithObservationStatus(r5.ObservationStatusFinal),
			r5.WithObservationValueQuantity(r5.Quantity{
				Value:  &value,
				Unit:   &unit,
				System: &system,
				Code:   &code,
			}),
		)

		require.NotNil(t, obs)
		require.NotNil(t, obs.ValueQuantity)
		assert.Equal(t, 72.0, *obs.ValueQuantity.Value)
		assert.Equal(t, "bpm", *obs.ValueQuantity.Unit)
	})
}

func TestPractitionerFunctionalOptions(t *testing.T) {
	t.Run("create practitioner with options", func(t *testing.T) {
		family := "Wilson"
		use := r5.NameUseOfficial

		practitioner := r5.NewPractitioner(
			r5.WithPractitionerId("prac-001"),
			r5.WithPractitionerActive(true),
			r5.WithPractitionerGender(r5.AdministrativeGenderFemale),
			r5.WithPractitionerName(r5.HumanName{
				Use:    &use,
				Family: &family,
				Given:  []string{"Sarah"},
				Prefix: []string{"Dr."},
			}),
		)

		require.NotNil(t, practitioner)
		assert.Equal(t, "prac-001", *practitioner.Id)
		assert.True(t, *practitioner.Active)
		assert.Equal(t, r5.AdministrativeGenderFemale, *practitioner.Gender)
		require.Len(t, practitioner.Name, 1)
		assert.Equal(t, "Wilson", *practitioner.Name[0].Family)
	})
}

func TestOrganizationFunctionalOptions(t *testing.T) {
	t.Run("create organization with options", func(t *testing.T) {
		org := r5.NewOrganization(
			r5.WithOrganizationId("org-001"),
			r5.WithOrganizationActive(true),
			r5.WithOrganizationName("General Hospital"),
		)

		require.NotNil(t, org)
		assert.Equal(t, "org-001", *org.Id)
		assert.True(t, *org.Active)
		assert.Equal(t, "General Hospital", *org.Name)
	})
}

func TestConditionFunctionalOptions(t *testing.T) {
	t.Run("create condition with options", func(t *testing.T) {
		clinicalSystem := "http://terminology.hl7.org/CodeSystem/condition-clinical"
		clinicalCode := "active"
		ref := "Patient/123"

		condition := r5.NewCondition(
			r5.WithConditionId("cond-001"),
			r5.WithConditionClinicalStatus(r5.CodeableConcept{
				Coding: []r5.Coding{{System: &clinicalSystem, Code: &clinicalCode}},
			}),
			r5.WithConditionSubject(r5.Reference{Reference: &ref}),
			r5.WithConditionOnsetDateTime("2024-01-15"),
		)

		require.NotNil(t, condition)
		assert.Equal(t, "cond-001", *condition.Id)
		require.NotNil(t, condition.ClinicalStatus)
		assert.Equal(t, "active", *condition.ClinicalStatus.Coding[0].Code)
		assert.Equal(t, "Patient/123", *condition.Subject.Reference)
		assert.Equal(t, "2024-01-15", *condition.OnsetDateTime)
	})
}

func TestBundleFunctionalOptions(t *testing.T) {
	t.Run("create transaction bundle with options", func(t *testing.T) {
		bundle := r5.NewBundle(
			r5.WithBundleId("bundle-001"),
			r5.WithBundleType(r5.BundleTypeTransaction),
		)

		require.NotNil(t, bundle)
		assert.Equal(t, "bundle-001", *bundle.Id)
		assert.Equal(t, r5.BundleTypeTransaction, *bundle.Type)
	})
}

package r4_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofhir/models/r4"
)

func TestPatientFunctionalOptions(t *testing.T) {
	t.Run("create patient with options", func(t *testing.T) {
		patient := r4.NewPatient(
			r4.WithPatientId("patient-123"),
			r4.WithPatientActive(true),
			r4.WithPatientGender(r4.AdministrativeGenderMale),
			r4.WithPatientBirthDate("1990-01-15"),
		)

		require.NotNil(t, patient)
		assert.Equal(t, "patient-123", *patient.Id)
		assert.True(t, *patient.Active)
		assert.Equal(t, r4.AdministrativeGenderMale, *patient.Gender)
		assert.Equal(t, "1990-01-15", *patient.BirthDate)
	})

	t.Run("add multiple names", func(t *testing.T) {
		family := "Smith"
		use := r4.NameUseOfficial

		patient := r4.NewPatient(
			r4.WithPatientId("patient-456"),
			r4.WithPatientName(r4.HumanName{
				Use:    &use,
				Family: &family,
				Given:  []string{"John"},
			}),
			r4.WithPatientName(r4.HumanName{
				Family: &family,
				Given:  []string{"Johnny"},
			}),
		)

		require.NotNil(t, patient)
		require.Len(t, patient.Name, 2)
		assert.Equal(t, "Smith", *patient.Name[0].Family)
		assert.Equal(t, r4.NameUseOfficial, *patient.Name[0].Use)
	})

	t.Run("add identifiers", func(t *testing.T) {
		system := "http://hospital.example.org/mrn"
		value := "12345"

		patient := r4.NewPatient(
			r4.WithPatientIdentifier(r4.Identifier{
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
		patient := r4.NewPatient()

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

		obs := r4.NewObservation(
			r4.WithObservationId("obs-123"),
			r4.WithObservationStatus(r4.ObservationStatusFinal),
			r4.WithObservationCode(r4.CodeableConcept{
				Coding: []r4.Coding{
					{System: &codeSystem, Code: &codeCode, Display: &codeDisplay},
				},
			}),
			r4.WithObservationEffectiveDateTime("2024-01-15T10:30:00Z"),
		)

		require.NotNil(t, obs)
		assert.Equal(t, "obs-123", *obs.Id)
		assert.Equal(t, r4.ObservationStatusFinal, *obs.Status)
		assert.Equal(t, "2024-01-15T10:30:00Z", *obs.EffectiveDateTime)
		require.Len(t, obs.Code.Coding, 1)
		assert.Equal(t, "http://loinc.org", *obs.Code.Coding[0].System)
	})

	t.Run("observation with value quantity", func(t *testing.T) {
		value := r4.NewDecimalFromFloat64(72.0)
		unit := "bpm"
		system := "http://unitsofmeasure.org"
		code := "/min"

		obs := r4.NewObservation(
			r4.WithObservationId("obs-456"),
			r4.WithObservationStatus(r4.ObservationStatusFinal),
			r4.WithObservationValueQuantity(r4.Quantity{
				Value:  value,
				Unit:   &unit,
				System: &system,
				Code:   &code,
			}),
		)

		require.NotNil(t, obs)
		require.NotNil(t, obs.ValueQuantity)
		assert.Equal(t, "72", obs.ValueQuantity.Value.String())
		assert.Equal(t, "bpm", *obs.ValueQuantity.Unit)
	})
}

func TestPractitionerFunctionalOptions(t *testing.T) {
	t.Run("create practitioner with options", func(t *testing.T) {
		family := "Wilson"
		use := r4.NameUseOfficial

		practitioner := r4.NewPractitioner(
			r4.WithPractitionerId("prac-001"),
			r4.WithPractitionerActive(true),
			r4.WithPractitionerGender(r4.AdministrativeGenderFemale),
			r4.WithPractitionerName(r4.HumanName{
				Use:    &use,
				Family: &family,
				Given:  []string{"Sarah"},
				Prefix: []string{"Dr."},
			}),
		)

		require.NotNil(t, practitioner)
		assert.Equal(t, "prac-001", *practitioner.Id)
		assert.True(t, *practitioner.Active)
		assert.Equal(t, r4.AdministrativeGenderFemale, *practitioner.Gender)
		require.Len(t, practitioner.Name, 1)
		assert.Equal(t, "Wilson", *practitioner.Name[0].Family)
	})
}

func TestOrganizationFunctionalOptions(t *testing.T) {
	t.Run("create organization with options", func(t *testing.T) {
		org := r4.NewOrganization(
			r4.WithOrganizationId("org-001"),
			r4.WithOrganizationActive(true),
			r4.WithOrganizationName("General Hospital"),
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

		condition := r4.NewCondition(
			r4.WithConditionId("cond-001"),
			r4.WithConditionClinicalStatus(r4.CodeableConcept{
				Coding: []r4.Coding{{System: &clinicalSystem, Code: &clinicalCode}},
			}),
			r4.WithConditionSubject(r4.Reference{Reference: &ref}),
			r4.WithConditionOnsetDateTime("2024-01-15"),
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
		bundle := r4.NewBundle(
			r4.WithBundleId("bundle-001"),
			r4.WithBundleType(r4.BundleTypeTransaction),
		)

		require.NotNil(t, bundle)
		assert.Equal(t, "bundle-001", *bundle.Id)
		assert.Equal(t, r4.BundleTypeTransaction, *bundle.Type)
	})
}

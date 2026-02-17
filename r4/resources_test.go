package r4

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatient(t *testing.T) {
	t.Run("create patient", func(t *testing.T) {
		id := "patient-123"
		gender := AdministrativeGenderMale
		birthDate := "1990-01-15"
		active := true
		family := "Smith"
		use := NameUseOfficial

		patient := Patient{
			Id:        &id,
			Gender:    &gender,
			BirthDate: &birthDate,
			Active:    &active,
			Name: []HumanName{
				{
					Use:    &use,
					Family: &family,
					Given:  []string{"John", "Robert"},
				},
			},
		}

		assert.Equal(t, "patient-123", *patient.Id)
		assert.Equal(t, AdministrativeGenderMale, *patient.Gender)
		assert.Equal(t, "1990-01-15", *patient.BirthDate)
		assert.True(t, *patient.Active)
		require.Len(t, patient.Name, 1)
		assert.Equal(t, "Smith", *patient.Name[0].Family)
	})

	t.Run("GetResourceType", func(t *testing.T) {
		patient := &Patient{}
		assert.Equal(t, "Patient", patient.GetResourceType())
	})

	t.Run("JSON round trip", func(t *testing.T) {
		id := "pt-456"
		gender := AdministrativeGenderFemale
		birthDate := "1985-06-20"
		family := "Johnson"
		city := "Boston"
		use := AddressUseHome

		original := Patient{
			Id:        &id,
			Gender:    &gender,
			BirthDate: &birthDate,
			Name: []HumanName{
				{Family: &family, Given: []string{"Jane"}},
			},
			Address: []Address{
				{Use: &use, City: &city},
			},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded Patient
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *original.Id, *decoded.Id)
		assert.Equal(t, *original.Gender, *decoded.Gender)
		assert.Equal(t, *original.BirthDate, *decoded.BirthDate)
		require.Len(t, decoded.Name, 1)
		assert.Equal(t, *original.Name[0].Family, *decoded.Name[0].Family)
		require.Len(t, decoded.Address, 1)
		assert.Equal(t, *original.Address[0].City, *decoded.Address[0].City)
	})

	t.Run("patient with deceased choice type", func(t *testing.T) {
		id := "pt-deceased"
		deceasedBool := true

		patient := Patient{
			Id:              &id,
			DeceasedBoolean: &deceasedBool,
		}

		assert.Equal(t, id, *patient.Id)
		assert.True(t, *patient.DeceasedBoolean)
		assert.Nil(t, patient.DeceasedDateTime)

		// Alternative: using datetime
		deceasedDT := "2024-01-15T10:30:00Z"
		patient2 := Patient{
			Id:               &id,
			DeceasedDateTime: &deceasedDT,
		}

		assert.Equal(t, id, *patient2.Id)
		assert.Nil(t, patient2.DeceasedBoolean)
		assert.Equal(t, "2024-01-15T10:30:00Z", *patient2.DeceasedDateTime)
	})

	t.Run("patient with multiple birth choice type", func(t *testing.T) {
		id := "pt-multiple"
		multipleBirthInt := 2 // Second of twins

		patient := Patient{
			Id:                   &id,
			MultipleBirthInteger: &multipleBirthInt,
		}

		assert.Equal(t, id, *patient.Id)
		assert.Equal(t, 2, *patient.MultipleBirthInteger)
		assert.Nil(t, patient.MultipleBirthBoolean)
	})
}

func TestObservation(t *testing.T) {
	t.Run("create observation with value quantity", func(t *testing.T) {
		id := "obs-123"
		status := ObservationStatusFinal
		value := NewDecimalFromFloat64(120.0)
		unit := "mmHg"
		system := "http://unitsofmeasure.org"
		code := "mm[Hg]"

		codeSystem := "http://loinc.org"
		codeCode := "8480-6"
		codeDisplay := "Systolic blood pressure"

		obs := Observation{
			Id:     &id,
			Status: &status,
			Code: CodeableConcept{
				Coding: []Coding{
					{System: &codeSystem, Code: &codeCode, Display: &codeDisplay},
				},
			},
			ValueQuantity: &Quantity{
				Value:  value,
				Unit:   &unit,
				System: &system,
				Code:   &code,
			},
		}

		assert.Equal(t, "obs-123", *obs.Id)
		assert.Equal(t, ObservationStatusFinal, *obs.Status)
		require.NotNil(t, obs.ValueQuantity)
		assert.Equal(t, 120.0, obs.ValueQuantity.Value.Float64())
		assert.Equal(t, "mmHg", *obs.ValueQuantity.Unit)
	})

	t.Run("GetResourceType", func(t *testing.T) {
		obs := &Observation{}
		assert.Equal(t, "Observation", obs.GetResourceType())
	})

	t.Run("observation with value codeable concept", func(t *testing.T) {
		id := "obs-456"
		status := ObservationStatusFinal
		system := "http://snomed.info/sct"
		code := "260385009"
		display := "Negative"

		obs := Observation{
			Id:     &id,
			Status: &status,
			Code:   CodeableConcept{},
			ValueCodeableConcept: &CodeableConcept{
				Coding: []Coding{
					{System: &system, Code: &code, Display: &display},
				},
			},
		}

		assert.Equal(t, id, *obs.Id)
		assert.Equal(t, status, *obs.Status)
		require.NotNil(t, obs.ValueCodeableConcept)
		require.Len(t, obs.ValueCodeableConcept.Coding, 1)
		assert.Equal(t, "Negative", *obs.ValueCodeableConcept.Coding[0].Display)
	})

	t.Run("JSON round trip with effective choice type", func(t *testing.T) {
		id := "obs-789"
		status := ObservationStatusPreliminary
		effectiveDT := "2024-06-15T14:30:00Z"

		original := Observation{
			Id:                &id,
			Status:            &status,
			Code:              CodeableConcept{},
			EffectiveDateTime: &effectiveDT,
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded Observation
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *original.Id, *decoded.Id)
		assert.Equal(t, *original.Status, *decoded.Status)
		assert.Equal(t, *original.EffectiveDateTime, *decoded.EffectiveDateTime)
	})
}

func TestAccount(t *testing.T) {
	t.Run("create account", func(t *testing.T) {
		id := "account-123"
		status := AccountStatusActive
		name := "John Smith's Account"

		account := Account{
			Id:     &id,
			Status: &status,
			Name:   &name,
		}

		assert.Equal(t, "account-123", *account.Id)
		assert.Equal(t, AccountStatusActive, *account.Status)
		assert.Equal(t, "John Smith's Account", *account.Name)
	})

	t.Run("GetResourceType", func(t *testing.T) {
		account := &Account{}
		assert.Equal(t, "Account", account.GetResourceType())
	})
}

func TestBundle(t *testing.T) {
	t.Run("create transaction bundle", func(t *testing.T) {
		id := "bundle-123"
		bundleType := BundleTypeTransaction

		bundle := Bundle{
			Id:   &id,
			Type: &bundleType,
		}

		assert.Equal(t, "bundle-123", *bundle.Id)
		assert.Equal(t, BundleTypeTransaction, *bundle.Type)
	})

	t.Run("GetResourceType", func(t *testing.T) {
		bundle := &Bundle{}
		assert.Equal(t, "Bundle", bundle.GetResourceType())
	})

	t.Run("bundle types", func(t *testing.T) {
		assert.Equal(t, BundleType("document"), BundleTypeDocument)
		assert.Equal(t, BundleType("message"), BundleTypeMessage)
		assert.Equal(t, BundleType("transaction"), BundleTypeTransaction)
		assert.Equal(t, BundleType("batch"), BundleTypeBatch)
		assert.Equal(t, BundleType("searchset"), BundleTypeSearchset)
		assert.Equal(t, BundleType("collection"), BundleTypeCollection)
	})
}

func TestCondition(t *testing.T) {
	t.Run("create condition", func(t *testing.T) {
		id := "condition-123"
		clinicalStatus := "active"
		system := "http://terminology.hl7.org/CodeSystem/condition-clinical"
		ref := "Patient/123"

		condition := Condition{
			Id: &id,
			ClinicalStatus: &CodeableConcept{
				Coding: []Coding{
					{System: &system, Code: &clinicalStatus},
				},
			},
			Subject: Reference{
				Reference: &ref,
			},
		}

		assert.Equal(t, "condition-123", *condition.Id)
		require.NotNil(t, condition.ClinicalStatus)
		require.Len(t, condition.ClinicalStatus.Coding, 1)
		assert.Equal(t, "active", *condition.ClinicalStatus.Coding[0].Code)
	})

	t.Run("GetResourceType", func(t *testing.T) {
		condition := &Condition{}
		assert.Equal(t, "Condition", condition.GetResourceType())
	})
}

func TestPractitioner(t *testing.T) {
	t.Run("create practitioner", func(t *testing.T) {
		id := "practitioner-123"
		active := true
		family := "Wilson"
		use := NameUseOfficial
		gender := AdministrativeGenderFemale

		practitioner := Practitioner{
			Id:     &id,
			Active: &active,
			Gender: &gender,
			Name: []HumanName{
				{
					Use:    &use,
					Family: &family,
					Given:  []string{"Sarah"},
					Prefix: []string{"Dr."},
				},
			},
		}

		assert.Equal(t, "practitioner-123", *practitioner.Id)
		assert.True(t, *practitioner.Active)
		assert.Equal(t, AdministrativeGenderFemale, *practitioner.Gender)
		require.Len(t, practitioner.Name, 1)
		assert.Equal(t, "Wilson", *practitioner.Name[0].Family)
	})

	t.Run("GetResourceType", func(t *testing.T) {
		practitioner := &Practitioner{}
		assert.Equal(t, "Practitioner", practitioner.GetResourceType())
	})
}

func TestOrganization(t *testing.T) {
	t.Run("create organization", func(t *testing.T) {
		id := "org-123"
		active := true
		name := "General Hospital"

		org := Organization{
			Id:     &id,
			Active: &active,
			Name:   &name,
		}

		assert.Equal(t, "org-123", *org.Id)
		assert.True(t, *org.Active)
		assert.Equal(t, "General Hospital", *org.Name)
	})

	t.Run("GetResourceType", func(t *testing.T) {
		org := &Organization{}
		assert.Equal(t, "Organization", org.GetResourceType())
	})
}

func TestMedicationRequest(t *testing.T) {
	t.Run("create medication request", func(t *testing.T) {
		id := "medrx-123"
		status := MedicationrequestStatus("active")
		intent := MedicationRequestIntent("order")
		subjectRef := "Patient/123"
		medRef := "Medication/456"

		medRequest := MedicationRequest{
			Id:     &id,
			Status: &status,
			Intent: &intent,
			Subject: Reference{
				Reference: &subjectRef,
			},
			MedicationReference: &Reference{
				Reference: &medRef,
			},
		}

		assert.Equal(t, "medrx-123", *medRequest.Id)
		assert.Equal(t, MedicationrequestStatus("active"), *medRequest.Status)
		assert.Equal(t, MedicationRequestIntent("order"), *medRequest.Intent)
		require.NotNil(t, medRequest.MedicationReference)
		assert.Equal(t, medRef, *medRequest.MedicationReference.Reference)
	})

	t.Run("GetResourceType", func(t *testing.T) {
		medRequest := &MedicationRequest{}
		assert.Equal(t, "MedicationRequest", medRequest.GetResourceType())
	})
}

func TestResourceInterface(t *testing.T) {
	t.Run("resources implement Resource interface", func(t *testing.T) {
		resources := []Resource{
			&Patient{},
			&Observation{},
			&Account{},
			&Bundle{},
			&Condition{},
			&Practitioner{},
			&Organization{},
		}

		expectedTypes := []string{
			"Patient",
			"Observation",
			"Account",
			"Bundle",
			"Condition",
			"Practitioner",
			"Organization",
		}

		for i, r := range resources {
			assert.Equal(t, expectedTypes[i], r.GetResourceType())
		}
	})
}

func TestResourceWithMeta(t *testing.T) {
	t.Run("patient with meta", func(t *testing.T) {
		id := "pt-meta"
		versionID := "1"
		lastUpdated := "2024-06-15T10:30:00Z"
		profile := "http://hl7.org/fhir/us/core/StructureDefinition/us-core-patient"

		patient := Patient{
			Id: &id,
			Meta: &Meta{
				VersionId:   &versionID,
				LastUpdated: &lastUpdated,
				Profile:     []string{profile},
			},
		}

		assert.Equal(t, id, *patient.Id)
		require.NotNil(t, patient.Meta)
		assert.Equal(t, "1", *patient.Meta.VersionId)
		assert.Equal(t, "2024-06-15T10:30:00Z", *patient.Meta.LastUpdated)
		require.Len(t, patient.Meta.Profile, 1)
		assert.Equal(t, profile, patient.Meta.Profile[0])
	})
}

func TestResourceWithPrimitiveExtensions(t *testing.T) {
	t.Run("patient birthDate with extension", func(t *testing.T) {
		id := "pt-ext"
		birthDate := "1990-01-15"
		extID := "birthdate-ext"
		extURL := "http://example.org/fhir/StructureDefinition/approximate-date"
		extValueBool := true

		patient := Patient{
			Id:        &id,
			BirthDate: &birthDate,
			BirthDateExt: &Element{
				Id: &extID,
				Extension: []Extension{
					{
						Url:          extURL,
						ValueBoolean: &extValueBool,
					},
				},
			},
		}

		assert.Equal(t, id, *patient.Id)
		assert.Equal(t, "1990-01-15", *patient.BirthDate)
		require.NotNil(t, patient.BirthDateExt)
		assert.Equal(t, "birthdate-ext", *patient.BirthDateExt.Id)
		require.Len(t, patient.BirthDateExt.Extension, 1)
		assert.Equal(t, extURL, patient.BirthDateExt.Extension[0].Url)
	})
}

func TestPatient_MarshalJSON_NoHTMLEscape(t *testing.T) {
	t.Run("narrative HTML is not escaped", func(t *testing.T) {
		id := "pt-narrative"
		status := NarrativeStatusGenerated
		divContent := "<div xmlns=\"http://www.w3.org/1999/xhtml\"><p>Test patient with <b>bold</b> and <i>italic</i> text</p></div>"

		patient := Patient{
			Id: &id,
			Text: &Narrative{
				Status: &status,
				Div:    &divContent,
			},
		}

		data, err := Marshal(patient)
		require.NoError(t, err)

		jsonStr := string(data)

		// Verify HTML tags are preserved (not escaped)
		assert.Contains(t, jsonStr, "<div")
		assert.Contains(t, jsonStr, "<b>bold</b>")
		assert.Contains(t, jsonStr, "<i>italic</i>")
		assert.Contains(t, jsonStr, "</div>")

		// Verify unicode escape sequences are NOT present
		assert.NotContains(t, jsonStr, `\u003c`)
		assert.NotContains(t, jsonStr, `\u003e`)
	})

	t.Run("narrative round trip preserves HTML", func(t *testing.T) {
		id := "pt-roundtrip"
		status := NarrativeStatusGenerated
		originalDiv := "<div xmlns=\"http://www.w3.org/1999/xhtml\"><h1>Patient Summary</h1><p>Status: <span style=\"color: red;\">Critical</span></p></div>"

		original := Patient{
			Id: &id,
			Text: &Narrative{
				Status: &status,
				Div:    &originalDiv,
			},
		}

		// Marshal using package helper
		data, err := Marshal(original)
		require.NoError(t, err)

		// Unmarshal
		var decoded Patient
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		// Verify content preserved exactly
		require.NotNil(t, decoded.Text)
		require.NotNil(t, decoded.Text.Div)
		assert.Equal(t, originalDiv, *decoded.Text.Div)
	})
}

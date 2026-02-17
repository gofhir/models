package r5_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofhir/models/r5"
)

// TestResourceBackboneElements tests backbone elements that are part of resources
func TestResourceBackboneElements(t *testing.T) {
	t.Run("PatientContact", func(t *testing.T) {
		contact := r5.PatientContact{
			Id: ptrStringB5("contact-1"),
			Relationship: []r5.CodeableConcept{
				{
					Text: ptrStringB5("Emergency Contact"),
				},
			},
			Name: &r5.HumanName{
				Family: ptrStringB5("Smith"),
				Given:  []string{"Jane"},
			},
		}

		data, err := json.Marshal(contact)
		require.NoError(t, err)

		var decoded r5.PatientContact
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "contact-1", *decoded.Id)
		require.Len(t, decoded.Relationship, 1)
		assert.Equal(t, "Emergency Contact", *decoded.Relationship[0].Text)
		assert.Equal(t, "Smith", *decoded.Name.Family)
	})

	t.Run("BundleEntryRequest", func(t *testing.T) {
		method := r5.HTTPVerb("POST")
		request := r5.BundleEntryRequest{
			Id:     ptrStringB5("request-1"),
			Method: &method,
			Url:    ptrStringB5("Patient"),
		}

		data, err := json.Marshal(request)
		require.NoError(t, err)

		var decoded r5.BundleEntryRequest
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "request-1", *decoded.Id)
		assert.Equal(t, r5.HTTPVerb("POST"), *decoded.Method)
		assert.Equal(t, "Patient", *decoded.Url)
	})

	t.Run("BundleEntryResponse", func(t *testing.T) {
		response := r5.BundleEntryResponse{
			Id:       ptrStringB5("response-1"),
			Status:   ptrStringB5("201 Created"),
			Location: ptrStringB5("Patient/123/_history/1"),
			Etag:     ptrStringB5("W/\"1\""),
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var decoded r5.BundleEntryResponse
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "response-1", *decoded.Id)
		assert.Equal(t, "201 Created", *decoded.Status)
	})

	t.Run("BundleEntrySearch", func(t *testing.T) {
		mode := r5.SearchEntryMode("match")
		search := r5.BundleEntrySearch{
			Id:    ptrStringB5("search-1"),
			Mode:  &mode,
			Score: ptrDecimalB5(0.95),
		}

		data, err := json.Marshal(search)
		require.NoError(t, err)

		var decoded r5.BundleEntrySearch
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "search-1", *decoded.Id)
		assert.Equal(t, r5.SearchEntryMode("match"), *decoded.Mode)
		assert.Equal(t, "0.95", decoded.Score.String())
	})

	t.Run("ObservationComponent", func(t *testing.T) {
		component := r5.ObservationComponent{
			Id: ptrStringB5("comp-1"),
			Code: r5.CodeableConcept{
				Coding: []r5.Coding{
					{
						System:  ptrStringB5("http://loinc.org"),
						Code:    ptrStringB5("8480-6"),
						Display: ptrStringB5("Systolic blood pressure"),
					},
				},
			},
			ValueQuantity: &r5.Quantity{
				Value:  ptrDecimalB5(120),
				Unit:   ptrStringB5("mmHg"),
				System: ptrStringB5("http://unitsofmeasure.org"),
				Code:   ptrStringB5("mm[Hg]"),
			},
		}

		data, err := json.Marshal(component)
		require.NoError(t, err)

		var decoded r5.ObservationComponent
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "comp-1", *decoded.Id)
		assert.Equal(t, "8480-6", *decoded.Code.Coding[0].Code)
		assert.Equal(t, "120", decoded.ValueQuantity.Value.String())
	})
}

// TestDatatypeBackboneElements tests backbone elements from datatype-style types
func TestDatatypeBackboneElements(t *testing.T) {
	t.Run("DosageDoseAndRate", func(t *testing.T) {
		doseAndRate := r5.DosageDoseAndRate{
			Id: ptrStringB5("dose-1"),
			Type: &r5.CodeableConcept{
				Coding: []r5.Coding{
					{
						System:  ptrStringB5("http://terminology.hl7.org/CodeSystem/dose-rate-type"),
						Code:    ptrStringB5("ordered"),
						Display: ptrStringB5("Ordered"),
					},
				},
			},
			DoseQuantity: &r5.Quantity{
				Value:  ptrDecimalB5(500),
				Unit:   ptrStringB5("mg"),
				System: ptrStringB5("http://unitsofmeasure.org"),
				Code:   ptrStringB5("mg"),
			},
		}

		data, err := json.Marshal(doseAndRate)
		require.NoError(t, err)

		var decoded r5.DosageDoseAndRate
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "dose-1", *decoded.Id)
		assert.Equal(t, "ordered", *decoded.Type.Coding[0].Code)
		assert.Equal(t, "500", decoded.DoseQuantity.Value.String())
	})

	t.Run("TimingRepeat", func(t *testing.T) {
		periodUnit := r5.UnitsOfTime("d")
		repeat := r5.TimingRepeat{
			Id:          ptrStringB5("repeat-1"),
			Frequency:   ptrUint32B5(2),
			Period:      ptrDecimalB5(1),
			PeriodUnit:  &periodUnit,
			DayOfWeek:   []r5.DaysOfWeek{"mon", "wed", "fri"},
			TimeOfDay:   []string{"08:00:00", "18:00:00"},
			Duration:    ptrDecimalB5(30),
			DurationMax: ptrDecimalB5(60),
		}

		data, err := json.Marshal(repeat)
		require.NoError(t, err)

		var decoded r5.TimingRepeat
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "repeat-1", *decoded.Id)
		assert.Equal(t, uint32(2), *decoded.Frequency)
		assert.Equal(t, r5.UnitsOfTime("d"), *decoded.PeriodUnit)
	})

	t.Run("ElementDefinitionSlicing", func(t *testing.T) {
		rules := r5.SlicingRules("open")
		slicing := r5.ElementDefinitionSlicing{
			Id:          ptrStringB5("slicing-1"),
			Description: ptrStringB5("Slice by code"),
			Ordered:     ptrBoolB5(false),
			Rules:       &rules,
		}

		data, err := json.Marshal(slicing)
		require.NoError(t, err)

		var decoded r5.ElementDefinitionSlicing
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "slicing-1", *decoded.Id)
		assert.Equal(t, "Slice by code", *decoded.Description)
		assert.False(t, *decoded.Ordered)
		assert.Equal(t, r5.SlicingRules("open"), *decoded.Rules)
	})

	t.Run("ElementDefinitionBinding", func(t *testing.T) {
		strength := r5.BindingStrength("required")
		binding := r5.ElementDefinitionBinding{
			Id:          ptrStringB5("binding-1"),
			Strength:    &strength,
			Description: ptrStringB5("The status of the encounter"),
			ValueSet:    ptrStringB5("http://hl7.org/fhir/ValueSet/encounter-status"),
		}

		data, err := json.Marshal(binding)
		require.NoError(t, err)

		var decoded r5.ElementDefinitionBinding
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "binding-1", *decoded.Id)
		assert.Equal(t, r5.BindingStrength("required"), *decoded.Strength)
	})

	t.Run("ElementDefinitionConstraint", func(t *testing.T) {
		severity := r5.ConstraintSeverity("error")
		constraint := r5.ElementDefinitionConstraint{
			Id:         ptrStringB5("constraint-1"),
			Key:        ptrStringB5("ele-1"),
			Severity:   &severity,
			Human:      ptrStringB5("All FHIR elements must have a @value or children"),
			Expression: ptrStringB5("hasValue() or (children().count() > id.count())"),
		}

		data, err := json.Marshal(constraint)
		require.NoError(t, err)

		var decoded r5.ElementDefinitionConstraint
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "ele-1", *decoded.Key)
		assert.Equal(t, r5.ConstraintSeverity("error"), *decoded.Severity)
	})

	// R5-specific: ElementDefinitionBindingAdditional
	t.Run("ElementDefinitionBindingAdditional", func(t *testing.T) {
		purpose := r5.AdditionalBindingPurposeVS("starter")
		additional := r5.ElementDefinitionBindingAdditional{
			Id:            ptrStringB5("additional-1"),
			Purpose:       &purpose,
			ValueSet:      ptrStringB5("http://hl7.org/fhir/ValueSet/example"),
			Documentation: ptrStringB5("Additional binding documentation"),
		}

		data, err := json.Marshal(additional)
		require.NoError(t, err)

		var decoded r5.ElementDefinitionBindingAdditional
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "additional-1", *decoded.Id)
		assert.Equal(t, r5.AdditionalBindingPurposeVS("starter"), *decoded.Purpose)
		assert.Equal(t, "http://hl7.org/fhir/ValueSet/example", *decoded.ValueSet)
	})
}

// TestBackboneWithExtensions tests that backbone elements support extensions
func TestBackboneWithExtensions(t *testing.T) {
	t.Run("BackboneWithExtensions", func(t *testing.T) {
		contact := r5.PatientContact{
			Id: ptrStringB5("contact-ext"),
			Extension: []r5.Extension{
				{
					Url:         "http://example.org/fhir/StructureDefinition/contact-priority",
					ValueString: ptrStringB5("high"),
				},
			},
			Name: &r5.HumanName{
				Family: ptrStringB5("Doe"),
			},
		}

		data, err := json.Marshal(contact)
		require.NoError(t, err)

		var decoded r5.PatientContact
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		require.Len(t, decoded.Extension, 1)
		assert.Equal(t, "http://example.org/fhir/StructureDefinition/contact-priority", decoded.Extension[0].Url)
		assert.Equal(t, "high", *decoded.Extension[0].ValueString)
	})
}

// TestBackboneJSONSerialization tests JSON round-trip
func TestBackboneJSONSerialization(t *testing.T) {
	t.Run("PatientContactRoundTrip", func(t *testing.T) {
		jsonInput := `{
			"id": "contact-json",
			"relationship": [
				{
					"coding": [
						{
							"system": "http://terminology.hl7.org/CodeSystem/v2-0131",
							"code": "C",
							"display": "Emergency Contact"
						}
					]
				}
			],
			"name": {
				"family": "Smith",
				"given": ["John"]
			}
		}`

		var contact r5.PatientContact
		err := json.Unmarshal([]byte(jsonInput), &contact)
		require.NoError(t, err)

		assert.Equal(t, "contact-json", *contact.Id)
		assert.Equal(t, "C", *contact.Relationship[0].Coding[0].Code)
		assert.Equal(t, "Smith", *contact.Name.Family)

		// Round-trip
		data, err := json.Marshal(contact)
		require.NoError(t, err)

		var decoded r5.PatientContact
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *contact.Id, *decoded.Id)
	})
}

// Helper functions - unique names to avoid redeclaration
func ptrStringB5(s string) *string {
	return &s
}

func ptrBoolB5(b bool) *bool {
	return &b
}

func ptrDecimalB5(f float64) *r5.Decimal {
	return r5.NewDecimalFromFloat64(f)
}

func ptrUint32B5(u uint32) *uint32 {
	return &u
}

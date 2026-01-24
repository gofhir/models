package r4b_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofhir/fhir/r4b"
)

// TestResourceBackboneElements tests backbone elements that are part of resources
func TestResourceBackboneElements(t *testing.T) {
	t.Run("PatientContact", func(t *testing.T) {
		contact := r4b.PatientContact{
			Id: ptrStringBB("contact-1"),
			Relationship: []r4b.CodeableConcept{
				{
					Text: ptrStringBB("Emergency Contact"),
				},
			},
			Name: &r4b.HumanName{
				Family: ptrStringBB("Smith"),
				Given:  []string{"Jane"},
			},
		}

		data, err := json.Marshal(contact)
		require.NoError(t, err)

		var decoded r4b.PatientContact
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "contact-1", *decoded.Id)
		require.Len(t, decoded.Relationship, 1)
		assert.Equal(t, "Emergency Contact", *decoded.Relationship[0].Text)
		assert.Equal(t, "Smith", *decoded.Name.Family)
	})

	t.Run("BundleEntryRequest", func(t *testing.T) {
		method := r4b.HTTPVerb("POST")
		request := r4b.BundleEntryRequest{
			Id:     ptrStringBB("request-1"),
			Method: &method,
			Url:    ptrStringBB("Patient"),
		}

		data, err := json.Marshal(request)
		require.NoError(t, err)

		var decoded r4b.BundleEntryRequest
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "request-1", *decoded.Id)
		assert.Equal(t, r4b.HTTPVerb("POST"), *decoded.Method)
		assert.Equal(t, "Patient", *decoded.Url)
	})

	t.Run("BundleEntryResponse", func(t *testing.T) {
		response := r4b.BundleEntryResponse{
			Id:       ptrStringBB("response-1"),
			Status:   ptrStringBB("201 Created"),
			Location: ptrStringBB("Patient/123/_history/1"),
			Etag:     ptrStringBB("W/\"1\""),
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var decoded r4b.BundleEntryResponse
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "response-1", *decoded.Id)
		assert.Equal(t, "201 Created", *decoded.Status)
	})

	t.Run("BundleEntrySearch", func(t *testing.T) {
		mode := r4b.SearchEntryMode("match")
		search := r4b.BundleEntrySearch{
			Id:    ptrStringBB("search-1"),
			Mode:  &mode,
			Score: ptrFloat64BB(0.95),
		}

		data, err := json.Marshal(search)
		require.NoError(t, err)

		var decoded r4b.BundleEntrySearch
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "search-1", *decoded.Id)
		assert.Equal(t, r4b.SearchEntryMode("match"), *decoded.Mode)
		assert.Equal(t, 0.95, *decoded.Score)
	})

	t.Run("ObservationComponent", func(t *testing.T) {
		component := r4b.ObservationComponent{
			Id: ptrStringBB("comp-1"),
			Code: r4b.CodeableConcept{
				Coding: []r4b.Coding{
					{
						System:  ptrStringBB("http://loinc.org"),
						Code:    ptrStringBB("8480-6"),
						Display: ptrStringBB("Systolic blood pressure"),
					},
				},
			},
			ValueQuantity: &r4b.Quantity{
				Value:  ptrFloat64BB(120),
				Unit:   ptrStringBB("mmHg"),
				System: ptrStringBB("http://unitsofmeasure.org"),
				Code:   ptrStringBB("mm[Hg]"),
			},
		}

		data, err := json.Marshal(component)
		require.NoError(t, err)

		var decoded r4b.ObservationComponent
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "comp-1", *decoded.Id)
		assert.Equal(t, "8480-6", *decoded.Code.Coding[0].Code)
		assert.Equal(t, float64(120), *decoded.ValueQuantity.Value)
	})
}

// TestDatatypeBackboneElements tests backbone elements from datatype-style types
func TestDatatypeBackboneElements(t *testing.T) {
	t.Run("DosageDoseAndRate", func(t *testing.T) {
		doseAndRate := r4b.DosageDoseAndRate{
			Id: ptrStringBB("dose-1"),
			Type: &r4b.CodeableConcept{
				Coding: []r4b.Coding{
					{
						System:  ptrStringBB("http://terminology.hl7.org/CodeSystem/dose-rate-type"),
						Code:    ptrStringBB("ordered"),
						Display: ptrStringBB("Ordered"),
					},
				},
			},
			DoseQuantity: &r4b.Quantity{
				Value:  ptrFloat64BB(500),
				Unit:   ptrStringBB("mg"),
				System: ptrStringBB("http://unitsofmeasure.org"),
				Code:   ptrStringBB("mg"),
			},
		}

		data, err := json.Marshal(doseAndRate)
		require.NoError(t, err)

		var decoded r4b.DosageDoseAndRate
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "dose-1", *decoded.Id)
		assert.Equal(t, "ordered", *decoded.Type.Coding[0].Code)
		assert.Equal(t, float64(500), *decoded.DoseQuantity.Value)
	})

	t.Run("TimingRepeat", func(t *testing.T) {
		periodUnit := r4b.UnitsOfTime("d")
		repeat := r4b.TimingRepeat{
			Id:          ptrStringBB("repeat-1"),
			Frequency:   ptrUint32BB(2),
			Period:      ptrFloat64BB(1),
			PeriodUnit:  &periodUnit,
			DayOfWeek:   []r4b.DaysOfWeek{"mon", "wed", "fri"},
			TimeOfDay:   []string{"08:00:00", "18:00:00"},
			Duration:    ptrFloat64BB(30),
			DurationMax: ptrFloat64BB(60),
		}

		data, err := json.Marshal(repeat)
		require.NoError(t, err)

		var decoded r4b.TimingRepeat
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "repeat-1", *decoded.Id)
		assert.Equal(t, uint32(2), *decoded.Frequency)
		assert.Equal(t, r4b.UnitsOfTime("d"), *decoded.PeriodUnit)
	})

	t.Run("ElementDefinitionSlicing", func(t *testing.T) {
		rules := r4b.SlicingRules("open")
		slicing := r4b.ElementDefinitionSlicing{
			Id:          ptrStringBB("slicing-1"),
			Description: ptrStringBB("Slice by code"),
			Ordered:     ptrBoolBB(false),
			Rules:       &rules,
		}

		data, err := json.Marshal(slicing)
		require.NoError(t, err)

		var decoded r4b.ElementDefinitionSlicing
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "slicing-1", *decoded.Id)
		assert.Equal(t, "Slice by code", *decoded.Description)
		assert.False(t, *decoded.Ordered)
		assert.Equal(t, r4b.SlicingRules("open"), *decoded.Rules)
	})

	t.Run("ElementDefinitionBinding", func(t *testing.T) {
		strength := r4b.BindingStrength("required")
		binding := r4b.ElementDefinitionBinding{
			Id:          ptrStringBB("binding-1"),
			Strength:    &strength,
			Description: ptrStringBB("The status of the encounter"),
			ValueSet:    ptrStringBB("http://hl7.org/fhir/ValueSet/encounter-status"),
		}

		data, err := json.Marshal(binding)
		require.NoError(t, err)

		var decoded r4b.ElementDefinitionBinding
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "binding-1", *decoded.Id)
		assert.Equal(t, r4b.BindingStrength("required"), *decoded.Strength)
	})

	t.Run("ElementDefinitionConstraint", func(t *testing.T) {
		severity := r4b.ConstraintSeverity("error")
		constraint := r4b.ElementDefinitionConstraint{
			Id:         ptrStringBB("constraint-1"),
			Key:        ptrStringBB("ele-1"),
			Severity:   &severity,
			Human:      ptrStringBB("All FHIR elements must have a @value or children"),
			Expression: ptrStringBB("hasValue() or (children().count() > id.count())"),
		}

		data, err := json.Marshal(constraint)
		require.NoError(t, err)

		var decoded r4b.ElementDefinitionConstraint
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "ele-1", *decoded.Key)
		assert.Equal(t, r4b.ConstraintSeverity("error"), *decoded.Severity)
	})
}

// TestBackboneWithExtensions tests that backbone elements support extensions
func TestBackboneWithExtensions(t *testing.T) {
	t.Run("BackboneWithExtensions", func(t *testing.T) {
		contact := r4b.PatientContact{
			Id: ptrStringBB("contact-ext"),
			Extension: []r4b.Extension{
				{
					Url:         "http://example.org/fhir/StructureDefinition/contact-priority",
					ValueString: ptrStringBB("high"),
				},
			},
			Name: &r4b.HumanName{
				Family: ptrStringBB("Doe"),
			},
		}

		data, err := json.Marshal(contact)
		require.NoError(t, err)

		var decoded r4b.PatientContact
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

		var contact r4b.PatientContact
		err := json.Unmarshal([]byte(jsonInput), &contact)
		require.NoError(t, err)

		assert.Equal(t, "contact-json", *contact.Id)
		assert.Equal(t, "C", *contact.Relationship[0].Coding[0].Code)
		assert.Equal(t, "Smith", *contact.Name.Family)

		// Round-trip
		data, err := json.Marshal(contact)
		require.NoError(t, err)

		var decoded r4b.PatientContact
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *contact.Id, *decoded.Id)
	})
}

// Helper functions - unique names to avoid redeclaration
func ptrStringBB(s string) *string {
	return &s
}

func ptrBoolBB(b bool) *bool {
	return &b
}

func ptrFloat64BB(f float64) *float64 {
	return &f
}

func ptrUint32BB(u uint32) *uint32 {
	return &u
}

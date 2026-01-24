package r4_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofhir/fhir/r4"
)

// TestResourceBackboneElements tests backbone elements that are part of resources
func TestResourceBackboneElements(t *testing.T) {
	t.Run("PatientContact", func(t *testing.T) {
		contact := r4.PatientContact{
			Id: ptrStringB("contact-1"),
			Relationship: []r4.CodeableConcept{
				{
					Text: ptrStringB("Emergency Contact"),
				},
			},
			Name: &r4.HumanName{
				Family: ptrStringB("Smith"),
				Given:  []string{"Jane"},
			},
		}

		data, err := json.Marshal(contact)
		require.NoError(t, err)

		var decoded r4.PatientContact
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "contact-1", *decoded.Id)
		require.Len(t, decoded.Relationship, 1)
		assert.Equal(t, "Emergency Contact", *decoded.Relationship[0].Text)
		assert.Equal(t, "Smith", *decoded.Name.Family)
	})

	t.Run("BundleEntryRequest", func(t *testing.T) {
		method := r4.HTTPVerb("POST")
		request := r4.BundleEntryRequest{
			Id:     ptrStringB("request-1"),
			Method: &method,
			Url:    ptrStringB("Patient"),
		}

		data, err := json.Marshal(request)
		require.NoError(t, err)

		var decoded r4.BundleEntryRequest
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "request-1", *decoded.Id)
		assert.Equal(t, r4.HTTPVerb("POST"), *decoded.Method)
		assert.Equal(t, "Patient", *decoded.Url)
	})

	t.Run("BundleEntryResponse", func(t *testing.T) {
		response := r4.BundleEntryResponse{
			Id:       ptrStringB("response-1"),
			Status:   ptrStringB("201 Created"),
			Location: ptrStringB("Patient/123/_history/1"),
			Etag:     ptrStringB("W/\"1\""),
		}

		data, err := json.Marshal(response)
		require.NoError(t, err)

		var decoded r4.BundleEntryResponse
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "response-1", *decoded.Id)
		assert.Equal(t, "201 Created", *decoded.Status)
		assert.Equal(t, "Patient/123/_history/1", *decoded.Location)
		assert.Equal(t, "W/\"1\"", *decoded.Etag)
	})

	t.Run("BundleEntrySearch", func(t *testing.T) {
		mode := r4.SearchEntryMode("match")
		search := r4.BundleEntrySearch{
			Id:    ptrStringB("search-1"),
			Mode:  &mode,
			Score: ptrFloat64B(0.95),
		}

		data, err := json.Marshal(search)
		require.NoError(t, err)

		var decoded r4.BundleEntrySearch
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "search-1", *decoded.Id)
		assert.Equal(t, r4.SearchEntryMode("match"), *decoded.Mode)
		assert.Equal(t, 0.95, *decoded.Score)
	})

	t.Run("ObservationComponent", func(t *testing.T) {
		component := r4.ObservationComponent{
			Id: ptrStringB("comp-1"),
			Code: r4.CodeableConcept{
				Coding: []r4.Coding{
					{
						System:  ptrStringB("http://loinc.org"),
						Code:    ptrStringB("8480-6"),
						Display: ptrStringB("Systolic blood pressure"),
					},
				},
			},
			ValueQuantity: &r4.Quantity{
				Value:  ptrFloat64B(120),
				Unit:   ptrStringB("mmHg"),
				System: ptrStringB("http://unitsofmeasure.org"),
				Code:   ptrStringB("mm[Hg]"),
			},
		}

		data, err := json.Marshal(component)
		require.NoError(t, err)

		var decoded r4.ObservationComponent
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "comp-1", *decoded.Id)
		assert.Equal(t, "8480-6", *decoded.Code.Coding[0].Code)
		assert.Equal(t, float64(120), *decoded.ValueQuantity.Value)
	})

	t.Run("ClaimItem", func(t *testing.T) {
		item := r4.ClaimItem{
			Id:       ptrStringB("item-1"),
			Sequence: ptrUint32B(1),
			ProductOrService: r4.CodeableConcept{
				Coding: []r4.Coding{
					{
						System: ptrStringB("http://example.org/fhir/CodeSystem/ex-USCLS"),
						Code:   ptrStringB("1205"),
					},
				},
			},
			Quantity: &r4.Quantity{
				Value: ptrFloat64B(1),
			},
			UnitPrice: &r4.Money{
				Value:    ptrFloat64B(135.57),
				Currency: ptrStringB("USD"),
			},
		}

		data, err := json.Marshal(item)
		require.NoError(t, err)

		var decoded r4.ClaimItem
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, uint32(1), *decoded.Sequence)
		assert.Equal(t, "1205", *decoded.ProductOrService.Coding[0].Code)
		assert.Equal(t, 135.57, *decoded.UnitPrice.Value)
	})

	t.Run("AllergyIntoleranceReaction", func(t *testing.T) {
		severity := r4.AllergyIntoleranceSeverity("mild")
		reaction := r4.AllergyIntoleranceReaction{
			Id: ptrStringB("reaction-1"),
			Substance: &r4.CodeableConcept{
				Text: ptrStringB("Peanuts"),
			},
			Manifestation: []r4.CodeableConcept{
				{
					Text: ptrStringB("Hives"),
				},
			},
			Severity:    &severity,
			Description: ptrStringB("Mild allergic reaction"),
		}

		data, err := json.Marshal(reaction)
		require.NoError(t, err)

		var decoded r4.AllergyIntoleranceReaction
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "reaction-1", *decoded.Id)
		assert.Equal(t, "Peanuts", *decoded.Substance.Text)
		assert.Equal(t, r4.AllergyIntoleranceSeverity("mild"), *decoded.Severity)
	})
}

// TestDatatypeBackboneElements tests backbone elements from datatype-style types
// These are types like Dosage, ElementDefinition, Timing that extend BackboneElement
func TestDatatypeBackboneElements(t *testing.T) {
	t.Run("DosageDoseAndRate", func(t *testing.T) {
		doseAndRate := r4.DosageDoseAndRate{
			Id: ptrStringB("dose-1"),
			Type: &r4.CodeableConcept{
				Coding: []r4.Coding{
					{
						System:  ptrStringB("http://terminology.hl7.org/CodeSystem/dose-rate-type"),
						Code:    ptrStringB("ordered"),
						Display: ptrStringB("Ordered"),
					},
				},
			},
			DoseQuantity: &r4.Quantity{
				Value:  ptrFloat64B(500),
				Unit:   ptrStringB("mg"),
				System: ptrStringB("http://unitsofmeasure.org"),
				Code:   ptrStringB("mg"),
			},
		}

		data, err := json.Marshal(doseAndRate)
		require.NoError(t, err)

		var decoded r4.DosageDoseAndRate
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "dose-1", *decoded.Id)
		assert.Equal(t, "ordered", *decoded.Type.Coding[0].Code)
		assert.Equal(t, float64(500), *decoded.DoseQuantity.Value)
		assert.Equal(t, "mg", *decoded.DoseQuantity.Unit)
	})

	t.Run("TimingRepeat", func(t *testing.T) {
		periodUnit := r4.UnitsOfTime("d")
		repeat := r4.TimingRepeat{
			Id:          ptrStringB("repeat-1"),
			Frequency:   ptrUint32B(2),
			Period:      ptrFloat64B(1),
			PeriodUnit:  &periodUnit,
			DayOfWeek:   []r4.DaysOfWeek{"mon", "wed", "fri"},
			TimeOfDay:   []string{"08:00:00", "18:00:00"},
			Duration:    ptrFloat64B(30),
			DurationMax: ptrFloat64B(60),
		}

		data, err := json.Marshal(repeat)
		require.NoError(t, err)

		var decoded r4.TimingRepeat
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "repeat-1", *decoded.Id)
		assert.Equal(t, uint32(2), *decoded.Frequency)
		assert.Equal(t, float64(1), *decoded.Period)
		assert.Equal(t, r4.UnitsOfTime("d"), *decoded.PeriodUnit)
		assert.Equal(t, []r4.DaysOfWeek{"mon", "wed", "fri"}, decoded.DayOfWeek)
		assert.Equal(t, []string{"08:00:00", "18:00:00"}, decoded.TimeOfDay)
	})

	t.Run("ElementDefinitionSlicing", func(t *testing.T) {
		rules := r4.SlicingRules("open")
		slicing := r4.ElementDefinitionSlicing{
			Id:          ptrStringB("slicing-1"),
			Description: ptrStringB("Slice by code"),
			Ordered:     ptrBoolB(false),
			Rules:       &rules,
		}

		data, err := json.Marshal(slicing)
		require.NoError(t, err)

		var decoded r4.ElementDefinitionSlicing
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "slicing-1", *decoded.Id)
		assert.Equal(t, "Slice by code", *decoded.Description)
		assert.False(t, *decoded.Ordered)
		assert.Equal(t, r4.SlicingRules("open"), *decoded.Rules)
	})

	t.Run("ElementDefinitionSlicingDiscriminator", func(t *testing.T) {
		discType := r4.DiscriminatorType("value")
		discriminator := r4.ElementDefinitionSlicingDiscriminator{
			Id:   ptrStringB("disc-1"),
			Type: &discType,
			Path: ptrStringB("code"),
		}

		data, err := json.Marshal(discriminator)
		require.NoError(t, err)

		var decoded r4.ElementDefinitionSlicingDiscriminator
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "disc-1", *decoded.Id)
		assert.Equal(t, r4.DiscriminatorType("value"), *decoded.Type)
		assert.Equal(t, "code", *decoded.Path)
	})

	t.Run("ElementDefinitionType", func(t *testing.T) {
		versioning := r4.ReferenceVersionRules("independent")
		elemType := r4.ElementDefinitionType{
			Id:            ptrStringB("type-1"),
			Code:          ptrStringB("Reference"),
			TargetProfile: []string{"http://hl7.org/fhir/StructureDefinition/Patient"},
			Aggregation:   []r4.AggregationMode{"referenced", "bundled"},
			Versioning:    &versioning,
		}

		data, err := json.Marshal(elemType)
		require.NoError(t, err)

		var decoded r4.ElementDefinitionType
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "type-1", *decoded.Id)
		assert.Equal(t, "Reference", *decoded.Code)
		assert.Contains(t, decoded.TargetProfile, "http://hl7.org/fhir/StructureDefinition/Patient")
		assert.Equal(t, []r4.AggregationMode{"referenced", "bundled"}, decoded.Aggregation)
	})

	t.Run("ElementDefinitionBinding", func(t *testing.T) {
		strength := r4.BindingStrength("required")
		binding := r4.ElementDefinitionBinding{
			Id:          ptrStringB("binding-1"),
			Strength:    &strength,
			Description: ptrStringB("The status of the encounter"),
			ValueSet:    ptrStringB("http://hl7.org/fhir/ValueSet/encounter-status"),
		}

		data, err := json.Marshal(binding)
		require.NoError(t, err)

		var decoded r4.ElementDefinitionBinding
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "binding-1", *decoded.Id)
		assert.Equal(t, r4.BindingStrength("required"), *decoded.Strength)
		assert.Equal(t, "http://hl7.org/fhir/ValueSet/encounter-status", *decoded.ValueSet)
	})

	t.Run("ElementDefinitionConstraint", func(t *testing.T) {
		severity := r4.ConstraintSeverity("error")
		constraint := r4.ElementDefinitionConstraint{
			Id:         ptrStringB("constraint-1"),
			Key:        ptrStringB("ele-1"),
			Severity:   &severity,
			Human:      ptrStringB("All FHIR elements must have a @value or children"),
			Expression: ptrStringB("hasValue() or (children().count() > id.count())"),
			Xpath:      ptrStringB("@value|f:*|h:div"),
			Source:     ptrStringB("http://hl7.org/fhir/StructureDefinition/Element"),
		}

		data, err := json.Marshal(constraint)
		require.NoError(t, err)

		var decoded r4.ElementDefinitionConstraint
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "ele-1", *decoded.Key)
		assert.Equal(t, r4.ConstraintSeverity("error"), *decoded.Severity)
		assert.Contains(t, *decoded.Human, "FHIR elements")
	})

	t.Run("ElementDefinitionMapping", func(t *testing.T) {
		mapping := r4.ElementDefinitionMapping{
			Id:       ptrStringB("mapping-1"),
			Identity: ptrStringB("rim"),
			Language: ptrStringB("application/xml"),
			Map:      ptrStringB("n/a"),
			Comment:  ptrStringB("RIM mapping"),
		}

		data, err := json.Marshal(mapping)
		require.NoError(t, err)

		var decoded r4.ElementDefinitionMapping
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "mapping-1", *decoded.Id)
		assert.Equal(t, "rim", *decoded.Identity)
		assert.Equal(t, "n/a", *decoded.Map)
	})

	t.Run("ElementDefinitionBase", func(t *testing.T) {
		base := r4.ElementDefinitionBase{
			Id:   ptrStringB("base-1"),
			Path: ptrStringB("Element.id"),
			Min:  ptrUint32B(0),
			Max:  ptrStringB("1"),
		}

		data, err := json.Marshal(base)
		require.NoError(t, err)

		var decoded r4.ElementDefinitionBase
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, "base-1", *decoded.Id)
		assert.Equal(t, "Element.id", *decoded.Path)
		assert.Equal(t, uint32(0), *decoded.Min)
		assert.Equal(t, "1", *decoded.Max)
	})
}

// TestBackboneWithExtensions tests that backbone elements properly support extensions
func TestBackboneWithExtensions(t *testing.T) {
	t.Run("BackboneWithExtensionsAndModifierExtensions", func(t *testing.T) {
		contact := r4.PatientContact{
			Id: ptrStringB("contact-ext"),
			Extension: []r4.Extension{
				{
					Url:         "http://example.org/fhir/StructureDefinition/contact-priority",
					ValueString: ptrStringB("high"),
				},
			},
			ModifierExtension: []r4.Extension{
				{
					Url:          "http://example.org/fhir/StructureDefinition/contact-inactive",
					ValueBoolean: ptrBoolB(false),
				},
			},
			Name: &r4.HumanName{
				Family: ptrStringB("Doe"),
			},
		}

		data, err := json.Marshal(contact)
		require.NoError(t, err)

		var decoded r4.PatientContact
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		require.Len(t, decoded.Extension, 1)
		assert.Equal(t, "http://example.org/fhir/StructureDefinition/contact-priority", decoded.Extension[0].Url)
		assert.Equal(t, "high", *decoded.Extension[0].ValueString)

		require.Len(t, decoded.ModifierExtension, 1)
		assert.Equal(t, "http://example.org/fhir/StructureDefinition/contact-inactive", decoded.ModifierExtension[0].Url)
		assert.False(t, *decoded.ModifierExtension[0].ValueBoolean)
	})
}

// TestBackboneJSONSerialization tests complete JSON round-trip for backbone elements
func TestBackboneJSONSerialization(t *testing.T) {
	t.Run("CompletePatientContactRoundTrip", func(t *testing.T) {
		jsonInput := `{
			"id": "contact-json",
			"extension": [
				{
					"url": "http://example.org/ext",
					"valueString": "test"
				}
			],
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
				"given": ["John", "Michael"]
			},
			"telecom": [
				{
					"system": "phone",
					"value": "+1-555-555-5555",
					"use": "home"
				}
			]
		}`

		var contact r4.PatientContact
		err := json.Unmarshal([]byte(jsonInput), &contact)
		require.NoError(t, err)

		assert.Equal(t, "contact-json", *contact.Id)
		require.Len(t, contact.Extension, 1)
		require.Len(t, contact.Relationship, 1)
		assert.Equal(t, "C", *contact.Relationship[0].Coding[0].Code)
		assert.Equal(t, "Smith", *contact.Name.Family)
		assert.Equal(t, []string{"John", "Michael"}, contact.Name.Given)
		require.Len(t, contact.Telecom, 1)
		assert.Equal(t, "+1-555-555-5555", *contact.Telecom[0].Value)

		// Round-trip back to JSON
		data, err := json.Marshal(contact)
		require.NoError(t, err)

		var decoded r4.PatientContact
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *contact.Id, *decoded.Id)
		assert.Equal(t, *contact.Name.Family, *decoded.Name.Family)
	})

	t.Run("CompleteDosageDoseAndRateRoundTrip", func(t *testing.T) {
		jsonInput := `{
			"id": "dose-json",
			"type": {
				"coding": [
					{
						"system": "http://terminology.hl7.org/CodeSystem/dose-rate-type",
						"code": "calculated",
						"display": "Calculated"
					}
				]
			},
			"doseQuantity": {
				"value": 250,
				"unit": "mg",
				"system": "http://unitsofmeasure.org",
				"code": "mg"
			},
			"rateRatio": {
				"numerator": {
					"value": 500,
					"unit": "mL"
				},
				"denominator": {
					"value": 1,
					"unit": "h"
				}
			}
		}`

		var doseAndRate r4.DosageDoseAndRate
		err := json.Unmarshal([]byte(jsonInput), &doseAndRate)
		require.NoError(t, err)

		assert.Equal(t, "dose-json", *doseAndRate.Id)
		assert.Equal(t, "calculated", *doseAndRate.Type.Coding[0].Code)
		assert.Equal(t, float64(250), *doseAndRate.DoseQuantity.Value)
		assert.NotNil(t, doseAndRate.RateRatio)
		assert.Equal(t, float64(500), *doseAndRate.RateRatio.Numerator.Value)
	})
}

// Helper functions - use unique names to avoid conflicts with other test files
func ptrStringB(s string) *string {
	return &s
}

func ptrBoolB(b bool) *bool {
	return &b
}

func ptrFloat64B(f float64) *float64 {
	return &f
}

func ptrUint32B(u uint32) *uint32 {
	return &u
}

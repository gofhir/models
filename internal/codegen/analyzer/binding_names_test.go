package analyzer

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gofhir/models/internal/codegen/parser"
)

// The specification names each required binding through an extension —
// elementdefinition-bindingName — and that name is a far better Go type name than
// the ValueSet's title: MedicationStatus rather than MedicationStatusCodes,
// CurrencyCode rather than Currencies.
//
// It is upstream data, though, and not written with Go in mind. These tests cover
// the four conditions under which it is refused, each of which was found by a
// binding that broke something.

// bindingSD builds a StructureDefinition whose elements carry required bindings,
// from (elementPath, valueSetURL, bindingName) triples. An empty bindingName omits
// the extension.
func bindingSD(t *testing.T, resource string, bindings ...[3]string) *parser.StructureDefinition {
	t.Helper()

	elements := make([]any, 0, len(bindings))
	for _, b := range bindings {
		binding := map[string]any{"strength": "required", "valueSet": b[1]}
		if b[2] != "" {
			binding["extension"] = []any{map[string]any{
				"url":         "http://hl7.org/fhir/StructureDefinition/elementdefinition-bindingName",
				"valueString": b[2],
			}}
		}
		elements = append(elements, map[string]any{
			"id": b[0], "path": b[0], "min": 0, "max": "1", "binding": binding,
		})
	}

	data, err := json.Marshal(map[string]any{
		"resourceType": "StructureDefinition",
		"id":           resource,
		"url":          "http://hl7.org/fhir/StructureDefinition/" + resource,
		"name":         resource,
		"status":       "active",
		"kind":         "resource",
		"abstract":     false,
		"type":         resource,
		"snapshot":     map[string]any{"element": elements},
	})
	require.NoError(t, err)

	sd, err := parser.ParseStructureDefinition(data)
	require.NoError(t, err)
	return sd
}

// valueSetRegistry builds a registry from (url, name) pairs, each carrying one
// code so it counts as a candidate for enum generation.
func valueSetRegistry(t *testing.T, pairs ...[2]string) *parser.ValueSetRegistry {
	t.Helper()

	entries := make([]string, 0, len(pairs))
	for i, p := range pairs {
		entries = append(entries, fmt.Sprintf(
			`{"resource":{"resourceType":"ValueSet","id":"vs%d","url":%q,"name":%q,`+
				`"compose":{"include":[{"system":"http://example.org/s%d",`+
				`"concept":[{"code":"active","display":"Active"}]}]}}}`,
			i, p[0], p[1], i))
	}

	registry := parser.NewValueSetRegistry()
	require.NoError(t, registry.LoadFromBundle(
		[]byte(`{"resourceType":"Bundle","entry":[`+strings.Join(entries, ",")+`]}`)))
	return registry
}

func TestBindingNameReplacesTheValueSetTitle(t *testing.T) {
	const url = "http://hl7.org/fhir/ValueSet/medication-status"

	a := NewAnalyzer(
		[]*parser.StructureDefinition{
			bindingSD(t, "Medication", [3]string{"Medication.status", url, "MedicationStatus"}),
		},
		valueSetRegistry(t, [2]string{url, "Medication Status Codes"}),
	)

	require.Equal(t, "MedicationStatus", a.ValueSetTypeName(url, "Medication Status Codes"),
		"the binding name is what the specification calls this binding")
}

func TestBindingNameIsRefusedWhenBindingsDisagree(t *testing.T) {
	// request-priority is bound from five places in R4 under five different names
	// — CommunicationPriority, TaskPriority, ServiceRequestPriority and two more.
	// A ValueSet becomes exactly one Go type, so there is no answer to pick.
	const url = "http://hl7.org/fhir/ValueSet/request-priority"

	a := NewAnalyzer(
		[]*parser.StructureDefinition{
			bindingSD(t, "Task", [3]string{"Task.priority", url, "TaskPriority"}),
			bindingSD(t, "Communication", [3]string{"Communication.priority", url, "CommunicationPriority"}),
		},
		valueSetRegistry(t, [2]string{url, "Request Priority"}),
	)

	require.Equal(t, "RequestPriority", a.ValueSetTypeName(url, "Request Priority"),
		"with several names claiming one ValueSet, the title is the only neutral choice")
}

func TestBindingNameIsRefusedWhenItNamesAResource(t *testing.T) {
	// R4B and R5 bind subscription-status as SubscriptionStatus, which is also a
	// resource. Taking the binding name declares that identifier twice and the
	// generated package stops compiling — found exactly that way.
	const url = "http://hl7.org/fhir/ValueSet/subscription-status"

	a := NewAnalyzer(
		[]*parser.StructureDefinition{
			bindingSD(t, "Subscription", [3]string{"Subscription.status", url, "SubscriptionStatus"}),
			bindingSD(t, "SubscriptionStatus"),
		},
		valueSetRegistry(t, [2]string{url, "Subscription Status Codes"}),
	)

	got := a.ValueSetTypeName(url, "Subscription Status Codes")
	require.NotEqual(t, "SubscriptionStatus", got,
		"the enum would collide with the resource of the same name")
	require.Equal(t, "SubscriptionStatusCodes", got)
}

func TestBindingNameIsRefusedWhenItSaysLess(t *testing.T) {
	// The artifact-assessment bindings are named Disposition, InformationType and
	// WorkflowStatus, and verificationresult-status is named plainly Status. Each
	// is a suffix of the title-derived name, so it says strictly less about what
	// the ValueSet is — and an exported Status among hundreds of enums is worse
	// than the name it would replace.
	for _, tc := range []struct{ url, title, bindingName string }{
		{"http://hl7.org/fhir/ValueSet/verificationresult-status", "Verification Result Status", "Status"},
		{"http://hl7.org/fhir/ValueSet/artifactassessment-disposition", "Artifact Assessment Disposition", "Disposition"},
		{"http://hl7.org/fhir/ValueSet/artifactassessment-workflow-status", "Artifact Assessment Workflow Status", "WorkflowStatus"},
	} {
		t.Run(tc.bindingName, func(t *testing.T) {
			a := NewAnalyzer(
				[]*parser.StructureDefinition{
					bindingSD(t, "Res", [3]string{"Res.field", tc.url, tc.bindingName}),
				},
				valueSetRegistry(t, [2]string{tc.url, tc.title}),
			)

			got := a.ValueSetTypeName(tc.url, tc.title)
			require.NotEqual(t, tc.bindingName, got, "the shorter name loses the context")
			require.Equal(t, sanitizeTypeName(tc.title), got)
		})
	}
}

func TestBindingNameIsRefusedWhenItCollidesWithAnotherValueSet(t *testing.T) {
	// Specimen.combined in R5 — grouped | pooled — is annotated
	// bindingName="PublicationStatus", which belongs to a different ValueSet
	// entirely. Taking it would put specimen-combined and publication-status on
	// one Go type.
	//
	// Both ValueSets had distinct names already, so the fix is to drop the binding
	// name rather than rename either of them. Generation fails outright on an
	// unresolved collision, so this is not a case that could pass silently — but
	// failing the build over an upstream annotation error is not a fix.
	const (
		publicationURL = "http://hl7.org/fhir/ValueSet/publication-status"
		specimenURL    = "http://hl7.org/fhir/ValueSet/specimen-combined"
	)

	a := NewAnalyzer(
		[]*parser.StructureDefinition{
			bindingSD(t, "Specimen", [3]string{"Specimen.combined", specimenURL, "PublicationStatus"}),
		},
		valueSetRegistry(t,
			[2]string{publicationURL, "PublicationStatus"},
			[2]string{specimenURL, "SpecimenCombined"},
		),
	)

	publication := a.ValueSetTypeName(publicationURL, "PublicationStatus")
	specimen := a.ValueSetTypeName(specimenURL, "SpecimenCombined")
	require.NotEqual(t, publication, specimen, "both ValueSets resolved to one type")
	require.Equal(t, "PublicationStatus", publication)
	require.Equal(t, "SpecimenCombined", specimen)
}

func TestBindingNameIsRefusedWhenItIsNotAnIdentifier(t *testing.T) {
	// The specification carries binding names like "appointment-type" and
	// "LOINC LL379-9 answerlist". Sanitizing them lands back on the name we
	// already had, so there is nothing to gain by trying.
	const url = "http://hl7.org/fhir/ValueSet/appointment-type"

	a := NewAnalyzer(
		[]*parser.StructureDefinition{
			bindingSD(t, "Appointment", [3]string{"Appointment.type", url, "appointment-type"}),
		},
		valueSetRegistry(t, [2]string{url, "Appointment Type"}),
	)

	require.Equal(t, "AppointmentType", a.ValueSetTypeName(url, "Appointment Type"))
}

func TestMisspelledBindingNamesAreRefused(t *testing.T) {
	// ConceptMap.additionalAttribute.type in R5 is bound as
	// ConceptMapmapAttributeType — "Map" twice, the second time in lower case.
	const url = "http://hl7.org/fhir/ValueSet/conceptmap-attribute-type"

	a := NewAnalyzer(
		[]*parser.StructureDefinition{
			bindingSD(t, "ConceptMap", [3]string{
				"ConceptMap.additionalAttribute.type", url, "ConceptMapmapAttributeType"}),
		},
		valueSetRegistry(t, [2]string{url, "Concept Map Attribute Type"}),
	)

	require.Equal(t, "ConceptMapAttributeType", a.ValueSetTypeName(url, "Concept Map Attribute Type"))
}

package conformance

// Builders existed for all 146 resources and none of the 58 datatypes, so the
// fluent chain stopped exactly where the data starts: you could build a Patient
// fluently right up to the moment you needed a HumanName, and then dropped into a
// struct literal with hand-written pointers.
//
// Task 6.6 said to extend the builders before removing the functional options and
// it was done in the reverse order. Nothing regressed — the options did not cover
// datatypes either — but this is the half that was owed.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestTheFluentChainReachesTheDatatypes(t *testing.T) {
	p := r4.NewPatientBuilder().
		SetId("p1").
		AddName(*r4.NewHumanNameBuilder().
			SetFamily("Smith").
			AddGiven("John").
			Build()).
		Build()

	out, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"resourceType":"Patient","id":"p1","name":[{"family":"Smith","given":["John"]}]}`
	if string(out) != want {
		t.Errorf("got  %s\nwant %s", out, want)
	}
}

func TestDatatypeBuilderCoversTheOnesActuallyUsed(t *testing.T) {
	// Measured over 1,200 published R4 examples: CodeableConcept appears 165,085
	// times, Reference 62,690, Extension 23,378, Quantity 22,434. These are what a
	// caller writes by hand.
	cc := r4.NewCodeableConceptBuilder().
		SetText("Diabetes").
		AddCoding(*r4.NewCodingBuilder().
			SetSystem("http://snomed.info/sct").
			SetCode("73211009").
			Build()).
		Build()

	out, err := json.Marshal(cc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"text":"Diabetes"`, `"code":"73211009"`, `"system":"http://snomed.info/sct"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("%s missing from %s", want, out)
		}
	}

	ref := r4.NewReferenceBuilder().SetReference("Patient/p1").SetDisplay("Smith").Build()
	if ref.Reference == nil || *ref.Reference != "Patient/p1" {
		t.Error("Reference builder did not set the reference")
	}
}

func TestChoiceExclusivityNowReachesDatatypes(t *testing.T) {
	// Extension.value[x] has 71 variants and no builder at all until now, so it
	// was the largest choice group with no protection.
	e := r4.NewExtensionBuilder().
		SetUrl("http://example.org/x").
		SetValueString("a").
		SetValueBoolean(true).
		Build()

	out, err := json.Marshal(e)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), "valueString") {
		t.Errorf("both variants survived on a datatype: %s", out)
	}
	if !strings.Contains(string(out), `"valueBoolean":true`) {
		t.Errorf("the last setter did not win: %s", out)
	}
	if !strings.Contains(string(out), `"url":"http://example.org/x"`) {
		t.Errorf("clearing the choice took an unrelated field with it: %s", out)
	}
}

func TestBuilderForATypeNamedAfterAKeyword(t *testing.T) {
	// Range lower-cases to "range", which is reserved, and the generated struct
	// field did not compile. Every resource name happens to be safe; this is the
	// only datatype that is not, in all three versions.
	rg := r4.NewRangeBuilder().
		SetLow(*r4.NewQuantityBuilder().SetValue(*r4.MustDecimal("1")).Build()).
		Build()

	out, err := json.Marshal(rg)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"low"`) {
		t.Errorf("got %s", out)
	}
}

func TestDatatypeBuildersInEveryVersion(t *testing.T) {
	t.Run("r4b", func(t *testing.T) {
		n := r4b.NewHumanNameBuilder().SetFamily("Smith").Build()
		if n.Family == nil || *n.Family != "Smith" {
			t.Error("family did not set")
		}
	})
	t.Run("r5", func(t *testing.T) {
		n := r5.NewHumanNameBuilder().SetFamily("Smith").Build()
		if n.Family == nil || *n.Family != "Smith" {
			t.Error("family did not set")
		}
	})
}

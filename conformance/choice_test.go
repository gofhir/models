package conformance

// A choice element holds exactly one variant: Observation.value[x] is a
// valueString or a valueBoolean or a valueQuantity, never two at once. Nothing
// enforced that, so a chain of builder calls produced a document with several
// present — which no FHIR server accepts and which nothing here reported.
//
// This is not data loss. It is the opposite: writing more than the specification
// allows, and being told nothing about it.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestChoiceSettersAreExclusive(t *testing.T) {
	o := r4.NewObservationBuilder().
		SetValueString("a").
		SetValueBoolean(true).
		Build()

	out, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), "valueString") {
		t.Errorf("both variants survived, which FHIR does not allow: %s", out)
	}
	if !strings.Contains(string(out), `"valueBoolean":true`) {
		t.Errorf("the last setter did not win: %s", out)
	}
}

func TestChoiceSetterClearsThePrimitiveCompanion(t *testing.T) {
	// A primitive variant has a _field companion carrying its extensions. Clearing
	// the value but leaving the companion would strand a _valueString beside a
	// valueQuantity — a member belonging to a variant that is no longer there.
	b := r4.NewObservationBuilder()
	b.SetValueString("a")
	built := b.Build()
	built.ValueStringExt = &r4.Element{Id: r4.Ptr("e1")}

	// Setting another variant on the same builder must take the companion with it.
	b.SetValueQuantity(r4.Quantity{})
	out, err := json.Marshal(b.Build())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), "_valueString") {
		t.Errorf("the companion outlived its variant: %s", out)
	}
	if !strings.Contains(string(out), "valueQuantity") {
		t.Errorf("the new variant is missing: %s", out)
	}
}

func TestChoiceExclusivityLeavesOtherFieldsAlone(t *testing.T) {
	// Clearing a choice group must not touch anything outside it.
	o := r4.NewObservationBuilder().
		SetId("o1").
		SetValueString("x").
		Build()

	if o.Id == nil || *o.Id != "o1" {
		t.Error("setting a choice cleared an unrelated field")
	}
	if o.ValueString == nil || *o.ValueString != "x" {
		t.Error("the variant that was set did not survive")
	}
}

func TestChoiceExclusivityInEveryVersion(t *testing.T) {
	t.Run("r4b", func(t *testing.T) {
		o := r4b.NewObservationBuilder().SetValueString("a").SetValueBoolean(true).Build()
		out, err := json.Marshal(o)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if strings.Contains(string(out), "valueString") {
			t.Errorf("both variants survived: %s", out)
		}
	})
	t.Run("r5", func(t *testing.T) {
		o := r5.NewObservationBuilder().SetValueString("a").SetValueBoolean(true).Build()
		out, err := json.Marshal(o)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if strings.Contains(string(out), "valueString") {
			t.Errorf("both variants survived: %s", out)
		}
	})
}

// TestStructLiteralCanStillHoldTwoVariants measures the limit of this, so the
// guarantee is not read as broader than it is.
//
// Builders exist only for the top level of a resource — not for backbone elements
// and not for datatypes — so of r4's 186 choice groups, 61 are reachable through a
// builder and protected here. Extension.value[x], with 71 variants, is a datatype
// and has no builder at all.
//
// Enforcing this on the struct would mean a MarshalJSON on every resource, which
// is the cost task 6.2 removed by deleting 437 of them.
func TestStructLiteralCanStillHoldTwoVariants(t *testing.T) {
	o := r4.Observation{
		ValueString:  r4.Ptr("a"),
		ValueBoolean: r4.Ptr(true),
	}
	out, err := json.Marshal(&o)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), "valueString") || !strings.Contains(string(out), "valueBoolean") {
		t.Fatal("a struct literal no longer emits two variants — if that was fixed, delete this test")
	}
	t.Logf("a struct literal still produces %s; the builder is what guards this", out)
}

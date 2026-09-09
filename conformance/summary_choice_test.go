package conformance

// summary.go had no entry for any choice element. analyzeChoiceType built the
// variant properties without copying isSummary off the element, so Observation
// appeared with neither value nor effective although the spec marks
// Observation.value[x] and Observation.effective[x] as summary elements.

import (
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestSummaryIncludesChoiceElements(t *testing.T) {
	// Each concrete variant, not the "[x]" form: a document holds one variant and
	// that is the name a caller filters on.
	for _, field := range []string{
		"valueQuantity", "valueString", "valueCodeableConcept",
		"effectiveDateTime", "effectivePeriod",
	} {
		if !r4.IsSummaryField("Observation", field) {
			t.Errorf("r4: Observation.%s is a summary element in the spec", field)
		}
		if !r4b.IsSummaryField("Observation", field) {
			t.Errorf("r4b: Observation.%s is a summary element in the spec", field)
		}
		if !r5.IsSummaryField("Observation", field) {
			t.Errorf("r5: Observation.%s is a summary element in the spec", field)
		}
	}

	// The non-choice elements it always had are still there.
	for _, field := range []string{"status", "code", "subject"} {
		if !r4.IsSummaryField("Observation", field) {
			t.Errorf("r4: Observation.%s went missing", field)
		}
	}

	// And a variant of a choice that is not a summary element stays out.
	if r4.IsSummaryField("Observation", "bodySite") {
		t.Error("Observation.bodySite is not a summary element")
	}
}

func TestSummaryStopsAtTheResourcesOwnElements(t *testing.T) {
	// Records a boundary rather than guarding an invariant.
	//
	// component is listed, because Observation.component is itself a summary
	// element. What is not listed is anything below it: Observation.component
	// .value[x] is a summary element in the spec and has no entry here, because
	// the names in this table are flat and "component" is the only name at this
	// level. Filtering a nested element means walking into it.
	//
	// Left as is rather than fixed: putting dotted paths in the same slice would
	// change what IsSummaryField's second argument means, and nothing calls these
	// accessors today. The flag itself is correct on the nested properties — see
	// TestSummaryFlagsReachNestedBackboneProperties in internal/codegen/analyzer.
	if !r4.IsSummaryField("Observation", "component") {
		t.Error("Observation.component is a summary element and should be listed")
	}
	for _, nested := range []string{
		"component.valueQuantity",
		"componentValueQuantity",
	} {
		if r4.IsSummaryField("Observation", nested) {
			t.Errorf("%q is listed; the table's scope changed and this test is stale", nested)
		}
	}
}

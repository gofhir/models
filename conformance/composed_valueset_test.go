package conformance

// A ValueSet may be built by including other ValueSets instead of listing codes.
// The parser ignored that, so a composed set resolved to whichever part happened
// to be written inline. R5's version-independent-all-resource-types came out as
// 41 retired type names — BodySite, Conformance, DataElement — with no Patient
// among them, and it types four real fields.
//
// Nothing caught it because the wire format never changes: these are string
// constants, so the corpus round-trips identically whether the enum is right or
// wrong. Only a reader looking for a constant would find out.

import (
	"testing"

	"github.com/gofhir/models/r5/v2"
)

func TestComposedValueSetsResolveThroughTheirIncludes(t *testing.T) {
	// The constant that did not exist. GraphDefinition.node.type is bound to this
	// value set as required, so "Patient" is a legal value the type had no name
	// for.
	if got := string(r5.VersionIndependentResourceTypesAllPatient); got != "Patient" {
		t.Errorf("got %q, want Patient", got)
	}

	// The retired names are still in it: the set is the union of the resource
	// types and the old ones, and dropping half of it was the bug.
	for _, want := range []r5.VersionIndependentResourceTypesAll{
		r5.VersionIndependentResourceTypesAllPatient,
		r5.VersionIndependentResourceTypesAllObservation,
		r5.VersionIndependentResourceTypesAllBodysite,
	} {
		if want.Display() == "" && want.System() == "" {
			t.Errorf("%q has neither display nor system; it is not in the table", want)
		}
	}

	// Size, so a partial resolution cannot pass. 203 is the union: 162 current
	// types plus 41 retired ones.
	if n := len(r5.VersionIndependentResourceTypesAllValues()); n < 200 {
		t.Errorf("got %d codes, want the full composed set of 203", n)
	}
}

func TestRequiredBindingsAreTypedWhateverTheirSize(t *testing.T) {
	// A required binding is the specification saying these are the only legal
	// values. A cap on how many constants are worth emitting should not decide
	// whether the caller gets help there.
	//
	// spdx-license is the largest at 346 codes; before this it was over the cap
	// and ImplementationGuide.license was a bare string.
	if n := len(r5.SPDXLicenseValues()); n < 300 {
		t.Errorf("got %d SPDX licenses, want 346", n)
	}
	if n := len(r5.ResourceTypeValues()); n < 150 {
		t.Errorf("got %d resource types, want 158", n)
	}

	// The enormous value sets — mimetypes, currencies, all of UCUM — stay out.
	// They are bound as example or preferred, never as required on a code
	// element, so they were never candidates and the cap was not what excluded
	// them. Their fields are still plain strings.
	var mime *string
	_ = mime
}

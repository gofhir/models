package conformance

// Finding an extension by URL is the most common thing anyone does with FHIR
// extensions — the slice order carries no meaning, so the URL is the only handle.
// Until now the library offered nothing for it and every consumer wrote the same
// loop, including the nil check on Url that is easy to forget.

import (
	"encoding/json"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

const (
	raceURL      = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-race"
	ethnicityURL = "http://hl7.org/fhir/us/core/StructureDefinition/us-core-ethnicity"
	absentURL    = "http://example.org/not-here"
)

func patientWithExtensions() *r4.Patient {
	return r4.NewPatientBuilder().
		SetId("p1").
		AddExtension(r4.NewExtensionBuilder().SetUrl(raceURL).SetValueString("2106-3").Build()).
		AddExtension(r4.NewExtensionBuilder().SetUrl(ethnicityURL).SetValueString("2186-5").Build()).
		AddExtension(r4.NewExtensionBuilder().SetUrl(raceURL).SetValueString("2054-5").Build()).
		Build()
}

func TestExtensionLookupByURL(t *testing.T) {
	p := patientWithExtensions()

	got := p.GetExtensionByURL(raceURL)
	if got == nil {
		t.Fatal("the race extension was not found")
	}
	if got.ValueString == nil || *got.ValueString != "2106-3" {
		t.Errorf("got the wrong one: %v", got.ValueString)
	}

	if p.GetExtensionByURL(absentURL) != nil {
		t.Error("a URL that is not present returned something")
	}
	if !p.HasExtensionByURL(ethnicityURL) || p.HasExtensionByURL(absentURL) {
		t.Error("HasExtensionByURL disagrees with GetExtensionByURL")
	}
}

func TestExtensionLookupReturnsEveryMatch(t *testing.T) {
	// A URL repeats where the extension's cardinality allows it, so returning only
	// the first would quietly drop data.
	p := patientWithExtensions()

	all := p.GetExtensionsByURL(raceURL)
	if len(all) != 2 {
		t.Fatalf("got %d extensions for a URL that appears twice, want 2", len(all))
	}
	if all[0].ValueString == nil || *all[0].ValueString != "2106-3" {
		t.Error("the first match is wrong")
	}
	if all[1].ValueString == nil || *all[1].ValueString != "2054-5" {
		t.Error("the second match is wrong")
	}
	if len(p.GetExtensionsByURL(absentURL)) != 0 {
		t.Error("a URL that is not present returned matches")
	}

	// Pointers, like the singular. Values would make the plural silently
	// read-only: a loop assigning through the results would compile, run and
	// change nothing.
	all[1].ValueString = r4.Ptr("edited")
	if again := p.GetExtensionsByURL(raceURL); again[1].ValueString == nil ||
		*again[1].ValueString != "edited" {
		t.Error("writing through a result of the plural did not reach the resource")
	}
}

func TestExtensionLookupPointsIntoTheSlice(t *testing.T) {
	// Documented behavior: the result points into the resource, so writing
	// through it edits the resource rather than a copy. That is what makes it
	// useful for changing an extension in place.
	p := patientWithExtensions()

	found := p.GetExtensionByURL(ethnicityURL)
	found.ValueString = r4.Ptr("changed")

	again := p.GetExtensionByURL(ethnicityURL)
	if again.ValueString == nil || *again.ValueString != "changed" {
		t.Error("writing through the result did not reach the resource")
	}
}

func TestExtensionLookupSurvivesAMissingURL(t *testing.T) {
	// Url is a pointer, so an extension with no URL at all is representable — and
	// a hand-written loop that forgets the nil check panics on it.
	p := r4.Patient{Extension: []r4.Extension{
		{ValueString: r4.Ptr("no url")},
		{Url: r4.Ptr(raceURL), ValueString: r4.Ptr("2106-3")},
	}}

	got := p.GetExtensionByURL(raceURL)
	if got == nil {
		t.Fatal("an extension without a URL stopped the search")
	}
	if got.ValueString == nil || *got.ValueString != "2106-3" {
		t.Error("wrong extension")
	}
}

func TestModifierExtensionLookupIsSeparate(t *testing.T) {
	// A modifier extension changes the meaning of the element it is on, so a
	// reader that does not recognize one must not process the element. Folding the
	// two searches together would hide that distinction.
	p := r4.Patient{
		Extension:         []r4.Extension{{Url: r4.Ptr(raceURL)}},
		ModifierExtension: []r4.Extension{{Url: r4.Ptr(ethnicityURL)}},
	}

	if p.GetExtensionByURL(ethnicityURL) != nil {
		t.Error("a modifier extension was returned by the ordinary lookup")
	}
	if p.GetModifierExtensionByURL(raceURL) != nil {
		t.Error("an ordinary extension was returned by the modifier lookup")
	}
	if p.GetModifierExtensionByURL(ethnicityURL) == nil {
		t.Error("the modifier extension was not found")
	}
}

func TestExtensionLookupReachesEveryLevel(t *testing.T) {
	// Extensions are not only on resources. They appear on datatypes, on backbone
	// elements, and nested inside other extensions — a complex extension is
	// expressed as sub-extensions rather than a value.
	doc := `{"resourceType":"Patient","name":[{"family":"Smith",
		"extension":[{"url":"` + raceURL + `","valueString":"on a datatype"}]}],
		"contact":[{"extension":[{"url":"` + raceURL + `","valueString":"on a backbone"}]}],
		"extension":[{"url":"` + ethnicityURL + `","extension":[
			{"url":"` + raceURL + `","valueString":"nested"}]}]}`

	var p r4.Patient
	if err := json.Unmarshal([]byte(doc), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got := p.Name[0].GetExtensionByURL(raceURL); got == nil {
		t.Error("a datatype has no lookup")
	} else if *got.ValueString != "on a datatype" {
		t.Errorf("datatype: %v", got.ValueString)
	}

	if got := p.Contact[0].GetExtensionByURL(raceURL); got == nil {
		t.Error("a backbone element has no lookup")
	} else if *got.ValueString != "on a backbone" {
		t.Errorf("backbone: %v", got.ValueString)
	}

	outer := p.GetExtensionByURL(ethnicityURL)
	if outer == nil {
		t.Fatal("the complex extension was not found")
	}
	if got := outer.GetExtensionByURL(raceURL); got == nil {
		t.Error("an extension cannot search its own sub-extensions")
	} else if *got.ValueString != "nested" {
		t.Errorf("nested: %v", got.ValueString)
	}
}

func TestExtensionHelpersWorkOnABareSlice(t *testing.T) {
	// The methods are generated from these, and they are exported so that code
	// holding a []Extension from somewhere else can use them too.
	exts := []r4.Extension{
		{Url: r4.Ptr(raceURL), ValueString: r4.Ptr("a")},
		{Url: r4.Ptr(raceURL), ValueString: r4.Ptr("b")},
	}
	if got := r4.ExtensionByURL(exts, raceURL); got == nil || *got.ValueString != "a" {
		t.Error("ExtensionByURL")
	}
	if got := r4.ExtensionsByURL(exts, raceURL); len(got) != 2 {
		t.Errorf("ExtensionsByURL returned %d, want 2", len(got))
	}
	if !r4.HasExtensionByURL(exts, raceURL) || r4.HasExtensionByURL(exts, absentURL) {
		t.Error("HasExtensionByURL")
	}
	if r4.ExtensionByURL(nil, raceURL) != nil {
		t.Error("a nil slice should find nothing rather than panic")
	}
}

func TestExtensionLookupInEveryVersion(t *testing.T) {
	t.Run("r4b", func(t *testing.T) {
		p := r4b.Patient{Extension: []r4b.Extension{{Url: r4b.Ptr(raceURL)}}}
		if p.GetExtensionByURL(raceURL) == nil {
			t.Error("not found")
		}
	})
	t.Run("r5", func(t *testing.T) {
		p := r5.Patient{Extension: []r5.Extension{{Url: r5.Ptr(raceURL)}}}
		if p.GetExtensionByURL(raceURL) == nil {
			t.Error("not found")
		}
	})
}

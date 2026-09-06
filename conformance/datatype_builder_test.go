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
		AddName(r4.NewHumanNameBuilder().
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
		AddCoding(r4.NewCodingBuilder().
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
		SetLow(r4.NewQuantityBuilder().SetValue(*r4.MustDecimal("1")).Build()).
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

// TestDatatypeBuildReturnsAValue pins the shape of the API, which is the thing
// that cannot be changed later without breaking callers.
//
// Build returns T for a datatype and *T for a resource. Every consumer of a
// datatype takes a value — Patient.Name is []HumanName, and a pointer field like
// Range.Low is set through SetLow(Quantity), which takes the address itself — so
// returning a pointer forced a dereference at every call site to undo a pointer
// nobody asked for. Resources stay pointers: they are handled as such and they
// implement the Resource interface on the pointer receiver.
func TestDatatypeBuildReturnsAValue(t *testing.T) {
	// The parameter types are the assertion: this call compiles only if Build
	// returns a HumanName value and a *Patient pointer. Stated as a signature
	// rather than as typed declarations, which staticcheck reads as redundant.
	assertReturnTypes := func(r4.HumanName, *r4.Patient) {}
	assertReturnTypes(r4.NewHumanNameBuilder().Build(), r4.NewPatientBuilder().Build())

	name := r4.NewHumanNameBuilder().SetFamily("Smith").Build()
	if name.Family == nil || *name.Family != "Smith" {
		t.Error("family did not survive Build")
	}
	p := r4.NewPatientBuilder().SetId("p1").Build()
	if p.Id == nil || *p.Id != "p1" {
		t.Error("id did not survive Build")
	}

	// The point of it: no dereference anywhere.
	built := r4.NewPatientBuilder().
		AddName(r4.NewHumanNameBuilder().SetFamily("Smith").Build()).
		Build()
	if len(built.Name) != 1 || built.Name[0].Family == nil {
		t.Error("the chain did not carry the name through")
	}

	// Build returns a copy, so replacing a field on the result does not reach the
	// builder. The copy is shallow, which is what a struct assignment does in Go
	// and not something the builder changes: the slices and pointers inside are
	// still shared. TestBuiltValuesShareTheirSlices states that explicitly, so
	// "a copy" is not read as more than it is.
	b := r4.NewHumanNameBuilder().SetFamily("Original")
	first := b.Build()
	first.Family = r4.Ptr("Mutated")
	if first.Family == nil || *first.Family != "Mutated" {
		t.Fatal("the copy was not writable")
	}
	if second := b.Build(); second.Family == nil || *second.Family != "Original" {
		t.Errorf("replacing a field on a built value reached back into the builder: %v", second.Family)
	}
}

// TestBuiltValuesShareTheirSlices records the limit of "returns a copy".
//
// Copying a struct copies the slice headers inside it, so two values built from
// one builder point at the same backing array. That is Go's semantics for `a := b`
// and not something the builder introduces; deep-copying instead would be
// surprising, and expensive on every Build.
//
// It is recorded because "Build returns a copy" reads as more of a guarantee than
// it is, and because reusing one builder for two values is the case where it bites.
func TestBuiltValuesShareTheirSlices(t *testing.T) {
	b := r4.NewHumanNameBuilder().SetFamily("Smith").AddGiven("John")
	first, second := b.Build(), b.Build()

	*first.Given[0] = "Mutated"
	if second.Given[0] == nil || *second.Given[0] != "Mutated" {
		t.Fatal("the slices are no longer shared — if Build now deep-copies, delete this test")
	}
	t.Log("two values built from one builder share their slices, as struct assignment does")
}

// TestResourceBuilderReturnsTheSameInstance is the resource-side counterpart.
//
// Build hands back the pointer it has been filling, so calling it twice yields the
// same object rather than two resources. A builder is meant to be used once; this
// is what happens if it is not.
func TestResourceBuilderReturnsTheSameInstance(t *testing.T) {
	b := r4.NewPatientBuilder().SetId("p1")
	first, second := b.Build(), b.Build()

	if first != second {
		t.Fatal("Build now returns distinct resources — if that changed deliberately, delete this test")
	}
	first.Id = r4.Ptr("changed")
	if second.Id == nil || *second.Id != "changed" {
		t.Error("the two are no longer the same object")
	}
}

// TestTheChainNeverBreaks is the point of completing the builder surface.
//
// A caller building a resource used to drop into a struct literal the moment they
// reached a backbone or a datatype, and again for any extension on a primitive.
// Those are not corner cases: a backbone is where most of a resource's data lives,
// and a _field companion is the only way FHIR expresses an extension on a
// primitive value.
//
// This builds one document through every layer without a single literal.
func TestTheChainNeverBreaks(t *testing.T) {
	p := r4.NewPatientBuilder().
		SetId("p1").
		// datatype
		AddName(r4.NewHumanNameBuilder().
			SetFamily("Smith").
			AddGiven("John").
			Build()).
		// backbone, with a datatype inside it
		AddContact(r4.NewPatientContactBuilder().
			SetName(r4.NewHumanNameBuilder().SetFamily("Doe").Build()).
			AddTelecom(r4.NewContactPointBuilder().
				SetSystem(r4.ContactPointSystemPhone).
				SetValue("555").
				Build()).
			Build()).
		// extension on a primitive, through its companion
		SetBirthDateExt(r4.NewElementBuilder().
			AddExtension(r4.NewExtensionBuilder().
				SetUrl("http://hl7.org/fhir/StructureDefinition/data-absent-reason").
				SetValueCode("asked-declined").
				Build()).
			Build()).
		Build()

	out, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{
		`"family":"Smith"`, `"given":["John"]`,
		`"contact":[{`, `"family":"Doe"`, `"value":"555"`,
		`"_birthDate":{`, `"asked-declined"`,
	} {
		if !strings.Contains(string(out), want) {
			t.Errorf("%s missing from:\n%s", want, out)
		}
	}

	// And it reads back as what was built.
	var back r4.Patient
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if len(back.Contact) != 1 || back.Contact[0].Name == nil {
		t.Error("the backbone did not survive the round trip")
	}
	if back.BirthDateExt == nil || len(back.BirthDateExt.Extension) != 1 {
		t.Error("the primitive's extension did not survive the round trip")
	}
	if back.BirthDate != nil {
		t.Error("birthDate has no value; only its extension was set")
	}
}

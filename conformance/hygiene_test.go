package conformance

// Three items the plan filed as "minor hygiene". Measuring them changed what each
// one turned out to be: one was not a defect, one was worse than described, and one
// had already been closed by unrelated work.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
)

func TestBuilderIgnoresANilContainedResource(t *testing.T) {
	// Contained holds an interface, so a nil entry marshals as JSON null — which
	// FHIR does not allow there, and which this library then drops on the way back
	// in. A document written with one re-reads shorter than it was written.
	org := &r4.Organization{Id: r4.Ptr("o1")}

	p := r4.NewPatientBuilder().
		AddContained(org).
		AddContained(nil).
		AddContained(org).
		Build()

	if len(p.Contained) != 2 {
		t.Fatalf("got %d contained, want 2 — the nil should not have been appended", len(p.Contained))
	}

	out, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), "null") {
		t.Errorf("a null reached the output: %s", out)
	}

	// And what it writes, it can read back unchanged.
	var back r4.Patient
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if len(back.Contained) != len(p.Contained) {
		t.Errorf("round trip changed the length: wrote %d, read %d",
			len(p.Contained), len(back.Contained))
	}
}

// TestStructLiteralWithNilContainedIsStillLossy measures a limitation rather than
// guarding an invariant.
//
// Filtering at serialization time would fix this everywhere, but giving
// ContainedList a MarshalJSON costs 2.9x on every marshal that contains no nil at
// all — 419 to 1209 ns/op, 1 to 3 allocs — because encoding/json re-compacts
// whatever a MarshalJSON returns. That is precisely the cost task 6.2 removed by
// deleting 437 of them, and paying it on every document to guard against a value
// the caller had to put there deliberately is not a good trade.
//
// So the builder is guarded and a struct literal is not. This test states that
// plainly, so the behavior is a known limitation rather than a surprise.
func TestStructLiteralWithNilContainedIsStillLossy(t *testing.T) {
	org := &r4.Organization{Id: r4.Ptr("o1")}
	p := r4.Patient{Contained: r4.ContainedList{org, nil, org}}

	out, err := json.Marshal(&p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), "null") {
		t.Fatal("a nil in the slice no longer emits null — if that was fixed, delete this test")
	}

	var back r4.Patient
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if len(back.Contained) != 2 {
		t.Errorf("got %d after the round trip, want 2", len(back.Contained))
	}
	t.Logf("wrote %d contained, read back %d — use the builder, which drops nils",
		len(p.Contained), len(back.Contained))
}

func TestDuplicateResourceTypeIsRefused(t *testing.T) {
	// Duplicate keys are valid JSON but ambiguous, and encoding/json takes the
	// last while a scanner might take the first. The plan listed this as a parser
	// differential; the resourceType check added for the major closes it, because
	// the marker sees every occurrence and refuses as soon as two disagree.
	//
	// What matters is that an ambiguous document does not decode to one arbitrary
	// reading of it.
	const dup = `{"resourceType":"Patient","resourceType":"Observation","id":"x"}`

	if _, err := r4.UnmarshalResource([]byte(dup)); err == nil {
		t.Error("a document declaring two different resource types was accepted")
	}

	var p r4.Patient
	if err := json.Unmarshal([]byte(dup), &p); err == nil {
		t.Error("decoding into a concrete type accepted two different resource types")
	}

	// The same key twice with the same value is not ambiguous, and must still work.
	const same = `{"resourceType":"Patient","resourceType":"Patient","id":"x"}`
	var ok r4.Patient
	if err := json.Unmarshal([]byte(same), &ok); err != nil {
		t.Errorf("a repeated but consistent resourceType was refused: %v", err)
	} else if ok.Id == nil || *ok.Id != "x" {
		t.Error("the document decoded but lost its id")
	}
}

// TestResourceIdTakesNoExtension records why there is no _id field, which reads
// like an omission until you check the specification.
//
// Resource.id and Element.id are declared as http://hl7.org/fhirpath/System.String
// — a System primitive, not a FHIR one. They have no Element structure behind them
// and therefore cannot carry extensions, so there is no _id companion to generate.
// The plan listed the missing field as a defect; it is not.
func TestResourceIdTakesNoExtension(t *testing.T) {
	const doc = `{"resourceType":"Patient","id":"p1","_id":{"extension":[
		{"url":"http://example.org/x","valueCode":"c"}]}}`

	var p r4.Patient
	if err := json.Unmarshal([]byte(doc), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.Id == nil || *p.Id != "p1" {
		t.Fatalf("id did not decode: %v", p.Id)
	}

	// _id is ignored, the way encoding/json ignores any member with no field.
	out, err := json.Marshal(&p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), "_id") {
		t.Errorf("an _id was emitted, which the specification does not define: %s", out)
	}

	// The contrast: birthDate is a FHIR primitive and does carry one.
	const withExt = `{"resourceType":"Patient","_birthDate":{"extension":[
		{"url":"http://example.org/x","valueCode":"c"}]}}`
	var q r4.Patient
	if err := json.Unmarshal([]byte(withExt), &q); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if q.BirthDateExt == nil {
		t.Error("a FHIR primitive lost its extension companion")
	}
}

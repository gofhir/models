package conformance

// resourceType is emitted by a zero-size marker type rather than by a
// per-resource MarshalJSON that overwrote a string field on every call.
//
// The corpus already proves the output is unchanged, since 8751 examples still
// round-trip. What it cannot show is the behavior around the edges of that
// change, which is what these cover: that a wrong or missing value on the way in
// is harmless, that the escaping rules did not move, and that a struct embedding
// a resource no longer loses its own fields.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestResourceTypeIsEmittedWithoutBeingSet(t *testing.T) {
	// No ResourceType field is set anywhere here — the zero value has to be
	// enough, or every struct literal in user code would need a magic string.
	for _, tt := range []struct {
		name string
		res  any
		want string
	}{
		{"r4 patient", &r4.Patient{Id: r4.Ptr("p1")}, `"resourceType":"Patient"`},
		{"r4 observation", &r4.Observation{Id: r4.Ptr("o1")}, `"resourceType":"Observation"`},
		{"r5 bundle", &r5.Bundle{}, `"resourceType":"Bundle"`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := json.Marshal(tt.res)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if !strings.Contains(string(out), tt.want) {
				t.Errorf("missing %s in %s", tt.want, out)
			}
			// It must also come first, as FHIR documents conventionally have it.
			if !strings.HasPrefix(string(out), "{"+tt.want) {
				t.Errorf("resourceType is not the first key: %s", out)
			}
		})
	}
}

func TestResourceTypeOnInputIsHarmless(t *testing.T) {
	// The Go type fixes the resource type, so nothing the document says about it
	// can change the outcome. A wrong value must not fail here either —
	// UnmarshalResource is the layer that validates it during dispatch, and
	// json.Unmarshal into a concrete type is a deliberate choice by the caller.
	for _, in := range []string{
		`{"resourceType":"Patient","id":"x"}`,
		`{"resourceType":"Observation","id":"x"}`,
		`{"resourceType":null,"id":"x"}`,
		`{"resourceType":"","id":"x"}`,
		`{"id":"x"}`,
	} {
		var p r4.Patient
		if err := json.Unmarshal([]byte(in), &p); err != nil {
			t.Errorf("%s: %v", in, err)
			continue
		}
		if p.Id == nil || *p.Id != "x" {
			t.Errorf("%s: id did not decode", in)
		}
		if got := p.GetResourceType(); got != "Patient" {
			t.Errorf("%s: GetResourceType() = %q", in, got)
		}
		// And it round-trips back to the correct type regardless of what came in.
		out, err := json.Marshal(&p)
		if err != nil {
			t.Fatalf("%s: re-marshal: %v", in, err)
		}
		if !strings.Contains(string(out), `"resourceType":"Patient"`) {
			t.Errorf("%s: re-marshaled as %s", in, out)
		}
	}
}

// TestEmbeddedResourceKeepsOuterFields is the regression test for what removing
// the per-resource MarshalJSON fixed.
//
// A method promoted from an embedded struct satisfies json.Marshaler for the
// outer type, so encoding/json called Patient's MarshalJSON and serialized only
// the Patient — every field the user added alongside it vanished from the
// output, with no error.
func TestEmbeddedResourceKeepsOuterFields(t *testing.T) {
	type withMetadata struct {
		r4.Patient
		Tenant string `json:"tenant"`
		Score  int    `json:"score"`
	}

	v := withMetadata{
		Patient: r4.Patient{Id: r4.Ptr("p1")},
		Tenant:  "acme",
		Score:   42,
	}

	for _, tc := range []struct {
		name    string
		marshal func(any) ([]byte, error)
	}{
		{"json.Marshal", json.Marshal},
		{"r4.Marshal", r4.Marshal},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, err := tc.marshal(v)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			var got map[string]any
			if err := json.Unmarshal(out, &got); err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			for _, key := range []string{"resourceType", "id", "tenant", "score"} {
				if _, ok := got[key]; !ok {
					t.Errorf("%q missing from embedded output: %s", key, out)
				}
			}
		})
	}
}

// TestEscapingIsUnchanged pins the split between the two entry points. The
// removed MarshalJSON called SetEscapeHTML(false), which never had any effect —
// encoding/json re-compacts a MarshalJSON result using the *outer* encoder's
// setting. Marshal() is where the decision is actually made, and it still holds.
func TestEscapingIsUnchanged(t *testing.T) {
	p := &r4.Patient{Id: r4.Ptr("a<b>c&d")}

	viaLib, err := r4.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(viaLib), "a<b>c&d") {
		t.Errorf("Marshal escaped HTML, which would corrupt narrative XHTML: %s", viaLib)
	}

	viaStd, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(viaStd), "a<b>c&d") {
		t.Errorf("json.Marshal stopped escaping; the docs tell users to prefer Marshal precisely because it does: %s", viaStd)
	}
}

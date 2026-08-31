package conformance

// Required complex fields are pointers.
//
// They used to be value structs, on the reasoning that the Go type should
// express the obligation. It could not: Go has no way to force a field to be
// set, so an incomplete resource marshaled with an empty object in place of the
// required element — `"code":{}` — which violates ele-1 and no validator
// accepts. The obligation belongs to a FHIR validator; what the type has to be
// able to say is "absent".
//
// The corpus cannot catch this, which is why these exist: every published
// example fills those fields in, so the defect only appears in resources built
// by hand, which is most of what library users do.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestIncompleteResourcesOmitRequiredComplexFields(t *testing.T) {
	for _, tt := range []struct {
		name   string
		value  any
		want   string
		absent string
	}{
		{
			name:   "observation without code",
			value:  &r4.Observation{Id: r4.Ptr("o1")},
			want:   `{"resourceType":"Observation","id":"o1"}`,
			absent: `"code":{}`,
		},
		{
			name:   "empty extension",
			value:  &r4.Extension{},
			want:   `{}`,
			absent: `"url":""`,
		},
		{
			name:   "r5 observation without code",
			value:  &r5.Observation{Id: r5.Ptr("o1")},
			want:   `{"resourceType":"Observation","id":"o1"}`,
			absent: `"code":{}`,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			out, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if got := string(out); got != tt.want {
				t.Errorf("got  %s\nwant %s", got, tt.want)
			}
			if strings.Contains(string(out), tt.absent) {
				t.Errorf("still emitting %s, which is invalid FHIR: %s", tt.absent, out)
			}
		})
	}
}

// TestPopulatedRequiredFieldsStillSerialize is the other half: making the fields
// optional in Go must not make them disappear when they are actually set.
func TestPopulatedRequiredFieldsStillSerialize(t *testing.T) {
	obs := &r4.Observation{
		Id:     r4.Ptr("o1"),
		Status: r4.Ptr(r4.ObservationStatusFinal),
		Code: &r4.CodeableConcept{
			Coding: []r4.Coding{{
				System: r4.Ptr("http://loinc.org"),
				Code:   r4.Ptr("8867-4"),
			}},
		},
	}

	out, err := json.Marshal(obs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"status":"final"`, `"8867-4"`, `"http://loinc.org"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("%s missing from %s", want, out)
		}
	}

	// And it survives a round-trip with the field still reachable.
	var back r4.Observation
	if err := json.Unmarshal(out, &back); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if back.Code == nil || len(back.Code.Coding) != 1 || back.Code.Coding[0].Code == nil {
		t.Fatalf("code did not survive: %+v", back.Code)
	}
	if got := *back.Code.Coding[0].Code; got != "8867-4" {
		t.Errorf("code = %q", got)
	}
}

// TestBuilderSignaturesUnchanged records what this change does *not* break.
//
// The builder takes the value and takes its address internally, so making the
// field a pointer left every Set* signature exactly as it was. That is the
// difference between a migration that is "add & to your struct literals" and one
// that touches every builder call site too.
func TestBuilderSignaturesUnchanged(t *testing.T) {
	obs := r4.NewObservationBuilder().
		SetId("o1").
		SetStatus(r4.ObservationStatusFinal).
		SetCode(r4.CodeableConcept{ // by value, as before
			Coding: []r4.Coding{{Code: r4.Ptr("8867-4")}},
		}).
		Build()

	if obs.Code == nil {
		t.Fatal("builder did not set code")
	}
	out, err := json.Marshal(obs)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"8867-4"`) {
		t.Errorf("code missing: %s", out)
	}
}

// TestExtensionURLRoundTripsInXML covers the one hand-written special case the
// change touched: Extension.url is an XML attribute, written by a template branch
// that assumed a value field.
func TestExtensionURLRoundTripsInXML(t *testing.T) {
	patient := &r4.Patient{
		Id: r4.Ptr("p1"),
		Extension: []r4.Extension{{
			Url:         r4.Ptr("http://example.org/ext"),
			ValueString: r4.Ptr("hello"),
		}},
	}

	out, err := r4.MarshalResourceXML(patient)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `url="http://example.org/ext"`) {
		t.Errorf("url attribute missing: %s", out)
	}

	back, err := r4.UnmarshalResourceXML(out)
	if err != nil {
		t.Fatalf("re-parse: %v", err)
	}
	ext := back.(*r4.Patient).Extension
	if len(ext) != 1 || ext[0].Url == nil {
		t.Fatalf("extension lost: %+v", ext)
	}
	if got := *ext[0].Url; got != "http://example.org/ext" {
		t.Errorf("url = %q", got)
	}

	// An extension with no url must not invent one.
	empty := &r4.Patient{Extension: []r4.Extension{{ValueString: r4.Ptr("x")}}}
	out2, err := r4.MarshalResourceXML(empty)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out2), `url=""`) {
		t.Errorf("emitted an empty url attribute: %s", out2)
	}
}

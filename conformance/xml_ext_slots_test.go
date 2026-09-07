package conformance

// Reading XML used to invent extension slots. A plain <given value="Maria"/> has
// no extension, but the decoder appended to the companion slice unconditionally,
// so the resource came back holding []*Element{nil} — which JSON then wrote out as
// "_given":[null], a member the document never had.
//
// The corpus could not see it. TestRoundTrip compares XML against XML and JSON
// against JSON; the spurious slot only shows when a document read as one is
// written as the other, which nothing exercised. It surfaced from building the
// same Patient nine different ways and finding that one of the nine disagreed.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestXMLDoesNotInventExtensionSlots(t *testing.T) {
	const doc = `<?xml version="1.0"?><Patient xmlns="http://hl7.org/fhir"><name>` +
		`<given value="A"/><given value="B"/></name></Patient>`

	res, err := r4.UnmarshalResourceXML([]byte(doc))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	p, ok := res.(*r4.Patient)
	if !ok {
		t.Fatalf("got %T", res)
	}
	if len(p.Name[0].GivenExt) != 0 {
		t.Errorf("%d extension slots for values that have no extension: %v",
			len(p.Name[0].GivenExt), p.Name[0].GivenExt)
	}

	out, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), "_given") {
		t.Errorf("a member the document never had reached the output: %s", out)
	}
}

func TestXMLKeepsExtensionSlotsInPosition(t *testing.T) {
	// The companion slice is parallel by position, so an extension on the second
	// value needs a nil in front of it. Suppressing the empty slots must not cost
	// that alignment.
	const doc = `<?xml version="1.0"?><Patient xmlns="http://hl7.org/fhir"><name>` +
		`<given value="A"/>` +
		`<given value="B"><extension url="http://x"><valueCode value="c"/></extension></given>` +
		`<given value="C"/>` +
		`</name></Patient>`

	res, err := r4.UnmarshalResourceXML([]byte(doc))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	p := res.(*r4.Patient)

	if len(p.Name[0].Given) != 3 {
		t.Fatalf("got %d given, want 3", len(p.Name[0].Given))
	}
	// Two slots, not three: the trailing value has no extension and needs none.
	if len(p.Name[0].GivenExt) != 2 {
		t.Fatalf("got %d slots, want 2", len(p.Name[0].GivenExt))
	}
	if p.Name[0].GivenExt[0] != nil {
		t.Error("the first value has no extension, so its slot must be nil")
	}
	if p.Name[0].GivenExt[1] == nil {
		t.Fatal("the extension did not land on the value that carried it")
	}

	out, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(out), `"_given":[null,{`) {
		t.Errorf("the wire form does not line up: %s", out)
	}
}

func TestTheSamePatientEveryWayAgrees(t *testing.T) {
	// The check that found the defect. A document built through each supported
	// route has to come out identical — otherwise one of the routes is adding or
	// losing something, and which one is not obvious from either alone.
	const want = `{"resourceType":"Patient","id":"p1","active":true,` +
		`"name":[{"family":"Garcia","given":["Maria"]}],"gender":"female"}`

	fromBuilder := r4.NewPatientBuilder().
		SetId("p1").
		SetActive(true).
		SetGender(r4.AdministrativeGenderFemale).
		AddName(r4.NewHumanNameBuilder().SetFamily("Garcia").AddGiven("Maria").Build()).
		Build()

	fromLiteral := &r4.Patient{
		Id:     r4.Ptr("p1"),
		Active: r4.Ptr(true),
		Gender: r4.Ptr(r4.AdministrativeGenderFemale),
		Name:   []r4.HumanName{{Family: r4.Ptr("Garcia"), Given: r4.PtrSlice("Maria")}},
	}

	var fromJSON r4.Patient
	if err := json.Unmarshal([]byte(want), &fromJSON); err != nil {
		t.Fatalf("json: %v", err)
	}

	fromXMLres, err := r4.UnmarshalResourceXML([]byte(`<?xml version="1.0"?>` +
		`<Patient xmlns="http://hl7.org/fhir"><id value="p1"/><active value="true"/>` +
		`<name><family value="Garcia"/><given value="Maria"/></name>` +
		`<gender value="female"/></Patient>`))
	if err != nil {
		t.Fatalf("xml: %v", err)
	}

	for name, built := range map[string]any{
		"builder": fromBuilder,
		"literal": fromLiteral,
		"json":    &fromJSON,
		"xml":     fromXMLres,
	} {
		out, err := json.Marshal(built)
		if err != nil {
			t.Errorf("%s: marshal: %v", name, err)
			continue
		}
		if string(out) != want {
			t.Errorf("%s disagrees:\n  got  %s\n  want %s", name, out, want)
		}
	}
}

func TestNoSpuriousSlotsInEveryVersion(t *testing.T) {
	const doc = `<?xml version="1.0"?><Patient xmlns="http://hl7.org/fhir"><name>` +
		`<given value="A"/></name></Patient>`

	t.Run("r4b", func(t *testing.T) {
		res, err := r4b.UnmarshalResourceXML([]byte(doc))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		out, err := json.Marshal(res)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if strings.Contains(string(out), "_given") {
			t.Errorf("%s", out)
		}
	})
	t.Run("r5", func(t *testing.T) {
		res, err := r5.UnmarshalResourceXML([]byte(doc))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		out, err := json.Marshal(res)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if strings.Contains(string(out), "_given") {
			t.Errorf("%s", out)
		}
	})
}

package conformance

// A server one version ahead is the ordinary case. An R5 server answering an R4
// client puts resources like InventoryItem in a searchset Bundle, and refusing the
// whole document over one entry meant a single unrecognized type destroyed
// everything around it.
//
// The failure was worse than it looked: Bundle.entry was left holding the entries
// decoded before the error while Unmarshal returned non-nil, so a caller that
// logged the error and carried on had a Bundle silently short some entries — which
// is harder to notice than an outright failure.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

const bundleWithUnknown = `{"resourceType":"Bundle","type":"searchset","entry":[
	{"resource":{"resourceType":"Patient","id":"p1"}},
	{"resource":{"resourceType":"InventoryItem","id":"i1","status":"active","name":[{"nameType":{"text":"brand"}}]}},
	{"resource":{"resourceType":"Patient","id":"p2"}}
]}`

func TestUnknownResourceDoesNotDestroyTheBundle(t *testing.T) {
	var b r4.Bundle
	if err := json.Unmarshal([]byte(bundleWithUnknown), &b); err != nil {
		t.Fatalf("one unrecognized entry failed the whole Bundle: %v", err)
	}
	if len(b.Entry) != 3 {
		t.Fatalf("got %d entries, want 3 — the entries around the unknown one were lost", len(b.Entry))
	}

	// The known entries decode as themselves.
	for _, i := range []int{0, 2} {
		p, ok := b.Entry[i].Resource.(*r4.Patient)
		if !ok {
			t.Errorf("entry[%d] is %T, want *r4.Patient", i, b.Entry[i].Resource)
			continue
		}
		if p.Id == nil {
			t.Errorf("entry[%d] lost its id", i)
		}
	}

	// The unknown one is preserved rather than dropped.
	u, ok := b.Entry[1].Resource.(*r4.UnknownResource)
	if !ok {
		t.Fatalf("entry[1] is %T, want *r4.UnknownResource", b.Entry[1].Resource)
	}
	if u.Type != "InventoryItem" {
		t.Errorf("Type = %q, want InventoryItem", u.Type)
	}
	if u.GetId() == nil || *u.GetId() != "i1" {
		t.Errorf("GetId() = %v, want i1 — id and meta are read because they mean the same in every version", u.GetId())
	}
	if u.GetResourceType() != "InventoryItem" {
		t.Errorf("GetResourceType() = %q", u.GetResourceType())
	}
}

func TestUnknownResourceRoundTripsVerbatim(t *testing.T) {
	// The library cannot interpret the resource, so reformatting it would be
	// inventing a representation for something it does not model. Untouched, the
	// bytes come back exactly as they arrived.
	const doc = `{"resourceType":"Nonesuch","id":"x1","weird":{"nested":[1,2,{"deep":true}]},"zzz":"last"}`

	res, err := r4.UnmarshalResource([]byte(doc))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	u, ok := res.(*r4.UnknownResource)
	if !ok {
		t.Fatalf("got %T, want *r4.UnknownResource", res)
	}

	out, err := json.Marshal(u)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if string(out) != doc {
		t.Errorf("the document was not preserved verbatim:\n  got  %s\n  want %s", out, doc)
	}
}

func TestUnknownResourceKeepsMembersItCannotModel(t *testing.T) {
	// The value of preserving the raw bytes is the members this version has no
	// field for. Before, they were lost along with the whole document.
	var b r4.Bundle
	if err := json.Unmarshal([]byte(bundleWithUnknown), &b); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := json.Marshal(&b)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{`"InventoryItem"`, `"nameType"`, `"brand"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("%s did not survive the round trip:\n%s", want, out)
		}
	}
}

func TestEditingAnUnknownResourceRebuildsIt(t *testing.T) {
	// SetId has to be honored or the Resource interface is a lie, and honoring it
	// costs the original formatting — nothing else.
	const doc = `{"resourceType":"Nonesuch","id":"old","keep":"this"}`

	res, err := r4.UnmarshalResource([]byte(doc))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	res.SetId("new")

	out, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("reparse: %v", err)
	}
	if got["id"] != "new" {
		t.Errorf(`id = %v, want "new"`, got["id"])
	}
	if got["keep"] != "this" {
		t.Error("rebuilding the document dropped a member it does not model")
	}
	if got["resourceType"] != "Nonesuch" {
		t.Errorf("resourceType = %v", got["resourceType"])
	}
}

func TestUnknownResourceWorksInContained(t *testing.T) {
	const doc = `{"resourceType":"Patient","id":"p1","contained":[{"resourceType":"Nonesuch","id":"c1"}]}`

	var p r4.Patient
	if err := json.Unmarshal([]byte(doc), &p); err != nil {
		t.Fatalf("a contained resource of an unknown type failed the parent: %v", err)
	}
	if len(p.Contained) != 1 {
		t.Fatalf("got %d contained, want 1", len(p.Contained))
	}
	if u, ok := p.Contained[0].(*r4.UnknownResource); !ok {
		t.Errorf("contained[0] is %T, want *r4.UnknownResource", p.Contained[0])
	} else if u.Type != "Nonesuch" {
		t.Errorf("Type = %q", u.Type)
	}
}

func TestAMissingResourceTypeIsStillAnError(t *testing.T) {
	// Unknown is not the same as absent. A document with no resourceType is not a
	// resource at all, and accepting it would turn a real defect into a shrug.
	for _, doc := range []string{
		`{"id":"x"}`,
		`{"resourceType":"","id":"x"}`,
		`{}`,
	} {
		if _, err := r4.UnmarshalResource([]byte(doc)); err == nil {
			t.Errorf("%s was accepted; a document with no resourceType is not a resource", doc)
		}
	}
}

func TestNewResourceStillRejectsUnknownTypes(t *testing.T) {
	// NewResource asks for an instance of a named type. Handing back an empty
	// UnknownResource would answer a question nobody asked — the fallback belongs
	// on the path that is decoding someone else's document.
	if _, err := r4.NewResource("Nonesuch"); err == nil {
		t.Error("NewResource invented a type that does not exist")
	}
	if !r4.IsKnownResourceType("Patient") || r4.IsKnownResourceType("Nonesuch") {
		t.Error("IsKnownResourceType should be unaffected by the fallback")
	}
}

func TestUnknownResourceInEveryVersion(t *testing.T) {
	// Each version has its own registry, so what counts as unknown differs — and
	// that is exactly the case this exists for.
	t.Run("r4b", func(t *testing.T) {
		res, err := r4b.UnmarshalResource([]byte(`{"resourceType":"InventoryItem","id":"i1"}`))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if _, ok := res.(*r4b.UnknownResource); !ok {
			t.Errorf("got %T, want *r4b.UnknownResource", res)
		}
	})

	t.Run("r5 knows InventoryItem", func(t *testing.T) {
		// R5 defines it, so it must decode as the real type, not the fallback.
		res, err := r5.UnmarshalResource([]byte(`{"resourceType":"InventoryItem","id":"i1","status":"active"}`))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if _, ok := res.(*r5.UnknownResource); ok {
			t.Fatal("R5 defines InventoryItem and must not fall back")
		}
		if res.GetResourceType() != "InventoryItem" {
			t.Errorf("GetResourceType() = %q", res.GetResourceType())
		}
	})

	t.Run("r5 falls back on a made-up type", func(t *testing.T) {
		res, err := r5.UnmarshalResource([]byte(`{"resourceType":"Nonesuch"}`))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if _, ok := res.(*r5.UnknownResource); !ok {
			t.Errorf("got %T, want *r5.UnknownResource", res)
		}
	})
}

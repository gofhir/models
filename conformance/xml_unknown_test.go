package conformance

// The JSON path kept a resource whose type this version does not define; the XML
// path refused it. So whether a document from a newer server was readable depended
// on which format it arrived in, which is not a distinction anyone would expect.
//
// Closing it needed the element captured verbatim and written back the same way.
// Re-emitting parsed XML re-injects namespace declarations — the defect the
// narrative rewrite existed to fix — so the capture goes through ",innerxml",
// which passes content through untouched rather than re-serializing it.

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

const xmlBundleWithUnknown = `<?xml version="1.0" encoding="UTF-8"?>` +
	`<Bundle xmlns="http://hl7.org/fhir"><type value="searchset"/>` +
	`<entry><resource><Patient><id value="p1"/></Patient></resource></entry>` +
	`<entry><resource><InventoryItem><id value="i1"/><status value="active"/>` +
	`<name><nameType><text value="brand"/></nameType></name></InventoryItem></resource></entry>` +
	`<entry><resource><Patient><id value="p2"/></Patient></resource></entry></Bundle>`

func TestXMLUnknownResourceDoesNotDestroyTheBundle(t *testing.T) {
	res, err := r4.UnmarshalResourceXML([]byte(xmlBundleWithUnknown))
	if err != nil {
		t.Fatalf("one unrecognized entry failed the whole Bundle: %v", err)
	}
	b, ok := res.(*r4.Bundle)
	if !ok {
		t.Fatalf("got %T, want *r4.Bundle", res)
	}
	if len(b.Entry) != 3 {
		t.Fatalf("got %d entries, want 3", len(b.Entry))
	}

	for _, i := range []int{0, 2} {
		if _, isPatient := b.Entry[i].Resource.(*r4.Patient); !isPatient {
			t.Errorf("entry[%d] is %T, want *r4.Patient", i, b.Entry[i].Resource)
		}
	}

	u, ok := b.Entry[1].Resource.(*r4.UnknownResource)
	if !ok {
		t.Fatalf("entry[1] is %T, want *r4.UnknownResource", b.Entry[1].Resource)
	}
	if u.Type != "InventoryItem" {
		t.Errorf("Type = %q", u.Type)
	}
	if u.RawXML == "" {
		t.Error("the element was not captured")
	}
}

func TestXMLUnknownResourceSurvivesTheRoundTrip(t *testing.T) {
	// The value of capturing it is the content this version has no field for.
	res, err := r4.UnmarshalResourceXML([]byte(xmlBundleWithUnknown))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := r4.MarshalResourceXML(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	for _, want := range []string{"InventoryItem", `value="i1"`, `value="active"`, "nameType", `value="brand"`} {
		if !strings.Contains(string(out), want) {
			t.Errorf("%s did not survive:\n%s", want, out)
		}
	}

	// And what came back can be read again.
	again, err := r4.UnmarshalResourceXML(out)
	if err != nil {
		t.Fatalf("the re-emitted document is not readable: %v", err)
	}
	if b, ok := again.(*r4.Bundle); !ok || len(b.Entry) != 3 {
		t.Errorf("re-read gave %T with a different shape", again)
	}
}

func TestUnknownResourceCannotCrossFormats(t *testing.T) {
	// Only one of Raw and RawXML is ever set: converting between the formats would
	// mean knowing which members are attributes, which are elements and which are
	// primitives carrying a value attribute — which is exactly what makes the type
	// unknown. Asking for the other format is an error rather than a guess.
	fromXML, err := r4.UnmarshalResourceXML([]byte(
		`<?xml version="1.0"?><Nonesuch xmlns="http://hl7.org/fhir"><id value="x"/></Nonesuch>`))
	if err != nil {
		t.Fatalf("unmarshal xml: %v", err)
	}
	if _, jsonErr := json.Marshal(fromXML); jsonErr == nil {
		t.Error("a resource read from XML was written as JSON, which cannot be right")
	} else if !strings.Contains(jsonErr.Error(), "read from XML") {
		t.Errorf("the error does not explain why: %v", jsonErr)
	}

	fromJSON, err := r4.UnmarshalResource([]byte(`{"resourceType":"Nonesuch","id":"x"}`))
	if err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if _, xmlErr := xml.Marshal(fromJSON); xmlErr == nil {
		t.Error("a resource read from JSON was written as XML")
	} else if !strings.Contains(xmlErr.Error(), "read from JSON") {
		t.Errorf("the error does not explain why: %v", xmlErr)
	}
}

func TestXMLUnknownResourceKeepsItsNamespace(t *testing.T) {
	// Re-emitting parsed XML is where namespace declarations get re-injected or
	// lost. The captured element is written through ",innerxml" so its content is
	// passed along rather than re-serialized.
	const doc = `<?xml version="1.0" encoding="UTF-8"?>` +
		`<Nonesuch xmlns="http://hl7.org/fhir"><id value="x"/></Nonesuch>`

	res, err := r4.UnmarshalResourceXML([]byte(doc))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	out, err := r4.MarshalResourceXML(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if n := strings.Count(string(out), "xmlns="); n != 1 {
		t.Errorf("the namespace appears %d times, want once:\n%s", n, out)
	}
	if !strings.Contains(string(out), `<id value="x"`) {
		t.Errorf("the content was lost:\n%s", out)
	}
}

func TestXMLUnknownResourceInEveryVersion(t *testing.T) {
	const doc = `<?xml version="1.0"?><InventoryItem xmlns="http://hl7.org/fhir"><id value="i1"/></InventoryItem>`

	t.Run("r4b treats it as unknown", func(t *testing.T) {
		res, err := r4b.UnmarshalResourceXML([]byte(doc))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if _, ok := res.(*r4b.UnknownResource); !ok {
			t.Errorf("got %T", res)
		}
	})

	t.Run("r5 knows it", func(t *testing.T) {
		res, err := r5.UnmarshalResourceXML([]byte(doc))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if _, ok := res.(*r5.UnknownResource); ok {
			t.Fatal("R5 defines InventoryItem and must not fall back")
		}
	})
}

// TestXMLCaptureIsSemanticNotByteForByte records the one way the XML capture
// differs from the JSON one, so "preserved" is not read as more than it is.
//
// The capture rebuilds the element from the decoder's token stream, and a token
// stream has no notion of how an empty element was spelled. The two forms are the
// same element in XML; keeping the spelling would mean holding the original bytes,
// which the decoder does not offer on every path this runs on.
func TestXMLCaptureIsSemanticNotByteForByte(t *testing.T) {
	const doc = `<?xml version="1.0"?><Nonesuch xmlns="http://hl7.org/fhir">` +
		`<empty></empty><self/><text>kept</text></Nonesuch>`

	res, err := r4.UnmarshalResourceXML([]byte(doc))
	if err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	u, ok := res.(*r4.UnknownResource)
	if !ok {
		t.Fatalf("got %T", res)
	}

	// Every member and its content survives.
	for _, want := range []string{"<empty>", "<self>", "kept"} {
		if !strings.Contains(u.RawXML, want) {
			t.Errorf("%s did not survive the capture: %s", want, u.RawXML)
		}
	}
	// But the self-closing spelling does not.
	if strings.Contains(u.RawXML, "<self/>") {
		t.Error("the capture now preserves self-closing form — if that was fixed, delete this test")
	}

	// And what it emits can be read back as the same thing.
	out, err := r4.MarshalResourceXML(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	again, err := r4.UnmarshalResourceXML(out)
	if err != nil {
		t.Fatalf("the re-emitted document is not readable: %v", err)
	}
	if u2, ok := again.(*r4.UnknownResource); !ok || u2.RawXML != u.RawXML {
		t.Error("a second round trip changed the capture, so it is not stable")
	}
}

package conformance

// XML used to drop the extension of every choice-type primitive. The struct has
// the companion field and JSON preserves it, but the encoder passed nil in its
// place and the decoder discarded it with `_ = ext`, so `_deceasedDateTime`
// survived a JSON round trip and vanished through XML.
//
// TestRoundTrip could not see it. It compares the re-emitted XML against the
// previous re-emission, never against the source, so anything dropped on the way
// in is dropped consistently and looks stable. The checks here compare against
// the source instead, which is what makes the loss visible.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

// stripXMLComments removes comments before counting, because comments can hold
// markup: patient-example-proband.xml carries two commented-out <extension>
// elements, which the decoder is right to ignore. Counting them made six files
// per version look lossy when nothing was lost.
func stripXMLComments(b []byte) []byte {
	var out []byte
	for {
		i := bytes.Index(b, []byte("<!--"))
		if i < 0 {
			return append(out, b...)
		}
		out = append(out, b[:i]...)
		j := bytes.Index(b[i:], []byte("-->"))
		if j < 0 {
			return out
		}
		b = b[i+j+3:]
	}
}

func TestXMLKeepsEveryExtensionInTheCorpus(t *testing.T) {
	// The whole XML corpus, source against re-emission. Before the fix this
	// reported v2-tables.xml losing 3882 extension elements in R4 and
	// valuesets.xml losing 2 in R5; R4B has neither file and lost none.
	versions := []struct {
		name      string
		roundTrip func([]byte) ([]byte, error)
	}{
		{"r4", func(b []byte) ([]byte, error) {
			res, err := r4.UnmarshalResourceXML(b)
			if err != nil {
				return nil, err
			}
			return r4.MarshalResourceXML(res)
		}},
		{"r4b", func(b []byte) ([]byte, error) {
			res, err := r4b.UnmarshalResourceXML(b)
			if err != nil {
				return nil, err
			}
			return r4b.MarshalResourceXML(res)
		}},
		{"r5", func(b []byte) ([]byte, error) {
			res, err := r5.UnmarshalResourceXML(b)
			if err != nil {
				return nil, err
			}
			return r5.MarshalResourceXML(res)
		}},
	}

	for _, v := range versions {
		t.Run(v.name, func(t *testing.T) {
			files, err := filepath.Glob("testdata/examples/" + v.name + "/xml/*.xml")
			if err != nil {
				t.Fatalf("glob: %v", err)
			}
			if len(files) == 0 {
				t.Skip("corpus not fetched; see scripts/fetch-examples.sh")
			}

			checked := 0
			for _, f := range files {
				src, err := os.ReadFile(f)
				if err != nil {
					t.Errorf("%s: %v", filepath.Base(f), err)
					continue
				}
				out, err := v.roundTrip(src)
				if err != nil {
					// Examples that do not decode are TestRoundTrip's business,
					// where the known-failure lists live. Not this test's.
					continue
				}
				checked++

				before := bytes.Count(stripXMLComments(src), []byte("<extension"))
				after := bytes.Count(out, []byte("<extension"))
				if after < before {
					t.Errorf("%s: %d extensions in, %d out — %d lost",
						filepath.Base(f), before, after, before-after)
				}
			}
			if checked == 0 {
				t.Fatal("no example was actually checked")
			}
			t.Logf("%d examples, no extension lost", checked)
		})
	}
}

func TestChoicePrimitivesCarryTheirExtensionThroughXML(t *testing.T) {
	// The narrow reproduction. deceasedDateTime is a variant of deceased[x], so
	// its companion used to be dropped in both directions on the XML path.
	const src = `{"resourceType":"Patient","id":"p1","deceasedDateTime":"2020-01-01",` +
		`"_deceasedDateTime":{"extension":[{"url":"http://x","valueString":"note"}]}}`

	var p r4.Patient
	if err := json.Unmarshal([]byte(src), &p); err != nil {
		t.Fatalf("json: %v", err)
	}
	viaJSON, err := json.Marshal(&p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	asXML, err := r4.MarshalResourceXML(&p)
	if err != nil {
		t.Fatalf("to xml: %v", err)
	}
	if !bytes.Contains(asXML, []byte(`<extension url="http://x">`)) {
		t.Errorf("the encoder did not write the extension: %s", asXML)
	}

	res, err := r4.UnmarshalResourceXML(asXML)
	if err != nil {
		t.Fatalf("from xml: %v", err)
	}
	viaXML, err := json.Marshal(res)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !bytes.Equal(viaXML, viaJSON) {
		t.Errorf("the extension did not survive the XML round trip:\n  json %s\n  xml  %s",
			viaJSON, viaXML)
	}
}

func TestChoiceExtensionsInEveryVersion(t *testing.T) {
	// Same check in R4B and R5, on a datatype rather than a resource: Extension's
	// own value[x] is the most-used choice in the spec.
	const doc = `<?xml version="1.0"?><Patient xmlns="http://hl7.org/fhir">` +
		`<deceasedDateTime value="2020-01-01">` +
		`<extension url="http://x"><valueString value="note"/></extension>` +
		`</deceasedDateTime></Patient>`

	t.Run("r4b", func(t *testing.T) {
		res, err := r4b.UnmarshalResourceXML([]byte(doc))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		p, ok := res.(*r4b.Patient)
		if !ok {
			t.Fatalf("got %T", res)
		}
		if p.DeceasedDateTimeExt == nil {
			t.Fatal("the extension was discarded on the way in")
		}
		out, err := r4b.MarshalResourceXML(res)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if !bytes.Contains(out, []byte(`url="http://x"`)) {
			t.Errorf("and not written on the way out: %s", out)
		}
	})
	t.Run("r5", func(t *testing.T) {
		res, err := r5.UnmarshalResourceXML([]byte(doc))
		if err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		p, ok := res.(*r5.Patient)
		if !ok {
			t.Fatalf("got %T", res)
		}
		if p.DeceasedDateTimeExt == nil {
			t.Fatal("the extension was discarded on the way in")
		}
		out, err := r5.MarshalResourceXML(res)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		if !bytes.Contains(out, []byte(`url="http://x"`)) {
			t.Errorf("and not written on the way out: %s", out)
		}
	})
}

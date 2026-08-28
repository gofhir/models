package conformance

// The XML half of TestRoundTrip only proves our output is self-consistent: parse,
// write, parse, write, converge. That check passed for a long time while the
// narrative was being dropped on every read, because a document with no narrative
// converges perfectly well.
//
// This one compares against the published file instead. For every example that
// carries a Narrative.div, it extracts the human-readable text from the source and
// from what we write back, and requires it to survive. That is the claim the
// <rawInner> fix actually makes, and self-stability cannot express it.
//
// Text is compared with whitespace collapsed rather than byte for byte: the
// decoder rebuilds the div from tokens, so attribute order and empty-element
// spelling can shift without any content being lost. What must not change is a
// single character the reader would see.

import (
	"bytes"
	"encoding/xml"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	// The corpus-wide checks drive all three versions through codecs(); the
	// targeted cases below need concrete types, and r4 is representative — the
	// XHTML handling is generated from one template for all three.
	"github.com/gofhir/models/r4/v2"
)

// narrativeText returns the visible text of the first <div> under <text>, and
// whether the document had one at all. Markup is discarded; only character data
// is kept, with runs of whitespace collapsed to a single space.
func narrativeText(doc []byte) (string, bool) {
	dec := xml.NewDecoder(bytes.NewReader(doc))

	inText := false
	depth := 0
	var sb strings.Builder

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			// Malformed input is the round-trip test's problem, not this one.
			return "", false
		}

		switch t := tok.(type) {
		case xml.StartElement:
			switch {
			case depth > 0:
				depth++
			case t.Name.Local == "text":
				inText = true
			case inText && t.Name.Local == "div":
				depth = 1
			}
		case xml.EndElement:
			if depth > 0 {
				depth--
				if depth == 0 {
					return collapseSpaces(sb.String()), true
				}
			} else if t.Name.Local == "text" {
				inText = false
			}
		case xml.CharData:
			if depth > 0 {
				sb.Write(t)
			}
		}
	}
	return "", false
}

func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func TestNarrativeSurvivesXMLRoundTrip(t *testing.T) {
	for _, c := range codecs() {
		t.Run(c.name, func(t *testing.T) {
			dir := filepath.Join(examplesDir, c.name, "xml")
			files, err := filepath.Glob(filepath.Join(dir, "*.xml"))
			if err != nil {
				t.Fatalf("glob %s: %v", dir, err)
			}
			if len(files) == 0 {
				t.Skipf("no corpus in %s; run scripts/fetch-examples.sh", dir)
			}

			var withNarrative, lost, changed int
			for _, path := range files {
				source, err := os.ReadFile(path)
				if err != nil {
					t.Fatalf("read %s: %v", path, err)
				}

				want, ok := narrativeText(source)
				if !ok || want == "" {
					continue // nothing to preserve
				}
				withNarrative++

				// Two full cycles, not one. The <rawInner> defect did not lose
				// the narrative on the way out — it lost it on the way back in,
				// because the decoder switches on the element name and never
				// matched. A single parse/write pass still has the div sitting
				// inside the wrapper and looks fine.
				resource, err := c.unmarshalXML(source)
				if err != nil {
					continue // a parse failure is the round-trip test's business
				}
				first, err := c.marshalXML(resource)
				if err != nil {
					continue
				}
				reread, err := c.unmarshalXML(first)
				if err != nil {
					continue
				}
				out, err := c.marshalXML(reread)
				if err != nil {
					continue
				}

				got, ok := narrativeText(out)
				switch {
				case !ok || got == "":
					lost++
					if lost <= 3 {
						t.Errorf("%s: narrative lost on round-trip", filepath.Base(path))
					}
				case got != want:
					changed++
					if changed <= 3 {
						t.Errorf("%s: narrative text changed\n  want: %s\n  got:  %s",
							filepath.Base(path), truncate(want), truncate(got))
					}
				}
			}

			if lost > 3 || changed > 3 {
				t.Errorf("%d narratives lost and %d altered in total", lost, changed)
			}

			// Without this the check would pass silently if the extractor stopped
			// finding narratives at all — the exact failure it exists to catch.
			if withNarrative < 500 {
				t.Errorf("only %d examples with a narrative found in %s; expected most of the corpus, so the extractor is not working",
					withNarrative, dir)
			}
			t.Logf("%d of %d examples carry a narrative; all survive", withNarrative, len(files))
		})
	}
}

func truncate(s string) string {
	const limit = 120
	if len(s) <= limit {
		return s
	}
	return s[:limit] + "…"
}

// TestNarrativeXHTMLIsNotRewritten pins the boundary between the self-closing
// rewrite and the narrative.
//
// MarshalResourceXML post-processes its output with a regular expression that
// turns <tag attr=""></tag> into <tag attr=""/>, because encoding/xml cannot emit
// self-closing elements and FHIR spells primitives that way. That rewrite used to
// run over the whole document, narrative included, so an author's
// <a href="q"></a> came back as <a href="q"/> — the same element in XML, a
// different one to an HTML parser, and either way not the verbatim carriage the
// specification asks for.
//
// The cases below are the ones that make the cut non-trivial: a div nested inside
// the narrative, a self-closing div, and the requirement that primitives outside
// the narrative still collapse.
func TestNarrativeXHTMLIsNotRewritten(t *testing.T) {
	const ns = `xmlns="http://www.w3.org/1999/xhtml"`

	tests := []struct {
		name     string
		div      string
		preserve []string
	}{
		{
			name:     "empty anchor keeps its closing tag",
			div:      `<div ` + ns + `><a href="q"></a></div>`,
			preserve: []string{`<a href="q"></a>`},
		},
		{
			name:     "nested div does not end the protected region early",
			div:      `<div ` + ns + `><div class="c"><a href="q"></a></div></div>`,
			preserve: []string{`<div class="c">`, `<a href="q"></a>`},
		},
		{
			name:     "mixed empty elements are left alone",
			div:      `<div ` + ns + `><br/><img src="a"></img></div>`,
			preserve: []string{`<br/>`, `<img src="a"></img>`},
		},
		{
			name:     "self-closing div does not swallow the rest",
			div:      `<div ` + ns + `/>`,
			preserve: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patient := &r4.Patient{
				ResourceType: "Patient",
				Id:           r4.Ptr("x"),
				Text: &r4.Narrative{
					Status: r4.Ptr(r4.NarrativeStatusGenerated),
					Div:    r4.Ptr(tt.div),
				},
			}

			out, err := r4.MarshalResourceXML(patient)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			got := string(out)

			for _, want := range tt.preserve {
				if !strings.Contains(got, want) {
					t.Errorf("narrative was rewritten: %s is missing from\n  %s", want, got)
				}
			}

			// The rewrite must still apply everywhere else, or this fix would have
			// traded one defect for a document full of expanded primitives.
			if !strings.Contains(got, `<id value="x"/>`) {
				t.Errorf("primitives outside the narrative stopped collapsing:\n  %s", got)
			}

			// And the document still has to survive a round-trip.
			back, err := r4.UnmarshalResourceXML(out)
			if err != nil {
				t.Fatalf("re-parse: %v", err)
			}
			if back.(*r4.Patient).Text == nil || back.(*r4.Patient).Text.Div == nil {
				t.Error("narrative lost on re-parse")
			}
		})
	}
}

// TestMalformedNarrativeIsRejected covers the other half of the old behavior:
// the div was written into the document unchecked, so broken markup produced
// output that was not well-formed XML while MarshalResourceXML reported success.
func TestMalformedNarrativeIsRejected(t *testing.T) {
	for _, tt := range []struct {
		name string
		div  string
	}{
		{"unclosed element", `<div><p>unclosed</div>`},
		{"not a div", `<span xmlns="http://www.w3.org/1999/xhtml">x</span>`},
		{"plain text", `just text`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			patient := &r4.Patient{
				ResourceType: "Patient",
				Text:         &r4.Narrative{Div: r4.Ptr(tt.div)},
			}
			if _, err := r4.MarshalResourceXML(patient); err == nil {
				t.Error("malformed narrative was accepted; it used to produce output that is not well-formed with a nil error")
			}
		})
	}
}

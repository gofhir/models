package conformance

// Marshal's own documentation, and the package doc, said the standard encoder
// "breaks" and "corrupts" the XHTML in Narrative.Div. It does not: < is
// ordinary JSON and decodes back to "<". What the escaping costs is bytes that
// no longer match the documents HL7 publishes, and a narrative nobody can read
// in the output.
//
// The claim was fixed. This is what holds the corrected version to something
// measured, since the previous one had stood unchallenged for as long as it took
// two other claims in the same paragraph to go stale.

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
)

// escapedLT is the six characters encoding/json writes for "<". Built rather
// than written as a literal, because a literal of it is exactly the thing this
// file keeps getting wrong.
var escapedLT = string([]byte{'\\', 'u', '0', '0', '3', 'c'})

func TestStandardEscapingIsUglyNotLossy(t *testing.T) {
	const div = `<div xmlns="http://www.w3.org/1999/xhtml"><p>a &amp; b &lt; c</p></div>`
	patient := &r4.Patient{Text: &r4.Narrative{
		Status: r4.Ptr(r4.NarrativeStatusGenerated),
		Div:    r4.Ptr(div),
	}}

	escaped, err := json.Marshal(patient)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	plain, err := r4.Marshal(patient)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	// The escaping happens: the standard encoder writes the six characters of a
	// < escape where the div has a "<", and Marshal writes the character.
	if !strings.Contains(string(escaped), escapedLT) || strings.Contains(string(escaped), "<") {
		t.Errorf("encoding/json no longer escapes; the doc needs revisiting: %s", escaped)
	}
	if !strings.Contains(string(plain), "<") {
		t.Errorf("Marshal is escaping after all: %s", plain)
	}

	// And it costs only bytes.
	if len(escaped) <= len(plain) {
		t.Errorf("escaped output is not longer: %d vs %d", len(escaped), len(plain))
	}

	// Nothing is lost: the escaped form reads back to the same div, through both
	// the plain decoder and the resource dispatcher.
	var back r4.Patient
	if err := json.Unmarshal(escaped, &back); err != nil {
		t.Fatalf("the escaped document does not decode: %v", err)
	}
	if got := r4.Val(back.Text.Div); got != div {
		t.Errorf("the XHTML did not survive:\n  got  %s\n  want %s", got, div)
	}
	if _, err := r4.UnmarshalResource(escaped); err != nil {
		t.Errorf("UnmarshalResource rejects the escaped document: %v", err)
	}
}

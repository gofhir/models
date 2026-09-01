package conformance

// GetResourceType scans for the member instead of unmarshaling the document, so
// UnmarshalResource no longer parses everything twice — once to learn the type
// and once to decode it, at every level of nesting.
//
// The scan handles only the shape it is sure about and falls through to
// encoding/json for everything else, which makes equivalence the property worth
// testing: for any input, the new path must return what the old one did,
// including the error. The table pins the cases that motivated the design; the
// fuzz target looks for the ones nobody thought of.

import (
	"encoding/json"
	"fmt"
	"testing"
	"unicode/utf8"

	"github.com/gofhir/models/r4/v2"
)

// legacyGetResourceType is the implementation this replaced, kept as the oracle.
func legacyGetResourceType(data []byte) (string, error) {
	var peek struct {
		ResourceType string `json:"resourceType"`
	}
	if err := json.Unmarshal(data, &peek); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %w", err)
	}
	if peek.ResourceType == "" {
		return "", fmt.Errorf("resourceType field is missing or empty")
	}
	return peek.ResourceType, nil
}

// sameResult requires full equivalence for documents encoding/json accepts, and
// only type agreement for those it rejects — the contract the scan makes.
func sameResult(t *testing.T, input string) {
	t.Helper()

	gotVal, gotErr := r4.GetResourceType([]byte(input))
	wantVal, wantErr := legacyGetResourceType([]byte(input))

	if !json.Valid([]byte(input)) {
		if gotErr == nil && wantErr == nil && gotVal != wantVal {
			t.Errorf("disagree about the type of invalid input %q: scan=%q legacy=%q", input, gotVal, wantVal)
		}
		return
	}

	if gotVal != wantVal {
		t.Errorf("value differs for %q:\n  scan:   %q\n  legacy: %q", input, gotVal, wantVal)
	}
	switch {
	case gotErr == nil && wantErr != nil:
		t.Errorf("scan accepted %q; legacy rejected it: %v", input, wantErr)
	case gotErr != nil && wantErr == nil:
		t.Errorf("scan rejected %q: %v; legacy accepted it", input, gotErr)
	case gotErr != nil && wantErr != nil && gotErr.Error() != wantErr.Error():
		t.Errorf("error text differs for %q:\n  scan:   %v\n  legacy: %v", input, gotErr, wantErr)
	}
}

func TestGetResourceTypeMatchesEncodingJSON(t *testing.T) {
	for _, input := range []string{
		// the ordinary shapes
		`{"resourceType":"Patient"}`,
		`{"id":"x","resourceType":"Patient"}`,
		`{ "resourceType" : "Patient" }`,
		"{\n\t\"resourceType\":\t\"Patient\"\n}",

		// absent, empty, wrong type
		`{}`,
		`{"id":"x"}`,
		`{"resourceType":""}`,
		`{"resourceType":null}`,
		`{"resourceType":5}`,
		`{"resourceType":true}`,
		`{"resourceType":[]}`,
		`{"resourceType":{}}`,

		// a later duplicate wins, and the scan must agree
		`{"resourceType":"A","resourceType":"B"}`,
		`{"resourceType":"A","x":1,"resourceType":"B"}`,
		`{"resourceType":"A","resourceType":null}`,
		`{"resourceType":"A","resourceType":5}`,

		// nested occurrences must not be mistaken for the top-level one
		`{"nested":{"resourceType":"Wrong"}}`,
		`{"nested":{"resourceType":"Wrong"},"resourceType":"Right"}`,
		`{"a":[{"resourceType":"Wrong"}],"resourceType":"Right"}`,
		`{"a":[[{"resourceType":"Wrong"}]],"resourceType":"Right"}`,

		// escapes, in the key and in the value
		`{"resource\u0054ype":"Patient"}`,
		`{"resourceType":"Pat\u0069ent"}`,
		`{"resourceType":"a\"b"}`,
		`{"resourceType":"a\\b"}`,
		`{"resourceType":"\u00e9"}`,

		// values that contain structural bytes
		`{"a":"}","resourceType":"Patient"}`,
		`{"a":"{\"resourceType\":\"Wrong\"}","resourceType":"Right"}`,
		`{"a":"[,]","resourceType":"Patient"}`,

		// not an object, or not JSON at all
		`[]`,
		`[{"resourceType":"Patient"}]`,
		`"x"`,
		`5`,
		`true`,
		`null`,
		``,
		`{`,
		`}`,
		`{"resourceType"}`,
		`{"resourceType":}`,
		`{"resourceType":"unterminated`,
		`{"resourceType":"Patient"`,
		`{"resourceType":"Patient"},`,
		`   `,

		// numbers and literals as sibling values, which the skipper has to step over
		`{"n":-1.5e10,"resourceType":"Patient"}`,
		`{"n":null,"b":true,"f":false,"resourceType":"Patient"}`,
		`{"deep":{"a":{"b":[1,2,{"c":"d"}]}},"resourceType":"Patient"}`,
	} {
		sameResult(t, input)
	}
}

// FuzzGetResourceType checks the guarantee the scan actually makes: for any
// document encoding/json accepts, it returns exactly what unmarshaling would,
// errors included.
//
// For documents encoding/json rejects, the scan is allowed to be more permissive
// — it does not validate what it steps over, and making it do so costs more than
// the scan saves. What it must never do is report a *different* type than
// encoding/json would, since UnmarshalResource picks the Go type from this answer
// and then decodes the same bytes with encoding/json; a disagreement there is a
// parser differential.
//
// Run with `go test -fuzz=FuzzGetResourceType ./conformance`.
func FuzzGetResourceType(f *testing.F) {
	for _, seed := range []string{
		`{"resourceType":"Patient"}`,
		`{"resourceType":"A","resourceType":"B"}`,
		`{"nested":{"resourceType":"Wrong"},"resourceType":"Right"}`,
		`{"resource\u0054ype":"Patient"}`,
		`{"a":"}","resourceType":"Patient"}`,
		`{"resourceType":5}`,
		`{"":A}`,
		`{"":"""":""}`,
		`{"\u0000":0}`,
		`{"":{""}}`,
		`{"resourceType":"Patient"},`,
		`{`,
		`[]`,
	} {
		f.Add([]byte(seed))
	}

	f.Fuzz(func(t *testing.T, data []byte) {
		if !utf8.Valid(data) {
			t.Skip() // neither path promises anything specific here
		}

		gotVal, gotErr := r4.GetResourceType(data)
		wantVal, wantErr := legacyGetResourceType(data)

		if json.Valid(data) {
			// Full equivalence is required.
			if gotVal != wantVal {
				t.Fatalf("value differs on valid JSON\n  input:  %q\n  scan:   %q\n  legacy: %q", data, gotVal, wantVal)
			}
			if (gotErr == nil) != (wantErr == nil) {
				t.Fatalf("error presence differs on valid JSON\n  input:  %q\n  scan:   %v\n  legacy: %v", data, gotErr, wantErr)
			}
			if gotErr != nil && gotErr.Error() != wantErr.Error() {
				t.Fatalf("error text differs on valid JSON\n  input:  %q\n  scan:   %v\n  legacy: %v", data, gotErr, wantErr)
			}
			return
		}

		// Invalid JSON: the scan may succeed where unmarshaling fails, but if both
		// produce a type it must be the same one.
		if gotErr == nil && wantErr == nil && gotVal != wantVal {
			t.Fatalf("the two disagree about the resource type of an invalid document\n  input:  %q\n  scan:   %q\n  legacy: %q", data, gotVal, wantVal)
		}
	})
}

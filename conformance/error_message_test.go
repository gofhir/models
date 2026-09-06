package conformance

// Decoding a resource that holds a polymorphic field goes through a local
// "type Alias X" so UnmarshalJSON does not recurse into itself. encoding/json
// names the type it was decoding into, so that internal detail reached the caller
// as a field that does not exist.

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestDecodeErrorsDoNotMentionTheAlias(t *testing.T) {
	// Parameters.parameter holds a resource, so it is one of the types that needs
	// the recursion guard.
	const doc = `{"resourceType":"Parameters","parameter":[{"name":42}]}`

	var p r4.Parameters
	err := json.Unmarshal([]byte(doc), &p)
	if err == nil {
		t.Fatal("a number in a string field was accepted")
	}
	if strings.Contains(err.Error(), "Alias") {
		t.Errorf("the error names an internal type: %v", err)
	}
	// The FHIR path, not the Go type name: Parameters.parameter.name is what a
	// reader of the specification recognizes, and what
	// OperationOutcome.issue.expression is expressed in.
	if !strings.Contains(err.Error(), "Parameters.parameter.name") {
		t.Errorf("the error does not give the FHIR path of the field: %v", err)
	}
	if strings.Contains(err.Error(), "ParametersParameter") {
		t.Errorf("the error gives the Go type name rather than the FHIR path: %v", err)
	}
	// And not the mangled form that deleting the alias alone would leave.
	if strings.Contains(err.Error(), "field .name") {
		t.Errorf("the field name lost its type: %v", err)
	}
}

func TestCleanedErrorsKeepTheirCause(t *testing.T) {
	// Rewriting the message must not cost the error chain: a caller inspecting the
	// failure with errors.As still needs the json error underneath.
	const doc = `{"resourceType":"Parameters","parameter":[{"name":42}]}`

	var p r4.Parameters
	err := json.Unmarshal([]byte(doc), &p)
	if err == nil {
		t.Fatal("expected an error")
	}

	var typeErr *json.UnmarshalTypeError
	if !errors.As(err, &typeErr) {
		t.Fatal("errors.As no longer reaches the json error")
	}
	if typeErr.Type.String() != "string" {
		t.Errorf("the underlying error lost its detail: %v", typeErr.Type)
	}
}

func TestUnaffectedErrorsAreLeftAlone(t *testing.T) {
	// A resource with no polymorphic field never goes through the alias, and its
	// errors must come through untouched rather than reformatted.
	var pat r4.Patient
	err := json.Unmarshal([]byte(`{"resourceType":"Patient","active":"no"}`), &pat)
	if err == nil {
		t.Fatal("expected an error")
	}
	const want = "json: cannot unmarshal string into Go struct field Patient.active of type bool"
	if err.Error() != want {
		t.Errorf("an untouched error was rewritten:\n  got  %v\n  want %s", err, want)
	}
}

func TestNestedErrorsKeepTheirIndex(t *testing.T) {
	// The dispatcher's own errors already carry position, which is what makes a
	// failure inside a Bundle or a contained list findable.
	for _, tt := range []struct{ name, doc, want string }{
		{
			"contained",
			`{"resourceType":"Patient","contained":[{"resourceType":"Organization"},{"id":"no-type"}]}`,
			"contained[1]",
		},
		{
			"bundle entry",
			`{"resourceType":"Bundle","entry":[{"resource":{"resourceType":"Patient","active":1}}]}`,
			"Patient",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var target any = &r4.Patient{}
			if strings.Contains(tt.doc, `"Bundle"`) {
				target = &r4.Bundle{}
			}
			err := json.Unmarshal([]byte(tt.doc), target)
			if err == nil {
				t.Fatal("expected an error")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("the error does not locate the failure: %v", err)
			}
		})
	}
}

func TestAliasIsHiddenInEveryVersion(t *testing.T) {
	const doc = `{"resourceType":"Parameters","parameter":[{"name":42}]}`

	t.Run("r4b", func(t *testing.T) {
		var p r4b.Parameters
		if err := json.Unmarshal([]byte(doc), &p); err == nil {
			t.Fatal("expected an error")
		} else if strings.Contains(err.Error(), "Alias") {
			t.Errorf("%v", err)
		}
	})
	t.Run("r5", func(t *testing.T) {
		var p r5.Parameters
		if err := json.Unmarshal([]byte(doc), &p); err == nil {
			t.Fatal("expected an error")
		} else if strings.Contains(err.Error(), "Alias") {
			t.Errorf("%v", err)
		}
	})
}

// TestFHIRFieldsNamedAliasAreUntouched guards the narrowness of the rewrite.
//
// FHIR has fields called alias — Organization.alias, Location.alias — and a looser
// pattern than "field .Alias." would rewrite them, or worse, rewrite a caller's own
// data that happened to contain the same text. They appear lower-cased in these
// messages, so the two cannot collide.
func TestFHIRFieldsNamedAliasAreUntouched(t *testing.T) {
	for _, tt := range []struct {
		name   string
		doc    string
		target func() any
		want   string
	}{
		{"Organization", `{"resourceType":"Organization","alias":[42]}`,
			func() any { return &r4.Organization{} },
			"json: cannot unmarshal number into Go struct field Organization.alias of type string"},
		{"Location", `{"resourceType":"Location","alias":[42]}`,
			func() any { return &r4.Location{} },
			"json: cannot unmarshal number into Go struct field Location.alias of type string"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			err := json.Unmarshal([]byte(tt.doc), tt.target())
			if err == nil {
				t.Fatal("expected an error")
			}
			if err.Error() != tt.want {
				t.Errorf("a real alias field was rewritten:\n  got  %v\n  want %s", err, tt.want)
			}
		})
	}
}

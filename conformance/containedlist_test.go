package conformance

// Contained is a ContainedList rather than a []Resource, so the dispatch that
// encoding/json cannot do for an interface lives on one type instead of being
// generated into every resource. That removed 437 UnmarshalJSON methods and
// 16,130 lines of generated code.
//
// A named slice type keeps its underlying type, so the change is meant to be
// invisible in source except for %T. "Meant to be" is what these check: each case
// below is a way client code touches the field, and the last one pins the single
// difference so it is documented rather than discovered.

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
)

func TestContainedListBehavesAsSlice(t *testing.T) {
	org := &r4.Organization{Id: r4.Ptr("o1")}

	t.Run("assignment from []Resource", func(t *testing.T) {
		// The explicit type is the assertion: it must be assignable to the field.
		var list []r4.Resource = []r4.Resource{org} //nolint:staticcheck // ST1023: the type is the point
		p := &r4.Patient{Contained: list}
		if len(p.Contained) != 1 {
			t.Fatalf("len = %d", len(p.Contained))
		}
	})

	t.Run("append", func(t *testing.T) {
		p := &r4.Patient{Contained: []r4.Resource{org}}
		p.Contained = append(p.Contained, &r4.Practitioner{Id: r4.Ptr("pr1")})
		if len(p.Contained) != 2 {
			t.Fatalf("len = %d", len(p.Contained))
		}
	})

	t.Run("range, index and slicing", func(t *testing.T) {
		p := &r4.Patient{Contained: []r4.Resource{org, &r4.Practitioner{}}}
		seen := 0
		for _, c := range p.Contained {
			if c == nil {
				t.Error("nil entry in range")
			}
			seen++
		}
		if seen != 2 {
			t.Errorf("range saw %d entries", seen)
		}
		if p.Contained[0].GetResourceType() != "Organization" {
			t.Error("indexing broken")
		}
		if len(p.Contained[:1]) != 1 {
			t.Error("slicing broken")
		}
	})

	t.Run("passing to func([]Resource)", func(t *testing.T) {
		count := func(rs []r4.Resource) int { return len(rs) }
		p := &r4.Patient{Contained: []r4.Resource{org}}
		if count(p.Contained) != 1 {
			t.Error("passing to a []Resource parameter lost entries")
		}
	})

	t.Run("GetContained still returns []Resource", func(t *testing.T) {
		p := &r4.Patient{Contained: []r4.Resource{org}}
		// Likewise: the getter must still yield a plain []Resource.
		var out []r4.Resource = p.GetContained() //nolint:staticcheck // ST1023: the type is the point
		if len(out) != 1 {
			t.Error("GetContained lost entries")
		}
	})

	// The one observable difference. Recorded here so that it is a documented
	// consequence rather than a surprise in someone's log output.
	t.Run("%T reports ContainedList", func(t *testing.T) {
		p := &r4.Patient{Contained: []r4.Resource{org}}
		if got := fmt.Sprintf("%T", p.Contained); got != "r4.ContainedList" {
			t.Errorf("%%T = %q, want r4.ContainedList", got)
		}
	})
}

// TestContainedListErrorsKeepTheirIndex guards the diagnostic quality that the
// per-resource methods provided. Moving the loop into one shared method is only
// acceptable if it still says which element failed.
func TestContainedListErrorsKeepTheirIndex(t *testing.T) {
	for _, tt := range []struct {
		name  string
		input string
		want  string
	}{
		// These used an unrecognized resourceType to force the failure. That is no
		// longer an error — it decodes to an UnknownResource so a document from a
		// newer server stays readable — so the failure now comes from a member
		// with no resourceType at all, which is not a resource in any version.
		{
			name:  "second element",
			input: `{"resourceType":"Patient","contained":[{"resourceType":"Organization"},{"id":"no-type"}]}`,
			want:  "contained[1]",
		},
		{
			name:  "nested inside a Bundle entry",
			input: `{"resourceType":"Bundle","entry":[{"resource":{"resourceType":"Patient","contained":[{"id":"no-type"}]}}]}`,
			want:  "contained[0]",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var target any
			if strings.Contains(tt.input, `"Bundle"`) {
				target = new(r4.Bundle)
			} else {
				target = new(r4.Patient)
			}
			err := json.Unmarshal([]byte(tt.input), target)
			if err == nil {
				t.Fatal("expected an error for the member with no resourceType")
			}
			if !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error does not identify the element: %v", err)
			}
		})
	}
}

// TestContainedListHandlesNullAndEmpty covers the two inputs that are not a list
// of resources. A nil entry left in the slice would marshal back out as null,
// which is not valid FHIR.
func TestContainedListHandlesNullAndEmpty(t *testing.T) {
	for _, tt := range []struct {
		name    string
		input   string
		wantLen int
	}{
		{"empty array", `{"resourceType":"Patient","contained":[]}`, 0},
		{"explicit null element", `{"resourceType":"Patient","contained":[null]}`, 0},
		{"null among resources", `{"resourceType":"Patient","contained":[null,{"resourceType":"Organization"}]}`, 1},
		{"absent", `{"resourceType":"Patient"}`, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var p r4.Patient
			if err := json.Unmarshal([]byte(tt.input), &p); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if len(p.Contained) != tt.wantLen {
				t.Errorf("len = %d, want %d", len(p.Contained), tt.wantLen)
			}
			out, err := json.Marshal(&p)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if strings.Contains(string(out), "null") {
				t.Errorf("re-emitted a null: %s", out)
			}
		})
	}
}

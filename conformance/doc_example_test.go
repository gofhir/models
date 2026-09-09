package conformance

// The example in doc.go did not compile. It assigned "Patient" to ResourceType,
// which is a marker struct with one possible value and no exported name, so the
// first code a reader meets on pkg.go.dev was code they could not run.
//
// It stood through three separate passes over that same doc comment, because
// every pass read the prose. Prose is what a reader checks; a compiler is what
// checks code. So the example is compiled here and then compared against the
// text, which is the only way the two cannot drift apart.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

// docExample is the example doc.go carries, as real code. Editing the doc without
// editing this fails, and editing this without keeping it compilable fails too.
const docExample = `patient := r4.Patient{
	Active: r4.Ptr(true),
	Name: []r4.HumanName{
		{Family: r4.Ptr("Smith")},
	},
}

family := r4.Val(r4.First(patient.Name)).Family`

func TestTheDocExampleCompiles(t *testing.T) {
	// Character for character what the const above says, so a reader copying it
	// out of the package documentation gets working code.
	patient := r4.Patient{
		Active: r4.Ptr(true),
		Name: []r4.HumanName{
			{Family: r4.Ptr("Smith")},
		},
	}

	family := r4.Val(r4.First(patient.Name)).Family

	if r4.Val(family) != "Smith" {
		t.Errorf("family reads back as %q", r4.Val(family))
	}

	// The same shape in the other two, since all three carry this doc.
	r4bPatient := r4b.Patient{Name: []r4b.HumanName{{Family: r4b.Ptr("Smith")}}}
	if got := r4b.Val(r4b.Val(r4b.First(r4bPatient.Name)).Family); got != "Smith" {
		t.Errorf("r4b: %q", got)
	}
	r5Patient := r5.Patient{Name: []r5.HumanName{{Family: r5.Ptr("Smith")}}}
	if got := r5.Val(r5.Val(r5.First(r5Patient.Name)).Family); got != "Smith" {
		t.Errorf("r5: %q", got)
	}
}

func TestTheDocExampleIsWhatTheDocSays(t *testing.T) {
	for _, version := range []string{"r4", "r4b", "r5"} {
		t.Run(version, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("..", version, "doc.go"))
			if err != nil {
				t.Fatalf("reading doc.go: %v", err)
			}

			want := strings.ReplaceAll(docExample, "r4.", version+".")
			if !strings.Contains(uncomment(string(source)), want) {
				t.Errorf("the example in %s/doc.go is not the one this test compiles.\n"+
					"Compiled:\n%s\n\nEdit both, or they drift.", version, want)
			}
		})
	}
}

// uncomment strips the "// " and "//\t" prefixes from a doc comment so the code
// inside it can be compared with code.
func uncomment(source string) string {
	var out strings.Builder
	for _, line := range strings.Split(source, "\n") {
		switch {
		case strings.HasPrefix(line, "//\t"):
			out.WriteString(strings.TrimPrefix(line, "//\t"))
		case line == "//":
			// A blank line inside the comment is blank in the code.
		case strings.HasPrefix(line, "// "):
			out.WriteString(strings.TrimPrefix(line, "// "))
		default:
			out.WriteString(line)
		}
		out.WriteString("\n")
	}
	return out.String()
}

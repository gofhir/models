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

// Each subtest is the const above character for character, with the package name
// swapped. All three are spelled out rather than shared, because a shared helper
// would compile one version's fields and leave the other two unchecked — which is
// the whole failure this file exists to prevent.
func TestTheDocExampleCompiles(t *testing.T) {
	t.Run("r4", func(t *testing.T) {
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
	})

	t.Run("r4b", func(t *testing.T) {
		patient := r4b.Patient{
			Active: r4b.Ptr(true),
			Name: []r4b.HumanName{
				{Family: r4b.Ptr("Smith")},
			},
		}

		family := r4b.Val(r4b.First(patient.Name)).Family

		if r4b.Val(family) != "Smith" {
			t.Errorf("family reads back as %q", r4b.Val(family))
		}
	})

	t.Run("r5", func(t *testing.T) {
		patient := r5.Patient{
			Active: r5.Ptr(true),
			Name: []r5.HumanName{
				{Family: r5.Ptr("Smith")},
			},
		}

		family := r5.Val(r5.First(patient.Name)).Family

		if r5.Val(family) != "Smith" {
			t.Errorf("family reads back as %q", r5.Val(family))
		}
	})
}

func TestTheDocExampleIsWhatTheDocSays(t *testing.T) {
	for _, version := range []string{"r4", "r4b", "r5"} {
		t.Run(version, func(t *testing.T) {
			source, err := os.ReadFile(filepath.Join("..", version, "doc.go"))
			if err != nil {
				t.Fatalf("reading doc.go: %v", err)
			}

			// A CRLF checkout carries the example just as faithfully, so it must
			// not be the thing that fails.
			text := strings.ReplaceAll(string(source), "\r\n", "\n")

			want := strings.ReplaceAll(docExample, "r4.", version+".")
			if !strings.Contains(uncomment(text), want) {
				t.Errorf("the example in %s/doc.go is not the one this test compiles.\n\n"+
					"Compiled here:\n%s\n\nIn %s/doc.go:\n%s\n\nEdit both, or they drift.",
					version, want, version, exampleIn(text))
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

// exampleIn pulls the indented block that holds the patient example out of a doc
// comment, so a failure can show what the file says rather than only what was
// expected of it.
func exampleIn(source string) string {
	lines := strings.Split(source, "\n")

	start := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "//\t") && strings.Contains(line, "patient :=") {
			start = i
			break
		}
	}
	if start < 0 {
		return "(no indented example found)"
	}

	var block []string
	for _, line := range lines[start:] {
		if line == "//" {
			block = append(block, "")
			continue
		}
		if !strings.HasPrefix(line, "//\t") {
			break
		}
		block = append(block, strings.TrimPrefix(line, "//\t"))
	}
	return strings.TrimRight(strings.Join(block, "\n"), "\n")
}

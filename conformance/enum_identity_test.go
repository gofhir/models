package conformance

// A generated enum was a bare string. The display text and the CodeSystem URL were
// both in the specification and both discarded — the display survived only as a
// comment, unreachable at run time — so anything showing a code to a person, or
// writing one into a CodeableConcept, carried its own copy of data the generator
// already had.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestEnumCarriesItsDisplayAndSystem(t *testing.T) {
	g := r4.AdministrativeGenderMale

	if got := g.Display(); got != "Male" {
		t.Errorf("Display() = %q, want Male", got)
	}
	if got := g.System(); got != "http://hl7.org/fhir/administrative-gender" {
		t.Errorf("System() = %q", got)
	}

	// A different CodeSystem, to show the system is per code rather than baked in.
	s := r4.ObservationStatusFinal
	if got := s.System(); got != "http://hl7.org/fhir/observation-status" {
		t.Errorf("ObservationStatus.System() = %q", got)
	}
}

func TestEnumProducesAUsableCoding(t *testing.T) {
	// Building this by hand is where the system URL gets copied wrong.
	cc := r4.NewCodeableConceptBuilder().
		AddCoding(r4.AdministrativeGenderMale.Coding()).
		Build()

	out, err := json.Marshal(cc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	const want = `{"coding":[{"system":"http://hl7.org/fhir/administrative-gender","code":"male","display":"Male"}]}`
	if string(out) != want {
		t.Errorf("got  %s\nwant %s", out, want)
	}
}

func TestEnumValuesAreCompleteAndOrdered(t *testing.T) {
	values := r4.AdministrativeGenderValues()
	want := []r4.AdministrativeGender{
		r4.AdministrativeGenderMale,
		r4.AdministrativeGenderFemale,
		r4.AdministrativeGenderOther,
		r4.AdministrativeGenderUnknown,
	}
	if len(values) != len(want) {
		t.Fatalf("got %d values, want %d", len(values), len(want))
	}
	for i := range want {
		if values[i] != want[i] {
			t.Errorf("values[%d] = %q, want %q — specification order", i, values[i], want[i])
		}
	}
}

func TestEnumValidatesWhatCameOffTheWire(t *testing.T) {
	// The type is a string, so anything can be assigned to it — including a value
	// that was never checked. Decoding does not reject it either.
	var p r4.Patient
	if err := json.Unmarshal([]byte(`{"resourceType":"Patient","gender":"not-a-gender"}`), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.Gender == nil {
		t.Fatal("gender did not decode")
	}
	if p.Gender.IsValid() {
		t.Error("IsValid accepted a code the specification does not define")
	}
	if !r4.AdministrativeGenderMale.IsValid() {
		t.Error("IsValid rejected a real code")
	}
}

func TestDisplayFallsBackToTheCode(t *testing.T) {
	// Falling back to the code rather than to an empty string means the result can
	// go straight in front of a person, even for a code we do not know.
	unknown := r4.AdministrativeGender("not-a-gender")
	if got := unknown.Display(); got != "not-a-gender" {
		t.Errorf("Display() = %q, want the code itself", got)
	}
	if got := unknown.System(); got != "" {
		t.Errorf("System() = %q, want empty for an unknown code", got)
	}

	// And a Coding for an unknown code carries no system rather than an empty one,
	// which would be a claim about a vocabulary that does not hold.
	out, err := json.Marshal(unknown.Coding())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(out), `"system"`) {
		t.Errorf("a code with no known system claimed one: %s", out)
	}
}

func TestEveryEnumHasIdentity(t *testing.T) {
	// The point is that this is not a handful of hand-written helpers: every
	// generated enum carries it.
	typeDecl := regexp.MustCompile(`(?m)^type ([A-Za-z0-9]+) string$`)
	codingMethod := regexp.MustCompile(`(?m)^func \(c ([A-Za-z0-9]+)\) Coding\(\) Coding \{$`)

	for _, module := range []string{"r4", "r4b", "r5"} {
		t.Run(module, func(t *testing.T) {
			source := readGenerated(t, module, "codesystems.go")

			types := map[string]bool{}
			for _, m := range typeDecl.FindAllStringSubmatch(source, -1) {
				types[m[1]] = true
			}
			withIdentity := map[string]bool{}
			for _, m := range codingMethod.FindAllStringSubmatch(source, -1) {
				withIdentity[m[1]] = true
			}

			if len(types) < 100 {
				t.Fatalf("only %d enums found; the check is not seeing the generated code", len(types))
			}
			var missing []string
			for name := range types {
				if !withIdentity[name] {
					missing = append(missing, name)
				}
			}
			if len(missing) > 0 {
				t.Errorf("%d enums have no Coding(): %v", len(missing), missing)
			}
			t.Logf("%d enums, all carrying display, system, Coding, Values and IsValid", len(types))
		})
	}
}

func TestEnumIdentityInEveryVersion(t *testing.T) {
	t.Run("r4b", func(t *testing.T) {
		if got := r4b.AdministrativeGenderFemale.Display(); got != "Female" {
			t.Errorf("Display() = %q", got)
		}
	})
	t.Run("r5", func(t *testing.T) {
		if got := r5.AdministrativeGenderFemale.Display(); got != "Female" {
			t.Errorf("Display() = %q", got)
		}
	})
}

// readGenerated returns the contents of a generated file, so a check can look at
// what was emitted rather than only at what one hand-picked type does.
func readGenerated(t *testing.T, module, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("..", module, name))
	if err != nil {
		t.Fatalf("reading %s/%s: %v", module, name, err)
	}
	return string(data)
}

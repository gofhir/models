package conformance

// Four defects that all shared a shape: the library accepted something wrong and
// carried on, so the caller got a plausible-looking result instead of an error.
//
// Each fix changes behavior rather than a signature, which is why they belong in
// the major: code that today gets a silently wrong answer will start getting an
// error, and that cannot be introduced in a minor.

import (
	"encoding/json"
	"encoding/xml"
	"strings"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

func TestDecodingRejectsAMismatchedResourceType(t *testing.T) {
	// A Practitioner document decoded into a Patient used to be accepted in
	// silence and written back out as a Patient: the type said one thing, the data
	// said another, and the conflict was resolved in favor of whichever the
	// caller happened to pass.
	//
	// UnmarshalResource always validated this, but json.Unmarshal into a known
	// type does not go through the dispatcher — and that is the ordinary way to
	// decode when you already know what you are expecting.
	var p r4.Patient
	err := json.Unmarshal([]byte(`{"resourceType":"Practitioner","id":"x"}`), &p)
	if err == nil {
		out, _ := json.Marshal(&p) //nolint:errcheck // only used to show what was decoded
		t.Fatalf("a Practitioner decoded into a Patient without complaint, and came back as %s", out)
	}
	for _, want := range []string{"Practitioner", "Patient"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("the error does not name %s, so it does not say what the conflict is: %v", want, err)
		}
	}
}

func TestMismatchedResourceTypeIsRejectedInEveryVersion(t *testing.T) {
	const input = `{"resourceType":"Observation","id":"x"}`

	t.Run("r4b", func(t *testing.T) {
		var p r4b.Patient
		if err := json.Unmarshal([]byte(input), &p); err == nil {
			t.Error("an Observation was accepted as a Patient")
		}
	})
	t.Run("r5", func(t *testing.T) {
		var p r5.Patient
		if err := json.Unmarshal([]byte(input), &p); err == nil {
			t.Error("an Observation was accepted as a Patient")
		}
	})
}

func TestXMLRejectsAValueWrittenAsElementText(t *testing.T) {
	// FHIR puts a primitive's value in the value attribute: <id value="p1"/>. The
	// text form was skipped without a word, so the whole document decoded to an
	// empty resource and reported success — the worst possible combination.
	const input = `<Patient xmlns="http://hl7.org/fhir"><id>p1</id><active>true</active></Patient>`

	var p r4.Patient
	err := xml.Unmarshal([]byte(input), &p)
	if err == nil {
		out, _ := json.Marshal(&p) //nolint:errcheck // only used to show what was decoded
		t.Fatalf("XML with values as element text was accepted and produced %s", out)
	}
	if !strings.Contains(err.Error(), "value attribute") {
		t.Errorf("the error does not say what the correct form is: %v", err)
	}
}

func TestXMLStillAcceptsTheFormsThatAreValid(t *testing.T) {
	// The check must not catch whitespace between child elements, nor a primitive
	// that carries an extension instead of a value — which is exactly how a
	// data-absent-reason is expressed.
	const input = `<Patient xmlns="http://hl7.org/fhir">
		<id value="p1"/>
		<active value="true"/>
		<birthDate>
			<extension url="http://hl7.org/fhir/StructureDefinition/data-absent-reason">
				<valueCode value="asked-declined"/>
			</extension>
		</birthDate>
	</Patient>`

	var p r4.Patient
	if err := xml.Unmarshal([]byte(input), &p); err != nil {
		t.Fatalf("valid FHIR XML was rejected: %v", err)
	}
	if p.Id == nil || *p.Id != "p1" {
		t.Error("id was lost")
	}
	if p.Active == nil || !*p.Active {
		t.Error("active was lost")
	}
	if p.BirthDate != nil {
		t.Error("birthDate has no value and must stay nil")
	}
	if p.BirthDateExt == nil || len(p.BirthDateExt.Extension) != 1 {
		t.Error("the data-absent-reason on birthDate was lost")
	}
}

func TestAllResourceTypesIsSorted(t *testing.T) {
	// It returned map iteration order, so two calls in one process gave two
	// different orders and anything printing, diffing or fixturing the list was
	// non-deterministic through no fault of its own.
	for name, got := range map[string][]string{
		"r4":  r4.AllResourceTypes(),
		"r4b": r4b.AllResourceTypes(),
		"r5":  r5.AllResourceTypes(),
	} {
		if len(got) < 100 {
			t.Errorf("%s: only %d resource types; the check is not seeing the registry", name, len(got))
			continue
		}
		for i := 1; i < len(got); i++ {
			if got[i-1] > got[i] {
				t.Errorf("%s: not sorted at %d: %q then %q", name, i, got[i-1], got[i])
				break
			}
		}
	}

	// Two calls must agree, which is the property callers actually depend on.
	a, b := r4.AllResourceTypes(), r4.AllResourceTypes()
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("two calls disagree at index %d: %q vs %q", i, a[i], b[i])
		}
	}
}

func TestSummaryFieldsCannotBeMutatedByCallers(t *testing.T) {
	// The map behind these was exported, so any consumer could rewrite the
	// specification's own data for every other consumer in the process — and
	// concurrent access to it was a plain data race.
	//
	// The map is now unreachable; what remains to check is that the accessor does
	// not hand out the stored slice, which would leave the same hazard in place.
	fields := r4.GetSummaryFields("Patient")
	if len(fields) == 0 {
		t.Fatal("Patient has no summary fields; the check is not seeing the data")
	}

	original := fields[0]
	fields[0] = "clobbered"

	if again := r4.GetSummaryFields("Patient"); again[0] != original {
		t.Errorf("mutating the returned slice changed the stored data: %q became %q",
			original, again[0])
	}
	if !r4.IsSummaryField("Patient", original) {
		t.Errorf("%q stopped being a summary field after a caller wrote to the returned slice", original)
	}
}

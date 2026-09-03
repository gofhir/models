package generator

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gofhir/models/internal/codegen/analyzer"
	"github.com/gofhir/models/internal/codegen/parser"
)

// -update rewrites the golden files instead of comparing against them. Review the
// resulting diff by hand: these fixtures are the only human-readable record of
// what the templates emit.
var update = flag.Bool("update", false, "rewrite golden files")

// The fixture bundles below are deliberately tiny and deliberately awkward. Each
// shape present here corresponds to a class of generator bug, so that a change in
// how any of them is emitted shows up as a readable diff rather than as forty
// thousand lines across r4, r4b and r5:
//
//   - a backbone element with a primitive child  -> missing _ext companion fields
//   - a choice type (value[x])                   -> choice handling and _ext suppression
//   - an array of primitives                     -> positional null alignment
//   - a required binding                         -> enum type resolution
//
// ValueSets sharing a FHIR name — the collision that gave Medication.status
// MedicationStatement's codes — have their own bundle and test further down,
// because that case now stops generation and so cannot be part of a fixture whose
// point is to generate successfully.

const goldenTypesBundle = `{
  "resourceType": "Bundle",
  "entry": [
    {
      "resource": {
        "resourceType": "StructureDefinition",
        "id": "boolean",
        "url": "http://hl7.org/fhir/StructureDefinition/boolean",
        "name": "boolean",
        "type": "boolean",
        "kind": "primitive-type",
        "fhirVersion": "4.0.1",
        "abstract": false,
        "snapshot": {"element": [{"path": "boolean", "min": 0, "max": "1"}]}
      }
    },
    {
      "resource": {
        "resourceType": "StructureDefinition",
        "id": "string",
        "url": "http://hl7.org/fhir/StructureDefinition/string",
        "name": "string",
        "type": "string",
        "kind": "primitive-type",
        "fhirVersion": "4.0.1",
        "abstract": false,
        "snapshot": {"element": [{"path": "string", "min": 0, "max": "1"}]}
      }
    },
    {
      "resource": {
        "resourceType": "StructureDefinition",
        "id": "code",
        "url": "http://hl7.org/fhir/StructureDefinition/code",
        "name": "code",
        "type": "code",
        "kind": "primitive-type",
        "fhirVersion": "4.0.1",
        "abstract": false,
        "snapshot": {"element": [{"path": "code", "min": 0, "max": "1"}]}
      }
    },
    {
      "resource": {
        "resourceType": "StructureDefinition",
        "id": "Element",
        "url": "http://hl7.org/fhir/StructureDefinition/Element",
        "name": "Element",
        "type": "Element",
        "kind": "complex-type",
        "fhirVersion": "4.0.1",
        "abstract": true,
        "snapshot": {
          "element": [
            {"path": "Element", "min": 0, "max": "*"},
            {"path": "Element.id", "min": 0, "max": "1", "type": [{"code": "string"}]}
          ]
        }
      }
    }
  ]
}`

const goldenResourcesBundle = `{
  "resourceType": "Bundle",
  "entry": [
    {
      "resource": {
        "resourceType": "StructureDefinition",
        "id": "Probe",
        "url": "http://hl7.org/fhir/StructureDefinition/Probe",
        "name": "Probe",
        "type": "Probe",
        "kind": "resource",
        "fhirVersion": "4.0.1",
        "abstract": false,
        "snapshot": {
          "element": [
            {"path": "Probe", "min": 0, "max": "*"},
            {"path": "Probe.id", "min": 0, "max": "1", "type": [{"code": "string"}]},
            {
              "path": "Probe.status", "min": 1, "max": "1",
              "short": "Status of the probe",
              "type": [{"code": "code"}],
              "binding": {
                "strength": "required",
                "valueSet": "http://hl7.org/fhir/ValueSet/probe-status|4.0.1"
              }
            },
            {
              "path": "Probe.alias", "min": 0, "max": "*",
              "short": "Repeating primitive",
              "type": [{"code": "string"}]
            },
            {
              "path": "Probe.marker", "min": 1, "max": "1",
              "short": "Required complex type: still a pointer, or an unset one marshals as {} and violates ele-1",
              "type": [{"code": "Element"}]
            },
            {
              "path": "Probe.contained", "min": 0, "max": "*",
              "short": "Contained resources: dispatched by ContainedList, not by a generated method",
              "type": [{"code": "Resource"}]
            },
            {
              "path": "Probe.value[x]", "min": 0, "max": "1",
              "short": "Choice of value",
              "type": [{"code": "boolean"}, {"code": "string"}]
            },
            {
              "path": "Probe.detail", "min": 0, "max": "*",
              "short": "Backbone with a primitive child",
              "type": [{"code": "BackboneElement"}]
            },
            {
              "path": "Probe.detail.label", "min": 0, "max": "1",
              "short": "Primitive inside a backbone",
              "type": [{"code": "string"}]
            }
          ]
        }
      }
    },
    {
      "resource": {
        "resourceType": "StructureDefinition",
        "id": "Sample",
        "url": "http://hl7.org/fhir/StructureDefinition/Sample",
        "name": "Sample",
        "type": "Sample",
        "kind": "resource",
        "fhirVersion": "4.0.1",
        "abstract": false,
        "snapshot": {
          "element": [
            {"path": "Sample", "min": 0, "max": "*"},
            {
              "path": "Sample.status", "min": 1, "max": "1",
              "short": "Status of the sample",
              "type": [{"code": "code"}],
              "binding": {
                "strength": "required",
                "valueSet": "http://hl7.org/fhir/ValueSet/sample-status|4.0.1"
              }
            }
          ]
        }
      }
    }
  ]
}`

// Two ValueSets with distinct names, so the golden fixture exercises ordinary
// generation. The colliding case has its own bundle and its own test below.
const goldenValueSetsBundle = `{
  "resourceType": "Bundle",
  "entry": [
    {
      "resource": {
        "resourceType": "ValueSet",
        "id": "probe-status",
        "url": "http://hl7.org/fhir/ValueSet/probe-status",
        "name": "Probe Status Codes",
        "compose": {
          "include": [
            {
              "system": "http://hl7.org/fhir/probe-status",
              "concept": [
                {"code": "active", "display": "Active"},
                {"code": "inactive", "display": "Inactive"}
              ]
            }
          ]
        }
      }
    },
    {
      "resource": {
        "resourceType": "ValueSet",
        "id": "sample-status",
        "url": "http://hl7.org/fhir/ValueSet/sample-status",
        "name": "Sample Status Codes",
        "compose": {
          "include": [
            {
              "system": "http://hl7.org/fhir/sample-status",
              "concept": [
                {"code": "draft", "display": "Draft"},
                {"code": "final", "display": "Final"}
              ]
            }
          ]
        }
      }
    }
  ]
}`

// generateGolden runs the full pipeline into a temp directory and returns it.
func generateGolden(t *testing.T) string {
	t.Helper()

	specs := writeSpecs(t, map[string]string{
		"profiles-types.json":     goldenTypesBundle,
		"profiles-resources.json": goldenResourcesBundle,
		"valuesets.json":          goldenValueSetsBundle,
	})
	out := t.TempDir()

	gen := New(Config{
		SpecsDir:    specs,
		OutputDir:   out,
		PackageName: "probe",
		Version:     "r4",
	})
	if err := gen.LoadTypes(); err != nil {
		t.Fatalf("LoadTypes: %v", err)
	}
	if err := gen.Generate(); err != nil {
		t.Fatalf("Generate: %v", err)
	}
	return out
}

// TestGolden pins the generated output for the fixture bundles. It is the only
// test that reads what the templates actually emit, so it is the net that catches
// unintended changes to generated code before they reach r4/, r4b/ and r5/.
func TestGolden(t *testing.T) {
	out := generateGolden(t)

	// Only the small, high-signal files are pinned. Pinning the consolidated
	// datatypes would make the fixture unreadable without adding coverage.
	for _, name := range []string{"codesystems.go", "resource_probe.go"} {
		t.Run(name, func(t *testing.T) {
			got, err := os.ReadFile(filepath.Join(out, name))
			if err != nil {
				t.Fatalf("reading generated %s: %v", name, err)
			}

			goldenPath := filepath.Join("testdata", "golden", name+".golden")
			if *update {
				if err = os.MkdirAll(filepath.Dir(goldenPath), 0o755); err != nil {
					t.Fatalf("mkdir testdata: %v", err)
				}
				if err = os.WriteFile(goldenPath, got, 0o644); err != nil {
					t.Fatalf("writing golden: %v", err)
				}
				t.Logf("updated %s", goldenPath)
				return
			}

			want, err := os.ReadFile(goldenPath)
			if err != nil {
				t.Fatalf("reading golden %s: %v (run: go test ./internal/codegen/generator -update)", goldenPath, err)
			}
			if !bytes.Equal(got, want) {
				t.Errorf("generated %s differs from golden.\nIf the change is intended, run:\n"+
					"  go test ./internal/codegen/generator -update\nand review the diff.\n\n%s",
					name, firstDiff(string(want), string(got)))
			}
		})
	}
}

// TestGoldenValueSetCollision states the invariant the collision bug violated:
// when two ValueSets share a FHIR name, the generator must not silently emit one
// and bind the other's fields to it.
//
// collidingValueSetsBundle mirrors the real R4 pair medication-status /
// medication-statement-status: different URLs, identical FHIR `name`, both
// sanitizing to the same Go identifier.
const collidingValueSetsBundle = `{
  "resourceType": "Bundle",
  "entry": [
    {
      "resource": {
        "resourceType": "ValueSet",
        "id": "probe-status",
        "url": "http://hl7.org/fhir/ValueSet/probe-status",
        "name": "Probe Status Codes",
        "compose": {
          "include": [
            {
              "system": "http://hl7.org/fhir/probe-status",
              "concept": [{"code": "active", "display": "Active"}]
            }
          ]
        }
      }
    },
    {
      "resource": {
        "resourceType": "ValueSet",
        "id": "sample-status",
        "url": "http://hl7.org/fhir/ValueSet/sample-status",
        "name": "Probe Status Codes",
        "compose": {
          "include": [
            {
              "system": "http://hl7.org/fhir/sample-status",
              "concept": [{"code": "draft", "display": "Draft"}]
            }
          ]
        }
      }
    }
  ]
}`

// TestValueSetCollisionFailsGeneration pins the resolution of the bug that this
// fixture was built to expose.
//
// Two ValueSets sharing a FHIR name used to be resolved by emitting whichever came
// first and dropping the other with a bare `continue` — while the analyzer had
// already pointed the dropped ValueSet's fields at the surviving type. In R4 that
// gave Medication.status MedicationStatement's codes, with no constant for
// `inactive`, its only other legal value.
//
// Generation now stops instead, naming both URLs, so a new collision in a future
// release is resolved deliberately rather than by bundle order.
func TestValueSetCollisionFailsGeneration(t *testing.T) {
	specs := writeSpecs(t, map[string]string{
		"profiles-types.json":     goldenTypesBundle,
		"profiles-resources.json": goldenResourcesBundle,
		"valuesets.json":          collidingValueSetsBundle,
	})

	gen := New(Config{
		SpecsDir:    specs,
		OutputDir:   t.TempDir(),
		PackageName: "probe",
		Version:     "r4",
	})
	if err := gen.LoadTypes(); err != nil {
		t.Fatalf("LoadTypes: %v", err)
	}

	err := gen.Generate()
	if err == nil {
		t.Fatal("generation succeeded despite two ValueSets mapping to one Go type;" +
			" one of them was silently dropped")
	}

	// The message has to identify both sides, or whoever hits this has no way to
	// tell which ValueSets are involved.
	for _, want := range []string{
		"ProbeStatusCodes",
		"http://hl7.org/fhir/ValueSet/probe-status",
		"http://hl7.org/fhir/ValueSet/sample-status",
		"valueSetTypeNameOverrides",
	} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error does not mention %q:\n%v", want, err)
		}
	}
}

// TestValueSetOverrideAppliesOnlyOnCollision covers the condition under which the
// R4 override fires.
//
// The first version keyed overrides on the URL alone, which renamed
// medication-status in every version. But the clash only exists in R4: upstream
// renamed one side to "MedicationStatement Status Codes" in R4B and to
// "MedicationStatementStatusCodes" in R5. The URL-keyed version therefore deleted
// the exported MedicationStatusCodes type from r4b and r5, two packages that never
// had the problem.
//
// The generated packages no longer contain MedicationStatusCodes at all — binding
// names call it MedicationStatus in every version, which is what the specification
// says it is. That is a separate, deliberate rename, and it does not make the
// condition tested here uninteresting: the override must still fire only on a real
// collision, or the next one added will quietly rename a type everywhere. The
// analyzer here is built without StructureDefinitions, so no binding name is in
// play and the override mechanism is exercised on its own.
func TestValueSetOverrideAppliesOnlyOnCollision(t *testing.T) {
	const (
		medicationURL = "http://hl7.org/fhir/ValueSet/medication-status"
		statementURL  = "http://hl7.org/fhir/ValueSet/medication-statement-status"
	)

	// valueSetBundle builds a ValueSet bundle from url/name pairs, each with one
	// code so it counts as a candidate for enum generation.
	valueSetBundle := func(pairs ...[2]string) string {
		entries := make([]string, 0, len(pairs))
		for i, p := range pairs {
			entries = append(entries, fmt.Sprintf(`{"resource":{"resourceType":"ValueSet",`+
				`"id":"vs%d","url":%q,"name":%q,"compose":{"include":[{"system":"http://example.org/s%d",`+
				`"concept":[{"code":"active","display":"Active"}]}]}}}`, i, p[0], p[1], i))
		}
		return `{"resourceType":"Bundle","entry":[` + strings.Join(entries, ",") + `]}`
	}

	newAnalyzerFor := func(t *testing.T, bundle string) *analyzer.Analyzer {
		t.Helper()
		registry := parser.NewValueSetRegistry()
		if err := registry.LoadFromBundle([]byte(bundle)); err != nil {
			t.Fatalf("loading value sets: %v", err)
		}
		return analyzer.NewAnalyzer(nil, registry)
	}

	t.Run("r4 shape: names collide, override applies", func(t *testing.T) {
		a := newAnalyzerFor(t, valueSetBundle(
			[2]string{statementURL, "Medication Status Codes"},
			[2]string{medicationURL, "Medication Status Codes"},
		))

		statement := a.ValueSetTypeName(statementURL, "Medication Status Codes")
		medication := a.ValueSetTypeName(medicationURL, "Medication Status Codes")

		if statement == medication {
			t.Fatalf("both ValueSets still resolve to %s", statement)
		}
		// The surviving name is deliberately unchanged, so existing constants keep
		// working; only the shadowed ValueSet is renamed.
		if statement != "MedicationStatusCodes" {
			t.Errorf("MedicationStatement's ValueSet became %s; renaming it breaks"+
				" existing constants for no reason", statement)
		}
		if medication != "MedicationStatus" {
			t.Errorf("Medication's ValueSet resolved to %s, want MedicationStatus", medication)
		}

		// The version suffix bindings carry must not defeat the override.
		if got := a.ValueSetTypeName(medicationURL+"|4.0.1", "Medication Status Codes"); got != "MedicationStatus" {
			t.Errorf("versioned URL resolved to %s; the |version suffix is not being stripped", got)
		}
	})

	t.Run("r4b/r5 shape: names differ, override must not fire", func(t *testing.T) {
		a := newAnalyzerFor(t, valueSetBundle(
			[2]string{statementURL, "MedicationStatement Status Codes"},
			[2]string{medicationURL, "Medication Status Codes"},
		))

		got := a.ValueSetTypeName(medicationURL, "Medication Status Codes")
		if got != "MedicationStatusCodes" {
			t.Errorf("medication-status resolved to %s where nothing collides with it;"+
				" that removes MedicationStatusCodes from packages that never had"+
				" the collision", got)
		}
	})
}

// firstDiff reports the first differing line, which is far more useful in test
// output than dumping two multi-kilobyte files.
func firstDiff(want, got string) string {
	wantLines := strings.Split(want, "\n")
	gotLines := strings.Split(got, "\n")
	for i := 0; i < len(wantLines) || i < len(gotLines); i++ {
		var w, g string
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if w != g {
			return "first difference at line " + strconv.Itoa(i+1) + ":\n  golden: " + w + "\n  got:    " + g
		}
	}
	return "files differ only in trailing content"
}

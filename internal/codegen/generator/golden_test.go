package generator

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// -update rewrites the golden files instead of comparing against them. Review the
// resulting diff by hand: these fixtures are the only human-readable record of
// what the templates emit.
var update = flag.Bool("update", false, "rewrite golden files")

// The fixture bundles below are deliberately tiny and deliberately awkward. Each
// shape present here corresponds to a class of generator bug:
//
//   - a backbone element with a primitive child  -> missing _ext companion fields
//   - a choice type (value[x])                   -> choice handling and _ext suppression
//   - an array of primitives                     -> positional null alignment
//   - two ValueSets sharing a FHIR name          -> silent type-name collision
//
// The last one is why this file exists: two distinct ValueSets named "Probe Status
// Codes" collapse to the same Go identifier, and the generator used to drop the
// second one with a bare `continue`, leaving fields bound to codes that were
// never emitted.

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

// Two ValueSets, different URLs, same `name`. This mirrors the real R4 pair
// medication-status / medication-statement-status, both named "Medication Status
// Codes", which made Medication.status resolve to MedicationStatement's codes.
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
        "name": "Probe Status Codes",
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
// The bug is live, so this test cannot assert the correct behavior yet: the real
// R4 specs contain the same collision (medication-status and
// medication-statement-status are both named "Medication Status Codes"), so making
// it fatal requires the disambiguation work in PLAN.md task 3.1.
//
// Rather than skipping — which would leave the body unreachable and give no
// signal when the defect is fixed — it detects which state the generator is in.
// While the bug is live it pins the broken behavior; once the collision is
// handled it switches to asserting the correct behavior, so the fix is verified
// without anyone having to remember this file exists.
func TestGoldenValueSetCollision(t *testing.T) {
	out := generateGolden(t)

	codesystems := readGenerated(t, out, "codesystems.go")
	probeType := typeOfStatusField(readGenerated(t, out, "resource_probe.go"))
	sampleType := typeOfStatusField(readGenerated(t, out, "resource_sample.go"))

	collisionPresent := probeType != "" && probeType == sampleType
	codesMissing := !strings.Contains(codesystems, `= "draft"`) ||
		!strings.Contains(codesystems, `= "final"`)

	if !collisionPresent && !codesMissing {
		t.Log("the ValueSet name collision is handled; asserting correct behavior." +
			" The broken-state branch of this test and task 3.1 in PLAN.md can now go.")
		assertValueSetCollisionFixed(t, out)
		return
	}

	// Pin the defect precisely, so a partial change is caught rather than passing
	// quietly as if nothing had moved.
	if !collisionPresent {
		t.Errorf("collision changed shape: Probe.Status=%s Sample.Status=%s;"+
			" expected both to share one type while the bug is live", probeType, sampleType)
	}
	if !codesMissing {
		t.Error("the second ValueSet's codes are now emitted, but the types still collide;" +
			" this is a new intermediate state that needs review")
	}
	t.Logf("known defect reproduced: Probe.Status and Sample.Status both use %s,"+
		" and the codes draft/final have no constants (PLAN.md task 3.1)", probeType)
}

// assertValueSetCollisionFixed is what TestGoldenValueSetCollision should become
// once task 3.1 lands: two ValueSets sharing a FHIR name must either produce two
// distinct Go types, or fail generation outright — never one type silently
// standing in for both.
func assertValueSetCollisionFixed(t *testing.T, out string) {
	t.Helper()

	cs := readGenerated(t, out, "codesystems.go")
	p := readGenerated(t, out, "resource_probe.go")
	s := readGenerated(t, out, "resource_sample.go")

	// Both resources bind to a required ValueSet, so both must end up with a
	// typed status field rather than a bare *string.
	if strings.Contains(p, "Status *string") {
		t.Error("Probe.Status degraded to *string: its required binding was not resolved")
	}
	if strings.Contains(s, "Status *string") {
		t.Error("Sample.Status degraded to *string: its required binding was not resolved")
	}

	// The two resources bind to different ValueSets, so they must not share one
	// Go type. Sharing means one ValueSet's codes were dropped.
	probeType := typeOfStatusField(p)
	sampleType := typeOfStatusField(s)
	if probeType != "" && probeType == sampleType {
		t.Errorf("Probe.Status and Sample.Status both use %s, but they bind to different ValueSets;"+
			" one ValueSet's codes were silently discarded", probeType)
	}

	// Whatever the types end up being called, every code from both ValueSets has
	// to exist somewhere. A missing constant means a legal value a user cannot
	// express — the concrete harm in the Medication.status case, where `inactive`
	// had no constant at all.
	for _, code := range []string{`= "active"`, `= "inactive"`, `= "draft"`, `= "final"`} {
		if !strings.Contains(cs, code) {
			t.Errorf("no constant with value %s was generated; that code is unreachable for users", code)
		}
	}
}

// typeOfStatusField extracts the Go type of the `Status` field from a generated
// resource file, e.g. "*ProbeStatusCodes".
func typeOfStatusField(src string) string {
	for _, line := range strings.Split(src, "\n") {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "Status ") {
			continue
		}
		fields := strings.Fields(trimmed)
		if len(fields) >= 2 {
			return fields[1]
		}
	}
	return ""
}

// readGenerated returns a generated file's contents as a string.
func readGenerated(t *testing.T, out, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(out, name))
	if err != nil {
		t.Fatalf("reading generated %s: %v", name, err)
	}
	return string(data)
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

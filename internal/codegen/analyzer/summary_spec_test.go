package analyzer

// isSummary was copied onto a property in exactly one of the four places that
// build them, so summary.go was missing every choice element, every
// backbone-typed element and every contentReference element: Observation had no
// value and no effective, and no component either, though the spec marks all
// three as summary elements.
//
// This compares the analyzer's answer against the specification itself rather
// than against a recorded fixture. A fixture would have been written from the
// same misreading and agreed with it.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/gofhir/models/internal/codegen/parser"
)

// summaryFieldsFromSpec reads the first-level summary elements straight out of a
// StructureDefinition's snapshot, expanding a choice into the concrete variant
// names a document would actually carry.
func summaryFieldsFromSpec(sd map[string]any) map[string]bool {
	out := map[string]bool{}
	snapshot, _ := sd["snapshot"].(map[string]any)
	elements, _ := snapshot["element"].([]any)
	for _, e := range elements {
		el, _ := e.(map[string]any)
		path, _ := el["path"].(string)
		if strings.Count(path, ".") != 1 {
			continue
		}
		if summary, _ := el["isSummary"].(bool); !summary {
			continue
		}
		name := path[strings.Index(path, ".")+1:]
		if !strings.HasSuffix(name, "[x]") {
			out[name] = true
			continue
		}
		base := strings.TrimSuffix(name, "[x]")
		types, _ := el["type"].([]any)
		for _, t := range types {
			tm, _ := t.(map[string]any)
			code, _ := tm["code"].(string)
			if code == "" {
				continue
			}
			out[base+strings.ToUpper(code[:1])+code[1:]] = true
		}
	}
	return out
}

func TestSummaryFlagsMatchTheSpecification(t *testing.T) {
	for _, version := range []string{"r4", "r4b", "r5"} {
		t.Run(version, func(t *testing.T) {
			path := filepath.Join("..", "..", "..", "specs", version, "profiles-resources.json")
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Skipf("specs not fetched; see scripts/fetch-specs.sh (%v)", err)
			}

			var bundle struct {
				Entry []struct {
					Resource json.RawMessage `json:"resource"`
				} `json:"entry"`
			}
			if err := json.Unmarshal(raw, &bundle); err != nil {
				t.Fatalf("parse bundle: %v", err)
			}

			// Every concrete resource in the bundle, analyzed the way the
			// generator analyzes it.
			expected := map[string]map[string]bool{}
			sds := make([]*parser.StructureDefinition, 0, len(bundle.Entry))
			for _, entry := range bundle.Entry {
				var head map[string]any
				if err := json.Unmarshal(entry.Resource, &head); err != nil {
					continue
				}
				if head["resourceType"] != "StructureDefinition" ||
					head["kind"] != "resource" ||
					head["derivation"] != "specialization" {
					continue
				}
				if abstract, _ := head["abstract"].(bool); abstract {
					continue
				}
				sd, err := parser.ParseStructureDefinition(entry.Resource)
				if err != nil {
					continue
				}
				sds = append(sds, sd)
				name, _ := head["name"].(string)
				expected[name] = summaryFieldsFromSpec(head)
			}
			if len(sds) == 0 {
				t.Fatal("no resource was read from the bundle")
			}

			a := NewAnalyzer(sds, nil)

			checked := 0
			for _, sd := range sds {
				at, err := a.Analyze(sd)
				if err != nil {
					t.Errorf("analyze %s: %v", sd.Name, err)
					continue
				}
				want, ok := expected[at.Name]
				if !ok {
					continue
				}
				got := map[string]bool{}
				for _, prop := range at.Properties {
					if prop.IsSummary {
						got[prop.JSONName] = true
					}
				}
				checked++
				if missing := diff(want, got); len(missing) > 0 {
					t.Errorf("%s: the spec marks these as summary elements and the analyzer does not: %s",
						at.Name, strings.Join(missing, ", "))
				}
				if extra := diff(got, want); len(extra) > 0 {
					t.Errorf("%s: the analyzer marks these as summary elements and the spec does not: %s",
						at.Name, strings.Join(extra, ", "))
				}
			}
			if checked == 0 {
				t.Fatal("no resource was actually compared")
			}
			t.Logf("%d resources agree with the specification", checked)
		})
	}
}

// TestSummaryFlagsReachNestedBackboneProperties covers the one place the
// comparison above cannot: it walks a resource's own properties, and a backbone
// nested inside another backbone belongs to the inner type.
//
// Bundle.entry is the case. Its search, request and response are all
// BackboneElement-typed and all isSummary — 41 such elements exist in R4 alone.
// An earlier version of this test used Observation.component.code, which is a
// CodeableConcept and reaches the analyzer through a different path, so it passed
// with the flag removed and proved nothing.
//
// Nothing reads the flag on a nested property today: template_loader is its only
// consumer and it visits resource-level properties only. This guards the fact
// rather than an output, because an untested flag is what produced the defect.
func TestSummaryFlagsReachNestedBackboneProperties(t *testing.T) {
	raw, readErr := os.ReadFile(filepath.Join("..", "..", "..", "specs", "r4", "profiles-resources.json"))
	if readErr != nil {
		t.Skipf("specs not fetched; see scripts/fetch-specs.sh (%v)", readErr)
	}

	var bundle struct {
		Entry []struct {
			Resource json.RawMessage `json:"resource"`
		} `json:"entry"`
	}
	if unmarshalErr := json.Unmarshal(raw, &bundle); unmarshalErr != nil {
		t.Fatalf("parse bundle: %v", unmarshalErr)
	}

	var bundleSD *parser.StructureDefinition
	for _, entry := range bundle.Entry {
		var head map[string]any
		if headErr := json.Unmarshal(entry.Resource, &head); headErr != nil {
			continue
		}
		if head["resourceType"] != "StructureDefinition" || head["name"] != "Bundle" {
			continue
		}
		sd, parseErr := parser.ParseStructureDefinition(entry.Resource)
		if parseErr != nil {
			t.Fatalf("parse Bundle: %v", parseErr)
		}
		bundleSD = sd
		break
	}
	if bundleSD == nil {
		t.Fatal("Bundle is not in the bundle")
	}

	at, err := NewAnalyzer([]*parser.StructureDefinition{bundleSD}, nil).Analyze(bundleSD)
	if err != nil {
		t.Fatalf("analyze: %v", err)
	}

	var bundleEntry *AnalyzedType
	for _, b := range at.BackboneTypes {
		if b.Name == "BundleEntry" {
			bundleEntry = b
			break
		}
	}
	if bundleEntry == nil {
		t.Fatal("BundleEntry was not analyzed")
	}

	summary := map[string]bool{}
	for _, prop := range bundleEntry.Properties {
		if prop.IsSummary {
			summary[prop.JSONName] = true
		}
	}

	// Straight from the R4 snapshot.
	for _, want := range []string{"search", "request", "response", "link", "fullUrl"} {
		if !summary[want] {
			t.Errorf("Bundle.entry.%s is a summary element in the spec", want)
		}
	}
	for _, notWant := range []string{"id", "extension"} {
		if summary[notWant] {
			t.Errorf("Bundle.entry.%s is not a summary element in the spec", notWant)
		}
	}
}

func diff(a, b map[string]bool) []string {
	var only []string
	for k := range a {
		if !b[k] {
			only = append(only, k)
		}
	}
	sort.Strings(only)
	if len(only) > 12 {
		only = append(only[:12], fmt.Sprintf("and %d more", len(only)-12))
	}
	return only
}

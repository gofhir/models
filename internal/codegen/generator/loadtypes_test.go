package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// minimalBundle is the smallest spec bundle LoadTypes will accept: one primitive
// datatype so the analyzer has something to chew on, plus a fhirVersion so the
// release can be derived.
const minimalBundle = `{
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
    }
  ]
}`

// minimalValueSets carries one ValueSet with one code, which is enough for
// ValueSetRegistry.Count() to be non-zero.
const minimalValueSets = `{
  "resourceType": "Bundle",
  "entry": [
    {
      "resource": {
        "resourceType": "ValueSet",
        "id": "administrative-gender",
        "url": "http://hl7.org/fhir/ValueSet/administrative-gender",
        "name": "AdministrativeGender",
        "compose": {
          "include": [
            {
              "system": "http://hl7.org/fhir/administrative-gender",
              "concept": [{"code": "male", "display": "Male"}]
            }
          ]
        }
      }
    }
  ]
}`

// writeSpecs lays out a specs directory for version "r4". Any file whose content
// is the empty string is omitted, which is how the missing-file cases are built.
func writeSpecs(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	dir := filepath.Join(root, "r4")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	for name, content := range files {
		if content == "" {
			continue
		}
		if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o644); err != nil {
			t.Fatalf("write %s: %v", name, err)
		}
	}
	return root
}

func loadFrom(t *testing.T, specsDir string) error {
	t.Helper()
	return New(Config{
		SpecsDir:    specsDir,
		OutputDir:   t.TempDir(),
		PackageName: "r4",
		Version:     "r4",
	}).LoadTypes()
}

// TestLoadTypesRequiresValueSets is the regression test for the silent-degradation
// bug: with valuesets.json absent, generation used to succeed and every required
// binding quietly became *string, dropping ~205 exported enum types while the
// package still compiled.
func TestLoadTypesRequiresValueSets(t *testing.T) {
	tests := []struct {
		name       string
		files      map[string]string
		wantErrSub string
	}{
		{
			name: "missing valuesets.json is fatal",
			files: map[string]string{
				"profiles-types.json":     minimalBundle,
				"profiles-resources.json": minimalBundle,
				"valuesets.json":          "", // omitted
			},
			wantErrSub: "failed to read required value sets",
		},
		{
			name: "unparseable valuesets.json is fatal",
			files: map[string]string{
				"profiles-types.json":     minimalBundle,
				"profiles-resources.json": minimalBundle,
				"valuesets.json":          `{"resourceType": "Bundle", "entry": [`,
			},
			wantErrSub: "failed to load value sets",
		},
		{
			name: "valuesets.json with no usable value sets is fatal",
			files: map[string]string{
				"profiles-types.json":     minimalBundle,
				"profiles-resources.json": minimalBundle,
				"valuesets.json":          `{"resourceType": "Bundle", "entry": []}`,
			},
			wantErrSub: "no value sets parsed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := loadFrom(t, writeSpecs(t, tt.files))
			if err == nil {
				t.Fatal("LoadTypes succeeded; expected it to refuse to generate")
			}
			if !strings.Contains(err.Error(), tt.wantErrSub) {
				t.Errorf("error %q does not mention %q", err, tt.wantErrSub)
			}
		})
	}
}

// TestLoadTypesSucceedsWithCompleteSpecs guards the happy path, so the checks
// above cannot be satisfied by failing for unrelated reasons.
func TestLoadTypesSucceedsWithCompleteSpecs(t *testing.T) {
	err := loadFrom(t, writeSpecs(t, map[string]string{
		"profiles-types.json":     minimalBundle,
		"profiles-resources.json": minimalBundle,
		"valuesets.json":          minimalValueSets,
	}))
	if err != nil {
		t.Fatalf("LoadTypes failed on a complete spec set: %v", err)
	}
}

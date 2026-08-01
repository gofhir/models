package generator

import (
	"strings"
	"testing"

	"github.com/gofhir/models/internal/codegen/parser"
)

func TestDetectFHIRVersion(t *testing.T) {
	tests := []struct {
		name    string
		sds     []*parser.StructureDefinition
		want    string
		wantErr string
	}{
		{
			name: "single version across definitions",
			sds: []*parser.StructureDefinition{
				{Name: "Patient", FHIRVersion: "4.0.1"},
				{Name: "Observation", FHIRVersion: "4.0.1"},
			},
			want: "4.0.1",
		},
		{
			name: "definitions without a fhirVersion are ignored",
			sds: []*parser.StructureDefinition{
				{Name: "Patient", FHIRVersion: "5.0.0"},
				{Name: "SomeLogicalModel"},
			},
			want: "5.0.0",
		},
		{
			name: "version is not inferred from the artifact version",
			sds: []*parser.StructureDefinition{
				{Name: "Patient", Version: "4.0.1"},
			},
			wantErr: "no StructureDefinition declares a fhirVersion",
		},
		{
			name:    "no definitions",
			sds:     nil,
			wantErr: "no StructureDefinition declares a fhirVersion",
		},
		{
			name: "conflicting versions are fatal",
			sds: []*parser.StructureDefinition{
				{Name: "Patient", FHIRVersion: "4.0.1"},
				{Name: "Observation", FHIRVersion: "5.0.0"},
			},
			wantErr: "conflicting fhirVersion values: 4.0.1 (e.g. Patient), 5.0.0 (e.g. Observation)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := detectFHIRVersion(tt.sds)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected error containing %q, got version %q", tt.wantErr, got)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr, err.Error())
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("expected %q, got %q", tt.want, got)
			}
		})
	}
}

// The conflict message lists versions in a stable order regardless of input order,
// so generation failures are reproducible.
func TestDetectFHIRVersion_ConflictMessageIsDeterministic(t *testing.T) {
	_, first := detectFHIRVersion([]*parser.StructureDefinition{
		{Name: "Patient", FHIRVersion: "4.0.1"},
		{Name: "Observation", FHIRVersion: "5.0.0"},
	})
	_, reversed := detectFHIRVersion([]*parser.StructureDefinition{
		{Name: "Observation", FHIRVersion: "5.0.0"},
		{Name: "Patient", FHIRVersion: "4.0.1"},
	})

	if first == nil || reversed == nil {
		t.Fatal("expected both inputs to conflict")
	}
	if first.Error() != reversed.Error() {
		t.Errorf("message depends on input order:\n  %s\n  %s", first.Error(), reversed.Error())
	}
}

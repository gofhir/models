package r4b_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofhir/models/r4b"
)

func TestNewResource(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		wantErr      bool
	}{
		{
			name:         "Patient",
			resourceType: "Patient",
			wantErr:      false,
		},
		{
			name:         "Observation",
			resourceType: "Observation",
			wantErr:      false,
		},
		{
			name:         "Bundle",
			resourceType: "Bundle",
			wantErr:      false,
		},
		{
			name:         "Unknown resource type",
			resourceType: "UnknownResource",
			wantErr:      true,
		},
		{
			name:         "Empty resource type",
			resourceType: "",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := r4b.NewResource(tt.resourceType)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, resource)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resource)
				assert.Equal(t, tt.resourceType, resource.GetResourceType())
			}
		})
	}
}

func TestUnmarshalResource(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		wantType    string
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid Patient",
			json: `{
				"resourceType": "Patient",
				"id": "example",
				"active": true
			}`,
			wantType: "Patient",
			wantErr:  false,
		},
		{
			name: "Valid Observation",
			json: `{
				"resourceType": "Observation",
				"id": "obs-1",
				"status": "final"
			}`,
			wantType: "Observation",
			wantErr:  false,
		},
		{
			name:        "Missing resourceType",
			json:        `{"id": "example"}`,
			wantErr:     true,
			errContains: "missing or empty",
		},
		{
			name:        "Empty resourceType",
			json:        `{"resourceType": "", "id": "example"}`,
			wantErr:     true,
			errContains: "missing or empty",
		},
		{
			name:        "Unknown resourceType",
			json:        `{"resourceType": "UnknownResource", "id": "example"}`,
			wantErr:     true,
			errContains: "unknown resource type",
		},
		{
			name:        "Invalid JSON",
			json:        `{not valid json}`,
			wantErr:     true,
			errContains: "failed to parse JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource, err := r4b.UnmarshalResource([]byte(tt.json))
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, resource)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, resource)
				assert.Equal(t, tt.wantType, resource.GetResourceType())
			}
		})
	}
}

func TestUnmarshalResourceRoundTrip(t *testing.T) {
	// Create a Patient with some data
	original := &r4b.Patient{
		Id:     ptrString("test-123"),
		Active: ptrBool(true),
		Name: []r4b.HumanName{
			{
				Family: ptrString("Smith"),
				Given:  []string{"John"},
			},
		},
	}

	// Marshal to JSON
	data, err := json.Marshal(original)
	require.NoError(t, err)

	// Add resourceType (not auto-populated by marshal)
	var m map[string]interface{}
	require.NoError(t, json.Unmarshal(data, &m))
	m["resourceType"] = "Patient"
	data, err = json.Marshal(m)
	require.NoError(t, err)

	// Unmarshal using registry
	resource, err := r4b.UnmarshalResource(data)
	require.NoError(t, err)

	// Assert it's a Patient
	patient, ok := resource.(*r4b.Patient)
	require.True(t, ok, "expected *r4b.Patient, got %T", resource)

	// Verify data
	assert.Equal(t, "test-123", *patient.Id)
	assert.True(t, *patient.Active)
	require.Len(t, patient.Name, 1)
	assert.Equal(t, "Smith", *patient.Name[0].Family)
	assert.Equal(t, []string{"John"}, patient.Name[0].Given)
}

func TestGetResourceType(t *testing.T) {
	tests := []struct {
		name        string
		json        string
		want        string
		wantErr     bool
		errContains string
	}{
		{
			name: "Valid Patient",
			json: `{"resourceType": "Patient", "id": "example"}`,
			want: "Patient",
		},
		{
			name: "Valid Observation",
			json: `{"resourceType": "Observation"}`,
			want: "Observation",
		},
		{
			name:        "Missing resourceType",
			json:        `{"id": "example"}`,
			wantErr:     true,
			errContains: "missing or empty",
		},
		{
			name:        "Invalid JSON",
			json:        `not json`,
			wantErr:     true,
			errContains: "failed to parse JSON",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := r4b.GetResourceType([]byte(tt.json))
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestIsKnownResourceType(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		want         bool
	}{
		{"Patient is known", "Patient", true},
		{"Observation is known", "Observation", true},
		{"Bundle is known", "Bundle", true},
		{"UnknownResource is not known", "UnknownResource", false},
		{"Empty is not known", "", false},
		{"lowercase patient is not known", "patient", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := r4b.IsKnownResourceType(tt.resourceType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAllResourceTypes(t *testing.T) {
	types := r4b.AllResourceTypes()

	// Should have many resource types (R4B has ~150)
	assert.Greater(t, len(types), 100, "expected more than 100 resource types")

	// Should include common types
	typeSet := make(map[string]bool)
	for _, t := range types {
		typeSet[t] = true
	}

	assert.True(t, typeSet["Patient"], "should include Patient")
	assert.True(t, typeSet["Observation"], "should include Observation")
	assert.True(t, typeSet["Bundle"], "should include Bundle")
	assert.True(t, typeSet["Condition"], "should include Condition")
	assert.True(t, typeSet["Medication"], "should include Medication")
}

// Helper functions
func ptrString(s string) *string {
	return &s
}

func ptrBool(b bool) *bool {
	return &b
}

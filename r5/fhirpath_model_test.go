package r5

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFHIRPathModel_NotNil(t *testing.T) {
	require.NotNil(t, FHIRPathModel())
}

func TestFHIRPathModel_FHIRVersion(t *testing.T) {
	// Must match StructureDefinition.fhirVersion in specs/r5, not the package name.
	assert.Equal(t, "5.0.0", FHIRPathModel().FHIRVersion())
}

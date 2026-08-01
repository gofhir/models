package r4b

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFHIRPathModel_NotNil(t *testing.T) {
	require.NotNil(t, FHIRPathModel())
}

func TestFHIRPathModel_FHIRVersion(t *testing.T) {
	// Must match StructureDefinition.fhirVersion in specs/r4b, not the package name.
	assert.Equal(t, "4.3.0", FHIRPathModel().FHIRVersion())
}

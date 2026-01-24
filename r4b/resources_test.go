package r4b

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPatient(t *testing.T) {
	t.Run("create patient with basic fields", func(t *testing.T) {
		id := "patient-123"
		active := true
		gender := AdministrativeGenderMale
		birthDate := "1990-01-15"
		family := "Smith"

		patient := Patient{
			Id:        &id,
			Active:    &active,
			Gender:    &gender,
			BirthDate: &birthDate,
			Name: []HumanName{
				{Family: &family, Given: []string{"John"}},
			},
		}

		assert.Equal(t, "patient-123", *patient.Id)
		assert.True(t, *patient.Active)
		assert.Equal(t, AdministrativeGenderMale, *patient.Gender)
		assert.Equal(t, "1990-01-15", *patient.BirthDate)
		require.Len(t, patient.Name, 1)
		assert.Equal(t, "Smith", *patient.Name[0].Family)
	})

	t.Run("JSON round trip", func(t *testing.T) {
		id := "pt-json"
		family := "Johnson"

		original := Patient{
			Id:   &id,
			Name: []HumanName{{Family: &family}},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded Patient
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *original.Id, *decoded.Id)
		require.Len(t, decoded.Name, 1)
		assert.Equal(t, *original.Name[0].Family, *decoded.Name[0].Family)
	})

	t.Run("GetResourceType", func(t *testing.T) {
		patient := &Patient{}
		assert.Equal(t, "Patient", patient.GetResourceType())
	})
}

func TestObservation(t *testing.T) {
	t.Run("create observation with value quantity", func(t *testing.T) {
		id := "obs-123"
		status := ObservationStatusFinal
		value := 72.0
		unit := "bpm"

		obs := Observation{
			Id:     &id,
			Status: &status,
			Code:   CodeableConcept{},
			ValueQuantity: &Quantity{
				Value: &value,
				Unit:  &unit,
			},
		}

		assert.Equal(t, "obs-123", *obs.Id)
		assert.Equal(t, ObservationStatusFinal, *obs.Status)
		require.NotNil(t, obs.ValueQuantity)
		assert.Equal(t, 72.0, *obs.ValueQuantity.Value)
		assert.Equal(t, "bpm", *obs.ValueQuantity.Unit)
	})

	t.Run("GetResourceType", func(t *testing.T) {
		obs := &Observation{}
		assert.Equal(t, "Observation", obs.GetResourceType())
	})
}

func TestResourceInterface(t *testing.T) {
	t.Run("resources implement Resource interface", func(t *testing.T) {
		resources := []Resource{
			&Patient{},
			&Observation{},
			&Practitioner{},
			&Organization{},
			&Bundle{},
		}

		expectedTypes := []string{"Patient", "Observation", "Practitioner", "Organization", "Bundle"}

		for i, r := range resources {
			assert.Equal(t, expectedTypes[i], r.GetResourceType())
		}
	})
}

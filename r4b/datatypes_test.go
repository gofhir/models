package r4b

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddress(t *testing.T) {
	t.Run("create address with all fields", func(t *testing.T) {
		use := AddressUseHome
		addrType := AddressTypePhysical
		text := "123 Main St, Springfield"
		city := "Springfield"
		state := "IL"
		postalCode := "62701"
		country := "USA"

		addr := Address{
			Use:        &use,
			Type:       &addrType,
			Text:       &text,
			Line:       []string{"123 Main St"},
			City:       &city,
			State:      &state,
			PostalCode: &postalCode,
			Country:    &country,
		}

		assert.Equal(t, AddressUseHome, *addr.Use)
		assert.Equal(t, AddressTypePhysical, *addr.Type)
		assert.Equal(t, "123 Main St, Springfield", *addr.Text)
		assert.Equal(t, []string{"123 Main St"}, addr.Line)
		assert.Equal(t, "Springfield", *addr.City)
		assert.Equal(t, "IL", *addr.State)
		assert.Equal(t, "62701", *addr.PostalCode)
		assert.Equal(t, "USA", *addr.Country)
	})

	t.Run("JSON marshal/unmarshal", func(t *testing.T) {
		use := AddressUseWork
		city := "Chicago"

		addr := Address{
			Use:  &use,
			City: &city,
			Line: []string{"456 Office Blvd", "Suite 100"},
		}

		data, err := json.Marshal(addr)
		require.NoError(t, err)

		var decoded Address
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, AddressUseWork, *decoded.Use)
		assert.Equal(t, "Chicago", *decoded.City)
		assert.Equal(t, []string{"456 Office Blvd", "Suite 100"}, decoded.Line)
	})

	t.Run("omitempty works correctly", func(t *testing.T) {
		addr := Address{}
		data, err := json.Marshal(addr)
		require.NoError(t, err)
		assert.Equal(t, "{}", string(data))
	})
}

func TestHumanName(t *testing.T) {
	t.Run("create human name", func(t *testing.T) {
		use := NameUseOfficial
		family := "Smith"
		text := "John Smith"

		name := HumanName{
			Use:    &use,
			Family: &family,
			Given:  []string{"John", "Robert"},
			Text:   &text,
		}

		assert.Equal(t, NameUseOfficial, *name.Use)
		assert.Equal(t, "Smith", *name.Family)
		assert.Equal(t, []string{"John", "Robert"}, name.Given)
		assert.Equal(t, "John Smith", *name.Text)
	})

	t.Run("JSON round trip", func(t *testing.T) {
		use := NameUseNickname
		family := "Jones"

		original := HumanName{
			Use:    &use,
			Family: &family,
			Given:  []string{"Bob"},
			Prefix: []string{"Mr."},
			Suffix: []string{"Jr."},
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded HumanName
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *original.Use, *decoded.Use)
		assert.Equal(t, *original.Family, *decoded.Family)
		assert.Equal(t, original.Given, decoded.Given)
		assert.Equal(t, original.Prefix, decoded.Prefix)
		assert.Equal(t, original.Suffix, decoded.Suffix)
	})
}

func TestCoding(t *testing.T) {
	t.Run("create coding", func(t *testing.T) {
		system := "http://loinc.org"
		code := "8867-4"
		display := "Heart rate"

		coding := Coding{
			System:  &system,
			Code:    &code,
			Display: &display,
		}

		assert.Equal(t, "http://loinc.org", *coding.System)
		assert.Equal(t, "8867-4", *coding.Code)
		assert.Equal(t, "Heart rate", *coding.Display)
	})

	t.Run("JSON serialization", func(t *testing.T) {
		system := "http://snomed.info/sct"
		code := "27113001"
		display := "Body weight"
		userSelected := true

		coding := Coding{
			System:       &system,
			Code:         &code,
			Display:      &display,
			UserSelected: &userSelected,
		}

		data, err := json.Marshal(coding)
		require.NoError(t, err)

		var decoded Coding
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		assert.Equal(t, *coding.System, *decoded.System)
		assert.Equal(t, *coding.Code, *decoded.Code)
		assert.Equal(t, *coding.Display, *decoded.Display)
		assert.Equal(t, *coding.UserSelected, *decoded.UserSelected)
	})
}

func TestCodeableConcept(t *testing.T) {
	t.Run("create codeable concept with codings", func(t *testing.T) {
		system := "http://terminology.hl7.org/CodeSystem/condition-clinical"
		code := "active"
		display := "Active"
		text := "Active condition"

		cc := CodeableConcept{
			Coding: []Coding{
				{
					System:  &system,
					Code:    &code,
					Display: &display,
				},
			},
			Text: &text,
		}

		require.Len(t, cc.Coding, 1)
		assert.Equal(t, "http://terminology.hl7.org/CodeSystem/condition-clinical", *cc.Coding[0].System)
		assert.Equal(t, "active", *cc.Coding[0].Code)
		assert.Equal(t, "Active condition", *cc.Text)
	})

	t.Run("JSON round trip with multiple codings", func(t *testing.T) {
		system1 := "http://snomed.info/sct"
		code1 := "38341003"
		system2 := "http://hl7.org/fhir/sid/icd-10"
		code2 := "I10"
		text := "Hypertension"

		original := CodeableConcept{
			Coding: []Coding{
				{System: &system1, Code: &code1},
				{System: &system2, Code: &code2},
			},
			Text: &text,
		}

		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded CodeableConcept
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)

		require.Len(t, decoded.Coding, 2)
		assert.Equal(t, *original.Coding[0].System, *decoded.Coding[0].System)
		assert.Equal(t, *original.Coding[1].Code, *decoded.Coding[1].Code)
		assert.Equal(t, *original.Text, *decoded.Text)
	})
}

func TestIdentifier(t *testing.T) {
	t.Run("create identifier", func(t *testing.T) {
		use := IdentifierUseOfficial
		system := "http://hospital.example.org/mrn"
		value := "12345"

		id := Identifier{
			Use:    &use,
			System: &system,
			Value:  &value,
		}

		assert.Equal(t, IdentifierUseOfficial, *id.Use)
		assert.Equal(t, "http://hospital.example.org/mrn", *id.System)
		assert.Equal(t, "12345", *id.Value)
	})
}

func TestContactPoint(t *testing.T) {
	t.Run("create phone contact", func(t *testing.T) {
		system := ContactPointSystemPhone
		value := "+1-555-123-4567"
		use := ContactPointUseHome
		rank := uint32(1)

		cp := ContactPoint{
			System: &system,
			Value:  &value,
			Use:    &use,
			Rank:   &rank,
		}

		assert.Equal(t, ContactPointSystemPhone, *cp.System)
		assert.Equal(t, "+1-555-123-4567", *cp.Value)
		assert.Equal(t, ContactPointUseHome, *cp.Use)
		assert.Equal(t, uint32(1), *cp.Rank)
	})

	t.Run("create email contact", func(t *testing.T) {
		system := ContactPointSystemEmail
		value := "john@example.com"
		use := ContactPointUseWork

		cp := ContactPoint{
			System: &system,
			Value:  &value,
			Use:    &use,
		}

		assert.Equal(t, ContactPointSystemEmail, *cp.System)
		assert.Equal(t, "john@example.com", *cp.Value)
		assert.Equal(t, ContactPointUseWork, *cp.Use)
	})
}

func TestQuantity(t *testing.T) {
	t.Run("create quantity", func(t *testing.T) {
		value := NewDecimalFromFloat64(72.5)
		unit := "kg"
		system := "http://unitsofmeasure.org"
		code := "kg"

		qty := Quantity{
			Value:  value,
			Unit:   &unit,
			System: &system,
			Code:   &code,
		}

		assert.Equal(t, 72.5, qty.Value.Float64())
		assert.Equal(t, "kg", *qty.Unit)
		assert.Equal(t, "http://unitsofmeasure.org", *qty.System)
		assert.Equal(t, "kg", *qty.Code)
	})

	t.Run("quantity with comparator", func(t *testing.T) {
		value := NewDecimalFromFloat64(100.0)
		comparator := QuantityComparatorGreaterThan
		unit := "mg/dL"

		qty := Quantity{
			Value:      value,
			Comparator: &comparator,
			Unit:       &unit,
		}

		assert.Equal(t, 100.0, qty.Value.Float64())
		assert.Equal(t, QuantityComparatorGreaterThan, *qty.Comparator)
		assert.Equal(t, unit, *qty.Unit)
	})
}

func TestPeriod(t *testing.T) {
	t.Run("create period", func(t *testing.T) {
		start := "2024-01-01"
		end := "2024-12-31"

		period := Period{
			Start: &start,
			End:   &end,
		}

		assert.Equal(t, "2024-01-01", *period.Start)
		assert.Equal(t, "2024-12-31", *period.End)
	})

	t.Run("ongoing period (no end)", func(t *testing.T) {
		start := "2024-01-01"

		period := Period{
			Start: &start,
		}

		assert.Equal(t, "2024-01-01", *period.Start)
		assert.Nil(t, period.End)
	})
}

func TestReference(t *testing.T) {
	t.Run("create reference", func(t *testing.T) {
		ref := "Patient/123"
		display := "John Smith"

		reference := Reference{
			Reference: &ref,
			Display:   &display,
		}

		assert.Equal(t, "Patient/123", *reference.Reference)
		assert.Equal(t, "John Smith", *reference.Display)
	})

	t.Run("reference with type", func(t *testing.T) {
		refType := "Practitioner"
		ref := "Practitioner/456"

		reference := Reference{
			Type:      &refType,
			Reference: &ref,
		}

		assert.Equal(t, "Practitioner", *reference.Type)
		assert.Equal(t, "Practitioner/456", *reference.Reference)
	})
}

func TestExtension(t *testing.T) {
	t.Run("extension with primitive extension", func(t *testing.T) {
		id := "ext-1"
		url := "http://example.org/fhir/StructureDefinition/custom"
		valueString := "test value"

		ext := Extension{
			Id:          &id,
			Url:         url,
			ValueString: &valueString,
		}

		assert.Equal(t, "ext-1", *ext.Id)
		assert.Equal(t, "http://example.org/fhir/StructureDefinition/custom", ext.Url)
		assert.Equal(t, "test value", *ext.ValueString)
	})
}

func TestElement(t *testing.T) {
	t.Run("element with extension", func(t *testing.T) {
		id := "elem-1"
		url := "http://example.org/ext"

		elem := Element{
			Id: &id,
			Extension: []Extension{
				{Url: url},
			},
		}

		assert.Equal(t, "elem-1", *elem.Id)
		require.Len(t, elem.Extension, 1)
		assert.Equal(t, "http://example.org/ext", elem.Extension[0].Url)
	})
}

func TestPrimitiveExtensionFields(t *testing.T) {
	t.Run("address has extension fields for primitives", func(t *testing.T) {
		city := "Boston"
		extID := "city-ext"

		addr := Address{
			City: &city,
			CityExt: &Element{
				Id: &extID,
			},
		}

		assert.Equal(t, "Boston", *addr.City)
		require.NotNil(t, addr.CityExt)
		assert.Equal(t, "city-ext", *addr.CityExt.Id)
	})

	t.Run("coding has extension fields", func(t *testing.T) {
		code := "12345"
		extID := "code-ext"

		coding := Coding{
			Code: &code,
			CodeExt: &Element{
				Id: &extID,
			},
		}

		assert.Equal(t, "12345", *coding.Code)
		require.NotNil(t, coding.CodeExt)
		assert.Equal(t, "code-ext", *coding.CodeExt.Id)
	})
}

package r4_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gofhir/models/r4"
)

func TestDecimal_MarshalJSON(t *testing.T) {
	tests := []struct {
		name string
		dec  r4.Decimal
		want string
	}{
		{"preserves trailing zero", *r4.MustDecimal("1.50"), "1.50"},
		{"integer value", *r4.MustDecimal("100"), "100"},
		{"negative value", *r4.MustDecimal("-3.14"), "-3.14"},
		{"small precision", *r4.MustDecimal("0.001"), "0.001"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := json.Marshal(tt.dec)
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(data))
		})
	}
}

func TestDecimal_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"preserves trailing zero", "1.50", "1.50", false},
		{"integer", "100", "100", false},
		{"negative", "-3.14", "-3.14", false},
		{"quoted number", `"1.50"`, "1.50", false},
		{"null", "null", "", false},
		{"invalid string", `"abc"`, "", true},
		{"NaN", `"NaN"`, "", true},
		{"Infinity", `"Infinity"`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var d r4.Decimal
			err := json.Unmarshal([]byte(tt.input), &d)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, d.String())
		})
	}
}

func TestDecimal_RoundTrip(t *testing.T) {
	values := []string{"1.50", "100", "0.001", "-3.14", "1234567890.123456"}

	for _, v := range values {
		t.Run(v, func(t *testing.T) {
			original := r4.MustDecimal(v)

			// Marshal
			data, err := json.Marshal(original)
			require.NoError(t, err)

			// Unmarshal
			var decoded r4.Decimal
			err = json.Unmarshal(data, &decoded)
			require.NoError(t, err)

			assert.Equal(t, v, decoded.String())
		})
	}
}

func TestDecimal_Float64(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"1.50", 1.5},
		{"100", 100.0},
		{"-3.14", -3.14},
		{"0", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			d := r4.MustDecimal(tt.input)
			assert.Equal(t, tt.want, d.Float64())
		})
	}
}

func TestDecimal_String(t *testing.T) {
	d := r4.MustDecimal("1.50")
	assert.Equal(t, "1.50", d.String())
}

func TestDecimal_IsZero(t *testing.T) {
	assert.True(t, r4.MustDecimal("0").IsZero())
	assert.True(t, r4.MustDecimal("0.0").IsZero())
	assert.True(t, r4.MustDecimal("0.00").IsZero())
	assert.False(t, r4.MustDecimal("1").IsZero())
	assert.False(t, r4.MustDecimal("0.001").IsZero())
}

func TestDecimal_Equal(t *testing.T) {
	assert.True(t, r4.MustDecimal("1.0").Equal(*r4.MustDecimal("1.00")))
	assert.True(t, r4.MustDecimal("100").Equal(*r4.MustDecimal("100.0")))
	assert.False(t, r4.MustDecimal("1.0").Equal(*r4.MustDecimal("2.0")))
}

func TestDecimal_InvalidInput(t *testing.T) {
	_, err := r4.NewDecimalFromString("abc")
	assert.Error(t, err)

	_, err = r4.NewDecimalFromString("NaN")
	assert.Error(t, err)

	_, err = r4.NewDecimalFromString("Infinity")
	assert.Error(t, err)

	_, err = r4.NewDecimalFromString("")
	assert.Error(t, err)
}

func TestDecimal_Constructors(t *testing.T) {
	t.Run("NewDecimalFromFloat64", func(t *testing.T) {
		d := r4.NewDecimalFromFloat64(1.5)
		assert.Equal(t, "1.5", d.String())
		assert.Equal(t, 1.5, d.Float64())
	})

	t.Run("NewDecimalFromString", func(t *testing.T) {
		d, err := r4.NewDecimalFromString("1.50")
		require.NoError(t, err)
		assert.Equal(t, "1.50", d.String())
	})

	t.Run("MustDecimal", func(t *testing.T) {
		d := r4.MustDecimal("1.50")
		assert.Equal(t, "1.50", d.String())
	})

	t.Run("MustDecimal panics on invalid", func(t *testing.T) {
		assert.Panics(t, func() { r4.MustDecimal("abc") })
	})

	t.Run("NewDecimalFromInt", func(t *testing.T) {
		d := r4.NewDecimalFromInt(42)
		assert.Equal(t, "42", d.String())
		assert.Equal(t, float64(42), d.Float64())
	})

	t.Run("NewDecimalFromInt64", func(t *testing.T) {
		d := r4.NewDecimalFromInt64(9999999999)
		assert.Equal(t, "9999999999", d.String())
	})
}

func TestQuantity_DecimalRoundTrip(t *testing.T) {
	// Create a Quantity with a precision-preserving Decimal value
	qty := r4.Quantity{
		Value:  r4.MustDecimal("1.50"),
		Unit:   ptr("mg"),
		System: ptr("http://unitsofmeasure.org"),
		Code:   ptr("mg"),
	}

	// Marshal to JSON
	data, err := json.Marshal(qty)
	require.NoError(t, err)

	// Verify the JSON contains the precise value
	assert.Contains(t, string(data), "1.50")

	// Unmarshal back
	var decoded r4.Quantity
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	// Verify precision is preserved
	require.NotNil(t, decoded.Value)
	assert.Equal(t, "1.50", decoded.Value.String())
}

func TestDecimal_XML_RoundTrip(t *testing.T) {
	// Create an Observation with a Decimal value
	status := r4.ObservationStatusFinal
	obs := r4.Observation{
		Status: &status,
		Code: r4.CodeableConcept{
			Text: ptr("test"),
		},
		ValueQuantity: &r4.Quantity{
			Value: r4.MustDecimal("120.50"),
			Unit:  ptr("mmHg"),
		},
	}

	// Marshal to XML
	xmlData, err := r4.MarshalResourceXML(&obs)
	require.NoError(t, err)

	// Verify the XML contains the precise value
	assert.Contains(t, string(xmlData), `value="120.50"`)

	// Unmarshal back
	res, err := r4.UnmarshalResourceXML(xmlData)
	require.NoError(t, err)

	decoded, ok := res.(*r4.Observation)
	require.True(t, ok)
	require.NotNil(t, decoded.ValueQuantity)
	require.NotNil(t, decoded.ValueQuantity.Value)
	assert.Equal(t, "120.50", decoded.ValueQuantity.Value.String())
}

func ptr(s string) *string {
	return &s
}

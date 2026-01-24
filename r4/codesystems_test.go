package r4

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCodeSystemTypes(t *testing.T) {
	t.Run("AdministrativeGender constants", func(t *testing.T) {
		assert.Equal(t, AdministrativeGender("male"), AdministrativeGenderMale)
		assert.Equal(t, AdministrativeGender("female"), AdministrativeGenderFemale)
		assert.Equal(t, AdministrativeGender("other"), AdministrativeGenderOther)
		assert.Equal(t, AdministrativeGender("unknown"), AdministrativeGenderUnknown)
	})

	t.Run("NameUse constants", func(t *testing.T) {
		assert.Equal(t, NameUse("usual"), NameUseUsual)
		assert.Equal(t, NameUse("official"), NameUseOfficial)
		assert.Equal(t, NameUse("temp"), NameUseTemp)
		assert.Equal(t, NameUse("nickname"), NameUseNickname)
		assert.Equal(t, NameUse("anonymous"), NameUseAnonymous)
		assert.Equal(t, NameUse("old"), NameUseOld)
	})

	t.Run("AddressUse constants", func(t *testing.T) {
		assert.Equal(t, AddressUse("home"), AddressUseHome)
		assert.Equal(t, AddressUse("work"), AddressUseWork)
		assert.Equal(t, AddressUse("temp"), AddressUseTemp)
		assert.Equal(t, AddressUse("old"), AddressUseOld)
		assert.Equal(t, AddressUse("billing"), AddressUseBilling)
	})

	t.Run("AddressType constants", func(t *testing.T) {
		assert.Equal(t, AddressType("postal"), AddressTypePostal)
		assert.Equal(t, AddressType("physical"), AddressTypePhysical)
		assert.Equal(t, AddressType("both"), AddressTypeBoth)
	})

	t.Run("AccountStatus constants", func(t *testing.T) {
		assert.Equal(t, AccountStatus("active"), AccountStatusActive)
		assert.Equal(t, AccountStatus("inactive"), AccountStatusInactive)
		assert.Equal(t, AccountStatus("entered-in-error"), AccountStatusEnteredInError)
		assert.Equal(t, AccountStatus("on-hold"), AccountStatusOnHold)
		assert.Equal(t, AccountStatus("unknown"), AccountStatusUnknown)
	})

	t.Run("BundleType constants", func(t *testing.T) {
		assert.Equal(t, BundleType("document"), BundleTypeDocument)
		assert.Equal(t, BundleType("message"), BundleTypeMessage)
		assert.Equal(t, BundleType("transaction"), BundleTypeTransaction)
		assert.Equal(t, BundleType("batch"), BundleTypeBatch)
		assert.Equal(t, BundleType("searchset"), BundleTypeSearchset)
		assert.Equal(t, BundleType("collection"), BundleTypeCollection)
	})
}

func TestCodeSystemTypeConversions(t *testing.T) {
	t.Run("string to code type", func(t *testing.T) {
		gender := AdministrativeGender("male")
		assert.Equal(t, "male", string(gender))
	})

	t.Run("code type comparison", func(t *testing.T) {
		gender := AdministrativeGenderFemale
		assert.True(t, gender == AdministrativeGenderFemale)
		assert.False(t, gender == AdministrativeGenderMale)
	})

	t.Run("code type in switch", func(t *testing.T) {
		gender := AdministrativeGenderOther
		var result string
		switch gender {
		case AdministrativeGenderMale:
			result = "M"
		case AdministrativeGenderFemale:
			result = "F"
		case AdministrativeGenderOther:
			result = "O"
		case AdministrativeGenderUnknown:
			result = "U"
		}
		assert.Equal(t, "O", result)
	})
}

func TestQuantityComparator(t *testing.T) {
	t.Run("comparator constants", func(t *testing.T) {
		assert.Equal(t, QuantityComparator("<"), QuantityComparatorLessThan)
		assert.Equal(t, QuantityComparator("<="), QuantityComparatorLessOrEqual)
		assert.Equal(t, QuantityComparator(">="), QuantityComparatorGreaterOrEqual)
		assert.Equal(t, QuantityComparator(">"), QuantityComparatorGreaterThan)
	})
}

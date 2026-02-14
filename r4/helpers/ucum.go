package helpers

import "github.com/gofhir/models/r4"

// UCUMSystem is the official UCUM code system URL.
const UCUMSystem = "http://unitsofmeasure.org"

// =============================================================================
// Weight Units
// =============================================================================

// QuantityKg creates a Quantity with kilograms.
func QuantityKg(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("kg"),
		System: ptr(UCUMSystem),
		Code:   ptr("kg"),
	}
}

// QuantityLb creates a Quantity with pounds.
func QuantityLb(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("[lb_av]"),
		System: ptr(UCUMSystem),
		Code:   ptr("[lb_av]"),
	}
}

// QuantityG creates a Quantity with grams.
func QuantityG(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("g"),
		System: ptr(UCUMSystem),
		Code:   ptr("g"),
	}
}

// =============================================================================
// Length/Height Units
// =============================================================================

// QuantityCm creates a Quantity with centimeters.
func QuantityCm(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("cm"),
		System: ptr(UCUMSystem),
		Code:   ptr("cm"),
	}
}

// QuantityM creates a Quantity with meters.
func QuantityM(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("m"),
		System: ptr(UCUMSystem),
		Code:   ptr("m"),
	}
}

// QuantityIn creates a Quantity with inches.
func QuantityIn(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("[in_i]"),
		System: ptr(UCUMSystem),
		Code:   ptr("[in_i]"),
	}
}

// QuantityFt creates a Quantity with feet.
func QuantityFt(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("[ft_i]"),
		System: ptr(UCUMSystem),
		Code:   ptr("[ft_i]"),
	}
}

// =============================================================================
// Temperature Units
// =============================================================================

// QuantityCelsius creates a Quantity with degrees Celsius.
func QuantityCelsius(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("Cel"),
		System: ptr(UCUMSystem),
		Code:   ptr("Cel"),
	}
}

// QuantityFahrenheit creates a Quantity with degrees Fahrenheit.
func QuantityFahrenheit(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("[degF]"),
		System: ptr(UCUMSystem),
		Code:   ptr("[degF]"),
	}
}

// =============================================================================
// Pressure Units (Blood Pressure)
// =============================================================================

// QuantityMmHg creates a Quantity with millimeters of mercury.
func QuantityMmHg(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("mm[Hg]"),
		System: ptr(UCUMSystem),
		Code:   ptr("mm[Hg]"),
	}
}

// =============================================================================
// Rate Units (Heart Rate, Respiratory Rate)
// =============================================================================

// QuantityBPM creates a Quantity with beats per minute (for heart rate).
func QuantityBPM(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("/min"),
		System: ptr(UCUMSystem),
		Code:   ptr("/min"),
	}
}

// QuantityBreathsPerMin creates a Quantity with breaths per minute.
func QuantityBreathsPerMin(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("/min"),
		System: ptr(UCUMSystem),
		Code:   ptr("/min"),
	}
}

// =============================================================================
// Percentage Units
// =============================================================================

// QuantityPercent creates a Quantity with percentage.
func QuantityPercent(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("%"),
		System: ptr(UCUMSystem),
		Code:   ptr("%"),
	}
}

// =============================================================================
// Concentration Units (Laboratory)
// =============================================================================

// QuantityMgDL creates a Quantity with milligrams per deciliter.
func QuantityMgDL(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("mg/dL"),
		System: ptr(UCUMSystem),
		Code:   ptr("mg/dL"),
	}
}

// QuantityMmolL creates a Quantity with millimoles per liter.
func QuantityMmolL(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("mmol/L"),
		System: ptr(UCUMSystem),
		Code:   ptr("mmol/L"),
	}
}

// QuantityGDL creates a Quantity with grams per deciliter.
func QuantityGDL(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("g/dL"),
		System: ptr(UCUMSystem),
		Code:   ptr("g/dL"),
	}
}

// QuantityUL creates a Quantity with units per liter.
func QuantityUL(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("U/L"),
		System: ptr(UCUMSystem),
		Code:   ptr("U/L"),
	}
}

// QuantityMeqL creates a Quantity with milliequivalents per liter.
func QuantityMeqL(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("meq/L"),
		System: ptr(UCUMSystem),
		Code:   ptr("meq/L"),
	}
}

// =============================================================================
// Renal Function Units
// =============================================================================

// QuantityMLMinPerM2 creates a Quantity for eGFR (mL/min/1.73m²).
func QuantityMLMinPerM2(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("mL/min/{1.73_m2}"),
		System: ptr(UCUMSystem),
		Code:   ptr("mL/min/{1.73_m2}"),
	}
}

// =============================================================================
// BMI Units
// =============================================================================

// QuantityKgM2 creates a Quantity for BMI (kg/m²).
func QuantityKgM2(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("kg/m2"),
		System: ptr(UCUMSystem),
		Code:   ptr("kg/m2"),
	}
}

// =============================================================================
// Time Units
// =============================================================================

// QuantitySeconds creates a Quantity with seconds.
func QuantitySeconds(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("s"),
		System: ptr(UCUMSystem),
		Code:   ptr("s"),
	}
}

// QuantityMinutes creates a Quantity with minutes.
func QuantityMinutes(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("min"),
		System: ptr(UCUMSystem),
		Code:   ptr("min"),
	}
}

// QuantityHours creates a Quantity with hours.
func QuantityHours(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("h"),
		System: ptr(UCUMSystem),
		Code:   ptr("h"),
	}
}

// QuantityDays creates a Quantity with days.
func QuantityDays(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("d"),
		System: ptr(UCUMSystem),
		Code:   ptr("d"),
	}
}

// QuantityWeeks creates a Quantity with weeks.
func QuantityWeeks(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("wk"),
		System: ptr(UCUMSystem),
		Code:   ptr("wk"),
	}
}

// QuantityMonths creates a Quantity with months.
func QuantityMonths(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("mo"),
		System: ptr(UCUMSystem),
		Code:   ptr("mo"),
	}
}

// QuantityYears creates a Quantity with years.
func QuantityYears(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("a"),
		System: ptr(UCUMSystem),
		Code:   ptr("a"),
	}
}

// =============================================================================
// Volume Units
// =============================================================================

// QuantityML creates a Quantity with milliliters.
func QuantityML(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("mL"),
		System: ptr(UCUMSystem),
		Code:   ptr("mL"),
	}
}

// QuantityL creates a Quantity with liters.
func QuantityL(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("L"),
		System: ptr(UCUMSystem),
		Code:   ptr("L"),
	}
}

// =============================================================================
// Medication Dosage Units
// =============================================================================

// QuantityMg creates a Quantity with milligrams.
func QuantityMg(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("mg"),
		System: ptr(UCUMSystem),
		Code:   ptr("mg"),
	}
}

// QuantityMcg creates a Quantity with micrograms.
func QuantityMcg(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("ug"),
		System: ptr(UCUMSystem),
		Code:   ptr("ug"),
	}
}

// QuantityMgKg creates a Quantity with milligrams per kilogram.
func QuantityMgKg(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("mg/kg"),
		System: ptr(UCUMSystem),
		Code:   ptr("mg/kg"),
	}
}

// QuantityUnits creates a Quantity with international units.
func QuantityUnits(value float64) r4.Quantity {
	return r4.Quantity{
		Value:  &value,
		Unit:   ptr("[iU]"),
		System: ptr(UCUMSystem),
		Code:   ptr("[iU]"),
	}
}

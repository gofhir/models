package helpers

import (
	"testing"

	"github.com/robertoaraneda/gofhir/r4"
)

// =============================================================================
// LOINC Code Tests
// =============================================================================

func TestLOINCVitalSignsCodes(t *testing.T) {
	tests := []struct {
		name        string
		code        r4.CodeableConcept
		wantSystem  string
		wantCode    string
		wantDisplay string
	}{
		{
			name:        "VitalSignsPanel",
			code:        VitalSignsPanel,
			wantSystem:  LOINCSystem,
			wantCode:    "85353-1",
			wantDisplay: "Vital signs, weight, height, head circumference, oxygen saturation and BMI panel",
		},
		{
			name:        "BodyWeight",
			code:        BodyWeight,
			wantSystem:  LOINCSystem,
			wantCode:    "29463-7",
			wantDisplay: "Body weight",
		},
		{
			name:        "BodyHeight",
			code:        BodyHeight,
			wantSystem:  LOINCSystem,
			wantCode:    "8302-2",
			wantDisplay: "Body height",
		},
		{
			name:        "BodyTemperature",
			code:        BodyTemperature,
			wantSystem:  LOINCSystem,
			wantCode:    "8310-5",
			wantDisplay: "Body temperature",
		},
		{
			name:        "HeartRate",
			code:        HeartRate,
			wantSystem:  LOINCSystem,
			wantCode:    "8867-4",
			wantDisplay: "Heart rate",
		},
		{
			name:        "RespiratoryRate",
			code:        RespiratoryRate,
			wantSystem:  LOINCSystem,
			wantCode:    "9279-1",
			wantDisplay: "Respiratory rate",
		},
		{
			name:        "BloodPressurePanel",
			code:        BloodPressurePanel,
			wantSystem:  LOINCSystem,
			wantCode:    "85354-9",
			wantDisplay: "Blood pressure panel with all children optional",
		},
		{
			name:        "SystolicBloodPressure",
			code:        SystolicBloodPressure,
			wantSystem:  LOINCSystem,
			wantCode:    "8480-6",
			wantDisplay: "Systolic blood pressure",
		},
		{
			name:        "DiastolicBloodPressure",
			code:        DiastolicBloodPressure,
			wantSystem:  LOINCSystem,
			wantCode:    "8462-4",
			wantDisplay: "Diastolic blood pressure",
		},
		{
			name:        "OxygenSaturation",
			code:        OxygenSaturation,
			wantSystem:  LOINCSystem,
			wantCode:    "2708-6",
			wantDisplay: "Oxygen saturation in Arterial blood",
		},
		{
			name:        "PulseOximetry",
			code:        PulseOximetry,
			wantSystem:  LOINCSystem,
			wantCode:    "59408-5",
			wantDisplay: "Oxygen saturation in Arterial blood by Pulse oximetry",
		},
		{
			name:        "BMI",
			code:        BMI,
			wantSystem:  LOINCSystem,
			wantCode:    "39156-5",
			wantDisplay: "Body mass index (BMI) [Ratio]",
		},
		{
			name:        "HeadCircumference",
			code:        HeadCircumference,
			wantSystem:  LOINCSystem,
			wantCode:    "9843-4",
			wantDisplay: "Head Occipital-frontal circumference",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCodeableConcept(t, tt.code, tt.wantSystem, tt.wantCode, tt.wantDisplay)
		})
	}
}

func TestLOINCLaboratoryCodes(t *testing.T) {
	tests := []struct {
		name        string
		code        r4.CodeableConcept
		wantSystem  string
		wantCode    string
		wantDisplay string
	}{
		{
			name:        "Glucose",
			code:        Glucose,
			wantSystem:  LOINCSystem,
			wantCode:    "2339-0",
			wantDisplay: "Glucose [Mass/volume] in Blood",
		},
		{
			name:        "GlucoseFasting",
			code:        GlucoseFasting,
			wantSystem:  LOINCSystem,
			wantCode:    "1558-6",
			wantDisplay: "Fasting glucose [Mass/volume] in Serum or Plasma",
		},
		{
			name:        "HemoglobinA1c",
			code:        HemoglobinA1c,
			wantSystem:  LOINCSystem,
			wantCode:    "4548-4",
			wantDisplay: "Hemoglobin A1c/Hemoglobin.total in Blood",
		},
		{
			name:        "Hemoglobin",
			code:        Hemoglobin,
			wantSystem:  LOINCSystem,
			wantCode:    "718-7",
			wantDisplay: "Hemoglobin [Mass/volume] in Blood",
		},
		{
			name:        "Hematocrit",
			code:        Hematocrit,
			wantSystem:  LOINCSystem,
			wantCode:    "4544-3",
			wantDisplay: "Hematocrit [Volume Fraction] of Blood by Automated count",
		},
		{
			name:        "Creatinine",
			code:        Creatinine,
			wantSystem:  LOINCSystem,
			wantCode:    "2160-0",
			wantDisplay: "Creatinine [Mass/volume] in Serum or Plasma",
		},
		{
			name:        "eGFR",
			code:        EGFR,
			wantSystem:  LOINCSystem,
			wantCode:    "33914-3",
			wantDisplay: "Glomerular filtration rate/1.73 sq M.predicted [Volume Rate/Area] in Serum or Plasma by Creatinine-based formula (MDRD)",
		},
		{
			name:        "Cholesterol",
			code:        Cholesterol,
			wantSystem:  LOINCSystem,
			wantCode:    "2093-3",
			wantDisplay: "Cholesterol [Mass/volume] in Serum or Plasma",
		},
		{
			name:        "HDLCholesterol",
			code:        HDLCholesterol,
			wantSystem:  LOINCSystem,
			wantCode:    "2085-9",
			wantDisplay: "Cholesterol in HDL [Mass/volume] in Serum or Plasma",
		},
		{
			name:        "LDLCholesterol",
			code:        LDLCholesterol,
			wantSystem:  LOINCSystem,
			wantCode:    "2089-1",
			wantDisplay: "Cholesterol in LDL [Mass/volume] in Serum or Plasma",
		},
		{
			name:        "Triglycerides",
			code:        Triglycerides,
			wantSystem:  LOINCSystem,
			wantCode:    "2571-8",
			wantDisplay: "Triglyceride [Mass/volume] in Serum or Plasma",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCodeableConcept(t, tt.code, tt.wantSystem, tt.wantCode, tt.wantDisplay)
		})
	}
}

func TestLOINCIPSSectionCodes(t *testing.T) {
	tests := []struct {
		name        string
		code        r4.CodeableConcept
		wantSystem  string
		wantCode    string
		wantDisplay string
	}{
		{
			name:        "IPSMedicationSummary",
			code:        IPSMedicationSummary,
			wantSystem:  LOINCSystem,
			wantCode:    "10160-0",
			wantDisplay: "History of Medication use Narrative",
		},
		{
			name:        "IPSAllergies",
			code:        IPSAllergies,
			wantSystem:  LOINCSystem,
			wantCode:    "48765-2",
			wantDisplay: "Allergies and adverse reactions Document",
		},
		{
			name:        "IPSProblems",
			code:        IPSProblems,
			wantSystem:  LOINCSystem,
			wantCode:    "11450-4",
			wantDisplay: "Problem list - Reported",
		},
		{
			name:        "IPSImmunizations",
			code:        IPSImmunizations,
			wantSystem:  LOINCSystem,
			wantCode:    "11369-6",
			wantDisplay: "History of Immunization Narrative",
		},
		{
			name:        "IPSProcedures",
			code:        IPSProcedures,
			wantSystem:  LOINCSystem,
			wantCode:    "47519-4",
			wantDisplay: "History of Procedures Document",
		},
		{
			name:        "IPSMedicalDevices",
			code:        IPSMedicalDevices,
			wantSystem:  LOINCSystem,
			wantCode:    "46264-8",
			wantDisplay: "History of medical device use",
		},
		{
			name:        "IPSDiagnosticResults",
			code:        IPSDiagnosticResults,
			wantSystem:  LOINCSystem,
			wantCode:    "30954-2",
			wantDisplay: "Relevant diagnostic tests/laboratory data Narrative",
		},
		{
			name:        "IPSVitalSigns",
			code:        IPSVitalSigns,
			wantSystem:  LOINCSystem,
			wantCode:    "8716-3",
			wantDisplay: "Vital signs",
		},
		{
			name:        "IPSPastIllness",
			code:        IPSPastIllness,
			wantSystem:  LOINCSystem,
			wantCode:    "11348-0",
			wantDisplay: "History of Past illness Narrative",
		},
		{
			name:        "IPSFunctionalStatus",
			code:        IPSFunctionalStatus,
			wantSystem:  LOINCSystem,
			wantCode:    "47420-5",
			wantDisplay: "Functional status assessment note",
		},
		{
			name:        "IPSPlanOfCare",
			code:        IPSPlanOfCare,
			wantSystem:  LOINCSystem,
			wantCode:    "18776-5",
			wantDisplay: "Plan of care note",
		},
		{
			name:        "IPSSocialHistory",
			code:        IPSSocialHistory,
			wantSystem:  LOINCSystem,
			wantCode:    "29762-2",
			wantDisplay: "Social history Narrative",
		},
		{
			name:        "IPSPregnancy",
			code:        IPSPregnancy,
			wantSystem:  LOINCSystem,
			wantCode:    "10162-6",
			wantDisplay: "History of pregnancies Narrative",
		},
		{
			name:        "IPSAdvanceDirectives",
			code:        IPSAdvanceDirectives,
			wantSystem:  LOINCSystem,
			wantCode:    "42348-3",
			wantDisplay: "Advance directives",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCodeableConcept(t, tt.code, tt.wantSystem, tt.wantCode, tt.wantDisplay)
		})
	}
}

// =============================================================================
// UCUM Quantity Tests
// =============================================================================

func TestUCUMWeightQuantities(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(float64) r4.Quantity
		value      float64
		wantUnit   string
		wantCode   string
		wantSystem string
	}{
		{
			name:       "QuantityKg",
			fn:         QuantityKg,
			value:      75.5,
			wantUnit:   "kg",
			wantCode:   "kg",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityLb",
			fn:         QuantityLb,
			value:      165.0,
			wantUnit:   "[lb_av]",
			wantCode:   "[lb_av]",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityG",
			fn:         QuantityG,
			value:      500.0,
			wantUnit:   "g",
			wantCode:   "g",
			wantSystem: UCUMSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fn(tt.value)
			assertQuantity(t, q, tt.value, tt.wantUnit, tt.wantCode, tt.wantSystem)
		})
	}
}

func TestUCUMLengthQuantities(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(float64) r4.Quantity
		value      float64
		wantUnit   string
		wantCode   string
		wantSystem string
	}{
		{
			name:       "QuantityCm",
			fn:         QuantityCm,
			value:      175.0,
			wantUnit:   "cm",
			wantCode:   "cm",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityM",
			fn:         QuantityM,
			value:      1.75,
			wantUnit:   "m",
			wantCode:   "m",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityIn",
			fn:         QuantityIn,
			value:      68.0,
			wantUnit:   "[in_i]",
			wantCode:   "[in_i]",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityFt",
			fn:         QuantityFt,
			value:      5.8,
			wantUnit:   "[ft_i]",
			wantCode:   "[ft_i]",
			wantSystem: UCUMSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fn(tt.value)
			assertQuantity(t, q, tt.value, tt.wantUnit, tt.wantCode, tt.wantSystem)
		})
	}
}

func TestUCUMTemperatureQuantities(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(float64) r4.Quantity
		value      float64
		wantUnit   string
		wantCode   string
		wantSystem string
	}{
		{
			name:       "QuantityCelsius",
			fn:         QuantityCelsius,
			value:      37.0,
			wantUnit:   "Cel",
			wantCode:   "Cel",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityFahrenheit",
			fn:         QuantityFahrenheit,
			value:      98.6,
			wantUnit:   "[degF]",
			wantCode:   "[degF]",
			wantSystem: UCUMSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fn(tt.value)
			assertQuantity(t, q, tt.value, tt.wantUnit, tt.wantCode, tt.wantSystem)
		})
	}
}

func TestUCUMPressureAndRateQuantities(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(float64) r4.Quantity
		value      float64
		wantUnit   string
		wantCode   string
		wantSystem string
	}{
		{
			name:       "QuantityMmHg",
			fn:         QuantityMmHg,
			value:      120.0,
			wantUnit:   "mm[Hg]",
			wantCode:   "mm[Hg]",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityBPM",
			fn:         QuantityBPM,
			value:      72.0,
			wantUnit:   "/min",
			wantCode:   "/min",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityBreathsPerMin",
			fn:         QuantityBreathsPerMin,
			value:      16.0,
			wantUnit:   "/min",
			wantCode:   "/min",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityPercent",
			fn:         QuantityPercent,
			value:      98.0,
			wantUnit:   "%",
			wantCode:   "%",
			wantSystem: UCUMSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fn(tt.value)
			assertQuantity(t, q, tt.value, tt.wantUnit, tt.wantCode, tt.wantSystem)
		})
	}
}

func TestUCUMConcentrationQuantities(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(float64) r4.Quantity
		value      float64
		wantUnit   string
		wantCode   string
		wantSystem string
	}{
		{
			name:       "QuantityMgDL",
			fn:         QuantityMgDL,
			value:      100.0,
			wantUnit:   "mg/dL",
			wantCode:   "mg/dL",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityMmolL",
			fn:         QuantityMmolL,
			value:      5.5,
			wantUnit:   "mmol/L",
			wantCode:   "mmol/L",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityGDL",
			fn:         QuantityGDL,
			value:      14.0,
			wantUnit:   "g/dL",
			wantCode:   "g/dL",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityUL",
			fn:         QuantityUL,
			value:      45.0,
			wantUnit:   "U/L",
			wantCode:   "U/L",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityMeqL",
			fn:         QuantityMeqL,
			value:      140.0,
			wantUnit:   "meq/L",
			wantCode:   "meq/L",
			wantSystem: UCUMSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fn(tt.value)
			assertQuantity(t, q, tt.value, tt.wantUnit, tt.wantCode, tt.wantSystem)
		})
	}
}

func TestUCUMTimeQuantities(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(float64) r4.Quantity
		value      float64
		wantUnit   string
		wantCode   string
		wantSystem string
	}{
		{
			name:       "QuantitySeconds",
			fn:         QuantitySeconds,
			value:      30.0,
			wantUnit:   "s",
			wantCode:   "s",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityMinutes",
			fn:         QuantityMinutes,
			value:      15.0,
			wantUnit:   "min",
			wantCode:   "min",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityHours",
			fn:         QuantityHours,
			value:      8.0,
			wantUnit:   "h",
			wantCode:   "h",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityDays",
			fn:         QuantityDays,
			value:      7.0,
			wantUnit:   "d",
			wantCode:   "d",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityWeeks",
			fn:         QuantityWeeks,
			value:      2.0,
			wantUnit:   "wk",
			wantCode:   "wk",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityMonths",
			fn:         QuantityMonths,
			value:      6.0,
			wantUnit:   "mo",
			wantCode:   "mo",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityYears",
			fn:         QuantityYears,
			value:      45.0,
			wantUnit:   "a",
			wantCode:   "a",
			wantSystem: UCUMSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fn(tt.value)
			assertQuantity(t, q, tt.value, tt.wantUnit, tt.wantCode, tt.wantSystem)
		})
	}
}

func TestUCUMVolumeAndDosageQuantities(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(float64) r4.Quantity
		value      float64
		wantUnit   string
		wantCode   string
		wantSystem string
	}{
		{
			name:       "QuantityML",
			fn:         QuantityML,
			value:      500.0,
			wantUnit:   "mL",
			wantCode:   "mL",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityL",
			fn:         QuantityL,
			value:      2.0,
			wantUnit:   "L",
			wantCode:   "L",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityMg",
			fn:         QuantityMg,
			value:      500.0,
			wantUnit:   "mg",
			wantCode:   "mg",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityMcg",
			fn:         QuantityMcg,
			value:      100.0,
			wantUnit:   "ug",
			wantCode:   "ug",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityMgKg",
			fn:         QuantityMgKg,
			value:      10.0,
			wantUnit:   "mg/kg",
			wantCode:   "mg/kg",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityUnits",
			fn:         QuantityUnits,
			value:      1000.0,
			wantUnit:   "[iU]",
			wantCode:   "[iU]",
			wantSystem: UCUMSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fn(tt.value)
			assertQuantity(t, q, tt.value, tt.wantUnit, tt.wantCode, tt.wantSystem)
		})
	}
}

func TestUCUMSpecialQuantities(t *testing.T) {
	tests := []struct {
		name       string
		fn         func(float64) r4.Quantity
		value      float64
		wantUnit   string
		wantCode   string
		wantSystem string
	}{
		{
			name:       "QuantityMLMinPerM2",
			fn:         QuantityMLMinPerM2,
			value:      90.0,
			wantUnit:   "mL/min/{1.73_m2}",
			wantCode:   "mL/min/{1.73_m2}",
			wantSystem: UCUMSystem,
		},
		{
			name:       "QuantityKgM2",
			fn:         QuantityKgM2,
			value:      24.5,
			wantUnit:   "kg/m2",
			wantCode:   "kg/m2",
			wantSystem: UCUMSystem,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.fn(tt.value)
			assertQuantity(t, q, tt.value, tt.wantUnit, tt.wantCode, tt.wantSystem)
		})
	}
}

// =============================================================================
// Category Tests
// =============================================================================

func TestObservationCategories(t *testing.T) {
	tests := []struct {
		name        string
		code        r4.CodeableConcept
		wantSystem  string
		wantCode    string
		wantDisplay string
	}{
		{
			name:        "ObservationCategoryVitalSigns",
			code:        ObservationCategoryVitalSigns,
			wantSystem:  ObservationCategorySystem,
			wantCode:    "vital-signs",
			wantDisplay: "Vital Signs",
		},
		{
			name:        "ObservationCategoryLaboratory",
			code:        ObservationCategoryLaboratory,
			wantSystem:  ObservationCategorySystem,
			wantCode:    "laboratory",
			wantDisplay: "Laboratory",
		},
		{
			name:        "ObservationCategorySocialHistory",
			code:        ObservationCategorySocialHistory,
			wantSystem:  ObservationCategorySystem,
			wantCode:    "social-history",
			wantDisplay: "Social History",
		},
		{
			name:        "ObservationCategoryImaging",
			code:        ObservationCategoryImaging,
			wantSystem:  ObservationCategorySystem,
			wantCode:    "imaging",
			wantDisplay: "Imaging",
		},
		{
			name:        "ObservationCategoryProcedure",
			code:        ObservationCategoryProcedure,
			wantSystem:  ObservationCategorySystem,
			wantCode:    "procedure",
			wantDisplay: "Procedure",
		},
		{
			name:        "ObservationCategorySurvey",
			code:        ObservationCategorySurvey,
			wantSystem:  ObservationCategorySystem,
			wantCode:    "survey",
			wantDisplay: "Survey",
		},
		{
			name:        "ObservationCategoryExam",
			code:        ObservationCategoryExam,
			wantSystem:  ObservationCategorySystem,
			wantCode:    "exam",
			wantDisplay: "Exam",
		},
		{
			name:        "ObservationCategoryTherapy",
			code:        ObservationCategoryTherapy,
			wantSystem:  ObservationCategorySystem,
			wantCode:    "therapy",
			wantDisplay: "Therapy",
		},
		{
			name:        "ObservationCategoryActivity",
			code:        ObservationCategoryActivity,
			wantSystem:  ObservationCategorySystem,
			wantCode:    "activity",
			wantDisplay: "Activity",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCodeableConcept(t, tt.code, tt.wantSystem, tt.wantCode, tt.wantDisplay)
		})
	}
}

func TestConditionCategories(t *testing.T) {
	tests := []struct {
		name        string
		code        r4.CodeableConcept
		wantSystem  string
		wantCode    string
		wantDisplay string
	}{
		{
			name:        "ConditionCategoryProblemListItem",
			code:        ConditionCategoryProblemListItem,
			wantSystem:  ConditionCategorySystem,
			wantCode:    "problem-list-item",
			wantDisplay: "Problem List Item",
		},
		{
			name:        "ConditionCategoryEncounterDiagnosis",
			code:        ConditionCategoryEncounterDiagnosis,
			wantSystem:  ConditionCategorySystem,
			wantCode:    "encounter-diagnosis",
			wantDisplay: "Encounter Diagnosis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCodeableConcept(t, tt.code, tt.wantSystem, tt.wantCode, tt.wantDisplay)
		})
	}
}

func TestAllergyCategories(t *testing.T) {
	tests := []struct {
		name        string
		code        r4.CodeableConcept
		wantSystem  string
		wantCode    string
		wantDisplay string
	}{
		{
			name:        "AllergyCategoryFood",
			code:        AllergyCategoryFood,
			wantSystem:  AllergyIntoleranceCategorySystem,
			wantCode:    "food",
			wantDisplay: "Food",
		},
		{
			name:        "AllergyCategoryMedication",
			code:        AllergyCategoryMedication,
			wantSystem:  AllergyIntoleranceCategorySystem,
			wantCode:    "medication",
			wantDisplay: "Medication",
		},
		{
			name:        "AllergyCategoryEnvironment",
			code:        AllergyCategoryEnvironment,
			wantSystem:  AllergyIntoleranceCategorySystem,
			wantCode:    "environment",
			wantDisplay: "Environment",
		},
		{
			name:        "AllergyCategoryBiologic",
			code:        AllergyCategoryBiologic,
			wantSystem:  AllergyIntoleranceCategorySystem,
			wantCode:    "biologic",
			wantDisplay: "Biologic",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCodeableConcept(t, tt.code, tt.wantSystem, tt.wantCode, tt.wantDisplay)
		})
	}
}

func TestDocumentTypes(t *testing.T) {
	tests := []struct {
		name        string
		code        r4.CodeableConcept
		wantSystem  string
		wantCode    string
		wantDisplay string
	}{
		{
			name:        "DocumentTypeIPS",
			code:        DocumentTypeIPS,
			wantSystem:  LOINCSystem,
			wantCode:    "60591-5",
			wantDisplay: "Patient summary Document",
		},
		{
			name:        "DocumentTypeCCD",
			code:        DocumentTypeCCD,
			wantSystem:  LOINCSystem,
			wantCode:    "34133-9",
			wantDisplay: "Summary of episode note",
		},
		{
			name:        "DocumentTypeDischarge",
			code:        DocumentTypeDischarge,
			wantSystem:  LOINCSystem,
			wantCode:    "18842-5",
			wantDisplay: "Discharge summary",
		},
		{
			name:        "DocumentTypeProgress",
			code:        DocumentTypeProgress,
			wantSystem:  LOINCSystem,
			wantCode:    "11506-3",
			wantDisplay: "Progress note",
		},
		{
			name:        "DocumentTypeHistory",
			code:        DocumentTypeHistory,
			wantSystem:  LOINCSystem,
			wantCode:    "34117-2",
			wantDisplay: "History and physical note",
		},
		{
			name:        "DocumentTypeConsult",
			code:        DocumentTypeConsult,
			wantSystem:  LOINCSystem,
			wantCode:    "11488-4",
			wantDisplay: "Consultation note",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assertCodeableConcept(t, tt.code, tt.wantSystem, tt.wantCode, tt.wantDisplay)
		})
	}
}

// =============================================================================
// Integration Tests - Real World Usage
// =============================================================================

func TestCreateVitalSignsObservation(t *testing.T) {
	// Test creating a body weight observation with helpers
	weight := QuantityKg(75.5)
	status := r4.ObservationStatus("final")

	obs := r4.Observation{
		Status:        &status,
		Category:      []r4.CodeableConcept{ObservationCategoryVitalSigns},
		Code:          BodyWeight,
		ValueQuantity: &weight,
	}

	// Verify the observation was created correctly
	if obs.Status == nil || *obs.Status != "final" {
		t.Errorf("expected status 'final', got %v", obs.Status)
	}

	if len(obs.Category) != 1 {
		t.Fatalf("expected 1 category, got %d", len(obs.Category))
	}

	if *obs.Category[0].Coding[0].Code != "vital-signs" {
		t.Errorf("expected category code 'vital-signs', got %s", *obs.Category[0].Coding[0].Code)
	}

	if *obs.Code.Coding[0].Code != "29463-7" {
		t.Errorf("expected LOINC code '29463-7', got %s", *obs.Code.Coding[0].Code)
	}

	if *obs.ValueQuantity.Value != 75.5 {
		t.Errorf("expected value 75.5, got %f", *obs.ValueQuantity.Value)
	}
}

func TestCreateBloodPressureObservation(t *testing.T) {
	// Test creating a blood pressure observation with components
	systolic := QuantityMmHg(120)
	diastolic := QuantityMmHg(80)
	status := r4.ObservationStatus("final")

	obs := r4.Observation{
		Status:   &status,
		Category: []r4.CodeableConcept{ObservationCategoryVitalSigns},
		Code:     BloodPressurePanel,
		Component: []r4.ObservationComponent{
			{
				Code:          SystolicBloodPressure,
				ValueQuantity: &systolic,
			},
			{
				Code:          DiastolicBloodPressure,
				ValueQuantity: &diastolic,
			},
		},
	}

	// Verify status
	if obs.Status == nil || *obs.Status != "final" {
		t.Errorf("expected status 'final', got %v", obs.Status)
	}

	// Verify category
	if len(obs.Category) != 1 || *obs.Category[0].Coding[0].Code != "vital-signs" {
		t.Errorf("expected category 'vital-signs'")
	}

	// Verify code
	if *obs.Code.Coding[0].Code != "85354-9" {
		t.Errorf("expected BP panel code '85354-9', got %s", *obs.Code.Coding[0].Code)
	}

	if len(obs.Component) != 2 {
		t.Fatalf("expected 2 components, got %d", len(obs.Component))
	}

	if *obs.Component[0].Code.Coding[0].Code != "8480-6" {
		t.Errorf("expected systolic LOINC code '8480-6', got %s", *obs.Component[0].Code.Coding[0].Code)
	}

	if *obs.Component[1].Code.Coding[0].Code != "8462-4" {
		t.Errorf("expected diastolic LOINC code '8462-4', got %s", *obs.Component[1].Code.Coding[0].Code)
	}
}

func TestCreateLabObservation(t *testing.T) {
	// Test creating a laboratory observation (glucose)
	glucose := QuantityMgDL(95)
	status := r4.ObservationStatus("final")

	obs := r4.Observation{
		Status:        &status,
		Category:      []r4.CodeableConcept{ObservationCategoryLaboratory},
		Code:          GlucoseFasting,
		ValueQuantity: &glucose,
	}

	// Verify status
	if obs.Status == nil || *obs.Status != "final" {
		t.Errorf("expected status 'final', got %v", obs.Status)
	}

	if *obs.Category[0].Coding[0].Code != "laboratory" {
		t.Errorf("expected category 'laboratory', got %s", *obs.Category[0].Coding[0].Code)
	}

	if *obs.Code.Coding[0].Code != "1558-6" {
		t.Errorf("expected LOINC code '1558-6', got %s", *obs.Code.Coding[0].Code)
	}

	if *obs.ValueQuantity.Code != "mg/dL" {
		t.Errorf("expected unit code 'mg/dL', got %s", *obs.ValueQuantity.Code)
	}
}

func TestCreateConditionWithCategory(t *testing.T) {
	// Test creating a condition with problem-list-item category
	condition := r4.Condition{
		Category: []r4.CodeableConcept{ConditionCategoryProblemListItem},
		Code: &r4.CodeableConcept{
			Coding: []r4.Coding{{
				System:  ptr("http://snomed.info/sct"),
				Code:    ptr("73211009"),
				Display: ptr("Diabetes mellitus"),
			}},
			Text: ptr("Diabetes"),
		},
	}

	// Verify category
	if *condition.Category[0].Coding[0].Code != "problem-list-item" {
		t.Errorf("expected category 'problem-list-item', got %s", *condition.Category[0].Coding[0].Code)
	}

	// Verify code
	if *condition.Code.Coding[0].Code != "73211009" {
		t.Errorf("expected SNOMED code '73211009', got %s", *condition.Code.Coding[0].Code)
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

func assertCodeableConcept(t *testing.T, cc r4.CodeableConcept, wantSystem, wantCode, wantDisplay string) {
	t.Helper()

	if len(cc.Coding) == 0 {
		t.Fatal("expected at least one coding, got none")
	}

	coding := cc.Coding[0]

	if coding.System == nil || *coding.System != wantSystem {
		got := "<nil>"
		if coding.System != nil {
			got = *coding.System
		}
		t.Errorf("system: expected %s, got %s", wantSystem, got)
	}

	if coding.Code == nil || *coding.Code != wantCode {
		got := "<nil>"
		if coding.Code != nil {
			got = *coding.Code
		}
		t.Errorf("code: expected %s, got %s", wantCode, got)
	}

	if coding.Display == nil || *coding.Display != wantDisplay {
		got := "<nil>"
		if coding.Display != nil {
			got = *coding.Display
		}
		t.Errorf("display: expected %s, got %s", wantDisplay, got)
	}
}

func assertQuantity(t *testing.T, q r4.Quantity, wantValue float64, wantUnit, wantCode, wantSystem string) {
	t.Helper()

	if q.Value == nil || *q.Value != wantValue {
		got := "<nil>"
		if q.Value != nil {
			got = string(rune(*q.Value))
		}
		t.Errorf("value: expected %f, got %s", wantValue, got)
	}

	if q.Unit == nil || *q.Unit != wantUnit {
		got := "<nil>"
		if q.Unit != nil {
			got = *q.Unit
		}
		t.Errorf("unit: expected %s, got %s", wantUnit, got)
	}

	if q.Code == nil || *q.Code != wantCode {
		got := "<nil>"
		if q.Code != nil {
			got = *q.Code
		}
		t.Errorf("code: expected %s, got %s", wantCode, got)
	}

	if q.System == nil || *q.System != wantSystem {
		got := "<nil>"
		if q.System != nil {
			got = *q.System
		}
		t.Errorf("system: expected %s, got %s", wantSystem, got)
	}
}

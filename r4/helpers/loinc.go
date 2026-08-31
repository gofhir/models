// Package helpers provides clinical helper functions and constants for FHIR R4.
// This includes LOINC codes, UCUM units, and other clinical coding standards.
package helpers

import "github.com/gofhir/models/r4/v2"

// LOINCSystem is the official LOINC code system URL.
const LOINCSystem = "http://loinc.org"

// =============================================================================
// Vital Signs - LOINC Codes
// =============================================================================

// VitalSignsPanel returns the LOINC code for the vital signs panel.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func VitalSignsPanel() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("85353-1"),
			Display: ptr("Vital signs, weight, height, head circumference, oxygen saturation and BMI panel"),
		}},
		Text: ptr("Vital Signs Panel"),
	}
}

// BodyWeight returns the LOINC code for body weight.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func BodyWeight() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("29463-7"),
			Display: ptr("Body weight"),
		}},
		Text: ptr("Body Weight"),
	}
}

// BodyHeight returns the LOINC code for body height.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func BodyHeight() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("8302-2"),
			Display: ptr("Body height"),
		}},
		Text: ptr("Body Height"),
	}
}

// BodyTemperature returns the LOINC code for body temperature.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func BodyTemperature() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("8310-5"),
			Display: ptr("Body temperature"),
		}},
		Text: ptr("Body Temperature"),
	}
}

// HeartRate returns the LOINC code for heart rate.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func HeartRate() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("8867-4"),
			Display: ptr("Heart rate"),
		}},
		Text: ptr("Heart Rate"),
	}
}

// RespiratoryRate returns the LOINC code for respiratory rate.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func RespiratoryRate() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("9279-1"),
			Display: ptr("Respiratory rate"),
		}},
		Text: ptr("Respiratory Rate"),
	}
}

// BloodPressurePanel returns the LOINC code for blood pressure panel.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func BloodPressurePanel() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("85354-9"),
			Display: ptr("Blood pressure panel with all children optional"),
		}},
		Text: ptr("Blood Pressure"),
	}
}

// SystolicBloodPressure returns the LOINC code for systolic blood pressure.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func SystolicBloodPressure() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("8480-6"),
			Display: ptr("Systolic blood pressure"),
		}},
		Text: ptr("Systolic Blood Pressure"),
	}
}

// DiastolicBloodPressure returns the LOINC code for diastolic blood pressure.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func DiastolicBloodPressure() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("8462-4"),
			Display: ptr("Diastolic blood pressure"),
		}},
		Text: ptr("Diastolic Blood Pressure"),
	}
}

// OxygenSaturation returns the LOINC code for oxygen saturation (SpO2).
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func OxygenSaturation() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("2708-6"),
			Display: ptr("Oxygen saturation in Arterial blood"),
		}},
		Text: ptr("Oxygen Saturation"),
	}
}

// PulseOximetry returns the LOINC code for pulse oximetry.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func PulseOximetry() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("59408-5"),
			Display: ptr("Oxygen saturation in Arterial blood by Pulse oximetry"),
		}},
		Text: ptr("Pulse Oximetry"),
	}
}

// BMI returns the LOINC code for Body Mass Index.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func BMI() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("39156-5"),
			Display: ptr("Body mass index (BMI) [Ratio]"),
		}},
		Text: ptr("BMI"),
	}
}

// HeadCircumference returns the LOINC code for head circumference.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func HeadCircumference() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("9843-4"),
			Display: ptr("Head Occipital-frontal circumference"),
		}},
		Text: ptr("Head Circumference"),
	}
}

// =============================================================================
// Laboratory - Common LOINC Codes
// =============================================================================

// Glucose returns the LOINC code for glucose in blood.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func Glucose() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("2339-0"),
			Display: ptr("Glucose [Mass/volume] in Blood"),
		}},
		Text: ptr("Blood Glucose"),
	}
}

// GlucoseFasting returns the LOINC code for fasting glucose.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func GlucoseFasting() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("1558-6"),
			Display: ptr("Fasting glucose [Mass/volume] in Serum or Plasma"),
		}},
		Text: ptr("Fasting Glucose"),
	}
}

// HemoglobinA1c returns the LOINC code for HbA1c.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func HemoglobinA1c() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("4548-4"),
			Display: ptr("Hemoglobin A1c/Hemoglobin.total in Blood"),
		}},
		Text: ptr("HbA1c"),
	}
}

// Hemoglobin returns the LOINC code for hemoglobin.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func Hemoglobin() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("718-7"),
			Display: ptr("Hemoglobin [Mass/volume] in Blood"),
		}},
		Text: ptr("Hemoglobin"),
	}
}

// Hematocrit returns the LOINC code for hematocrit.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func Hematocrit() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("4544-3"),
			Display: ptr("Hematocrit [Volume Fraction] of Blood by Automated count"),
		}},
		Text: ptr("Hematocrit"),
	}
}

// Creatinine returns the LOINC code for creatinine.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func Creatinine() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("2160-0"),
			Display: ptr("Creatinine [Mass/volume] in Serum or Plasma"),
		}},
		Text: ptr("Creatinine"),
	}
}

// EGFR returns the LOINC code for estimated glomerular filtration rate (eGFR).
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func EGFR() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("33914-3"),
			Display: ptr("Glomerular filtration rate/1.73 sq M.predicted [Volume Rate/Area] in Serum or Plasma by Creatinine-based formula (MDRD)"),
		}},
		Text: ptr("eGFR"),
	}
}

// Cholesterol returns the LOINC code for total cholesterol.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func Cholesterol() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("2093-3"),
			Display: ptr("Cholesterol [Mass/volume] in Serum or Plasma"),
		}},
		Text: ptr("Total Cholesterol"),
	}
}

// HDLCholesterol returns the LOINC code for HDL cholesterol.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func HDLCholesterol() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("2085-9"),
			Display: ptr("Cholesterol in HDL [Mass/volume] in Serum or Plasma"),
		}},
		Text: ptr("HDL Cholesterol"),
	}
}

// LDLCholesterol returns the LOINC code for LDL cholesterol.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func LDLCholesterol() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("2089-1"),
			Display: ptr("Cholesterol in LDL [Mass/volume] in Serum or Plasma"),
		}},
		Text: ptr("LDL Cholesterol"),
	}
}

// Triglycerides returns the LOINC code for triglycerides.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func Triglycerides() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("2571-8"),
			Display: ptr("Triglyceride [Mass/volume] in Serum or Plasma"),
		}},
		Text: ptr("Triglycerides"),
	}
}

// =============================================================================
// IPS (International Patient Summary) Section LOINC Codes
// =============================================================================

// IPSMedicationSummary returns the LOINC code for IPS Medication Summary section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSMedicationSummary() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("10160-0"),
			Display: ptr("History of Medication use Narrative"),
		}},
		Text: ptr("Medication Summary"),
	}
}

// IPSAllergies returns the LOINC code for IPS Allergies and Intolerances section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSAllergies() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("48765-2"),
			Display: ptr("Allergies and adverse reactions Document"),
		}},
		Text: ptr("Allergies and Intolerances"),
	}
}

// IPSProblems returns the LOINC code for IPS Problem List section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSProblems() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("11450-4"),
			Display: ptr("Problem list - Reported"),
		}},
		Text: ptr("Problem List"),
	}
}

// IPSImmunizations returns the LOINC code for IPS Immunizations section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSImmunizations() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("11369-6"),
			Display: ptr("History of Immunization Narrative"),
		}},
		Text: ptr("Immunizations"),
	}
}

// IPSProcedures returns the LOINC code for IPS History of Procedures section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSProcedures() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("47519-4"),
			Display: ptr("History of Procedures Document"),
		}},
		Text: ptr("History of Procedures"),
	}
}

// IPSMedicalDevices returns the LOINC code for IPS Medical Devices section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSMedicalDevices() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("46264-8"),
			Display: ptr("History of medical device use"),
		}},
		Text: ptr("Medical Devices"),
	}
}

// IPSDiagnosticResults returns the LOINC code for IPS Diagnostic Results section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSDiagnosticResults() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("30954-2"),
			Display: ptr("Relevant diagnostic tests/laboratory data Narrative"),
		}},
		Text: ptr("Diagnostic Results"),
	}
}

// IPSVitalSigns returns the LOINC code for IPS Vital Signs section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSVitalSigns() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("8716-3"),
			Display: ptr("Vital signs"),
		}},
		Text: ptr("Vital Signs"),
	}
}

// IPSPastIllness returns the LOINC code for IPS History of Past Illness section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSPastIllness() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("11348-0"),
			Display: ptr("History of Past illness Narrative"),
		}},
		Text: ptr("History of Past Illness"),
	}
}

// IPSFunctionalStatus returns the LOINC code for IPS Functional Status section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSFunctionalStatus() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("47420-5"),
			Display: ptr("Functional status assessment note"),
		}},
		Text: ptr("Functional Status"),
	}
}

// IPSPlanOfCare returns the LOINC code for IPS Plan of Care section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSPlanOfCare() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("18776-5"),
			Display: ptr("Plan of care note"),
		}},
		Text: ptr("Plan of Care"),
	}
}

// IPSSocialHistory returns the LOINC code for IPS Social History section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSSocialHistory() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("29762-2"),
			Display: ptr("Social history Narrative"),
		}},
		Text: ptr("Social History"),
	}
}

// IPSPregnancy returns the LOINC code for IPS Pregnancy section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSPregnancy() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("10162-6"),
			Display: ptr("History of pregnancies Narrative"),
		}},
		Text: ptr("Pregnancy History"),
	}
}

// IPSAdvanceDirectives returns the LOINC code for IPS Advance Directives section.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func IPSAdvanceDirectives() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("42348-3"),
			Display: ptr("Advance directives"),
		}},
		Text: ptr("Advance Directives"),
	}
}

// ptr is a helper function to create a pointer to a string.
func ptr(s string) *string {
	return &s
}

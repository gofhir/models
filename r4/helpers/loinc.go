// Package helpers provides clinical helper functions and constants for FHIR R4.
// This includes LOINC codes, UCUM units, and other clinical coding standards.
package helpers

import "github.com/robertoaraneda/gofhir/r4"

// LOINCSystem is the official LOINC code system URL.
const LOINCSystem = "http://loinc.org"

// =============================================================================
// Vital Signs - LOINC Codes
// =============================================================================

// VitalSignsPanel is the LOINC code for the vital signs panel.
var VitalSignsPanel = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("85353-1"),
		Display: ptr("Vital signs, weight, height, head circumference, oxygen saturation and BMI panel"),
	}},
	Text: ptr("Vital Signs Panel"),
}

// BodyWeight is the LOINC code for body weight.
var BodyWeight = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("29463-7"),
		Display: ptr("Body weight"),
	}},
	Text: ptr("Body Weight"),
}

// BodyHeight is the LOINC code for body height.
var BodyHeight = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("8302-2"),
		Display: ptr("Body height"),
	}},
	Text: ptr("Body Height"),
}

// BodyTemperature is the LOINC code for body temperature.
var BodyTemperature = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("8310-5"),
		Display: ptr("Body temperature"),
	}},
	Text: ptr("Body Temperature"),
}

// HeartRate is the LOINC code for heart rate.
var HeartRate = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("8867-4"),
		Display: ptr("Heart rate"),
	}},
	Text: ptr("Heart Rate"),
}

// RespiratoryRate is the LOINC code for respiratory rate.
var RespiratoryRate = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("9279-1"),
		Display: ptr("Respiratory rate"),
	}},
	Text: ptr("Respiratory Rate"),
}

// BloodPressurePanel is the LOINC code for blood pressure panel.
var BloodPressurePanel = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("85354-9"),
		Display: ptr("Blood pressure panel with all children optional"),
	}},
	Text: ptr("Blood Pressure"),
}

// SystolicBloodPressure is the LOINC code for systolic blood pressure.
var SystolicBloodPressure = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("8480-6"),
		Display: ptr("Systolic blood pressure"),
	}},
	Text: ptr("Systolic Blood Pressure"),
}

// DiastolicBloodPressure is the LOINC code for diastolic blood pressure.
var DiastolicBloodPressure = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("8462-4"),
		Display: ptr("Diastolic blood pressure"),
	}},
	Text: ptr("Diastolic Blood Pressure"),
}

// OxygenSaturation is the LOINC code for oxygen saturation (SpO2).
var OxygenSaturation = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("2708-6"),
		Display: ptr("Oxygen saturation in Arterial blood"),
	}},
	Text: ptr("Oxygen Saturation"),
}

// PulseOximetry is the LOINC code for pulse oximetry.
var PulseOximetry = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("59408-5"),
		Display: ptr("Oxygen saturation in Arterial blood by Pulse oximetry"),
	}},
	Text: ptr("Pulse Oximetry"),
}

// BMI is the LOINC code for Body Mass Index.
var BMI = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("39156-5"),
		Display: ptr("Body mass index (BMI) [Ratio]"),
	}},
	Text: ptr("BMI"),
}

// HeadCircumference is the LOINC code for head circumference.
var HeadCircumference = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("9843-4"),
		Display: ptr("Head Occipital-frontal circumference"),
	}},
	Text: ptr("Head Circumference"),
}

// =============================================================================
// Laboratory - Common LOINC Codes
// =============================================================================

// Glucose is the LOINC code for glucose in blood.
var Glucose = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("2339-0"),
		Display: ptr("Glucose [Mass/volume] in Blood"),
	}},
	Text: ptr("Blood Glucose"),
}

// GlucoseFasting is the LOINC code for fasting glucose.
var GlucoseFasting = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("1558-6"),
		Display: ptr("Fasting glucose [Mass/volume] in Serum or Plasma"),
	}},
	Text: ptr("Fasting Glucose"),
}

// HemoglobinA1c is the LOINC code for HbA1c.
var HemoglobinA1c = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("4548-4"),
		Display: ptr("Hemoglobin A1c/Hemoglobin.total in Blood"),
	}},
	Text: ptr("HbA1c"),
}

// Hemoglobin is the LOINC code for hemoglobin.
var Hemoglobin = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("718-7"),
		Display: ptr("Hemoglobin [Mass/volume] in Blood"),
	}},
	Text: ptr("Hemoglobin"),
}

// Hematocrit is the LOINC code for hematocrit.
var Hematocrit = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("4544-3"),
		Display: ptr("Hematocrit [Volume Fraction] of Blood by Automated count"),
	}},
	Text: ptr("Hematocrit"),
}

// Creatinine is the LOINC code for creatinine.
var Creatinine = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("2160-0"),
		Display: ptr("Creatinine [Mass/volume] in Serum or Plasma"),
	}},
	Text: ptr("Creatinine"),
}

// eGFR is the LOINC code for estimated glomerular filtration rate.
var EGFR = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("33914-3"),
		Display: ptr("Glomerular filtration rate/1.73 sq M.predicted [Volume Rate/Area] in Serum or Plasma by Creatinine-based formula (MDRD)"),
	}},
	Text: ptr("eGFR"),
}

// Cholesterol is the LOINC code for total cholesterol.
var Cholesterol = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("2093-3"),
		Display: ptr("Cholesterol [Mass/volume] in Serum or Plasma"),
	}},
	Text: ptr("Total Cholesterol"),
}

// HDLCholesterol is the LOINC code for HDL cholesterol.
var HDLCholesterol = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("2085-9"),
		Display: ptr("Cholesterol in HDL [Mass/volume] in Serum or Plasma"),
	}},
	Text: ptr("HDL Cholesterol"),
}

// LDLCholesterol is the LOINC code for LDL cholesterol.
var LDLCholesterol = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("2089-1"),
		Display: ptr("Cholesterol in LDL [Mass/volume] in Serum or Plasma"),
	}},
	Text: ptr("LDL Cholesterol"),
}

// Triglycerides is the LOINC code for triglycerides.
var Triglycerides = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("2571-8"),
		Display: ptr("Triglyceride [Mass/volume] in Serum or Plasma"),
	}},
	Text: ptr("Triglycerides"),
}

// =============================================================================
// IPS (International Patient Summary) Section LOINC Codes
// =============================================================================

// IPSMedicationSummary is the LOINC code for IPS Medication Summary section.
var IPSMedicationSummary = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("10160-0"),
		Display: ptr("History of Medication use Narrative"),
	}},
	Text: ptr("Medication Summary"),
}

// IPSAllergies is the LOINC code for IPS Allergies and Intolerances section.
var IPSAllergies = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("48765-2"),
		Display: ptr("Allergies and adverse reactions Document"),
	}},
	Text: ptr("Allergies and Intolerances"),
}

// IPSProblems is the LOINC code for IPS Problem List section.
var IPSProblems = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("11450-4"),
		Display: ptr("Problem list - Reported"),
	}},
	Text: ptr("Problem List"),
}

// IPSImmunizations is the LOINC code for IPS Immunizations section.
var IPSImmunizations = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("11369-6"),
		Display: ptr("History of Immunization Narrative"),
	}},
	Text: ptr("Immunizations"),
}

// IPSProcedures is the LOINC code for IPS History of Procedures section.
var IPSProcedures = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("47519-4"),
		Display: ptr("History of Procedures Document"),
	}},
	Text: ptr("History of Procedures"),
}

// IPSMedicalDevices is the LOINC code for IPS Medical Devices section.
var IPSMedicalDevices = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("46264-8"),
		Display: ptr("History of medical device use"),
	}},
	Text: ptr("Medical Devices"),
}

// IPSDiagnosticResults is the LOINC code for IPS Diagnostic Results section.
var IPSDiagnosticResults = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("30954-2"),
		Display: ptr("Relevant diagnostic tests/laboratory data Narrative"),
	}},
	Text: ptr("Diagnostic Results"),
}

// IPSVitalSigns is the LOINC code for IPS Vital Signs section.
var IPSVitalSigns = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("8716-3"),
		Display: ptr("Vital signs"),
	}},
	Text: ptr("Vital Signs"),
}

// IPSPastIllness is the LOINC code for IPS History of Past Illness section.
var IPSPastIllness = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("11348-0"),
		Display: ptr("History of Past illness Narrative"),
	}},
	Text: ptr("History of Past Illness"),
}

// IPSFunctionalStatus is the LOINC code for IPS Functional Status section.
var IPSFunctionalStatus = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("47420-5"),
		Display: ptr("Functional status assessment note"),
	}},
	Text: ptr("Functional Status"),
}

// IPSPlanOfCare is the LOINC code for IPS Plan of Care section.
var IPSPlanOfCare = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("18776-5"),
		Display: ptr("Plan of care note"),
	}},
	Text: ptr("Plan of Care"),
}

// IPSSocialHistory is the LOINC code for IPS Social History section.
var IPSSocialHistory = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("29762-2"),
		Display: ptr("Social history Narrative"),
	}},
	Text: ptr("Social History"),
}

// IPSPregnancy is the LOINC code for IPS Pregnancy section.
var IPSPregnancy = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("10162-6"),
		Display: ptr("History of pregnancies Narrative"),
	}},
	Text: ptr("Pregnancy History"),
}

// IPSAdvanceDirectives is the LOINC code for IPS Advance Directives section.
var IPSAdvanceDirectives = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("42348-3"),
		Display: ptr("Advance directives"),
	}},
	Text: ptr("Advance Directives"),
}

// ptr is a helper function to create a pointer to a string.
func ptr(s string) *string {
	return &s
}

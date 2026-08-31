package helpers

import "github.com/gofhir/models/r4/v2"

// ObservationCategorySystem is the HL7 observation category code system.
const ObservationCategorySystem = "http://terminology.hl7.org/CodeSystem/observation-category"

// ConditionCategorySystem is the HL7 condition category code system.
const ConditionCategorySystem = "http://terminology.hl7.org/CodeSystem/condition-category"

// AllergyIntoleranceCategorySystem is the FHIR allergy intolerance category code system.
const AllergyIntoleranceCategorySystem = "http://hl7.org/fhir/allergy-intolerance-category"

// =============================================================================
// Observation Categories
// =============================================================================

// ObservationCategoryVitalSigns returns the category for vital signs observations.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ObservationCategoryVitalSigns() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ObservationCategorySystem),
			Code:    ptr("vital-signs"),
			Display: ptr("Vital Signs"),
		}},
		Text: ptr("Vital Signs"),
	}
}

// ObservationCategoryLaboratory returns the category for laboratory observations.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ObservationCategoryLaboratory() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ObservationCategorySystem),
			Code:    ptr("laboratory"),
			Display: ptr("Laboratory"),
		}},
		Text: ptr("Laboratory"),
	}
}

// ObservationCategorySocialHistory returns the category for social history observations.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ObservationCategorySocialHistory() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ObservationCategorySystem),
			Code:    ptr("social-history"),
			Display: ptr("Social History"),
		}},
		Text: ptr("Social History"),
	}
}

// ObservationCategoryImaging returns the category for imaging observations.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ObservationCategoryImaging() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ObservationCategorySystem),
			Code:    ptr("imaging"),
			Display: ptr("Imaging"),
		}},
		Text: ptr("Imaging"),
	}
}

// ObservationCategoryProcedure returns the category for procedure observations.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ObservationCategoryProcedure() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ObservationCategorySystem),
			Code:    ptr("procedure"),
			Display: ptr("Procedure"),
		}},
		Text: ptr("Procedure"),
	}
}

// ObservationCategorySurvey returns the category for survey observations.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ObservationCategorySurvey() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ObservationCategorySystem),
			Code:    ptr("survey"),
			Display: ptr("Survey"),
		}},
		Text: ptr("Survey"),
	}
}

// ObservationCategoryExam returns the category for exam observations.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ObservationCategoryExam() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ObservationCategorySystem),
			Code:    ptr("exam"),
			Display: ptr("Exam"),
		}},
		Text: ptr("Exam"),
	}
}

// ObservationCategoryTherapy returns the category for therapy observations.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ObservationCategoryTherapy() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ObservationCategorySystem),
			Code:    ptr("therapy"),
			Display: ptr("Therapy"),
		}},
		Text: ptr("Therapy"),
	}
}

// ObservationCategoryActivity returns the category for activity observations.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ObservationCategoryActivity() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ObservationCategorySystem),
			Code:    ptr("activity"),
			Display: ptr("Activity"),
		}},
		Text: ptr("Activity"),
	}
}

// =============================================================================
// Condition Categories
// =============================================================================

// ConditionCategoryProblemListItem returns the category for problem list items.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ConditionCategoryProblemListItem() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ConditionCategorySystem),
			Code:    ptr("problem-list-item"),
			Display: ptr("Problem List Item"),
		}},
		Text: ptr("Problem List Item"),
	}
}

// ConditionCategoryEncounterDiagnosis returns the category for encounter diagnoses.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func ConditionCategoryEncounterDiagnosis() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(ConditionCategorySystem),
			Code:    ptr("encounter-diagnosis"),
			Display: ptr("Encounter Diagnosis"),
		}},
		Text: ptr("Encounter Diagnosis"),
	}
}

// =============================================================================
// Allergy/Intolerance Categories
// =============================================================================

// AllergyCategoryFood returns the category for food allergies.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func AllergyCategoryFood() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(AllergyIntoleranceCategorySystem),
			Code:    ptr("food"),
			Display: ptr("Food"),
		}},
		Text: ptr("Food Allergy"),
	}
}

// AllergyCategoryMedication returns the category for medication allergies.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func AllergyCategoryMedication() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(AllergyIntoleranceCategorySystem),
			Code:    ptr("medication"),
			Display: ptr("Medication"),
		}},
		Text: ptr("Medication Allergy"),
	}
}

// AllergyCategoryEnvironment returns the category for environmental allergies.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func AllergyCategoryEnvironment() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(AllergyIntoleranceCategorySystem),
			Code:    ptr("environment"),
			Display: ptr("Environment"),
		}},
		Text: ptr("Environmental Allergy"),
	}
}

// AllergyCategoryBiologic returns the category for biologic allergies.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func AllergyCategoryBiologic() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(AllergyIntoleranceCategorySystem),
			Code:    ptr("biologic"),
			Display: ptr("Biologic"),
		}},
		Text: ptr("Biologic Allergy"),
	}
}

// =============================================================================
// Document Type Codes (for Composition/DocumentReference)
// =============================================================================

// DocumentTypeIPS returns the LOINC code for International Patient Summary.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func DocumentTypeIPS() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("60591-5"),
			Display: ptr("Patient summary Document"),
		}},
		Text: ptr("International Patient Summary"),
	}
}

// DocumentTypeCCD returns the LOINC code for Continuity of Care Document.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func DocumentTypeCCD() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("34133-9"),
			Display: ptr("Summary of episode note"),
		}},
		Text: ptr("Continuity of Care Document"),
	}
}

// DocumentTypeDischarge returns the LOINC code for Discharge Summary.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func DocumentTypeDischarge() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("18842-5"),
			Display: ptr("Discharge summary"),
		}},
		Text: ptr("Discharge Summary"),
	}
}

// DocumentTypeProgress returns the LOINC code for Progress Note.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func DocumentTypeProgress() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("11506-3"),
			Display: ptr("Progress note"),
		}},
		Text: ptr("Progress Note"),
	}
}

// DocumentTypeHistory returns the LOINC code for History and Physical.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func DocumentTypeHistory() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("34117-2"),
			Display: ptr("History and physical note"),
		}},
		Text: ptr("History and Physical"),
	}
}

// DocumentTypeConsult returns the LOINC code for Consultation Note.
//
// Each call builds a new value, so the result can be mutated without
// affecting any other caller.
func DocumentTypeConsult() *r4.CodeableConcept {
	return &r4.CodeableConcept{
		Coding: []r4.Coding{{
			System:  ptr(LOINCSystem),
			Code:    ptr("11488-4"),
			Display: ptr("Consultation note"),
		}},
		Text: ptr("Consultation Note"),
	}
}

package helpers

import "github.com/robertoaraneda/gofhir/r4"

// ObservationCategorySystem is the HL7 observation category code system.
const ObservationCategorySystem = "http://terminology.hl7.org/CodeSystem/observation-category"

// ConditionCategorySystem is the HL7 condition category code system.
const ConditionCategorySystem = "http://terminology.hl7.org/CodeSystem/condition-category"

// AllergyIntoleranceCategorySystem is the FHIR allergy intolerance category code system.
const AllergyIntoleranceCategorySystem = "http://hl7.org/fhir/allergy-intolerance-category"

// =============================================================================
// Observation Categories
// =============================================================================

// ObservationCategoryVitalSigns is the category for vital signs observations.
var ObservationCategoryVitalSigns = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ObservationCategorySystem),
		Code:    ptr("vital-signs"),
		Display: ptr("Vital Signs"),
	}},
	Text: ptr("Vital Signs"),
}

// ObservationCategoryLaboratory is the category for laboratory observations.
var ObservationCategoryLaboratory = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ObservationCategorySystem),
		Code:    ptr("laboratory"),
		Display: ptr("Laboratory"),
	}},
	Text: ptr("Laboratory"),
}

// ObservationCategorySocialHistory is the category for social history observations.
var ObservationCategorySocialHistory = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ObservationCategorySystem),
		Code:    ptr("social-history"),
		Display: ptr("Social History"),
	}},
	Text: ptr("Social History"),
}

// ObservationCategoryImaging is the category for imaging observations.
var ObservationCategoryImaging = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ObservationCategorySystem),
		Code:    ptr("imaging"),
		Display: ptr("Imaging"),
	}},
	Text: ptr("Imaging"),
}

// ObservationCategoryProcedure is the category for procedure observations.
var ObservationCategoryProcedure = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ObservationCategorySystem),
		Code:    ptr("procedure"),
		Display: ptr("Procedure"),
	}},
	Text: ptr("Procedure"),
}

// ObservationCategorySurvey is the category for survey observations.
var ObservationCategorySurvey = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ObservationCategorySystem),
		Code:    ptr("survey"),
		Display: ptr("Survey"),
	}},
	Text: ptr("Survey"),
}

// ObservationCategoryExam is the category for exam observations.
var ObservationCategoryExam = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ObservationCategorySystem),
		Code:    ptr("exam"),
		Display: ptr("Exam"),
	}},
	Text: ptr("Exam"),
}

// ObservationCategoryTherapy is the category for therapy observations.
var ObservationCategoryTherapy = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ObservationCategorySystem),
		Code:    ptr("therapy"),
		Display: ptr("Therapy"),
	}},
	Text: ptr("Therapy"),
}

// ObservationCategoryActivity is the category for activity observations.
var ObservationCategoryActivity = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ObservationCategorySystem),
		Code:    ptr("activity"),
		Display: ptr("Activity"),
	}},
	Text: ptr("Activity"),
}

// =============================================================================
// Condition Categories
// =============================================================================

// ConditionCategoryProblemListItem is the category for problem list items.
var ConditionCategoryProblemListItem = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ConditionCategorySystem),
		Code:    ptr("problem-list-item"),
		Display: ptr("Problem List Item"),
	}},
	Text: ptr("Problem List Item"),
}

// ConditionCategoryEncounterDiagnosis is the category for encounter diagnoses.
var ConditionCategoryEncounterDiagnosis = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(ConditionCategorySystem),
		Code:    ptr("encounter-diagnosis"),
		Display: ptr("Encounter Diagnosis"),
	}},
	Text: ptr("Encounter Diagnosis"),
}

// =============================================================================
// Allergy/Intolerance Categories
// =============================================================================

// AllergyCategoryFood is the category for food allergies.
var AllergyCategoryFood = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(AllergyIntoleranceCategorySystem),
		Code:    ptr("food"),
		Display: ptr("Food"),
	}},
	Text: ptr("Food Allergy"),
}

// AllergyCategoryMedication is the category for medication allergies.
var AllergyCategoryMedication = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(AllergyIntoleranceCategorySystem),
		Code:    ptr("medication"),
		Display: ptr("Medication"),
	}},
	Text: ptr("Medication Allergy"),
}

// AllergyCategoryEnvironment is the category for environmental allergies.
var AllergyCategoryEnvironment = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(AllergyIntoleranceCategorySystem),
		Code:    ptr("environment"),
		Display: ptr("Environment"),
	}},
	Text: ptr("Environmental Allergy"),
}

// AllergyCategoryBiologic is the category for biologic allergies.
var AllergyCategoryBiologic = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(AllergyIntoleranceCategorySystem),
		Code:    ptr("biologic"),
		Display: ptr("Biologic"),
	}},
	Text: ptr("Biologic Allergy"),
}

// =============================================================================
// Document Type Codes (for Composition/DocumentReference)
// =============================================================================

// DocumentTypeIPS is the LOINC code for International Patient Summary.
var DocumentTypeIPS = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("60591-5"),
		Display: ptr("Patient summary Document"),
	}},
	Text: ptr("International Patient Summary"),
}

// DocumentTypeCCD is the LOINC code for Continuity of Care Document.
var DocumentTypeCCD = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("34133-9"),
		Display: ptr("Summary of episode note"),
	}},
	Text: ptr("Continuity of Care Document"),
}

// DocumentTypeDischarge is the LOINC code for Discharge Summary.
var DocumentTypeDischarge = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("18842-5"),
		Display: ptr("Discharge summary"),
	}},
	Text: ptr("Discharge Summary"),
}

// DocumentTypeProgress is the LOINC code for Progress Note.
var DocumentTypeProgress = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("11506-3"),
		Display: ptr("Progress note"),
	}},
	Text: ptr("Progress Note"),
}

// DocumentTypeHistory is the LOINC code for History and Physical.
var DocumentTypeHistory = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("34117-2"),
		Display: ptr("History and physical note"),
	}},
	Text: ptr("History and Physical"),
}

// DocumentTypeConsult is the LOINC code for Consultation Note.
var DocumentTypeConsult = r4.CodeableConcept{
	Coding: []r4.Coding{{
		System:  ptr(LOINCSystem),
		Code:    ptr("11488-4"),
		Display: ptr("Consultation note"),
	}},
	Text: ptr("Consultation Note"),
}

package conformance

// fhirpath_model.go is the largest generated file in every version — 9,900 lines in
// r4, 12,221 in r5 — and it was the least tested: 13 test functions in r4 against 4
// in r4b and r5, on content that genuinely differs between versions. Testing r4
// says nothing about r5 here, unlike decimal.go, which is the same code three times.
//
// The gap is closed once rather than by cloning r4's file into two packages. The
// existing r4b and r5 test files already carry function-for-function identical names
// — they were copied, not written — and adding more of that would raise the line
// count without raising the guarantee.
//
// So the model is exercised through an interface, with the facts that differ by
// version stated per version. A case that holds everywhere is written once; a case
// that does not is where the versions actually disagree, and saying so is the point.

import (
	"slices"
	"testing"

	"github.com/gofhir/models/r4/v2"
	"github.com/gofhir/models/r4b/v2"
	"github.com/gofhir/models/r5/v2"
)

// fhirPathModel is the surface every version's model exposes. The generated types
// are per-package, so this is what lets one test drive all three.
type fhirPathModel interface {
	FHIRVersion() string
	ChoiceTypes(path string) []string
	TypeOf(path string) string
	ReferenceTargets(path string) []string
	ParentType(typeName string) string
	IsSubtype(child, parent string) bool
	ResolvePath(path string) string
	HasType(typeName string) bool
	IsResource(typeName string) bool
	ChildElements(path string) []string
}

type modelCase struct {
	name    string
	model   fhirPathModel
	version string
	// valueChoices is Observation.value[x]: R5 adds Attachment and Reference.
	valueChoices []string
	// present and absent are types that exist in some versions and not others,
	// which is the sharpest evidence that each version has its own model.
	present []string
	absent  []string
}

func models() []modelCase {
	return []modelCase{
		{
			name: "r4", model: r4.FHIRPathModel(), version: "4.0.1",
			valueChoices: []string{"Quantity", "CodeableConcept", "string", "boolean", "integer",
				"Range", "Ratio", "SampledData", "time", "dateTime", "Period"},
			present: []string{"MedicinalProduct", "Patient", "HumanName"},
			absent:  []string{"Citation", "InventoryItem"},
		},
		{
			name: "r4b", model: r4b.FHIRPathModel(), version: "4.3.0",
			valueChoices: []string{"Quantity", "CodeableConcept", "string", "boolean", "integer",
				"Range", "Ratio", "SampledData", "time", "dateTime", "Period"},
			present: []string{"Citation", "Patient", "HumanName"},
			absent:  []string{"MedicinalProduct", "InventoryItem"},
		},
		{
			name: "r5", model: r5.FHIRPathModel(), version: "5.0.0",
			valueChoices: []string{"Quantity", "CodeableConcept", "string", "boolean", "integer",
				"Range", "Ratio", "SampledData", "time", "dateTime", "Period", "Attachment", "Reference"},
			present: []string{"Citation", "InventoryItem", "Patient", "HumanName"},
			absent:  []string{"MedicinalProduct"},
		},
	}
}

// compareSets reports which of want is absent from got, and which of got was not
// in want. Comparing lengths and then checking one direction is not enough: with
// want {A,B,C} and got [A,A,B] both checks pass and C is silently missing.
func compareSets(want, got []string) (missing, unexpected []string) {
	inWant := make(map[string]bool, len(want))
	for _, w := range want {
		inWant[w] = true
	}
	seen := make(map[string]bool, len(got))
	for _, g := range got {
		seen[g] = true
		if !inWant[g] {
			unexpected = append(unexpected, g)
		}
	}
	for _, w := range want {
		if !seen[w] {
			missing = append(missing, w)
		}
	}
	return missing, unexpected
}

func TestModelReportsItsFHIRVersion(t *testing.T) {
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.model.FHIRVersion(); got != tc.version {
				t.Errorf("FHIRVersion() = %q, want %q", got, tc.version)
			}
		})
	}
}

func TestModelKnowsWhichTypesItsVersionHas(t *testing.T) {
	// A model that answered the same for every version would be a model built from
	// the wrong spec, and nothing else in the suite would notice.
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			for _, name := range tc.present {
				if !tc.model.HasType(name) {
					t.Errorf("HasType(%q) = false, but %s has it", name, tc.name)
				}
			}
			for _, name := range tc.absent {
				if tc.model.HasType(name) {
					t.Errorf("HasType(%q) = true, but %s does not have it", name, tc.name)
				}
			}
		})
	}
}

func TestModelResolvesChoiceTypes(t *testing.T) {
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.model.ChoiceTypes("Observation.value")
			if missing, unexpected := compareSets(tc.valueChoices, got); len(missing) > 0 || len(unexpected) > 0 {
				t.Errorf("Observation.value choice types are wrong:\n  missing:    %v\n  unexpected: %v\n  got:        %v",
					missing, unexpected, got)
			}

			// A path that is not a choice has none.
			if got := tc.model.ChoiceTypes("Patient.name"); len(got) != 0 {
				t.Errorf("Patient.name is not a choice, got %v", got)
			}
		})
	}
}

func TestModelTypesPaths(t *testing.T) {
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			for path, want := range map[string]string{
				"Patient.name":         "HumanName",
				"Patient.active":       "boolean",
				"Observation.status":   "code",
				"Patient.contact.name": "HumanName",
			} {
				if got := tc.model.TypeOf(path); got != want {
					t.Errorf("TypeOf(%q) = %q, want %q", path, got, want)
				}
			}
			if got := tc.model.TypeOf("Patient.notAField"); got != "" {
				t.Errorf("TypeOf on an unknown path returned %q, want empty", got)
			}
		})
	}
}

func TestModelKnowsReferenceTargets(t *testing.T) {
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			want := []string{"Organization", "Practitioner", "PractitionerRole"}
			got := tc.model.ReferenceTargets("Patient.generalPractitioner")
			if missing, unexpected := compareSets(want, got); len(missing) > 0 || len(unexpected) > 0 {
				t.Errorf("Patient.generalPractitioner targets are wrong:\n  missing:    %v\n  unexpected: %v\n  got:        %v",
					missing, unexpected, got)
			}

			// A non-reference field has no targets.
			if got := tc.model.ReferenceTargets("Patient.name"); len(got) != 0 {
				t.Errorf("Patient.name is not a reference, got %v", got)
			}
		})
	}
}

func TestModelWalksTheTypeHierarchy(t *testing.T) {
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.model.ParentType("Patient"); got != "DomainResource" {
				t.Errorf("ParentType(Patient) = %q, want DomainResource", got)
			}
			if got := tc.model.ParentType("DomainResource"); got != "Resource" {
				t.Errorf("ParentType(DomainResource) = %q, want Resource", got)
			}

			// IsSubtype has to climb, not just check the immediate parent —
			// Patient's parent is DomainResource, and Resource is above that.
			if !tc.model.IsSubtype("Patient", "Resource") {
				t.Error("IsSubtype(Patient, Resource) = false; the walk stops at the direct parent")
			}
			if !tc.model.IsSubtype("Patient", "DomainResource") {
				t.Error("IsSubtype(Patient, DomainResource) = false")
			}
			if tc.model.IsSubtype("Patient", "HumanName") {
				t.Error("IsSubtype(Patient, HumanName) = true, which is not a hierarchy that exists")
			}
			if tc.model.IsSubtype("Resource", "Patient") {
				t.Error("IsSubtype is answering in both directions")
			}
		})
	}
}

func TestModelResolvesRecursivePaths(t *testing.T) {
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			// Questionnaire.item.item is a contentReference back to
			// Questionnaire.item — a recursive structure, and the reason
			// ResolvePath exists at all.
			if got := tc.model.ResolvePath("Questionnaire.item.item"); got != "Questionnaire.item" {
				t.Errorf("ResolvePath(Questionnaire.item.item) = %q, want Questionnaire.item", got)
			}
			// A path defined where it appears comes back unchanged.
			for _, path := range []string{"Patient.name", "Observation.value"} {
				if got := tc.model.ResolvePath(path); got != path {
					t.Errorf("ResolvePath(%q) = %q, want it unchanged", path, got)
				}
			}
		})
	}
}

func TestModelSeparatesResourcesFromDatatypes(t *testing.T) {
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			for _, name := range []string{"Patient", "Observation", "Bundle"} {
				if !tc.model.IsResource(name) {
					t.Errorf("IsResource(%q) = false", name)
				}
			}
			for _, name := range []string{"HumanName", "CodeableConcept", "Period", "Element"} {
				if tc.model.IsResource(name) {
					t.Errorf("IsResource(%q) = true, but it is a datatype", name)
				}
			}
			// Abstract bases are in the hierarchy but are not resources one can
			// instantiate, and holding both halves of that at once is what
			// IsResource is for — HasType alone would not tell them apart.
			for _, name := range []string{"Resource", "DomainResource"} {
				if !tc.model.HasType(name) {
					t.Errorf("%s is missing from the type hierarchy", name)
				}
				if tc.model.IsResource(name) {
					t.Errorf("IsResource(%q) = true, but it is an abstract base", name)
				}
			}
		})
	}
}

// JSON does not order an object's keys, so the model is the only record a
// FHIRPath engine has of the order a definition lists an element's children
// in, which children() and descendants() return them in, as the HL7 validator
// does. Every type and every element path with children is listed, in its
// snapshot's order, with a choice element's [x] kept.
func TestModelListsChildrenInTheDefinitionsOrder(t *testing.T) {
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			for path, want := range map[string][]string{
				"Reference": {"id", "extension", "reference", "type", "identifier", "display"},
				"Extension": {"id", "extension", "url", "value[x]"},
				// A primitive type's own elements, which the element FHIR
				// writes beside a primitive is ordered by.
				"date":    {"id", "extension", "value"},
				"boolean": {"id", "extension", "value"},
				// Roots, which R4 writes with no derivation.
				"Element":  {"id", "extension"},
				"Resource": {"id", "meta", "implicitRules", "language"},
			} {
				if got := tc.model.ChildElements(path); !slices.Equal(got, want) {
					t.Errorf("ChildElements(%q) = %v, want %v", path, got, want)
				}
			}

			// A backbone element is listed by its path.
			if got := tc.model.ChildElements("Observation.component"); len(got) < 5 ||
				!slices.Equal(got[:5], []string{"id", "extension", "modifierExtension", "code", "value[x]"}) {
				t.Errorf("ChildElements(Observation.component) = %v", got)
			}
			// A type that constrains another has its base's children.
			if got, want := tc.model.ChildElements("SimpleQuantity"), tc.model.ChildElements("Quantity"); !slices.Equal(got, want) || len(want) == 0 {
				t.Errorf("ChildElements(SimpleQuantity) = %v, want Quantity's %v", got, want)
			}
			if got := tc.model.ChildElements("Patient.notAField"); got != nil {
				t.Errorf("ChildElements on an unknown path = %v, want nil", got)
			}
		})
	}
}

// Only a definition of a type contributes its snapshot, so no constraint
// written under the same paths can repeat or reorder a resource's children;
// and each version lists its own, as Observation's differ between R4 and R5.
func TestModelListsEachVersionsOwnChildren(t *testing.T) {
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			observation := tc.model.ChildElements("Observation")
			want := map[string]struct {
				count int
				tenth string
			}{"r4": {32, "basedOn"}, "r4b": {32, "basedOn"}, "r5": {35, "instantiates[x]"}}[tc.name]
			if len(observation) != want.count {
				t.Fatalf("ChildElements(Observation) has %d children, want %d: %v", len(observation), want.count, observation)
			}
			if observation[9] != want.tenth {
				t.Errorf("ChildElements(Observation)'s tenth child is %q, want %q", observation[9], want.tenth)
			}
			for _, path := range []string{"Observation", "Patient", "Bundle", "Bundle.entry", "Questionnaire.item"} {
				names := tc.model.ChildElements(path)
				seen := map[string]bool{}
				for _, name := range names {
					if seen[name] {
						t.Errorf("ChildElements(%q) lists %q twice", path, name)
					}
					seen[name] = true
				}
			}
		})
	}
}

// A primitive type's own elements are typed: the element FHIR writes beside a
// primitive, _birthDate, has an id and extensions, read beneath date.
func TestModelTypesAPrimitivesElements(t *testing.T) {
	for _, tc := range models() {
		t.Run(tc.name, func(t *testing.T) {
			for path, want := range map[string]string{
				"date.extension":   "Extension",
				"date.id":          "http://hl7.org/fhirpath/System.String",
				"date.value":       "http://hl7.org/fhirpath/System.Date",
				"string.extension": "Extension",
			} {
				if got := tc.model.TypeOf(path); got != want {
					t.Errorf("TypeOf(%q) = %q, want %q", path, got, want)
				}
			}
		})
	}
}

package generator

import (
	"path/filepath"
	"sort"
	"strings"

	"github.com/gofhir/models/internal/codegen/analyzer"
)

// FHIRPathModelTemplateData holds all data needed by fhirpath_model.go.tmpl.
type FHIRPathModelTemplateData struct {
	TemplateData
	ChoiceTypePaths       []FHIRPathKVMulti
	Path2Type             []FHIRPathKV
	Path2RefType          []FHIRPathKVMulti
	Type2Parent           []FHIRPathKV
	PathsDefinedElsewhere []FHIRPathKV
	Resources             []string
}

// FHIRPathKV is a key-value pair used in map literals within generated code.
type FHIRPathKV struct {
	Key   string
	Value string
}

// FHIRPathKVMulti is a key-multi-value pair used in map[string][]string literals.
type FHIRPathKVMulti struct {
	Key    string
	Values []string
}

// generateFHIRPathModel generates fhirpath_model.go, which provides runtime
// metadata (type info, choice types, reference targets, type hierarchy, etc.)
// for a FHIRPath engine to evaluate expressions against this FHIR version.
func (c *CodeGen) generateFHIRPathModel() error {
	choiceTypeMap := make(map[string][]string)
	path2TypeMap := make(map[string]string)
	path2RefMap := make(map[string][]string)
	contentRefMap := make(map[string]string)
	resourceSet := make(map[string]bool)

	// collectProps iterates the properties of a type (or backbone) and populates
	// the intermediate maps. fhirName is the fully qualified FHIR path prefix,
	// e.g. "Patient" or "Patient.contact".
	collectProps := func(fhirName string, props []analyzer.AnalyzedProperty) {
		for _, prop := range props {
			// Skip JSON extension companion fields (e.g. _birthDate, _status).
			// These are not addressable FHIRPath elements.
			if strings.HasPrefix(prop.JSONName, "_") {
				continue
			}

			path := fhirName + "." + prop.JSONName

			// path2Type: store the FHIR type code as-is.
			// BackboneElement properties stay as "BackboneElement" — this matches
			// the FHIR spec and what FHIRPath engines expect.
			if prop.FHIRType != "ContentReference" && prop.FHIRType != "" {
				path2TypeMap[path] = prop.FHIRType
			}

			// choiceTypePaths: record the base choice path once per baseName
			// (all variants share the same ChoiceTypes and ChoiceBaseName).
			if prop.IsChoice && prop.ChoiceBaseName != "" {
				choiceKey := fhirName + "." + prop.ChoiceBaseName
				if _, exists := choiceTypeMap[choiceKey]; !exists {
					choiceTypeMap[choiceKey] = prop.ChoiceTypes
				}
			}

			// path2RefType: Reference and canonical elements with target constraints.
			if len(prop.TargetTypes) > 0 {
				path2RefMap[path] = prop.TargetTypes
			}

			// pathsDefinedElsewhere: contentReference aliases.
			if prop.ContentRef != "" {
				contentRefMap[path] = prop.ContentRef
			}
		}
	}

	for _, t := range c.types {
		if t.Kind == kindResource {
			resourceSet[t.Name] = true
		}

		collectProps(t.FHIRName, t.Properties)

		for _, bb := range t.BackboneTypes {
			collectProps(bb.FHIRName, bb.Properties)
		}
	}

	data := FHIRPathModelTemplateData{
		TemplateData: TemplateData{
			PackageName: c.config.PackageName,
			Version:     strings.ToUpper(c.config.Version),
			FileType:    "fhirpath_model",
		},
		ChoiceTypePaths:       sortedKVMulti(choiceTypeMap),
		Path2Type:             sortedKV(path2TypeMap),
		Path2RefType:          sortedKVMulti(path2RefMap),
		Type2Parent:           sortedKV(c.buildTypeHierarchy()),
		PathsDefinedElsewhere: sortedKV(contentRefMap),
		Resources:             sortedBoolMapKeys(resourceSet),
	}

	outputPath := filepath.Join(c.config.OutputDir, "fhirpath_model.go")
	return writeTemplateFile(outputPath, "fhirpath_model.go.tmpl", data)
}

// sortedKV converts a map[string]string to a []FHIRPathKV sorted by key.
func sortedKV(m map[string]string) []FHIRPathKV {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]FHIRPathKV, 0, len(keys))
	for _, k := range keys {
		result = append(result, FHIRPathKV{Key: k, Value: m[k]})
	}
	return result
}

// sortedKVMulti converts a map[string][]string to a []FHIRPathKVMulti sorted by key.
// The inner value slices retain their original order (choice types preserve declaration order).
func sortedKVMulti(m map[string][]string) []FHIRPathKVMulti {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	result := make([]FHIRPathKVMulti, 0, len(keys))
	for _, k := range keys {
		result = append(result, FHIRPathKVMulti{Key: k, Values: m[k]})
	}
	return result
}

// buildTypeHierarchy builds a map of FHIR type name → parent type name from
// all raw StructureDefinitions (including abstract types like DomainResource
// and Resource that are filtered out of the code generator).
func (c *CodeGen) buildTypeHierarchy() map[string]string {
	result := make(map[string]string)
	seen := make(map[string]bool)
	for _, sd := range c.rawSDs {
		if seen[sd.Name] {
			continue
		}
		seen[sd.Name] = true
		if sd.BaseDefinition == "" {
			continue
		}
		parts := strings.Split(sd.BaseDefinition, "/")
		parent := parts[len(parts)-1]
		if parent != "" && parent != sd.Name {
			result[sd.Name] = parent
		}
	}
	return result
}

// sortedBoolMapKeys returns sorted keys from a map[string]bool.
func sortedBoolMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

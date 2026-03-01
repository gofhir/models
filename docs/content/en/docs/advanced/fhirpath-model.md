---
title: "FHIRPath Model"
linkTitle: "FHIRPath Model"
description: "Runtime metadata for FHIRPath expression evaluation, including type information, choice type paths, reference targets, and type hierarchy data."
weight: 1
---

The `FHIRPathModelData` struct provides runtime metadata that a FHIRPath engine needs to correctly evaluate expressions against FHIR resources. Each version package (`r4`, `r4b`, `r5`) exposes its own singleton instance via the `FHIRPathModel()` function.

## The Singleton

Access the model metadata through the package-level function:

```go
import "github.com/gofhir/models/r4"

model := r4.FHIRPathModel()
```

The returned `*FHIRPathModelData` is initialized once at package load time and is safe for concurrent use. It contains all the type metadata extracted from the FHIR StructureDefinitions for that version.

## Internal Data Maps

The `FHIRPathModelData` struct holds six internal maps, each serving a specific purpose during FHIRPath evaluation:

| Map | Type | Purpose |
|-----|------|---------|
| `choiceTypePaths` | `map[string][]string` | Maps polymorphic element base paths to their permitted type codes |
| `path2Type` | `map[string]string` | Maps every element path to its FHIR type code |
| `path2RefType` | `map[string][]string` | Maps Reference element paths to allowed target resource types |
| `type2Parent` | `map[string]string` | Maps type names to their parent in the FHIR type hierarchy |
| `pathsDefinedElsewhere` | `map[string]string` | Resolves element paths that are defined on a different type |
| `resources` | `map[string]bool` | Set of all resource type names |

These maps are not exported directly. Instead, seven accessor methods provide a clean API for querying the metadata.

## Accessor Methods

### ChoiceTypes

Returns the permitted type codes for a polymorphic (choice) element. The path must use the base name without the type suffix.

```go
model := r4.FHIRPathModel()

// Observation.value is a choice type (value[x])
types := model.ChoiceTypes("Observation.value")
// Returns: ["Quantity", "CodeableConcept", "string", "boolean", "integer",
//           "Range", "Ratio", "SampledData", "time", "dateTime", "Period"]
```

### TypeOf

Returns the FHIR type code for a given element path. Primitive types use lowercase (`"string"`, `"boolean"`), complex types use PascalCase (`"HumanName"`, `"CodeableConcept"`).

```go
model := r4.FHIRPathModel()

t := model.TypeOf("Patient.name")
// Returns: "HumanName"

t = model.TypeOf("Patient.active")
// Returns: "boolean"
```

### ReferenceTargets

Returns the allowed target resource types for a Reference element path.

```go
model := r4.FHIRPathModel()

targets := model.ReferenceTargets("Observation.subject")
// Returns: ["Device", "Group", "Location", "Patient"]
```

### ParentType

Returns the immediate parent type name in the FHIR type hierarchy.

```go
model := r4.FHIRPathModel()

parent := model.ParentType("Patient")
// Returns: "DomainResource"

parent = model.ParentType("Age")
// Returns: "Quantity"
```

### IsSubtype

Reports whether `child` is the same as or a subtype of `parent` by walking the type hierarchy.

```go
model := r4.FHIRPathModel()

model.IsSubtype("Patient", "Resource")       // true
model.IsSubtype("Patient", "DomainResource") // true
model.IsSubtype("Age", "Quantity")           // true
model.IsSubtype("Quantity", "Age")           // false
```

### ResolvePath

Resolves element paths that are defined elsewhere (e.g., backbone elements shared across types). If the path is not defined elsewhere, it returns the input unchanged.

```go
model := r4.FHIRPathModel()

resolved := model.ResolvePath("Bundle.entry.resource")
// May return the canonical path if this element is defined on another type
```

### IsResource

Reports whether a given type name is a FHIR resource type.

```go
model := r4.FHIRPathModel()

model.IsResource("Patient")   // true
model.IsResource("HumanName") // false
model.IsResource("Bundle")    // true
```

## Integration with gofhir/fhirpath

The primary use case for `FHIRPathModelData` is to provide type information to a FHIRPath evaluation engine. The companion library [`gofhir/fhirpath`](https://github.com/gofhir/fhirpath) accepts the model via the `WithModel` option:

```go
import (
    "fmt"
    "github.com/gofhir/fhirpath"
    "github.com/gofhir/models/r4"
)

// Build a resource
patient := r4.NewPatient(
    r4.WithPatientId("example-1"),
    r4.WithPatientBirthDate("1990-01-15"),
)

// Evaluate a FHIRPath expression against it
result, err := fhirpath.Evaluate(patient, "Patient.birthDate",
    fhirpath.WithModel(r4.FHIRPathModel()))
if err != nil {
    panic(err)
}
fmt.Println(result) // ["1990-01-15"]
```

The model enables the FHIRPath engine to:

- **Resolve choice types**: When an expression references `Observation.value`, the engine uses `ChoiceTypes` to know which concrete fields (`valueQuantity`, `valueString`, etc.) to check.
- **Validate paths**: `TypeOf` confirms that a path like `Patient.name` exists and returns `HumanName`.
- **Navigate the type hierarchy**: `IsSubtype` and `ParentType` support the FHIRPath `is` and `as` operators.
- **Check reference targets**: `ReferenceTargets` validates `.resolve()` calls against allowed target types.

{{< callout type="info" >}}
The FHIRPath model is generated alongside all other code. When you regenerate models for a new FHIR version, the model metadata is automatically updated to reflect any changes in the specification.
{{< /callout >}}

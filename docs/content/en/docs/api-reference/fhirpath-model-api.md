---
title: "FHIRPath Model API"
linkTitle: "FHIRPath Model API"
description: "Complete API reference for the FHIRPathModelData struct and its accessor methods."
weight: 4
---

The `FHIRPathModelData` struct provides runtime metadata required by FHIRPath expression engines. This page documents every exported function and method with its full signature, description, and usage example.

## FHIRPathModel

Returns the package-level singleton containing all FHIRPath metadata for this FHIR version.

### Signature

```go
func FHIRPathModel() *FHIRPathModelData
```

### Returns

A pointer to the singleton `FHIRPathModelData` instance. This instance is initialized at package load time and is safe for concurrent read access.

### Example

```go
import "github.com/gofhir/models/r4"

model := r4.FHIRPathModel()
// model is ready to use -- no initialization needed
```

---

## ChoiceTypes

Returns the permitted FHIR type codes for a polymorphic (choice) element path. The path must be the base name without a type suffix.

### Signature

```go
func (m *FHIRPathModelData) ChoiceTypes(path string) []string
```

### Parameters

- `path` -- The base element path using dot notation, e.g., `"Observation.value"` (not `"Observation.valueQuantity"`).

### Returns

- A slice of type code strings representing the permitted types for this choice element.
- `nil` if the path is not a choice type or is unknown.

### Example

```go
model := r4.FHIRPathModel()

types := model.ChoiceTypes("Observation.value")
// ["Quantity", "CodeableConcept", "string", "boolean", "integer",
//  "Range", "Ratio", "SampledData", "time", "dateTime", "Period"]

types = model.ChoiceTypes("Patient.deceased")
// ["boolean", "dateTime"]

types = model.ChoiceTypes("Patient.name")
// nil -- not a choice type
```

---

## TypeOf

Returns the FHIR type code for the given element path. Primitive types use lowercase (`"string"`, `"boolean"`, `"dateTime"`), complex types use PascalCase (`"HumanName"`, `"CodeableConcept"`).

### Signature

```go
func (m *FHIRPathModelData) TypeOf(path string) string
```

### Parameters

- `path` -- The element path using dot notation, e.g., `"Patient.name"`, `"Observation.status"`.

### Returns

- The FHIR type code as a string.
- An empty string if the path is unknown.

### Example

```go
model := r4.FHIRPathModel()

model.TypeOf("Patient.name")              // "HumanName"
model.TypeOf("Patient.active")            // "boolean"
model.TypeOf("Observation.status")        // "code"
model.TypeOf("Observation.valueQuantity") // "Quantity"
model.TypeOf("Patient.contact")           // "BackboneElement"
model.TypeOf("Unknown.path")             // ""
```

---

## ReferenceTargets

Returns the allowed target resource type names for a Reference or canonical element at the given path.

### Signature

```go
func (m *FHIRPathModelData) ReferenceTargets(path string) []string
```

### Parameters

- `path` -- The element path of a Reference-typed field, e.g., `"Observation.subject"`.

### Returns

- A slice of resource type names that are valid targets for this reference.
- `nil` if the path is not a reference or has no constrained targets.

### Example

```go
model := r4.FHIRPathModel()

targets := model.ReferenceTargets("Observation.subject")
// ["Device", "Group", "Location", "Patient"]

targets = model.ReferenceTargets("Observation.performer")
// ["CareTeam", "Organization", "Patient", "Practitioner", "PractitionerRole", "RelatedPerson"]

targets = model.ReferenceTargets("Patient.name")
// nil -- not a Reference field
```

---

## ParentType

Returns the immediate parent type name in the FHIR type hierarchy.

### Signature

```go
func (m *FHIRPathModelData) ParentType(typeName string) string
```

### Parameters

- `typeName` -- A FHIR type name, e.g., `"Patient"`, `"Age"`, `"DomainResource"`.

### Returns

- The parent type name as a string.
- An empty string if the type has no parent (e.g., the root `Element` type) or is unknown.

### Example

```go
model := r4.FHIRPathModel()

model.ParentType("Patient")        // "DomainResource"
model.ParentType("DomainResource") // "Resource"
model.ParentType("Bundle")         // "Resource"
model.ParentType("Age")            // "Quantity"
model.ParentType("Quantity")       // "Element"
model.ParentType("Element")        // ""
```

---

## IsSubtype

Reports whether `child` is the same type as, or a subtype of, `parent` in the FHIR type hierarchy. It walks the `type2Parent` chain from `child` upward until it either finds `parent` or reaches the root.

### Signature

```go
func (m *FHIRPathModelData) IsSubtype(child, parent string) bool
```

### Parameters

- `child` -- The type name to check.
- `parent` -- The potential ancestor type name.

### Returns

- `true` if `child == parent` or if `child` is a descendant of `parent` in the type hierarchy.
- `false` otherwise.

### Example

```go
model := r4.FHIRPathModel()

model.IsSubtype("Patient", "Patient")        // true  (same type)
model.IsSubtype("Patient", "DomainResource") // true
model.IsSubtype("Patient", "Resource")       // true
model.IsSubtype("Age", "Quantity")           // true
model.IsSubtype("Quantity", "Age")           // false (parent, not child)
model.IsSubtype("Patient", "Observation")    // false (siblings)
```

---

## ResolvePath

Resolves an element path that may be defined elsewhere. Some FHIR element paths are defined on a shared backbone type and referenced from multiple resources. This method maps such paths to their canonical definition location.

### Signature

```go
func (m *FHIRPathModelData) ResolvePath(path string) string
```

### Parameters

- `path` -- The element path to resolve.

### Returns

- The canonical path if the element is defined elsewhere.
- The original `path` unchanged if it is not remapped.

### Example

```go
model := r4.FHIRPathModel()

// Most paths resolve to themselves
resolved := model.ResolvePath("Patient.name")
// "Patient.name"

// Some paths may resolve to a shared definition
resolved = model.ResolvePath("Bundle.entry.resource")
// Returns the canonical definition path
```

---

## IsResource

Reports whether the given type name is a FHIR resource type (as opposed to a data type or backbone element).

### Signature

```go
func (m *FHIRPathModelData) IsResource(typeName string) bool
```

### Parameters

- `typeName` -- The FHIR type name to check.

### Returns

- `true` if the type is a resource (e.g., `"Patient"`, `"Bundle"`, `"Observation"`).
- `false` if it is a data type (e.g., `"HumanName"`, `"CodeableConcept"`) or unknown.

### Example

```go
model := r4.FHIRPathModel()

model.IsResource("Patient")         // true
model.IsResource("Observation")     // true
model.IsResource("Bundle")          // true
model.IsResource("HumanName")       // false
model.IsResource("CodeableConcept") // false
model.IsResource("BackboneElement") // false
```

{{< callout type="info" >}}
All accessor methods perform simple map lookups and return zero values for unknown keys. They never return errors or panic. This makes them safe to call with user-provided input without validation.
{{< /callout >}}

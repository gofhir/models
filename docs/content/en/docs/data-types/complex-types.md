---
title: "Complex Types"
linkTitle: "Complex Types"
description: "Structured FHIR data types including HumanName, CodeableConcept, Reference, Quantity, and more."
weight: 2
---

FHIR complex types are structured data types composed of multiple elements. The `gofhir/models` library generates a Go struct for each complex type, with fields corresponding to every element defined in the FHIR specification. All complex types inherit from either `Element` (for general-purpose data types) or `BackboneElement` (for types used only within a specific resource context).

## Base Types

### Element

`Element` is the base type for all FHIR data types. It provides an optional `Id` for inter-element referencing and a list of extensions:

```go
type Element struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}
```

### BackboneElement

`BackboneElement` extends `Element` with `ModifierExtension` support. It is used for nested components that are defined inline within a resource (such as `Patient.contact` or `Observation.component`):

```go
type BackboneElement struct {
    Id                *string     `json:"id,omitempty"`
    Extension         []Extension `json:"extension,omitempty"`
    ModifierExtension []Extension `json:"modifierExtension,omitempty"`
}
```

## Key Complex Types

### HumanName

Represents a person's name with structured components:

```go
type HumanName struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
    Use       *NameUse    `json:"use,omitempty"`
    UseExt    *Element    `json:"_use,omitempty"`
    Text      *string     `json:"text,omitempty"`
    TextExt   *Element    `json:"_text,omitempty"`
    Family    *string     `json:"family,omitempty"`
    FamilyExt *Element    `json:"_family,omitempty"`
    Given     []string    `json:"given,omitempty"`
    GivenExt  []Element   `json:"_given,omitempty"`
    Prefix    []string    `json:"prefix,omitempty"`
    PrefixExt []Element   `json:"_prefix,omitempty"`
    Suffix    []string    `json:"suffix,omitempty"`
    SuffixExt []Element   `json:"_suffix,omitempty"`
    Period    *Period     `json:"period,omitempty"`
}
```

Usage example:

```go
name := r4.HumanName{
    Use:    ptrTo(r4.NameUseOfficial),
    Family: ptrTo("Smith"),
    Given:  []string{"John", "Michael"},
    Prefix: []string{"Dr."},
}

patient := &r4.Patient{
    ResourceType: "Patient",
    Name:         []r4.HumanName{name},
}
```

### CodeableConcept

Represents a concept that may be defined by one or more coding systems, along with a human-readable text:

```go
type CodeableConcept struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
    Coding    []Coding    `json:"coding,omitempty"`
    Text      *string     `json:"text,omitempty"`
    TextExt   *Element    `json:"_text,omitempty"`
}
```

A `CodeableConcept` typically contains one or more `Coding` entries that identify the concept in different terminology systems:

```go
condition := &r4.Condition{
    ResourceType: "Condition",
    Code: &r4.CodeableConcept{
        Coding: []r4.Coding{
            {
                System:  ptrTo("http://snomed.info/sct"),
                Code:    ptrTo("73211009"),
                Display: ptrTo("Diabetes mellitus"),
            },
            {
                System:  ptrTo("http://hl7.org/fhir/sid/icd-10-cm"),
                Code:    ptrTo("E11.9"),
                Display: ptrTo("Type 2 diabetes mellitus without complications"),
            },
        },
        Text: ptrTo("Type 2 Diabetes"),
    },
}
```

### Coding

Represents a single code from a terminology system:

```go
type Coding struct {
    Id           *string     `json:"id,omitempty"`
    Extension    []Extension `json:"extension,omitempty"`
    System       *string     `json:"system,omitempty"`
    SystemExt    *Element    `json:"_system,omitempty"`
    Version      *string     `json:"version,omitempty"`
    VersionExt   *Element    `json:"_version,omitempty"`
    Code         *string     `json:"code,omitempty"`
    CodeExt      *Element    `json:"_code,omitempty"`
    Display      *string     `json:"display,omitempty"`
    DisplayExt   *Element    `json:"_display,omitempty"`
    UserSelected *bool       `json:"userSelected,omitempty"`
    UserSelectedExt *Element `json:"_userSelected,omitempty"`
}
```

### Reference

Represents a reference to another FHIR resource. References can be literal (a URL), logical (an identifier), or display-only:

```go
type Reference struct {
    Id           *string     `json:"id,omitempty"`
    Extension    []Extension `json:"extension,omitempty"`
    Reference    *string     `json:"reference,omitempty"`
    ReferenceExt *Element    `json:"_reference,omitempty"`
    Type         *string     `json:"type,omitempty"`
    TypeExt      *Element    `json:"_type,omitempty"`
    Identifier   *Identifier `json:"identifier,omitempty"`
    Display      *string     `json:"display,omitempty"`
    DisplayExt   *Element    `json:"_display,omitempty"`
}
```

Usage examples:

```go
// Literal reference
ref := r4.Reference{
    Reference: ptrTo("Patient/123"),
}

// Reference with display text
ref := r4.Reference{
    Reference: ptrTo("Practitioner/dr-smith"),
    Display:   ptrTo("Dr. Jane Smith"),
}

// Logical reference via identifier
ref := r4.Reference{
    Type: ptrTo("Patient"),
    Identifier: &r4.Identifier{
        System: ptrTo("http://hospital.example.org/mrn"),
        Value:  ptrTo("MRN-12345"),
    },
}
```

### Identifier

Represents a business identifier for a resource:

```go
type Identifier struct {
    Id        *string          `json:"id,omitempty"`
    Extension []Extension      `json:"extension,omitempty"`
    Use       *IdentifierUse   `json:"use,omitempty"`
    UseExt    *Element         `json:"_use,omitempty"`
    Type      *CodeableConcept `json:"type,omitempty"`
    System    *string          `json:"system,omitempty"`
    SystemExt *Element         `json:"_system,omitempty"`
    Value     *string          `json:"value,omitempty"`
    ValueExt  *Element         `json:"_value,omitempty"`
    Period    *Period          `json:"period,omitempty"`
    Assigner  *Reference       `json:"assigner,omitempty"`
}
```

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    Identifier: []r4.Identifier{
        {
            Use:    ptrTo(r4.IdentifierUseOfficial),
            System: ptrTo("http://hospital.example.org/mrn"),
            Value:  ptrTo("MRN-67890"),
        },
    },
}
```

### Address

Represents a postal address:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    Address: []r4.Address{
        {
            Use:        ptrTo(r4.AddressUseHome),
            Type:       ptrTo(r4.AddressTypePhysical),
            Line:       []string{"123 Main Street", "Apt 4B"},
            City:       ptrTo("Springfield"),
            State:      ptrTo("IL"),
            PostalCode: ptrTo("62704"),
            Country:    ptrTo("US"),
        },
    },
}
```

### ContactPoint

Represents a phone number, email, or other contact mechanism:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    Telecom: []r4.ContactPoint{
        {
            System: ptrTo(r4.ContactPointSystemPhone),
            Value:  ptrTo("+1-555-0100"),
            Use:    ptrTo(r4.ContactPointUseHome),
        },
        {
            System: ptrTo(r4.ContactPointSystemEmail),
            Value:  ptrTo("john.smith@example.com"),
            Use:    ptrTo(r4.ContactPointUseWork),
        },
    },
}
```

### Period

Represents a time period with start and/or end boundaries:

```go
type Period struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
    Start     *string     `json:"start,omitempty"`
    StartExt  *Element    `json:"_start,omitempty"`
    End       *string     `json:"end,omitempty"`
    EndExt    *Element    `json:"_end,omitempty"`
}
```

### Quantity

Represents a measured amount with a unit, system, and code:

```go
type Quantity struct {
    Id            *string             `json:"id,omitempty"`
    Extension     []Extension         `json:"extension,omitempty"`
    Value         *Decimal            `json:"value,omitempty"`
    ValueExt      *Element            `json:"_value,omitempty"`
    Comparator    *QuantityComparator `json:"comparator,omitempty"`
    ComparatorExt *Element            `json:"_comparator,omitempty"`
    Unit          *string             `json:"unit,omitempty"`
    UnitExt       *Element            `json:"_unit,omitempty"`
    System        *string             `json:"system,omitempty"`
    SystemExt     *Element            `json:"_system,omitempty"`
    Code          *string             `json:"code,omitempty"`
    CodeExt       *Element            `json:"_code,omitempty"`
}
```

```go
weight := r4.Quantity{
    Value:  r4.MustDecimal("72.5"),
    Unit:   ptrTo("kg"),
    System: ptrTo("http://unitsofmeasure.org"),
    Code:   ptrTo("kg"),
}
```

### Narrative

Contains the human-readable summary of a resource as XHTML:

```go
type Narrative struct {
    Id        *string          `json:"id,omitempty"`
    Extension []Extension      `json:"extension,omitempty"`
    Status    *NarrativeStatus `json:"status,omitempty"`
    StatusExt *Element         `json:"_status,omitempty"`
    Div       *string          `json:"div,omitempty"`
    DivExt    *Element         `json:"_div,omitempty"`
}
```

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    Text: &r4.Narrative{
        Status: ptrTo(r4.NarrativeStatusGenerated),
        Div:    ptrTo(`<div xmlns="http://www.w3.org/1999/xhtml"><p>John Smith, Male</p></div>`),
    },
}
```

## Backbone Elements

Resources also contain inline backbone elements that are unique to that resource. These are generated as separate structs named after the resource and path. For example, `Patient` has `PatientContact` and `PatientCommunication`:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    Contact: []r4.PatientContact{
        {
            Relationship: []r4.CodeableConcept{
                {Text: ptrTo("Emergency Contact")},
            },
            Name: &r4.HumanName{
                Family: ptrTo("Smith"),
                Given:  []string{"Jane"},
            },
        },
    },
}
```

{{< callout type="info" >}}
All complex types implement `MarshalJSON`, `UnmarshalJSON`, `MarshalXML`, and `UnmarshalXML` for full serialization support. When composing resources, you do not need to handle serialization of individual data types -- the resource-level marshaling methods handle everything recursively.
{{< /callout >}}

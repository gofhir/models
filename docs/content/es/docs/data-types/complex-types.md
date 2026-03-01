---
title: "Tipos Complejos"
linkTitle: "Tipos Complejos"
description: "Tipos de datos FHIR estructurados incluyendo HumanName, CodeableConcept, Reference, Quantity y más."
weight: 2
---

Los tipos complejos de FHIR son tipos de datos estructurados compuestos por múltiples elementos. La biblioteca `gofhir/models` genera un struct de Go para cada tipo complejo, con campos correspondientes a cada elemento definido en la especificación FHIR. Todos los tipos complejos heredan de `Element` (para tipos de datos de propósito general) o `BackboneElement` (para tipos usados solo dentro de un contexto de recurso específico).

## Tipos Base

### Element

`Element` es el tipo base para todos los tipos de datos FHIR. Proporciona un `Id` opcional para referencia entre elementos y una lista de extensiones:

```go
type Element struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
}
```

### BackboneElement

`BackboneElement` extiende `Element` con soporte para `ModifierExtension`. Se usa para componentes anidados que se definen en línea dentro de un recurso (como `Patient.contact` u `Observation.component`):

```go
type BackboneElement struct {
    Id                *string     `json:"id,omitempty"`
    Extension         []Extension `json:"extension,omitempty"`
    ModifierExtension []Extension `json:"modifierExtension,omitempty"`
}
```

## Tipos Complejos Clave

### HumanName

Representa el nombre de una persona con componentes estructurados:

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

Ejemplo de uso:

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

Representa un concepto que puede estar definido por uno o más sistemas de codificación, junto con un texto legible por humanos:

```go
type CodeableConcept struct {
    Id        *string     `json:"id,omitempty"`
    Extension []Extension `json:"extension,omitempty"`
    Coding    []Coding    `json:"coding,omitempty"`
    Text      *string     `json:"text,omitempty"`
    TextExt   *Element    `json:"_text,omitempty"`
}
```

Un `CodeableConcept` típicamente contiene una o más entradas `Coding` que identifican el concepto en diferentes sistemas de terminología:

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

Representa un único código de un sistema de terminología:

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

Representa una referencia a otro recurso FHIR. Las referencias pueden ser literales (una URL), lógicas (un identificador) o solo de visualización:

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

Ejemplos de uso:

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

Representa un identificador de negocio para un recurso:

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

Representa una dirección postal:

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

Representa un número de teléfono, correo electrónico u otro mecanismo de contacto:

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

Representa un período de tiempo con límites de inicio y/o fin:

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

Representa una cantidad medida con una unidad, sistema y código:

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

Contiene el resumen legible por humanos de un recurso como XHTML:

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

## Elementos Backbone

Los recursos también contienen elementos backbone en línea que son únicos de ese recurso. Estos se generan como structs separados nombrados según el recurso y la ruta. Por ejemplo, `Patient` tiene `PatientContact` y `PatientCommunication`:

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
Todos los tipos complejos implementan `MarshalJSON`, `UnmarshalJSON`, `MarshalXML` y `UnmarshalXML` para soporte completo de serialización. Al componer recursos, no necesitas manejar la serialización de tipos de datos individuales -- los métodos de marshaling a nivel de recurso manejan todo de forma recursiva.
{{< /callout >}}

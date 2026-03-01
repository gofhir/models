---
title: "API del Builder"
linkTitle: "API del Builder"
description: "Patron builder fluido y API de opciones funcionales para construir recursos FHIR."
weight: 3
---

Cada tipo de recurso en `gofhir/models` proporciona dos APIs de construccion complementarias: un **builder fluido** y **opciones funcionales**. Ambos se generan automaticamente para todos los recursos en R4, R4B y R5.

## Patron Builder Fluido

El patron builder proporciona una API encadenable para construir recursos paso a paso.

### Estructura

Para cada tipo de recurso `<Resource>`, la biblioteca genera:

| Exportacion | Tipo | Descripcion |
|-------------|------|-------------|
| `<Resource>Builder` | struct | La struct del builder que contiene el recurso en construccion |
| `New<Resource>Builder()` | `*<Resource>Builder` | Constructor que crea un nuevo builder con un recurso de valor cero |
| `Set<Field>(v T)` | `*<Resource>Builder` | Establece un campo singular (puntero o escalar) |
| `Add<Field>(v T)` | `*<Resource>Builder` | Agrega a un campo repetido (slice) |
| `Build()` | `*<Resource>` | Devuelve el recurso construido |

### Convencion de Nomenclatura

- **`Set`** se usa para campos singulares -- campos con una cardinalidad maxima de 1 (por ejemplo, `Id`, `Gender`, `BirthDate`, `Status`).
- **`Add`** se usa para campos repetidos -- campos con una cardinalidad maxima mayor a 1 (por ejemplo, `Name`, `Identifier`, `Telecom`, `Extension`).

### Ejemplo con Patient

```go
import "github.com/gofhir/models/r4"

patient := r4.NewPatientBuilder().
    SetId("patient-123").
    SetActive(true).
    SetGender(r4.AdministrativeGenderMale).
    SetBirthDate("1990-05-15").
    AddName(r4.HumanName{
        Family: ptrTo("Doe"),
        Given:  []string{"John", "Michael"},
    }).
    AddIdentifier(r4.Identifier{
        System: ptrTo("http://hospital.example.org/mrn"),
        Value:  ptrTo("MRN-12345"),
    }).
    AddTelecom(r4.ContactPoint{
        System: ptrTo(r4.ContactPointSystemPhone),
        Value:  ptrTo("+1-555-0100"),
        Use:    ptrTo(r4.ContactPointUseHome),
    }).
    AddAddress(r4.Address{
        Use:        ptrTo(r4.AddressUseHome),
        Line:       []string{"123 Main St"},
        City:       ptrTo("Springfield"),
        State:      ptrTo("IL"),
        PostalCode: ptrTo("62701"),
        Country:    ptrTo("US"),
    }).
    Build()

// Funcion auxiliar para crear punteros a strings
func ptrTo[T any](v T) *T {
    return &v
}
```

### Ejemplo con Observation

```go
import "github.com/gofhir/models/r4"

status := r4.ObservationStatusFinal
observation := r4.NewObservationBuilder().
    SetId("obs-001").
    SetStatus(status).
    SetCode(r4.CodeableConcept{
        Coding: []r4.Coding{{
            System:  ptrTo("http://loinc.org"),
            Code:    ptrTo("29463-7"),
            Display: ptrTo("Body weight"),
        }},
        Text: ptrTo("Body Weight"),
    }).
    SetSubject(r4.Reference{
        Reference: ptrTo("Patient/patient-123"),
    }).
    SetEffectiveDateTime("2024-01-15T10:30:00Z").
    SetValueQuantity(r4.Quantity{
        Value:  r4.NewDecimalFromFloat64(72.5),
        Unit:   ptrTo("kg"),
        System: ptrTo("http://unitsofmeasure.org"),
        Code:   ptrTo("kg"),
    }).
    AddCategory(r4.CodeableConcept{
        Coding: []r4.Coding{{
            System:  ptrTo("http://terminology.hl7.org/CodeSystem/observation-category"),
            Code:    ptrTo("vital-signs"),
            Display: ptrTo("Vital Signs"),
        }},
    }).
    Build()
```

## Patron de Opciones Funcionales

El patron de opciones funcionales proporciona una sintaxis mas concisa para crear recursos en una sola llamada a funcion.

### Estructura

Para cada tipo de recurso `<Resource>`, la biblioteca genera:

| Exportacion | Tipo | Descripcion |
|-------------|------|-------------|
| `<Resource>Option` | `func(*<Resource>)` | El tipo de funcion de opcion |
| `New<Resource>(opts ...Option)` | `*<Resource>` | Constructor que aplica todas las opciones |
| `With<Resource><Field>(v T)` | `<Resource>Option` | Funcion de opcion para cada campo |

### Ejemplo con Patient

```go
import "github.com/gofhir/models/r4"

patient := r4.NewPatient(
    r4.WithPatientId("patient-456"),
    r4.WithPatientActive(true),
    r4.WithPatientGender(r4.AdministrativeGenderFemale),
    r4.WithPatientBirthDate("1985-11-20"),
    r4.WithPatientName(r4.HumanName{
        Family: ptrTo("Smith"),
        Given:  []string{"Jane"},
    }),
    r4.WithPatientIdentifier(r4.Identifier{
        System: ptrTo("http://hospital.example.org/mrn"),
        Value:  ptrTo("MRN-67890"),
    }),
)
```

### Ejemplo con Observation

```go
import "github.com/gofhir/models/r4"

observation := r4.NewObservation(
    r4.WithObservationId("obs-002"),
    r4.WithObservationStatus(r4.ObservationStatusFinal),
    r4.WithObservationCode(r4.CodeableConcept{
        Coding: []r4.Coding{{
            System:  ptrTo("http://loinc.org"),
            Code:    ptrTo("8310-5"),
            Display: ptrTo("Body temperature"),
        }},
    }),
    r4.WithObservationValueQuantity(r4.Quantity{
        Value:  r4.NewDecimalFromFloat64(37.2),
        Unit:   ptrTo("Cel"),
        System: ptrTo("http://unitsofmeasure.org"),
        Code:   ptrTo("Cel"),
    }),
)
```

## Elegir entre Patrones

| Criterio | Builder Fluido | Opciones Funcionales |
|----------|:--------------:|:--------------------:|
| Encadenamiento de metodos | Si | No |
| Expresion unica | Cadena multi-linea | Llamada a funcion unica |
| Construccion incremental | Ajuste natural | Menos natural |
| Campos condicionales | Agregar despues de crear el builder | Componer slices de opciones |
| Testing/mocking | El builder se puede inyectar | Las opciones se pueden recopilar |

### Cuando Usar el Builder

El builder es adecuado cuando necesitas construir un recurso de forma incremental, especialmente cuando algunos campos dependen de condiciones:

```go
builder := r4.NewPatientBuilder().
    SetId(id).
    SetActive(true)

if hasName {
    builder.AddName(name)
}
if hasAddress {
    builder.AddAddress(address)
}

patient := builder.Build()
```

### Cuando Usar Opciones Funcionales

Las opciones funcionales funcionan bien cuando se construye un recurso en una sola declaracion declarativa, o cuando las opciones se componen desde diferentes fuentes:

```go
opts := []r4.PatientOption{
    r4.WithPatientId(id),
    r4.WithPatientActive(true),
}

if hasName {
    opts = append(opts, r4.WithPatientName(name))
}

patient := r4.NewPatient(opts...)
```

{{< callout type="info" >}}
Tanto el builder como las opciones funcionales producen resultados identicos. Ambos establecen campos en la misma struct subyacente. Elige el estilo que mejor se adapte a los requisitos de legibilidad de tu codigo.
{{< /callout >}}

## Literales de Struct

Siempre puedes construir recursos directamente usando literales de struct de Go. Esto proporciona el mayor control y es el patron Go mas familiar:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    Id:           ptrTo("patient-789"),
    Active:       ptrTo(true),
    Gender:       ptrTo(r4.AdministrativeGenderMale),
    Name: []r4.HumanName{{
        Family: ptrTo("Johnson"),
        Given:  []string{"Robert"},
    }},
}
```

Ten en cuenta que al usar literales de struct, debes establecer `ResourceType` tu mismo. El builder y las opciones funcionales establecen `ResourceType` automaticamente durante el marshaling JSON.

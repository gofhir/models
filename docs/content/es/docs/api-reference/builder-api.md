---
title: "API del Builder"
linkTitle: "API del Builder"
description: "Patrón builder fluido para construir recursos FHIR."
weight: 3
---

Cada tipo de recurso en `gofhir/models` proporciona un **builder fluido**, generado automáticamente para todos los recursos en R4, R4B y R5.

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
import "github.com/gofhir/models/r4/v2"

patient := r4.NewPatientBuilder().
    SetId("patient-123").
    SetActive(true).
    SetGender(r4.AdministrativeGenderMale).
    SetBirthDate("1990-05-15").
    AddName(r4.HumanName{
        Family: ptrTo("Doe"),
        Given:  r4.PtrSlice("John", "Michael"),
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
        Line:       r4.PtrSlice("123 Main St"),
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
import "github.com/gofhir/models/r4/v2"

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

## Elegir entre Patrones

| Criterio | Builder Fluido | Literal de Struct |
|----------|:--------------:|:-----------------:|
| Encadenamiento de métodos | Sí | No |
| Construcción incremental | Encaja de forma natural | No es posible |
| Campos condicionales | Añadir tras crear el builder | Requiere un if antes del literal |
| Envoltura de punteros | La hacen `Set*`/`Add*` | Usa `Ptr` |
| Todo conocido de antemano | Funciona | Es lo más corto |

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

### Cuándo Usar Literales de Struct

Para un recurso que ya tienes completo, un literal de struct es más corto que cualquiera de los dos y no necesita builder:

```go
patient := &r4.Patient{
    Id:     r4.Ptr(id),
    Active: r4.Ptr(true),
    Name:   []r4.HumanName{name},
}
```

Ambos estilos producen la misma salida —el builder solo asigna campos en el mismo struct—, así que elige el que se lea mejor donde estés.

{{< callout type="info" >}}
**Las opciones funcionales eran el tercer estilo, y desaparecen en la v2.** Cada `With<Recurso><Campo>` tenía un método de builder con firma y comportamiento idénticos, así que la v1.6.0 marcó las 11.952 como `Deprecated` nombrando su reemplazo, y la v2 las eliminó: `Set*` para los campos simples, `Add*` para los repetibles.

Si vienes de la v1, `staticcheck` sobre tu build de v1.6.0 o posterior te lista cada llamada con su reemplazo. Consulta la [guía de migración de v1 a v2](../../migration/v1-to-v2/).
{{< /callout >}}

## Literales de Struct

Siempre puedes construir recursos directamente usando literales de struct de Go. Esto proporciona el mayor control y es el patron Go mas familiar:

```go
patient := &r4.Patient{
    Id:           ptrTo("patient-789"),
    Active:       ptrTo(true),
    Gender:       ptrTo(r4.AdministrativeGenderMale),
    Name: []r4.HumanName{{
        Family: ptrTo("Johnson"),
        Given:  r4.PtrSlice("Robert"),
    }},
}
```

`resourceType` no requiere atención en ninguno de los dos estilos: es un marcador de tamaño cero en el struct, así que un simple `r4.Patient{}` ya se serializa con el valor correcto. Léelo con `GetResourceType()`.

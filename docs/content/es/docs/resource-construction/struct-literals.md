---
title: "Literales de Struct"
linkTitle: "Literales de Struct"
description: "Crea recursos FHIR usando la inicialización directa de structs Go con control total sobre cada campo."
weight: 1
---

Los literales de struct te dan el control más directo sobre la creación de recursos. Inicializas los campos del struct Go explícitamente, lo cual es familiar para cualquier desarrollador Go y proporciona verificación completa de tipos en tiempo de compilación.

## Ejemplo Básico

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4"
)

func main() {
    patient := r4.Patient{
        ResourceType: "Patient",
        Id:           ptrTo("123"),
        Active:       ptrTo(true),
        Name: []r4.HumanName{
            {Family: ptrTo("Smith"), Given: []string{"John"}},
        },
    }

    data, _ := r4.Marshal(&patient)
    fmt.Println(string(data))
}
```

{{< callout type="info" >}}
El campo `ResourceType` se inyecta automáticamente durante el marshaling JSON. Puedes establecerlo explícitamente para mayor claridad, pero no es obligatorio.
{{< /callout >}}

## Manejo de Punteros

En FHIR, la mayoría de los campos primitivos son opcionales. Los structs Go representan esto con tipos puntero (`*string`, `*bool`, `*int`). No puedes tomar la dirección de un valor literal directamente en Go, por lo que necesitas un patrón auxiliar.

### El Auxiliar `ptrTo`

El enfoque más común es una función auxiliar genérica:

```go
func ptrTo[T any](v T) *T {
    return &v
}
```

Esto funciona con cualquier tipo:

```go
patient := r4.Patient{
    Id:        ptrTo("patient-1"),
    Active:    ptrTo(true),
    BirthDate: ptrTo("1990-05-15"),
    Gender:    ptrTo(r4.AdministrativeGenderMale),
}
```

### Variables Locales

Alternativamente, puedes usar variables locales y tomar su dirección:

```go
id := "patient-1"
active := true
gender := r4.AdministrativeGenderFemale

patient := r4.Patient{
    Id:     &id,
    Active: &active,
    Gender: &gender,
}
```

Este enfoque es útil cuando necesitas reutilizar valores o cuando el tipo del campo es un tipo personalizado de sistema de códigos.

## Recursos Complejos

Los literales de struct funcionan bien para construir recursos con tipos de datos anidados:

```go
system := "http://hospital.example.org/mrn"
value := "MRN-12345"
use := r4.NameUseOfficial
family := "Johnson"
homeUse := r4.AddressUseHome

patient := r4.Patient{
    Id:     ptrTo("patient-complex"),
    Active: ptrTo(true),
    Identifier: []r4.Identifier{
        {
            System: &system,
            Value:  &value,
        },
    },
    Name: []r4.HumanName{
        {
            Use:    &use,
            Family: &family,
            Given:  []string{"Robert", "James"},
        },
    },
    Gender:    ptrTo(r4.AdministrativeGenderMale),
    BirthDate: ptrTo("1978-11-03"),
    Address: []r4.Address{
        {
            Use:        &homeUse,
            Line:       []string{"123 Main St", "Apt 4B"},
            City:       ptrTo("Springfield"),
            State:      ptrTo("IL"),
            PostalCode: ptrTo("62704"),
            Country:    ptrTo("US"),
        },
    },
    Telecom: []r4.ContactPoint{
        {
            System: ptrTo(r4.ContactPointSystemPhone),
            Value:  ptrTo("555-0123"),
            Use:    ptrTo(r4.ContactPointUseHome),
        },
        {
            System: ptrTo(r4.ContactPointSystemEmail),
            Value:  ptrTo("robert.johnson@example.com"),
        },
    },
}
```

## Observations con Literales de Struct

```go
codeSystem := "http://loinc.org"
codeCode := "8480-6"
codeDisplay := "Systolic blood pressure"
unitSystem := "http://unitsofmeasure.org"

obs := r4.Observation{
    Status: ptrTo(r4.ObservationStatusFinal),
    Code: r4.CodeableConcept{
        Coding: []r4.Coding{
            {
                System:  &codeSystem,
                Code:    &codeCode,
                Display: &codeDisplay,
            },
        },
    },
    ValueQuantity: &r4.Quantity{
        Value:  r4.NewDecimalFromFloat64(120.0),
        Unit:   ptrTo("mmHg"),
        System: &unitSystem,
        Code:   ptrTo("mm[Hg]"),
    },
    EffectiveDateTime: ptrTo("2024-01-15T10:30:00Z"),
}
```

## Cuándo Usar Literales de Struct

Los literales de struct son una buena elección cuando:

- Quieres visibilidad completa de cada campo que se establece.
- Estás construyendo un recurso a partir de un conjunto conocido y fijo de valores (como fixtures de prueba o datos semilla).
- Prefieres los modismos explícitos de Go sobre las abstracciones de builders.

Para situaciones donde el código repetitivo de punteros es engorroso, considera el [Patrón Builder](../builder-pattern) o las [Opciones Funcionales](../functional-options) en su lugar.

---
title: "Inicio Rápido"
linkTitle: "Inicio Rápido"
description: "Crea, serializa y deserializa recursos FHIR con ejemplos de código Go funcionales."
weight: 2
---

Esta guía demuestra las operaciones más comunes: crear un recurso, serializarlo a JSON y deserializarlo de vuelta. Todos los ejemplos usan el paquete R4, pero la API es idéntica para R4B y R5.

## 1. Crear un Patient con un Literal de Struct

La forma más directa de crear un recurso FHIR es inicializando los campos del struct. Los campos FHIR opcionales se representan como punteros Go, por lo que necesitas una función auxiliar o el operador de dirección para establecerlos.

```go
package main

import (
    "fmt"

    "github.com/gofhir/models/r4"
)

// ptrTo is a generic helper that returns a pointer to the given value.
func ptrTo[T any](v T) *T {
    return &v
}

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

Salida:

```json
{"resourceType":"Patient","id":"123","active":true,"name":[{"family":"Smith","given":["John"]}]}
```

## 2. Crear un Patient con Opciones Funcionales

Las opciones funcionales eliminan el código repetitivo de punteros. Cada campo tiene una función `With<Resource><Field>()` correspondiente que establece el valor y maneja el envolvimiento del puntero internamente.

```go
patient := r4.NewPatient(
    r4.WithPatientId("patient-123"),
    r4.WithPatientActive(true),
    r4.WithPatientGender(r4.AdministrativeGenderMale),
    r4.WithPatientBirthDate("1990-01-15"),
)
```

También puedes agregar campos anidados complejos como nombres e identificadores:

```go
use := r4.NameUseOfficial
family := "Smith"

patient := r4.NewPatient(
    r4.WithPatientId("patient-456"),
    r4.WithPatientName(r4.HumanName{
        Use:    &use,
        Family: &family,
        Given:  []string{"John"},
    }),
    r4.WithPatientIdentifier(r4.Identifier{
        System: ptrTo("http://hospital.example.org/mrn"),
        Value:  ptrTo("12345"),
    }),
)
```

## 3. Crear un Patient con el Builder Fluido

El patrón builder proporciona una API encadenable. Comienza con `New<Resource>Builder()`, establece campos con métodos `Set` y `Add`, y finaliza con `.Build()`.

```go
family := "Garcia"
use := r4.NameUseOfficial

patient := r4.NewPatientBuilder().
    SetId("patient-789").
    SetActive(true).
    SetGender(r4.AdministrativeGenderFemale).
    SetBirthDate("1985-06-20").
    AddName(r4.HumanName{
        Use:    &use,
        Family: &family,
        Given:  []string{"Maria"},
    }).
    Build()
```

## 4. Serializar a JSON

Usa `r4.Marshal` en lugar de `json.Marshal`. La función de la biblioteca deshabilita el escapado HTML para que el XHTML narrativo de FHIR en los campos `text.div` se preserve correctamente. El campo `resourceType` siempre se inyecta automáticamente.

```go
data, err := r4.Marshal(patient)
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

Para una salida con formato legible, usa `MarshalIndent`:

```go
data, err := r4.MarshalIndent(patient, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

Salida:

```json
{
  "resourceType": "Patient",
  "id": "patient-789",
  "active": true,
  "name": [
    {
      "use": "official",
      "family": "Garcia",
      "given": ["Maria"]
    }
  ],
  "gender": "female",
  "birthDate": "1985-06-20"
}
```

## 5. Deserializar desde JSON

### Tipo de Recurso Conocido

Si conoces el tipo de recurso en tiempo de compilación, deserializa directamente en el struct:

```go
var patient r4.Patient
err := json.Unmarshal(data, &patient)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Patient ID: %s\n", *patient.Id)
```

### Tipo de Recurso Desconocido

Cuando el tipo de recurso no se conoce de antemano (por ejemplo, al leer una respuesta de un servidor FHIR), usa `UnmarshalResource`. Esta función inspecciona el campo `resourceType` y devuelve el struct Go correcto detrás de la interfaz `Resource`:

```go
resource, err := r4.UnmarshalResource(data)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Type: %s\n", resource.GetResourceType())

// Type-assert to access resource-specific fields
if patient, ok := resource.(*r4.Patient); ok {
    fmt.Printf("Patient ID: %s\n", *patient.Id)
}
```

También puedes consultar el tipo de recurso sin una deserialización completa:

```go
resourceType, err := r4.GetResourceType(data)
if err != nil {
    log.Fatal(err)
}
fmt.Println(resourceType) // "Patient"
```

## 6. Crear una Observation

Aquí hay un ejemplo más completo que muestra una Observation con un código, valor y categoría:

```go
codeSystem := "http://loinc.org"
codeCode := "8867-4"
codeDisplay := "Heart rate"

obs := r4.NewObservation(
    r4.WithObservationId("obs-123"),
    r4.WithObservationStatus(r4.ObservationStatusFinal),
    r4.WithObservationCode(r4.CodeableConcept{
        Coding: []r4.Coding{
            {System: &codeSystem, Code: &codeCode, Display: &codeDisplay},
        },
    }),
    r4.WithObservationEffectiveDateTime("2024-01-15T10:30:00Z"),
)

data, _ := r4.Marshal(obs)
fmt.Println(string(data))
```

## Siguientes Pasos

- Aprende sobre las [Versiones FHIR](../fhir-versions) y cómo difieren los tres paquetes.
- Explora los patrones de [Construcción de Recursos](../../resource-construction) en profundidad.

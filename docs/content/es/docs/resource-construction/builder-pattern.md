---
title: "Patrón Builder"
linkTitle: "Patrón Builder"
description: "Usa la API de builder fluido para construir recursos FHIR con llamadas de métodos encadenables."
weight: 2
---

El patrón builder proporciona una API fluida y encadenable para construir recursos FHIR. Cada tipo de recurso tiene un builder correspondiente con métodos `Set` para campos singulares y métodos `Add` para campos repetitivos. El builder maneja el envolvimiento de punteros internamente, eliminando el código repetitivo requerido por los literales de struct.

## Cómo Funciona

Cada tipo de recurso en el paquete tiene un builder:

1. Crea un builder con `New<Resource>Builder()`.
2. Encadena llamadas `Set<Field>()` para campos singulares.
3. Encadena llamadas `Add<Field>()` para campos repetitivos (slice).
4. Llama a `.Build()` para obtener el struct del recurso final.

El builder devuelve `*<Resource>Builder` desde cada setter, por lo que las llamadas pueden encadenarse.

## Ejemplo Básico

```go
patient := r4.NewPatientBuilder().
    SetId("patient-789").
    SetActive(true).
    SetGender(r4.AdministrativeGenderFemale).
    SetBirthDate("1985-06-20").
    Build()

data, _ := r4.Marshal(patient)
fmt.Println(string(data))
```

Salida:

```json
{"resourceType":"Patient","id":"patient-789","active":true,"gender":"female","birthDate":"1985-06-20"}
```

## Agregar Campos Complejos

Para campos que contienen structs de tipos de datos (como `HumanName`, `Identifier` o `Address`), usa los métodos `Add`. Estos agregan al slice subyacente:

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

Ten en cuenta que los structs de tipos de datos pasados a los métodos `Add` aún usan punteros para campos opcionales. El builder elimina el código repetitivo de punteros para campos primitivos del recurso (string, bool, tipos code), pero los structs de tipos de datos complejos conservan su representación estándar de Go.

## Agregar Múltiples Elementos

Llama a los métodos `Add` múltiples veces para agregar a campos repetitivos:

```go
system := "http://hospital.example.org/mrn"
value1 := "MRN-001"
value2 := "MRN-002"

patient := r4.NewPatientBuilder().
    SetId("patient-multi").
    AddIdentifier(r4.Identifier{System: &system, Value: &value1}).
    AddIdentifier(r4.Identifier{System: &system, Value: &value2}).
    Build()

// patient.Identifier has 2 elements
```

## Construir una Observation

El patrón builder funciona para todos los tipos de recursos. Aquí hay una Observation con una medición de signos vitales:

```go
codeSystem := "http://loinc.org"
codeCode := "8480-6"
codeDisplay := "Systolic blood pressure"
value := r4.NewDecimalFromFloat64(120.0)
unit := "mmHg"
unitSystem := "http://unitsofmeasure.org"
unitCode := "mm[Hg]"

obs := r4.NewObservationBuilder().
    SetId("obs-bp-001").
    SetStatus(r4.ObservationStatusFinal).
    SetCode(r4.CodeableConcept{
        Coding: []r4.Coding{
            {System: &codeSystem, Code: &codeCode, Display: &codeDisplay},
        },
    }).
    SetValueQuantity(r4.Quantity{
        Value:  value,
        Unit:   &unit,
        System: &unitSystem,
        Code:   &unitCode,
    }).
    SetEffectiveDateTime("2024-01-15T10:30:00Z").
    Build()
```

## Viaje de Ida y Vuelta JSON

Los recursos construidos con el builder se serializan y deserializan exactamente igual que los literales de struct:

```go
family := "Johnson"
city := "Boston"
use := r4.AddressUseHome

original := r4.NewPatientBuilder().
    SetId("pt-json").
    SetActive(true).
    SetGender(r4.AdministrativeGenderMale).
    AddName(r4.HumanName{Family: &family, Given: []string{"Robert"}}).
    AddAddress(r4.Address{Use: &use, City: &city}).
    Build()

// Marshal
data, err := r4.Marshal(original)
if err != nil {
    log.Fatal(err)
}

// Unmarshal
var decoded r4.Patient
err = json.Unmarshal(data, &decoded)
if err != nil {
    log.Fatal(err)
}

fmt.Println(*decoded.Id)          // "pt-json"
fmt.Println(*decoded.Name[0].Family) // "Johnson"
```

## Builder Vacío

Llamar a `Build()` sin establecer ningún campo devuelve un recurso válido y vacío:

```go
patient := r4.NewPatientBuilder().Build()
// patient.Id is nil, patient.Active is nil, patient.Name is empty
```

Esto es útil como punto de partida cuando necesitas poblar campos condicionalmente.

## Métodos Disponibles

Cada builder sigue la misma convención de nombres:

| Patrón de Método | Propósito | Ejemplo |
|----------------|---------|---------|
| `Set<Field>(v)` | Establecer un campo singular | `SetId("123")`, `SetActive(true)` |
| `Add<Field>(v)` | Agregar a un campo repetitivo | `AddName(humanName)`, `AddIdentifier(id)` |
| `Build()` | Devolver el recurso construido | `Build()` |

Los métodos `Set` aceptan valores sin envolver (por ejemplo, `string` en lugar de `*string`) y manejan la creación de punteros internamente. Los métodos `Add` aceptan el struct del tipo de dato directamente y lo agregan al slice correspondiente.

## Cuándo Usar el Builder

El patrón builder es ideal cuando:

- Estás construyendo recursos paso a paso, posiblemente a través de múltiples llamadas de función.
- Quieres una cadena fluida y legible de asignaciones de campos.
- Quieres evitar el código repetitivo de punteros para campos primitivos.

Para una inicialización de una sola vez con control total, considera los [Literales de Struct](../struct-literals). Para configuración componible, consulta las [Opciones Funcionales](../functional-options).

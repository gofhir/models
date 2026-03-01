---
title: "Opciones Funcionales"
linkTitle: "Opciones Funcionales"
description: "Crea recursos FHIR usando funciones de opciones funcionales componibles para una construcción limpia y configurable."
weight: 3
---

El patrón de opciones funcionales usa funciones Go para configurar campos de recursos. Cada tipo de recurso tiene un constructor `New<Resource>(opts...)` y un conjunto de funciones `With<Resource><Field>()`. Este patrón, popularizado por Dave Cheney y Rob Pike, produce sitios de llamada limpios y hace que las opciones sean componibles.

## Cómo Funciona

Cada tipo de recurso proporciona:

1. Un tipo de opción: `<Resource>Option` (una función que modifica el recurso).
2. Un constructor: `New<Resource>(opts ...<Resource>Option)` que crea un recurso y aplica las opciones.
3. Funciones de opción: `With<Resource><Field>(value)` para cada campo del recurso.

```go
// PatientOption is a functional option for configuring a Patient.
type PatientOption func(*Patient)

// NewPatient creates a new Patient with the given options.
func NewPatient(opts ...PatientOption) *Patient {
    r := &Patient{}
    for _, opt := range opts {
        opt(r)
    }
    return r
}
```

## Ejemplo Básico

```go
patient := r4.NewPatient(
    r4.WithPatientId("patient-123"),
    r4.WithPatientActive(true),
    r4.WithPatientGender(r4.AdministrativeGenderMale),
    r4.WithPatientBirthDate("1990-01-15"),
)

fmt.Println(*patient.Id)     // "patient-123"
fmt.Println(*patient.Active) // true
fmt.Println(*patient.Gender) // "male"
```

Las funciones `With` manejan el envolvimiento de punteros internamente, por lo que pasas valores simples (`string`, `bool`, etc.) y la función crea el puntero por ti.

## Agregar Nombres e Identificadores

Para tipos de datos complejos como `HumanName` e `Identifier`, las funciones de opción aceptan el struct directamente. Cada llamada agrega al slice:

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
    r4.WithPatientName(r4.HumanName{
        Family: &family,
        Given:  []string{"Johnny"},
    }),
)

// patient.Name has 2 elements
```

Agregar identificadores funciona de la misma manera:

```go
system := "http://hospital.example.org/mrn"
value := "12345"

patient := r4.NewPatient(
    r4.WithPatientIdentifier(r4.Identifier{
        System: &system,
        Value:  &value,
    }),
)
```

## Construir una Observation

Las opciones funcionales están disponibles para cada tipo de recurso. Aquí hay una Observation con un código LOINC:

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
    r4.WithObservationSubject(r4.Reference{
        Reference: ptrTo("Patient/patient-123"),
    }),
)
```

## Componer Opciones

Como las opciones son funciones simples, puedes almacenarlas en slices, pasarlas entre funciones y componerlas:

```go
// Define reusable option sets
func defaultPatientOptions() []r4.PatientOption {
    return []r4.PatientOption{
        r4.WithPatientActive(true),
        r4.WithPatientLanguage("en"),
    }
}

func main() {
    // Start with defaults and add specific options
    opts := defaultPatientOptions()
    opts = append(opts,
        r4.WithPatientId("patient-composed"),
        r4.WithPatientGender(r4.AdministrativeGenderFemale),
    )

    patient := r4.NewPatient(opts...)
}
```

Este patrón es particularmente útil en código de pruebas donde quieres definir fixtures base y sobreescribir campos específicos por caso de prueba.

## Opciones Condicionales

Puedes incluir opciones condicionalmente usando el flujo de control estándar de Go:

```go
func createPatient(id string, birthDate string, deceased bool) *r4.Patient {
    opts := []r4.PatientOption{
        r4.WithPatientId(id),
        r4.WithPatientActive(true),
        r4.WithPatientBirthDate(birthDate),
    }

    if deceased {
        opts = append(opts, r4.WithPatientDeceasedBoolean(true))
    }

    return r4.NewPatient(opts...)
}
```

## Recurso Vacío

Llamar al constructor sin opciones devuelve un recurso válido y vacío:

```go
patient := r4.NewPatient()
// patient.Id is nil
// patient.Active is nil
// patient.Name is empty
```

## Convención de Nombres

Todas las funciones de opción siguen un patrón de nombres consistente:

| Patrón | Propósito | Ejemplo |
|---------|---------|---------|
| `With<Resource><Field>(v)` | Establecer un campo singular | `WithPatientId("123")` |
| `With<Resource><Field>(v)` | Agregar a un campo repetitivo | `WithPatientName(humanName)` |

Para campos singulares (como `Id`, `Active`, `Gender`), la función de opción establece el campo. Para campos repetitivos (como `Name`, `Identifier`, `Telecom`), la función de opción agrega al slice existente. Llamar a `WithPatientName` dos veces agrega dos nombres.

## Cuándo Usar Opciones Funcionales

Las opciones funcionales son ideales cuando:

- Quieres una construcción limpia y legible sin código repetitivo de punteros.
- Necesitas componer opciones de diferentes fuentes (valores por defecto, sobreescrituras, lógica condicional).
- Estás escribiendo código de biblioteca que acepta configuración de los llamadores.
- Quieres construir fixtures de prueba con configuraciones base reutilizables.

Para control total sobre cada campo, considera los [Literales de Struct](../struct-literals). Para construcción paso a paso con una API encadenable, consulta el [Patrón Builder](../builder-pattern).

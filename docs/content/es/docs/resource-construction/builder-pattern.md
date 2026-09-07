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
        Given:  r4.PtrSlice("Maria"),
    }).
    Build()
```

Los tipos de datos tienen sus propios builders, así que la cadena no se corta ahí:

```go
patient := r4.NewPatientBuilder().
    SetId("patient-789").
    SetGender(r4.AdministrativeGenderFemale).
    SetBirthDate("1985-06-20").
    AddName(r4.NewHumanNameBuilder().
        SetUse(r4.NameUseOfficial).
        SetFamily("Garcia").
        AddGiven("Maria").
        Build()).
    Build()
```

`Build()` en un tipo de datos devuelve un **valor**, no un puntero, porque es lo
que toma todo consumidor: `Patient.Name` es `[]HumanName`, e incluso un campo
puntero como `Range.Low` se asigna con `SetLow(Quantity)`, que toma la dirección
por su cuenta. Así no hace falta desreferenciar en la llamada.

Los recursos van al contrario —`Build()` devuelve `*Patient`— porque se manejan
como punteros y satisfacen la interfaz `Resource` con receptor puntero.

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
    AddName(r4.HumanName{Family: &family, Given: r4.PtrSlice("Robert")}).
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
| `Set<Campo>(v)` | Establecer un campo singular | `SetId("123")`, `SetActive(true)` |
| `Add<Campo>(v)` | Agregar a un campo repetitivo | `AddName(humanName)` |
| `Set<Campo>Ext(v)` | Establecer las extensiones de un primitivo | `SetBirthDateExt(element)` |
| `Add<Campo>Ext(v)` | Agregar una ranura de extensión a un primitivo repetido | `AddGivenExt(&element)` |
| `Build()` | Devolver el valor construido | `Build()` |

Los métodos `Set` aceptan valores sin envolver (`string` en lugar de `*string`) y
crean el puntero internamente.

### Todo tipo tiene el suyo

Los builders se generan para recursos, tipos de datos **y** backbone elements:
663 en R4, 681 en R4B, 834 en R5. Un backbone es donde vive la mayor parte de los
datos de un recurso, así que detenerse en el nivel del recurso obligaría a usar un
literal para `Patient.contact`, `Bundle.entry` u `Observation.component`:

```go
patient := r4.NewPatientBuilder().
    AddContact(r4.NewPatientContactBuilder().
        SetName(r4.NewHumanNameBuilder().SetFamily("Doe").Build()).
        AddTelecom(r4.NewContactPointBuilder().
            SetSystem(r4.ContactPointSystemPhone).
            SetValue("555-0100").
            Build()).
        Build()).
    Build()
```

### Extensiones sobre primitivos

Un primitivo lleva sus extensiones en un campo compañero —`BirthDateExt` para
`birthDate`, serializado como `_birthDate`—. Es la única forma en que FHIR expresa
una extensión sobre un valor primitivo, y el builder llega hasta ahí:

```go
patient := r4.NewPatientBuilder().
    SetBirthDateExt(r4.NewElementBuilder().
        AddExtension(r4.NewExtensionBuilder().
            SetUrl("http://hl7.org/fhir/StructureDefinition/data-absent-reason").
            SetValueCode("asked-declined").
            Build()).
        Build()).
    Build()

// {"resourceType":"Patient","_birthDate":{"extension":[{"url":"...","valueCode":"asked-declined"}]}}
```

Fíjate en que no hay valor de `birthDate` en esa salida: la extensión dice *por
qué* falta la fecha, que es justo el propósito.

En un primitivo **repetido** los dos slices son paralelos por posición, así que
`Add<Campo>Ext` se asocia al elemento añadido más recientemente y rellena con nil
cualquier hueco anterior:

```go
name := r4.NewHumanNameBuilder().
    AddGiven("A").
    AddGiven("B").
    AddGivenExt(&ext).      // pertenece a "B"
    Build()

// {"given":["A","B"],"_given":[null,{...}]}
```

Pasar `nil` ahí es significativo: es una posición sin extensión.

### Los campos choice son exclusivos

Un elemento choice contiene exactamente una variante, así que establecer una
limpia las demás, incluido el campo compañero `_field` de una variante primitiva:

```go
obs := r4.NewObservationBuilder().
    SetValueString("a").
    SetValueBoolean(true).      // limpia valueString
    Build()

// {"resourceType":"Observation","valueBoolean":true}
```

Sin eso, una cadena de setters producía un documento con varias variantes
presentes, que ningún servidor FHIR acepta. Un literal de struct todavía puede
hacerlo: la guarda está en el builder.

## Cuándo Usar el Builder

El patrón builder es ideal cuando:

- Estás construyendo recursos paso a paso, posiblemente a través de múltiples llamadas de función.
- Quieres una cadena fluida y legible de asignaciones de campos.
- Quieres evitar el código repetitivo de punteros para campos primitivos.

Para una inicialización de una sola vez con control total, considera los [Literales de Struct](../struct-literals).

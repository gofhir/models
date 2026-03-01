---
title: "Patrones Comunes"
linkTitle: "Patrones Comunes"
description: "Patrones de codigo del mundo real para crear patients, observations, bundles y trabajar con tipos de datos codificados y helpers."
weight: 1
---

Esta pagina demuestra los patrones de construccion de recursos mas comunes que encontraras al usar `gofhir/models`. Todos los ejemplos usan el paquete R4.

## Funcion Auxiliar

La mayoria de los ejemplos usan un helper generico de punteros. Definelo una vez en tu proyecto:

```go
func ptrTo[T any](v T) *T {
    return &v
}
```

## 1. Creando un Patient con Demografia Completa

Un paciente completo con nombre, direccion, identificadores, telecom y campos demograficos:

```go
import "github.com/gofhir/models/r4"

patient := r4.NewPatientBuilder().
    SetId("patient-001").
    SetActive(true).
    SetGender(r4.AdministrativeGenderFemale).
    SetBirthDate("1992-03-14").
    AddName(r4.HumanName{
        Use:    ptrTo(r4.NameUseOfficial),
        Family: ptrTo("Garcia"),
        Given:  []string{"Maria", "Elena"},
        Prefix: []string{"Dr."},
    }).
    AddName(r4.HumanName{
        Use:    ptrTo(r4.NameUseNickname),
        Given:  []string{"Mari"},
    }).
    AddIdentifier(r4.Identifier{
        Use:    ptrTo(r4.IdentifierUseOfficial),
        System: ptrTo("http://hospital.example.org/mrn"),
        Value:  ptrTo("MRN-2024-001"),
    }).
    AddIdentifier(r4.Identifier{
        Use:    ptrTo(r4.IdentifierUseSecondary),
        System: ptrTo("http://hl7.org/fhir/sid/us-ssn"),
        Value:  ptrTo("123-45-6789"),
    }).
    AddTelecom(r4.ContactPoint{
        System: ptrTo(r4.ContactPointSystemPhone),
        Value:  ptrTo("+1-555-0199"),
        Use:    ptrTo(r4.ContactPointUseMobile),
    }).
    AddTelecom(r4.ContactPoint{
        System: ptrTo(r4.ContactPointSystemEmail),
        Value:  ptrTo("maria.garcia@example.com"),
        Use:    ptrTo(r4.ContactPointUseWork),
    }).
    AddAddress(r4.Address{
        Use:        ptrTo(r4.AddressUseHome),
        Type:       ptrTo(r4.AddressTypeBoth),
        Line:       []string{"742 Evergreen Terrace", "Apt 4B"},
        City:       ptrTo("Portland"),
        State:      ptrTo("OR"),
        PostalCode: ptrTo("97201"),
        Country:    ptrTo("US"),
    }).
    SetMaritalStatus(r4.CodeableConcept{
        Coding: []r4.Coding{{
            System:  ptrTo("http://terminology.hl7.org/CodeSystem/v3-MaritalStatus"),
            Code:    ptrTo("M"),
            Display: ptrTo("Married"),
        }},
        Text: ptrTo("Married"),
    }).
    SetManagingOrganization(r4.Reference{
        Reference: ptrTo("Organization/org-001"),
        Display:   ptrTo("Portland General Hospital"),
    }).
    Build()
```

## 2. Construyendo una Observation (Signo Vital con Valor Quantity)

Una observacion de peso corporal con todos los campos que un servidor FHIR tipicamente espera:

```go
import "github.com/gofhir/models/r4"

observation := r4.NewObservationBuilder().
    SetId("obs-weight-001").
    SetStatus(r4.ObservationStatusFinal).
    AddCategory(r4.CodeableConcept{
        Coding: []r4.Coding{{
            System:  ptrTo("http://terminology.hl7.org/CodeSystem/observation-category"),
            Code:    ptrTo("vital-signs"),
            Display: ptrTo("Vital Signs"),
        }},
    }).
    SetCode(r4.CodeableConcept{
        Coding: []r4.Coding{{
            System:  ptrTo("http://loinc.org"),
            Code:    ptrTo("29463-7"),
            Display: ptrTo("Body weight"),
        }},
        Text: ptrTo("Body Weight"),
    }).
    SetSubject(r4.Reference{
        Reference: ptrTo("Patient/patient-001"),
    }).
    SetEncounter(r4.Reference{
        Reference: ptrTo("Encounter/enc-001"),
    }).
    SetEffectiveDateTime("2024-06-15T09:30:00Z").
    SetValueQuantity(r4.Quantity{
        Value:  r4.NewDecimalFromFloat64(72.5),
        Unit:   ptrTo("kg"),
        System: ptrTo("http://unitsofmeasure.org"),
        Code:   ptrTo("kg"),
    }).
    AddPerformer(r4.Reference{
        Reference: ptrTo("Practitioner/pract-001"),
        Display:   ptrTo("Dr. Smith"),
    }).
    Build()
```

## 3. Construyendo un Bundle de Recursos

Un bundle de transaccion conteniendo un Patient y una Observation relacionada:

```go
import "github.com/gofhir/models/r4"

bundleType := r4.BundleTypeTransaction

bundle := r4.NewBundleBuilder().
    SetId("bundle-001").
    SetType(bundleType).
    AddEntry(r4.BundleEntry{
        FullUrl: ptrTo("urn:uuid:patient-temp-1"),
        Resource: r4.NewPatient(
            r4.WithPatientName(r4.HumanName{
                Family: ptrTo("Doe"),
                Given:  []string{"Jane"},
            }),
        ),
        Request: &r4.BundleEntryRequest{
            Method: ptrTo(r4.HTTPVerbPOST),
            Url:    ptrTo("Patient"),
        },
    }).
    AddEntry(r4.BundleEntry{
        FullUrl: ptrTo("urn:uuid:obs-temp-1"),
        Resource: r4.NewObservation(
            r4.WithObservationStatus(r4.ObservationStatusFinal),
            r4.WithObservationCode(r4.CodeableConcept{
                Coding: []r4.Coding{{
                    System: ptrTo("http://loinc.org"),
                    Code:   ptrTo("8867-4"),
                    Display: ptrTo("Heart rate"),
                }},
            }),
            r4.WithObservationSubject(r4.Reference{
                Reference: ptrTo("urn:uuid:patient-temp-1"),
            }),
            r4.WithObservationValueQuantity(r4.Quantity{
                Value:  r4.NewDecimalFromFloat64(72),
                Unit:   ptrTo("/min"),
                System: ptrTo("http://unitsofmeasure.org"),
                Code:   ptrTo("/min"),
            }),
        ),
        Request: &r4.BundleEntryRequest{
            Method: ptrTo(r4.HTTPVerbPOST),
            Url:    ptrTo("Observation"),
        },
    }).
    Build()
```

## 4. Trabajando con CodeableConcept y Coding

`CodeableConcept` es uno de los tipos de datos FHIR mas utilizados. Representa un concepto que puede estar definido por uno o mas sistemas de codificacion:

```go
import "github.com/gofhir/models/r4"

// Un CodeableConcept con multiples codings (por ejemplo, mismo concepto en SNOMED e ICD-10)
diagnosisCode := r4.CodeableConcept{
    Coding: []r4.Coding{
        {
            System:  ptrTo("http://snomed.info/sct"),
            Code:    ptrTo("73211009"),
            Display: ptrTo("Diabetes mellitus"),
        },
        {
            System:  ptrTo("http://hl7.org/fhir/sid/icd-10-cm"),
            Code:    ptrTo("E11"),
            Display: ptrTo("Type 2 diabetes mellitus"),
        },
    },
    Text: ptrTo("Type 2 Diabetes Mellitus"),
}

// Un Coding simple (para uso en bindings de sistemas de codigos)
genderCoding := r4.Coding{
    System:  ptrTo("http://hl7.org/fhir/administrative-gender"),
    Code:    ptrTo("female"),
    Display: ptrTo("Female"),
}

// Buscando a traves de codings
func hasCode(cc r4.CodeableConcept, system, code string) bool {
    for _, coding := range cc.Coding {
        if coding.System != nil && *coding.System == system &&
            coding.Code != nil && *coding.Code == code {
            return true
        }
    }
    return false
}

// Uso
if hasCode(diagnosisCode, "http://snomed.info/sct", "73211009") {
    fmt.Println("Patient has diabetes")
}
```

## 5. Usando el Paquete Helpers

El paquete `helpers` proporciona valores `CodeableConcept` preconstruidos para necesidades comunes de codificacion clinica:

### Categorias de Observation

```go
import (
    "github.com/gofhir/models/r4"
    "github.com/gofhir/models/r4/helpers"
)

observation := r4.NewObservationBuilder().
    SetStatus(r4.ObservationStatusFinal).
    AddCategory(helpers.ObservationCategoryVitalSigns).
    SetCode(helpers.BodyWeight).
    SetValueQuantity(helpers.QuantityKg(75.0)).
    Build()
```

### Constantes de Categorias Disponibles

| Variable | Sistema | Codigo |
|----------|---------|--------|
| `helpers.ObservationCategoryVitalSigns` | observation-category | vital-signs |
| `helpers.ObservationCategoryLaboratory` | observation-category | laboratory |
| `helpers.ObservationCategorySocialHistory` | observation-category | social-history |
| `helpers.ObservationCategoryImaging` | observation-category | imaging |
| `helpers.ObservationCategorySurvey` | observation-category | survey |
| `helpers.ObservationCategoryExam` | observation-category | exam |
| `helpers.ObservationCategoryTherapy` | observation-category | therapy |
| `helpers.ObservationCategoryActivity` | observation-category | activity |

### Codigos LOINC para Signos Vitales

```go
import "github.com/gofhir/models/r4/helpers"

// CodeableConcepts preconstruidos para signos vitales comunes
helpers.BodyWeight       // LOINC 29463-7
helpers.BodyHeight       // LOINC 8302-2
helpers.BodyTemperature  // LOINC 8310-5
helpers.HeartRate         // LOINC 8867-4
helpers.VitalSignsPanel  // LOINC 85353-1
```

### Helpers de Quantity UCUM

El paquete `helpers` tambien proporciona funciones para crear valores `Quantity` con unidades UCUM estandar:

```go
import "github.com/gofhir/models/r4/helpers"

weight := helpers.QuantityKg(72.5)    // 72.5 kg
height := helpers.QuantityCm(175.0)   // 175 cm
temp := helpers.QuantityCel(37.2)     // 37.2 Cel (Celsius)
```

### Ejemplo Completo de Signos Vitales

```go
import (
    "github.com/gofhir/models/r4"
    "github.com/gofhir/models/r4/helpers"
)

// Construir una observacion de temperatura corporal usando helpers
tempObs := r4.NewObservationBuilder().
    SetId("temp-001").
    SetStatus(r4.ObservationStatusFinal).
    AddCategory(helpers.ObservationCategoryVitalSigns).
    SetCode(helpers.BodyTemperature).
    SetSubject(r4.Reference{Reference: ptrTo("Patient/patient-001")}).
    SetEffectiveDateTime("2024-06-15T14:30:00Z").
    SetValueQuantity(helpers.QuantityCel(38.1)).
    Build()
```

{{< callout type="info" >}}
El paquete `helpers` esta escrito manualmente (no es generado) y actualmente solo esta disponible para R4. Esta disenado para reducir el codigo repetitivo para los codigos clinicos y unidades mas comunmente utilizados.
{{< /callout >}}

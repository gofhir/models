---
title: "Versiones FHIR"
linkTitle: "Versiones FHIR"
description: "Comprende las diferencias entre FHIR R4, R4B y R5 y cómo cada versión se empaqueta como un módulo Go."
weight: 3
---

El proyecto **gofhir/models** soporta tres versiones de la especificación FHIR. Cada versión se publica como un módulo Go independiente con sus propios tipos, builders, sistemas de códigos y funciones de serialización.

## Versiones Soportadas

| Versión | Especificación FHIR | Ruta del Módulo Go | Estado |
|---------|-----------|----------------|--------|
| **R4** | 4.0.1 | `github.com/gofhir/models/r4` | Estable, la más ampliamente adoptada |
| **R4B** | 4.3.0 | `github.com/gofhir/models/r4b` | Estable, versión de transición |
| **R5** | 5.0.0 | `github.com/gofhir/models/r5` | Estable, última versión normativa |

## Diferencias entre Versiones

### FHIR R4 (4.0.1)

R4 es la versión FHIR más ampliamente desplegada en sistemas de producción. Fue publicada en 2019 y es utilizada por la Guía de Implementación US Core, el framework SMART on FHIR y la mayoría de las implementaciones comerciales de servidores FHIR.

```go
import "github.com/gofhir/models/r4"

patient := r4.NewPatient(
    r4.WithPatientId("r4-example"),
    r4.WithPatientGender(r4.AdministrativeGenderMale),
)
```

El paquete R4 también incluye un sub-paquete opcional `helpers` con valores `CodeableConcept` preconstruidos:

```go
import "github.com/gofhir/models/r4/helpers"

// Use a pre-built vital signs category
category := helpers.ObservationCategoryVitalSigns
```

### FHIR R4B (4.3.0)

R4B es una versión de transición publicada en 2022. Es retrocompatible con R4 para la mayoría de los recursos, pero introduce nuevos recursos y actualizaciones en los recursos relacionados con terminología (CodeSystem, ValueSet, ConceptMap) que se alinean con la dirección de R5.

```go
import "github.com/gofhir/models/r4b"

patient := r4b.NewPatient(
    r4b.WithPatientId("r4b-example"),
    r4b.WithPatientGender(r4b.AdministrativeGenderFemale),
)
```

R4B se utiliza típicamente cuando necesitas soportar sistemas que están en transición de R4 hacia R5.

### FHIR R5 (5.0.0)

R5 es la última versión normativa, publicada en 2023. Incluye cambios significativos en varios recursos, nuevos recursos, sistemas de códigos actualizados y cambios estructurales en los elementos backbone.

```go
import "github.com/gofhir/models/r5"

patient := r5.NewPatient(
    r5.WithPatientId("r5-example"),
    r5.WithPatientGender(r5.AdministrativeGenderMale),
)
```

Las diferencias clave en R5 incluyen cambios en las estructuras de componentes de Observation, nuevos recursos como `SubscriptionTopic` y recursos de terminología actualizados.

## Estructura de Rutas de Módulos

Cada versión es su propio módulo Go con un archivo `go.mod` separado. Esto significa:

1. **Versionado independiente** -- Cada módulo tiene su propia versión semántica. Una actualización del paquete R5 no requiere actualizar R4.
2. **Sin conflictos de dependencias** -- Importar múltiples versiones en el mismo proyecto no crea problemas de dependencias diamante.
3. **Tamaño mínimo del binario** -- Tu binario compilado solo incluye la(s) versión(es) FHIR que importas.

```
github.com/gofhir/models/
    r4/                # module: github.com/gofhir/models/r4
        go.mod
        helpers/       # sub-package within the r4 module
    r4b/               # module: github.com/gofhir/models/r4b
        go.mod
    r5/                # module: github.com/gofhir/models/r5
        go.mod
```

## Elegir una Versión

- **Usa R4** si estás construyendo contra US Core, SMART on FHIR o cualquier sistema que implemente FHIR R4. Esta es la opción predeterminada más segura para la mayoría de los proyectos.
- **Usa R4B** si tu sistema objetivo requiere específicamente la especificación 4.3.0, especialmente para recursos de terminología actualizados.
- **Usa R5** si estás construyendo aplicaciones nuevas desde cero o apuntando a sistemas que han adoptado el último estándar FHIR.

## Consistencia de API entre Versiones

Los tres paquetes exponen los mismos patrones de API:

- Tipos de struct: `r4.Patient`, `r4b.Patient`, `r5.Patient`
- Opciones funcionales: `r4.NewPatient(opts...)`, `r4b.NewPatient(opts...)`, `r5.NewPatient(opts...)`
- Builders: `r4.NewPatientBuilder()`, `r4b.NewPatientBuilder()`, `r5.NewPatientBuilder()`
- Serialización: `r4.Marshal(v)`, `r4b.Marshal(v)`, `r5.Marshal(v)`
- Deserialización: `r4.UnmarshalResource(data)`, `r4b.UnmarshalResource(data)`, `r5.UnmarshalResource(data)`
- Sistemas de códigos: `r4.AdministrativeGenderMale`, `r4b.AdministrativeGenderMale`, `r5.AdministrativeGenderMale`

Los campos de struct y los valores de sistemas de códigos difieren entre versiones para coincidir con las StructureDefinitions FHIR correspondientes, pero los patrones de construcción y serialización permanecen iguales.

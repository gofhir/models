---
title: "Temas Avanzados"
linkTitle: "Avanzado"
description: "Funcionalidades avanzadas de gofhir/models incluyendo metadatos del modelo FHIRPath, campos de resumen, generacion de codigo y soporte multi-version."
weight: 5
---

Mas alla de la construccion basica de recursos y la serializacion, la biblioteca `gofhir/models` proporciona varias funcionalidades avanzadas que soportan la construccion de infraestructura FHIR para produccion. Estas capacidades se generan junto con las structs de recursos y exponen los metadatos enriquecidos incorporados en la especificacion FHIR.

## Temas

{{< cards >}}
  {{< card link="fhirpath-model" title="Modelo FHIRPath" subtitle="Metadatos de tipo en tiempo de ejecucion para la evaluacion de expresiones FHIRPath, incluyendo choice types, jerarquias de tipos y objetivos de referencia." icon="academic-cap" >}}
  {{< card link="summary-fields" title="Campos de Resumen" subtitle="Listas de campos de resumen precalculadas para implementar el parametro de busqueda _summary=true en un servidor FHIR." icon="document-text" >}}
  {{< card link="code-generation" title="Generacion de Codigo" subtitle="Como el generador lee las StructureDefinitions de FHIR y produce todo el codigo fuente Go para recursos, builders y metadatos." icon="cog" >}}
  {{< card link="multi-version" title="Soporte Multi-Version" subtitle="Arquitectura de workspace de Go para importar multiples versiones de FHIR en paralelo con versionado independiente de modulos." icon="collection" >}}
{{< /cards >}}

## Descripcion General

| Funcionalidad | Exportacion del Paquete | Proposito |
|---------------|------------------------|-----------|
| Modelo FHIRPath | `FHIRPathModel()` | Informacion de tipos en tiempo de ejecucion para motores FHIRPath |
| Campos de resumen | `SummaryFields` | Listas de campos para el comportamiento `_summary=true` del servidor |
| Generador de codigo | `cmd/generator` | Regenerar todo el codigo Go desde las StructureDefinitions de FHIR |
| Multi-version | `go.work` | Importar R4, R4B y R5 en el mismo proyecto |

Cada una de estas funcionalidades esta disenada para usarse de forma independiente. No necesitas un motor FHIRPath para usar los campos de resumen, y no necesitas el generador de codigo para usar la biblioteca en absoluto -- el codigo generado ya esta committeado y publicado como modulos de Go.

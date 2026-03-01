---
title: "Documentación"
linkTitle: "Documentación"
description: "Documentación completa de gofhir/models -- structs Go con tipado seguro para recursos FHIR R4, R4B y R5."
weight: 1
---

Bienvenido a la documentación de **gofhir/models**. Esta biblioteca proporciona structs Go autogenerados y con tipado seguro para cada recurso y tipo de dato definido en la especificación HL7 FHIR. Cubre FHIR R4, R4B y R5, cada uno publicado como un módulo Go independiente.

## Por Dónde Empezar

{{< cards cols="2" >}}
  {{< card link="getting-started" title="Primeros Pasos" subtitle="Instala la biblioteca, crea tu primer recurso FHIR y aprende sobre las versiones FHIR soportadas." icon="play" >}}
  {{< card link="resource-construction" title="Construcción de Recursos" subtitle="Explora tres patrones para construir recursos: literales de struct, builders fluidos y opciones funcionales." icon="puzzle" >}}
{{< /cards >}}

## Características Principales

- **Todas las versiones FHIR** -- R4 (4.0.1), R4B (4.3.0) y R5 (5.0.0) con cada recurso, elemento backbone, tipo de dato y sistema de códigos.
- **Tres patrones de construcción** -- Elige entre literales de struct directos, cadenas de builder fluido u opciones funcionales según tu caso de uso.
- **Serialización JSON** -- Marshaling seguro para HTML que preserva el contenido XHTML narrativo de FHIR, con inyección automática de `resourceType`.
- **Serialización XML** -- Marshaling y unmarshaling XML completamente conforme con FHIR mediante `MarshalXML` y `UnmarshalXML` en cada tipo.
- **Deserialización polimórfica** -- `UnmarshalResource(data)` detecta automáticamente el `resourceType` y devuelve el struct Go correcto.
- **Decimales con preservación de precisión** -- Un tipo `Decimal` personalizado almacena la representación textual exacta (por ejemplo, `"1.50"` conserva el cero final).
- **Metadatos del modelo FHIRPath** -- Información de tipos en tiempo de ejecución para motores FHIRPath, incluyendo rutas de tipos choice, jerarquías de tipos y objetivos de referencia.
- **Constantes auxiliares** -- Valores `CodeableConcept` preconstruidos para categorías comunes de observación, categorías de condición, códigos LOINC y unidades UCUM.
- **Interfaces Resource y DomainResource** -- Interfaces estándar de Go que coinciden con la jerarquía de tipos FHIR para el manejo genérico de recursos.

## Resumen de Paquetes

| Paquete | Descripción |
|---------|-------------|
| `github.com/gofhir/models/r4` | Recursos, tipos de datos, sistemas de códigos, builders y serialización de FHIR R4 (4.0.1) |
| `github.com/gofhir/models/r4b` | Recursos, tipos de datos, sistemas de códigos, builders y serialización de FHIR R4B (4.3.0) |
| `github.com/gofhir/models/r5` | Recursos, tipos de datos, sistemas de códigos, builders y serialización de FHIR R5 (5.0.0) |
| `github.com/gofhir/models/r4/helpers` | CodeableConcepts preconstruidos para categorías comunes (observación, condición, LOINC, UCUM) |

## Estructura del Proyecto

Cada versión FHIR reside en su propio directorio en la raíz del repositorio y se publica como un módulo Go separado:

```
models/
  r4/            # github.com/gofhir/models/r4
    helpers/     # github.com/gofhir/models/r4/helpers
  r4b/           # github.com/gofhir/models/r4b
  r5/            # github.com/gofhir/models/r5
```

Todos los tipos se generan automáticamente a partir de las StructureDefinitions oficiales de FHIR mediante el generador de código en el directorio `cmd/generator`. Nunca debes editar los archivos fuente generados directamente.

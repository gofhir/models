---
title: "Referencia de API"
linkTitle: "Referencia de API"
description: "Referencia completa de API para las interfaces, funciones del registro, builders y accesores del modelo FHIRPath de gofhir/models."
weight: 6
---

Esta seccion proporciona una referencia detallada de API para los tipos y funciones principales exportados por los paquetes de `gofhir/models`. La misma superficie de API esta disponible en cada paquete de version (`r4`, `r4b`, `r5`), con definiciones de tipos que reflejan la especificacion FHIR correspondiente.

## Temas

{{< cards >}}
  {{< card link="resource-interfaces" title="Interfaces de Recursos" subtitle="Las interfaces Resource y DomainResource que todos los recursos FHIR implementan." icon="code" >}}
  {{< card link="registry-functions" title="Funciones del Registro" subtitle="Funciones de fabrica, deserializacion e introspeccion para el manejo dinamico de recursos." icon="server" >}}
  {{< card link="builder-api" title="API del Builder" subtitle="Patron builder fluido con metodos Set/Add y opciones funcionales para cada tipo de recurso." icon="puzzle" >}}
  {{< card link="fhirpath-model-api" title="API del Modelo FHIRPath" subtitle="Referencia completa para los metodos de acceso de FHIRPathModelData." icon="academic-cap" >}}
{{< /cards >}}

## Resumen de la API

La biblioteca exporta una superficie de API pequena y enfocada sobre las structs de recursos generadas:

| Categoria | Exportaciones Clave | Proposito |
|-----------|---------------------|-----------|
| Interfaces | `Resource`, `DomainResource` | Manejo generico de recursos sin type assertions |
| Registro | `NewResource`, `UnmarshalResource`, `GetResourceType`, `IsKnownResourceType`, `AllResourceTypes` | Creacion y deserializacion dinamica de recursos |
| Builders | `New<Resource>Builder`, `Set*`, `Add*`, `Build` | Construccion fluida de recursos |
| Opciones Funcionales | `New<Resource>`, `With<Resource><Field>` | Creacion concisa de recursos con opciones |
| Serializacion | `Marshal`, `MarshalIndent`, `MarshalResourceXML`, `MarshalResourceXMLIndent`, `UnmarshalResourceXML` | Codificacion JSON y XML conforme a FHIR |
| Metadatos | `FHIRPathModel`, `SummaryFields` | Informacion de tipos en tiempo de ejecucion y listas de campos de resumen |

Todos los ejemplos en esta seccion usan el paquete `r4`. Los mismos patrones aplican a `r4b` y `r5` con sus respectivas definiciones de tipos.

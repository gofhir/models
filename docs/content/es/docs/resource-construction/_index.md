---
title: "Construcción de Recursos"
linkTitle: "Construcción de Recursos"
description: "Aprende dos patrones para construir recursos FHIR en Go: literales de struct y builders fluidos."
weight: 2
---

La biblioteca **gofhir/models** proporciona dos patrones distintos para crear recursos FHIR. Ambos producen los mismos tipos de struct Go, así que puedes mezclarlos dentro de un mismo proyecto. Elige el que mejor se adapte a tu estilo de programación y caso de uso.

## Descripción General

| Patrón | Ideal Para | Verbosidad | Seguridad en Compilación |
|---------|----------|-----------|---------------------|
| [Literales de Struct](struct-literals) | Control total, inicialización de una sola vez | Media (requiere punteros) | Completa |
| [Patrón Builder](builder-pattern) | Construcción paso a paso, cadenas fluidas | Baja | Completa |

Ambos patrones establecen campos en el mismo struct subyacente (por ejemplo, `r4.Patient`), por lo que la serialización y deserialización funcionan de forma idéntica independientemente de cómo se haya creado el recurso.

## Comparación Rápida

### Literal de Struct

```go
active := true
patient := r4.Patient{
    Id:           &id,
    Active:       &active,
}
```

### Patrón Builder

```go
patient := r4.NewPatientBuilder().
    SetId("patient-1").
    SetActive(true).
    Build()
```

## Guías

{{< cards >}}
  {{< card link="struct-literals" title="Literales de Struct" subtitle="Inicialización directa de structs con control total sobre cada campo." icon="code" >}}
  {{< card link="builder-pattern" title="Patrón Builder" subtitle="API fluida y encadenable para la construcción paso a paso de recursos." icon="cube" >}}
  {{< card link="working-with-primitives" title="Trabajo con Primitivos" subtitle="Tipos primitivos FHIR, punteros, el tipo Decimal y elementos de extensión." icon="variable" >}}
{{< /cards >}}

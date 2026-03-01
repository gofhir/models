---
title: "Primeros Pasos"
linkTitle: "Primeros Pasos"
description: "Instala gofhir/models, crea tu primer recurso FHIR y aprende sobre las versiones FHIR soportadas."
weight: 1
---

Esta sección te guía a través de la instalación de la biblioteca, la construcción de tu primer recurso FHIR en Go y la comprensión de cómo se organizan las tres versiones FHIR soportadas.

## Descripción General

**gofhir/models** proporciona structs Go con tipado seguro para todos los recursos, tipos de datos y sistemas de códigos FHIR. Cada versión FHIR (R4, R4B, R5) se publica como un módulo Go independiente, así que solo importas lo que necesitas.

La biblioteca está diseñada en torno a tres principios fundamentales:

1. **Tipado seguro** -- Cada campo FHIR se mapea a un campo de struct Go con tipado fuerte y semántica adecuada de punteros para valores opcionales.
2. **Múltiples patrones de construcción** -- Elige entre literales de struct, builders fluidos u opciones funcionales según tu estilo de programación.
3. **Serialización conforme con FHIR** -- El marshaling de JSON y XML sigue la especificación FHIR exactamente, incluyendo la preservación de XHTML narrativo y la precisión decimal.

## Guías

{{< cards >}}
  {{< card link="installation" title="Instalación" subtitle="Instala el módulo Go para tu versión FHIR objetivo y configura tu proyecto." icon="download" >}}
  {{< card link="quick-start" title="Inicio Rápido" subtitle="Crea, serializa y deserializa recursos FHIR con ejemplos de código funcionales." icon="play" >}}
  {{< card link="fhir-versions" title="Versiones FHIR" subtitle="Comprende las diferencias entre R4, R4B y R5 y cómo están empaquetadas." icon="book-open" >}}
{{< /cards >}}

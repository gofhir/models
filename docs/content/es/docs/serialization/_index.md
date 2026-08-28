---
title: "Serialización"
linkTitle: "Serialización"
description: "Soporte de serialización JSON y XML para recursos FHIR en la biblioteca gofhir/models."
weight: 3
---

La biblioteca `gofhir/models` serializa recursos FHIR a JSON y XML. Cada struct de recurso generado implementa las interfaces estándar de marshaling de Go, y la biblioteca también proporciona funciones personalizadas para los requisitos de serialización específicos de FHIR.

{{< callout type="warning" >}}
Medido sobre los corpus oficiales de ejemplos, **los 3653 ejemplos XML publicados sobreviven un round-trip**, igual que el 99,9 % de los 8758 en JSON. Consulta [XML Marshaling](xml-marshaling/) para ver qué mide y qué no mide esa cifra de XML.
{{< /callout >}}

## Descripción General de la Serialización

Los recursos FHIR pueden intercambiarse en dos formatos de transmisión: JSON y XML. La biblioteca `gofhir/models` soporta ambos, con especial atención a los requisitos de la especificación FHIR para cada formato.

### Serialización JSON

Todos los structs de recursos implementan `json.Marshaler` y `json.Unmarshaler` del paquete estándar `encoding/json` de Go. Esto significa que puedes usar `json.Marshal()` y `json.Unmarshal()` directamente con cualquier tipo de recurso.

```go
import (
    "encoding/json"
    "github.com/gofhir/models/r4"
)

patient := &r4.Patient{
    ResourceType: "Patient",
    Id:           ptrTo("123"),
}
data, err := json.Marshal(patient)
```

Además, la biblioteca proporciona las funciones `r4.Marshal()` y `r4.MarshalIndent()` que resuelven un problema específico del codificador JSON estándar de Go: el escape de entidades HTML. Estas funciones personalizadas preservan el contenido HTML en los campos narrativos de FHIR exactamente como lo requiere la especificación FHIR.

### Serialización XML

La serialización XML se maneja a través de funciones auxiliares dedicadas en el módulo `xml_helpers.go`: `MarshalResourceXML()`, `MarshalResourceXMLIndent()` y `UnmarshalResourceXML()`. El manejo de namespaces, la convención FHIR de codificar primitivos como atributos `<name value="..."/>` y la narrativa XHTML funcionan. Consulta [XML Marshaling](xml-marshaling/).

### Deserialización Polimórfica

Al trabajar con datos FHIR sin procesar donde el tipo de recurso no se conoce en tiempo de compilación, la biblioteca proporciona un registro de recursos con funciones como `UnmarshalResource()`, `GetResourceType()` y `NewResource()` que permiten el despacho dinámico al struct de Go correcto basándose en el campo `resourceType`.

## Temas

{{< cards >}}
  {{< card link="json-marshaling" title="Serialización JSON" subtitle="Compatibilidad estándar con encoding/json mediante MarshalJSON y UnmarshalJSON." >}}
  {{< card link="xml-marshaling" title="Serialización XML" subtitle="Formato XML de FHIR con manejo de namespaces y codificación de primitivos." >}}
  {{< card link="custom-marshal" title="Marshal Personalizado" subtitle="Salida JSON segura para HTML con r4.Marshal() para contenido narrativo FHIR." >}}
  {{< card link="polymorphic-deserialization" title="Deserialización Polimórfica" subtitle="Registro de recursos para resolución dinámica de tipos desde JSON o XML sin procesar." >}}
{{< /cards >}}

## Comparación Rápida

| Método | Seguro para HTML | Indentado | Caso de Uso |
|--------|------------------|-----------|-------------|
| `json.Marshal()` | No | No | Salida JSON general sin contenido HTML |
| `json.MarshalIndent()` | No | Sí | Salida de depuración/visualización sin contenido HTML |
| `r4.Marshal()` | Sí | No | Salida JSON FHIR de producción |
| `r4.MarshalIndent()` | Sí | Sí | Salida JSON FHIR legible por humanos |
| `r4.MarshalResourceXML()` | N/A | No | Salida XML FHIR compacta |
| `r4.MarshalResourceXMLIndent()` | N/A | Sí | Salida XML FHIR legible por humanos |

{{< callout type="info" >}}
Para sistemas en producción que intercambian recursos FHIR, usa `r4.Marshal()` en lugar de `json.Marshal()` para asegurar que el contenido XHTML narrativo se preserve correctamente. Consulta la página de [Marshal Personalizado](custom-marshal) para más detalles.
{{< /callout >}}

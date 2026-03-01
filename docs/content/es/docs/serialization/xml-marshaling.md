---
title: "Serialización XML"
linkTitle: "Serialización XML"
description: "Serialización y deserialización XML de FHIR con manejo de namespaces y codificación de primitivos."
weight: 2
---

La biblioteca `gofhir/models` proporciona soporte completo para la serialización XML de FHIR a través de funciones auxiliares dedicadas definidas en `xml_helpers.go`. Cada struct de recurso implementa `MarshalXML` y `UnmarshalXML` del paquete `encoding/xml` de Go, y las funciones de nivel superior manejan la declaración XML, el namespace de FHIR y las convenciones de elementos auto-cerrados.

## Funciones Auxiliares XML

La biblioteca expone tres funciones principales para la serialización XML:

### MarshalResourceXML

Serializa un recurso FHIR a bytes XML con la declaración XML estándar y el namespace de FHIR.

```go
package main

import (
    "fmt"
    "log"

    "github.com/gofhir/models/r4"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    patient := &r4.Patient{
        ResourceType: "Patient",
        Id:           ptrTo("xml-example"),
        Active:       ptrTo(true),
        Gender:       ptrTo(r4.AdministrativeGenderMale),
        BirthDate:    ptrTo("1990-06-15"),
    }

    data, err := r4.MarshalResourceXML(patient)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(string(data))
}
```

La salida incluye la declaración XML y el namespace de FHIR:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Patient xmlns="http://hl7.org/fhir">
  <id value="xml-example"/>
  <active value="true"/>
  <gender value="male"/>
  <birthDate value="1990-06-15"/>
</Patient>
```

### MarshalResourceXMLIndent

Produce la misma salida que `MarshalResourceXML` pero con indentación personalizada para una salida legible por humanos:

```go
data, err := r4.MarshalResourceXMLIndent(patient, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

Los parámetros `prefix` e `indent` funcionan de la misma manera que en `xml.Encoder.Indent()`.

### UnmarshalResourceXML

Deserializa bytes XML de FHIR al tipo de recurso correcto. Lee el nombre del elemento raíz para determinar el tipo de recurso, crea el struct apropiado a través del registro de recursos y llama a `UnmarshalXML`:

```go
xmlData := []byte(`<?xml version="1.0" encoding="UTF-8"?>
<Patient xmlns="http://hl7.org/fhir">
  <id value="from-xml"/>
  <active value="true"/>
  <name>
    <use value="official"/>
    <family value="Smith"/>
    <given value="John"/>
  </name>
  <gender value="male"/>
</Patient>`)

resource, err := r4.UnmarshalResourceXML(xmlData)
if err != nil {
    log.Fatal(err)
}

patient := resource.(*r4.Patient)
fmt.Println(*patient.Id)             // "from-xml"
fmt.Println(*patient.Active)         // true
fmt.Println(*patient.Name[0].Given[0]) // "John"
```

## Manejo del Namespace FHIR

La especificación FHIR requiere que las representaciones XML usen el namespace `http://hl7.org/fhir`. Las funciones `MarshalResourceXML` y `MarshalResourceXMLIndent` agregan automáticamente este namespace al elemento raíz:

```xml
<Patient xmlns="http://hl7.org/fhir">
```

Durante la deserialización, `UnmarshalResourceXML` determina el tipo de recurso a partir del nombre local del elemento raíz, independientemente del prefijo de namespace.

## Codificación de Elementos Primitivos

FHIR XML codifica valores primitivos (cadenas, booleanos, enteros, decimales, fechas) como atributos XML en lugar de como contenido de texto del elemento. El valor se coloca en un atributo `value` en el elemento:

```xml
<!-- String primitive -->
<id value="example-123"/>

<!-- Boolean primitive -->
<active value="true"/>

<!-- Code primitive -->
<gender value="male"/>

<!-- Date primitive -->
<birthDate value="1990-06-15"/>

<!-- Decimal primitive -->
<value value="72.5"/>
```

Esto difiere del XML típico donde los valores son contenido de texto del elemento. La biblioteca maneja esto automáticamente a través de funciones auxiliares internas como `xmlEncodePrimitiveString`, `xmlEncodePrimitiveBool`, `xmlEncodePrimitiveInt`, `xmlEncodePrimitiveDecimal` y `xmlEncodePrimitiveCode`.

## Elementos Auto-cerrados

La especificación FHIR usa elementos auto-cerrados para primitivos sin hijos: `<id value="123"/>` en lugar de `<id value="123"></id>`. La biblioteca post-procesa la salida XML para colapsar elementos vacíos en forma auto-cerrada usando la función `collapseEmptyElements`.

## Codificación de Tipos Complejos

Los tipos complejos (como `HumanName`, `CodeableConcept`, `Reference`) se codifican como elementos XML anidados con sus elementos hijos:

```xml
<Patient xmlns="http://hl7.org/fhir">
  <id value="complex-example"/>
  <name>
    <use value="official"/>
    <family value="Johnson"/>
    <given value="Alice"/>
    <given value="Marie"/>
  </name>
  <telecom>
    <system value="phone"/>
    <value value="+1-555-0100"/>
    <use value="home"/>
  </telecom>
</Patient>
```

Nota que los elementos repetidos (como múltiples nombres `given`) aparecen como elementos XML separados con el mismo nombre de etiqueta, siguiendo la convención XML de FHIR.

## Recursos Contenidos en XML

Los recursos contenidos se envuelven en un elemento `<contained>`, con el tipo de recurso como un elemento anidado:

```xml
<Patient xmlns="http://hl7.org/fhir">
  <id value="with-contained"/>
  <contained>
    <Organization>
      <id value="org-1"/>
      <name value="Example Hospital"/>
    </Organization>
  </contained>
  <managingOrganization>
    <reference value="#org-1"/>
  </managingOrganization>
</Patient>
```

La biblioteca maneja esto a través de las funciones auxiliares `xmlEncodeContainedResource` y `xmlDecodeContainedResource`.

## Narrativa XHTML en XML

El campo `Narrative.Div` contiene contenido XHTML que debe preservarse tal cual en la salida XML. La biblioteca usa `xmlEncodeRawXHTML` para inyectar el contenido XHTML sin procesar directamente en el flujo XML sin re-codificarlo:

```go
patient := &r4.Patient{
    ResourceType: "Patient",
    Id:           ptrTo("with-narrative"),
    Text: &r4.Narrative{
        Status: ptrTo(r4.NarrativeStatusGenerated),
        Div:    ptrTo(`<div xmlns="http://www.w3.org/1999/xhtml"><p>John Smith</p></div>`),
    },
}

data, _ := r4.MarshalResourceXMLIndent(patient, "", "  ")
fmt.Println(string(data))
```

## Fidelidad de Ida y Vuelta en XML

La biblioteca soporta serialización XML de ida y vuelta (round-trip). Puedes serializar un recurso a XML, luego deserializarlo de vuelta, y el struct resultante contendrá los mismos datos:

```go
// Marshal to XML
xmlBytes, err := r4.MarshalResourceXML(patient)
if err != nil {
    log.Fatal(err)
}

// Unmarshal back
resource, err := r4.UnmarshalResourceXML(xmlBytes)
if err != nil {
    log.Fatal(err)
}

decoded := resource.(*r4.Patient)
fmt.Println(*decoded.Id) // same as original
```

{{< callout type="info" >}}
La serialización XML usa el mismo registro de recursos que la deserialización JSON. El nombre del elemento raíz en el XML corresponde al campo `resourceType` en JSON. Todos los tipos de recursos registrados en `resourceFactories` son compatibles con viajes de ida y vuelta XML.
{{< /callout >}}

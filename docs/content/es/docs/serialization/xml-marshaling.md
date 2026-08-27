---
title: "Serialización XML"
linkTitle: "Serialización XML"
description: "Serialización y deserialización XML de FHIR con manejo de namespaces y codificación de primitivos."
weight: 2
---

{{< callout type="info" >}}
**Los 3653 ejemplos XML publicados ahora sobreviven un round-trip.** Frente a 74 de ellos en la v1.6.0.

Un solo defecto explicaba casi toda la diferencia: `Narrative.Div` se emitía dentro de un elemento `<rawInner>` —el nombre de un tipo interno de Go filtrándose al documento— y, como el decodificador discrimina por el nombre del elemento, nunca coincidía al volver a leer y la narrativa **se descartaba sin ningún error**.

| | ejemplos | round-trip |
|---|---:|---:|
| XML r4 | 1138 | **100 %** |
| XML r4b | 1156 | **100 %** |
| XML r5 | 1359 | **100 %** |
| JSON, las tres | 8758 | 99,5 % |

Verificado como funcional: primitivos como atributos `value=`, orden de elementos según la StructureDefinition, el namespace FHIR en el elemento raíz, recursos `contained`, choice types, `Element.id` como atributo, elementos con valor de recurso como `Bundle.entry.resource`, la precisión de los decimales y la narrativa XHTML.

Quedan dos cosas que conviene saber:

| Limitación | Consecuencia |
|---|---|
| El namespace no se valida al leer | Un documento en cualquier namespace se parsea como FHIR |
| Extensiones sobre primitivos dentro de backbone elements | No son representables, así que se descartan — esto también afecta a JSON |

{{< /callout >}}

{{< callout type="warning" >}}
**Qué mide ese «100 %».** La suite de XML comprueba que nuestra propia salida sea estable: parsear, serializar, parsear, serializar, converger. **No** compara contra los bytes publicados, lo que exigiría una comparación canónica de XML que aquí todavía no existe. Un documento que perdiera siempre el mismo campo convergería igual de bien y pasaría.

Ese hueco es justo la razón de comprobar la narrativa aparte y directamente contra el archivo fuente: para cada uno de los 3572 ejemplos que la llevan, se extrae el texto legible del documento publicado y del que escribimos tras dos ciclos completos, y se exige que coincidan. La pérdida de *otro* contenido respecto al documento original sigue **sin medir** en el camino XML. El JSON sí se compara en ambos sentidos.
{{< /callout >}}


La biblioteca `gofhir/models` serializa XML de FHIR a través de funciones auxiliares dedicadas definidas en `xml_helpers.go`. Cada struct de recurso implementa `MarshalXML` y `UnmarshalXML` del paquete `encoding/xml` de Go, y las funciones de nivel superior manejan la declaración XML, el namespace de FHIR y las convenciones de elementos auto-cerrados.

## Funciones Auxiliares XML

La biblioteca expone tres funciones principales para la serialización XML:

### MarshalResourceXML

Serializa un recurso FHIR a bytes XML con la declaración XML estándar y el namespace de FHIR.

```go
package main

import (
    "fmt"
    "log"

    "github.com/gofhir/models/r4/v2"
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

`MarshalResourceXML` escribe la declaración y luego el recurso completo en una línea:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Patient xmlns="http://hl7.org/fhir"><id value="xml-example"/><active value="true"/><gender value="male"/><birthDate value="1990-06-15"/></Patient>
```

Para la forma indentada usa `MarshalResourceXMLIndent`: ver la sección siguiente.

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

El campo `Narrative.Div` contiene XHTML que la especificación exige preservar tal cual. Se escribe como un `<div>` real, con su marcado interior intacto:

```xml
<text>
  <status value="generated"/>
  <div xmlns="http://www.w3.org/1999/xhtml"><p>John Smith</p></div>
</text>
```

El namespace XHTML se añade si tu marcado lo omite, ya que un `div` sin él no es conforme. Todo lo que está dentro del `div` se traslada byte a byte —incluidos elementos vacíos como `<a href="q"></a>`, que la reescritura a etiquetas autocerradas aplicada al resto del documento deja deliberadamente en paz.

{{< callout type="info" >}}
**Arreglado en la v1.7.0.** Antes, la narrativa salía dentro de un elemento `<rawInner>` —el nombre de un tipo interno de Go filtrándose al documento— y, como el decodificador discrimina por el nombre del elemento, nunca coincidía al releer y la narrativa se descartaba sin devolver ningún error. La especificación FHIR considera `text.div` la forma legible autoritativa de un recurso cuando el receptor no puede procesar los datos estructurados, así que aquello era una pérdida de significado clínico, no un detalle de formato.

El XHTML malformado también se rechaza ahora. Antes se escribía en el documento sin comprobar, así que un `div` roto producía una salida que no estaba bien formada mientras `MarshalResourceXML` informaba de éxito:

```go
patient.Text = &r4.Narrative{Div: r4.Ptr(`<div><p>sin cerrar</div>`)}
_, err := r4.MarshalResourceXML(patient)
// err: div is not well-formed XML: XML syntax error on line 1: element <p> closed by </div>
```

{{< /callout >}}

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

Medido sobre los corpus de ejemplos publicados:

| | ejemplos | sobreviven el round-trip |
|---|---:|---:|
| XML r4 | 1138 | **100 %** |
| XML r4b | 1156 | **100 %** |
| XML r5 | 1359 | **100 %** |
| JSON, las tres versiones | 8758 | 99,5 % |

La suite de conformidad en `conformance/` lleva la cuenta archivo por archivo, así que estas cifras son medidas, no estimadas.

{{< callout type="warning" >}}
**Lee las cifras de XML por lo que son.** La suite comprueba que nuestra propia salida sea estable —parsear, serializar, parsear, serializar, converger— porque comparar contra los bytes publicados exige una comparación canónica de XML que aquí todavía no existe. Un documento que perdiera siempre el mismo campo convergería igual de limpiamente, que es exactamente cómo el defecto de la narrativa pasó tanto tiempo sin detectarse.

Por eso la narrativa se comprueba aparte y directamente contra el origen: para cada uno de los 3572 ejemplos que la llevan, se extrae el texto legible del archivo publicado y del nuestro tras dos ciclos completos, y se exige que coincidan. La pérdida de *otro* contenido respecto al documento de origen sigue sin medir en el camino XML. El JSON se compara en ambos sentidos.
{{< /callout >}}

Un recurso con narrativa hace round-trip con sus datos intactos, narrativa incluida.

Una diferencia a tener en cuenta: `UnmarshalResourceXML` deja vacío el campo
`ResourceType`, así que el struct decodificado no es idéntico al original. Usa
`GetResourceType()`, que devuelve el valor correcto de todas formas, y ten en
cuenta que la serialización no se ve afectada porque `MarshalJSON` rellena el
campo:

```go
plain := &r4.Patient{
    ResourceType: "Patient",
    Id:           ptrTo("round-trip"),
    Active:       ptrTo(true),
    Gender:       ptrTo(r4.AdministrativeGenderMale),
}

xmlBytes, err := r4.MarshalResourceXML(plain)
if err != nil {
    log.Fatal(err)
}

resource, err := r4.UnmarshalResourceXML(xmlBytes)
if err != nil {
    log.Fatal(err)
}

decoded := resource.(*r4.Patient)
fmt.Println(*decoded.Id)     // "round-trip"
fmt.Println(*decoded.Active) // true
```

{{< callout type="info" >}}
La serialización XML usa el mismo registro de recursos que la deserialización JSON, así que el nombre del elemento raíz corresponde al campo `resourceType` de JSON. Todo tipo de recurso registrado se puede serializar y deserializar, pero consulta los avisos de arriba para saber qué no sobrevive el viaje, y ten en cuenta que el struct decodificado queda con el campo `ResourceType` vacío.
{{< /callout >}}

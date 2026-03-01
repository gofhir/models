---
title: "XML Marshaling"
linkTitle: "XML Marshaling"
description: "FHIR XML serialization and deserialization with namespace handling and primitive encoding."
weight: 2
---

The `gofhir/models` library provides full support for FHIR XML serialization through dedicated helper functions defined in `xml_helpers.go`. Every resource struct implements `MarshalXML` and `UnmarshalXML` from Go's `encoding/xml` package, and top-level functions handle the XML declaration, FHIR namespace, and self-closing element conventions.

## XML Helper Functions

The library exposes three primary functions for XML serialization:

### MarshalResourceXML

Serializes a FHIR resource to XML bytes with the standard XML declaration header and the FHIR namespace.

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

The output includes the XML declaration and the FHIR namespace:

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

Produces the same output as `MarshalResourceXML` but with custom indentation for human-readable output:

```go
data, err := r4.MarshalResourceXMLIndent(patient, "", "  ")
if err != nil {
    log.Fatal(err)
}
fmt.Println(string(data))
```

The `prefix` and `indent` parameters work the same way as in `xml.Encoder.Indent()`.

### UnmarshalResourceXML

Deserializes FHIR XML bytes to the correct resource type. It reads the root element name to determine the resource type, creates the appropriate struct via the resource registry, and calls `UnmarshalXML`:

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

## FHIR Namespace Handling

The FHIR specification requires that XML representations use the namespace `http://hl7.org/fhir`. The `MarshalResourceXML` and `MarshalResourceXMLIndent` functions automatically add this namespace to the root element:

```xml
<Patient xmlns="http://hl7.org/fhir">
```

During deserialization, `UnmarshalResourceXML` determines the resource type from the local name of the root element, regardless of namespace prefix.

## Primitive Element Encoding

FHIR XML encodes primitive values (strings, booleans, integers, decimals, dates) as XML attributes rather than as element content. The value is placed in a `value` attribute on the element:

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

This differs from typical XML where values are element text content. The library handles this automatically through internal helper functions like `xmlEncodePrimitiveString`, `xmlEncodePrimitiveBool`, `xmlEncodePrimitiveInt`, `xmlEncodePrimitiveDecimal`, and `xmlEncodePrimitiveCode`.

## Self-Closing Elements

The FHIR specification uses self-closing elements for primitives without children: `<id value="123"/>` rather than `<id value="123"></id>`. The library post-processes the XML output to collapse empty elements into self-closing form using the `collapseEmptyElements` function.

## Complex Type Encoding

Complex types (such as `HumanName`, `CodeableConcept`, `Reference`) are encoded as nested XML elements with their child elements:

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

Note that repeating elements (like multiple `given` names) each appear as separate XML elements with the same tag name, following the FHIR XML convention.

## Contained Resources in XML

Contained resources are wrapped in a `<contained>` element, with the resource type as a nested element:

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

The library handles this through the `xmlEncodeContainedResource` and `xmlDecodeContainedResource` helper functions.

## XHTML Narrative in XML

The `Narrative.Div` field contains XHTML content that must be preserved verbatim in the XML output. The library uses `xmlEncodeRawXHTML` to inject the raw XHTML content directly into the XML stream without re-encoding:

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

## Round-Trip XML Fidelity

The library supports XML round-trip serialization. You can marshal a resource to XML, then unmarshal it back, and the resulting struct will contain the same data:

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
XML serialization uses the same resource registry as JSON deserialization. The root element name in the XML corresponds to the `resourceType` field in JSON. All resource types registered in `resourceFactories` are supported for XML round-trips.
{{< /callout >}}

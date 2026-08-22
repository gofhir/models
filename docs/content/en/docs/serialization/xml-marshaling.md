---
title: "XML Marshaling"
linkTitle: "XML Marshaling"
description: "FHIR XML serialization and deserialization with namespace handling and primitive encoding."
weight: 2
---

{{< callout type="warning" >}}
**XML support is experimental. Do not rely on it for data you cannot afford to lose.**

Round-tripping the official FHIR example corpora through this library, **3579 of 3653 XML files (98%) do not survive**. Over the same three versions, 8683 of 8758 JSON files (99.1%) do. One defect accounts for nearly all of it: `Narrative.Div` is emitted inside a spurious `<rawInner>` element and is then **silently discarded** when the document is read back. No error is returned.

Verified as working: primitives as `value=` attributes, element ordering per the StructureDefinition, the FHIR namespace on the root element, `contained` resources, choice types, `Element.id` as an attribute, resource-valued elements such as `Bundle.entry.resource`, and decimal precision.

Verified as broken today:

| Defect | Consequence |
|---|---|
| `Narrative.Div` wrapped in `<rawInner>` | Non-conformant output; the narrative is lost on re-parse, with no error |
| Namespace is not validated on input | A document in any namespace parses as FHIR |
| `Narrative.Div` is written unvalidated | A malformed `div` produces XML that is not well-formed, and `MarshalResourceXML` returns `nil` error |
| Extensions on primitives inside backbone elements | Not representable, so they are dropped — this affects JSON too |
| Post-processing by `collapseEmptyElements` | A regex rewrites the user's XHTML: an empty element **with attributes** is collapsed, so `<a href="q"></a>` becomes `<a href="q"/>` |

JSON is the supported serialization. Use XML for interchange with a system that requires it, and treat the narrative as unsafe until these are fixed.
{{< /callout >}}

The `gofhir/models` library serializes FHIR XML through dedicated helper functions defined in `xml_helpers.go`. Every resource struct implements `MarshalXML` and `UnmarshalXML` from Go's `encoding/xml` package, and top-level functions handle the XML declaration, FHIR namespace, and self-closing element conventions.

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

The output includes the XML declaration and the FHIR namespace (`MarshalResourceXML` emits a single line; the indented form below comes from `MarshalResourceXMLIndent`.)

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

{{< callout type="warning" >}}
**This does not currently work.** The narrative is emitted inside a `<rawInner>` element — the name of an internal Go type leaking into the document — and is discarded when the XML is read back, without an error:

```xml
<text>
  <status value="generated"/>
  <rawInner><div xmlns="http://www.w3.org/1999/xhtml"><p>John Smith</p></div></rawInner>
</text>
```

No other FHIR implementation will accept that, and a round-trip through this library loses the narrative entirely. The FHIR specification treats `text.div` as the authoritative human-readable form of a resource when the receiver cannot process the structured data, so this is a loss of clinical meaning, not a formatting quirk.

The cause is that `xml.Encoder` offers no way to write raw bytes, so writing the XHTML verbatim requires replacing the encoder — tracked as the XML writer task in the remediation plan. Until then, if you need the narrative to survive, use JSON.
{{< /callout >}}

The `Narrative.Div` field contains XHTML that must be preserved verbatim in the XML output. The library attempts this through `xmlEncodeRawXHTML`:

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

{{< callout type="warning" >}}
**XML round-trip fidelity does not hold.** Measured over the published example corpora:

| | examples | surviving a round-trip |
|---|---:|---:|
| r4 JSON | 2912 | 99.9% |
| r4 XML | 1138 | **0.7%** |
| r4b XML | 1156 | 3.1% |
| r5 XML | 1359 | 2.2% |

Practically every published example carries a narrative, and the narrative does not survive (see above). A resource with no narrative and no primitive extensions inside backbone elements does round-trip correctly.

The conformance suite in `conformance/` tracks this file by file, so these numbers are measured rather than estimated. One caveat on what they measure: for XML the suite checks only that our own output is stable (parse, serialize, parse, serialize, converge), because comparing against the published bytes needs a canonical XML comparison that does not exist yet. Loss against the source document is therefore **unmeasured** on the XML path — the real figure can only be worse than the one above. JSON is checked both ways.
{{< /callout >}}

A resource without a narrative round-trips with its data intact. Note that this
example builds its own resource rather than reusing the `patient` from the
narrative section above — that one carries a `Text.Div`, and it would come back
nil.

One difference to be aware of: `UnmarshalResourceXML` leaves the `ResourceType`
field empty, so the decoded struct is not byte-identical to the original. Use
`GetResourceType()`, which returns the correct value regardless, and note that
serialization is unaffected because `MarshalJSON` supplies the field:

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
XML serialization uses the same resource registry as JSON deserialization, so the root element name corresponds to the `resourceType` field in JSON. Every registered resource type can be marshalled and unmarshalled — but see the warnings above for what does not survive the trip, and note that the decoded struct has an empty `ResourceType` field.
{{< /callout >}}

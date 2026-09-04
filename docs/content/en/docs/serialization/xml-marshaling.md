---
title: "XML Marshaling"
linkTitle: "XML Marshaling"
description: "FHIR XML serialization and deserialization with namespace handling and primitive encoding."
weight: 2
---

{{< callout type="info" >}}
**All 3653 published XML examples now round-trip.** Up from 74 of them in v1.6.0.

One defect accounted for nearly all of the difference: `Narrative.Div` was emitted inside a `<rawInner>` element — the name of an internal Go type leaking into the document — and because the decoder switches on the element name, it never matched on the way back and the narrative was **discarded without an error**.

| | examples | round-trip |
|---|---:|---:|
| r4 XML | 1138 | **100%** |
| r4b XML | 1156 | **100%** |
| r5 XML | 1359 | **100%** |
| JSON, all three | 8758 | 99.9% |

Verified as working: primitives as `value=` attributes, element ordering per the StructureDefinition, the FHIR namespace on the root element, `contained` resources, choice types, `Element.id` as an attribute, resource-valued elements such as `Bundle.entry.resource`, decimal precision, and the XHTML narrative.

Two things are still worth knowing:

| Limitation | Consequence |
|---|---|
| The namespace is not validated on input | A document in any namespace parses as FHIR |
| Extensions on primitives inside backbone elements | Not representable, so they are dropped — this affects JSON too |

{{< /callout >}}

{{< callout type="warning" >}}
**What "100%" measures.** The XML suite checks that our own output is stable: parse, serialize, parse, serialize, converge. It does **not** compare against the published bytes, which would need a canonical XML comparison that does not exist here yet. A document that lost the same field on every pass would converge perfectly and still pass.

That gap is why the narrative is checked separately and directly against the source file: for each of the 3572 examples that carry one, the readable text is extracted from the published document and from what we write back after two full cycles, and required to match. Loss of *other* content against the source remains unmeasured on the XML path. JSON is checked both ways.
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

    "github.com/gofhir/models/r4/v2"
)

func ptrTo[T any](v T) *T {
    return &v
}

func main() {
    patient := &r4.Patient{
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

`MarshalResourceXML` writes the declaration, then the whole resource on one line:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<Patient xmlns="http://hl7.org/fhir"><id value="xml-example"/><active value="true"/><gender value="male"/><birthDate value="1990-06-15"/></Patient>
```

For the indented form, use `MarshalResourceXMLIndent` — see the next section.

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

The `Narrative.Div` field contains XHTML that the specification requires to be preserved verbatim. It is written as a real `<div>`, with its inner markup carried through unchanged:

```xml
<text>
  <status value="generated"/>
  <div xmlns="http://www.w3.org/1999/xhtml"><p>John Smith</p></div>
</text>
```

The XHTML namespace is supplied if your markup omits it, since a `div` without it is not conformant. Everything inside the `div` is passed through byte for byte — including empty elements such as `<a href="q"></a>`, which the self-closing rewrite applied to the rest of the document deliberately leaves alone.

{{< callout type="info" >}}
**Fixed in v1.7.0.** Before that, the narrative went out inside a `<rawInner>` element — the name of an internal Go type leaking into the document — and since the decoder switches on the element name, it never matched on re-read and the narrative was dropped with no error returned. The FHIR specification treats `text.div` as the authoritative human-readable form when a receiver cannot process the structured data, so that was a loss of clinical meaning rather than a formatting quirk.

Malformed XHTML is now rejected, too. It used to be written into the document unchecked, so a broken `div` produced output that was not well-formed XML while `MarshalResourceXML` reported success:

```go
patient.Text = &r4.Narrative{Div: r4.Ptr(`<div><p>unclosed</div>`)}
_, err := r4.MarshalResourceXML(patient)
// err: div is not well-formed XML: XML syntax error on line 1: element <p> closed by </div>
```

{{< /callout >}}

```go
patient := &r4.Patient{
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

Measured over the published example corpora:

| | examples | surviving a round-trip |
|---|---:|---:|
| r4 XML | 1138 | **100%** |
| r4b XML | 1156 | **100%** |
| r5 XML | 1359 | **100%** |
| JSON, all three versions | 8758 | 99.9% |

The conformance suite in `conformance/` tracks this file by file, so these numbers are measured rather than estimated.

{{< callout type="warning" >}}
**Read the XML figures for what they are.** The suite checks that our own output is stable — parse, serialize, parse, serialize, converge — because comparing against the published bytes needs a canonical XML comparison that does not exist here yet. A document that dropped the same field on every pass would converge just as cleanly, which is exactly how the narrative defect survived undetected for so long.

The narrative is therefore checked separately and directly against the source: for each of the 3572 examples that carry one, the readable text is extracted from the published file and from our output after two full cycles, and required to match. Loss of *other* content relative to the source document remains unmeasured on the XML path. JSON is compared both ways.
{{< /callout >}}

A resource carrying a narrative round-trips with its data intact, narrative included.

The decoded struct is equal to the original: `resourceType` is a zero-size marker
carried by the Go type rather than a string the decoder has to fill in, so there
is no field left inconsistent. Read it with `GetResourceType()`:

```go
plain := &r4.Patient{
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
XML serialization uses the same resource registry as JSON deserialization, so the root element name corresponds to the `resourceType` field in JSON. Every registered resource type can be marshalled and unmarshalled. Note that the decoded struct has an empty `ResourceType` field — use `GetResourceType()`.
{{< /callout >}}

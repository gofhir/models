---
title: "v1 to v2"
linkTitle: "v1 to v2"
description: "What changes in v2, what to do about it, and what is still being decided."
weight: 1
---

{{< callout type="info" >}}
**v2 is not released yet.** This page exists so that the deprecation warnings in v1.6.0 point somewhere real. The section on functional options is final and safe to act on today. The rest is a forecast, marked as such, and will be completed when v2 is cut.
{{< /callout >}}

## Start here: turn on the warnings

v1.6.0 marks every API that v2 removes with Go's `Deprecated:` convention. You do not need to read this page to find your call sites — your tooling will list them:

```shell
go vet ./...              # does not report deprecations
staticcheck ./...         # does: SA1019
golangci-lint run ./...   # does, via staticcheck
```

Editors using `gopls` show the same symbols struck through.

Each warning names its own replacement, so most of the migration is mechanical:

```text
main.go:11:3: r4.WithPatientActive is deprecated: use PatientBuilder.SetActive
              instead; removed in v2. (SA1019)
```

If you see no warnings, you are not using anything that v2 removes.

## Functional options become builders

**Status: final.** This is the largest change by volume — 11,952 functions across the three modules — and the most mechanical.

Every `With<Resource><Field>` option has a builder method with the same parameter type and the same effect. The correspondence was verified pair by pair, not assumed: a test parses the generated code and compares the parameter type and the assignment for all 11,952.

| v1 | v2 |
|---|---|
| `NewPatient(opts...)` | `NewPatientBuilder()` … `.Build()` |
| `WithPatientActive(true)` | `.SetActive(true)` |
| `WithPatientIdentifier(id)` | `.AddIdentifier(id)` |
| `PatientOption` (type) | removed — the builder needs no option type |

The rule for the method name: **`Add` for repeating fields, `Set` for everything else.** That mirrors what the option already did — repeating fields appended, single fields assigned.

Before:

```go
p := r4.NewPatient(
    r4.WithPatientId("p1"),
    r4.WithPatientActive(true),
    r4.WithPatientIdentifier(r4.Identifier{System: r4.Ptr("urn:mrn")}),
    r4.WithPatientName(r4.HumanName{Family: r4.Ptr("Smith")}),
)
```

After:

```go
p := r4.NewPatientBuilder().
    SetId("p1").
    SetActive(true).
    AddIdentifier(r4.Identifier{System: r4.Ptr("urn:mrn")}).
    AddName(r4.HumanName{Family: r4.Ptr("Smith")}).
    Build()
```

Struct literals are unaffected and remain the shortest form for a resource you have entirely in hand:

```go
p := &r4.Patient{ResourceType: "Patient", Id: r4.Ptr("p1"), Active: r4.Ptr(true)}
```

### If you pass options around

An option value is a `func(*Patient)`, and code that stores or passes them needs restructuring rather than renaming:

```go
// v1
func configure() []r4.PatientOption {
    return []r4.PatientOption{r4.WithPatientActive(true)}
}

// v2 — take the builder instead
func configure(b *r4.PatientBuilder) *r4.PatientBuilder {
    return b.SetActive(true)
}
```

This is the only part of the option migration that is not a rename. If your code does not name `PatientOption` anywhere, you will not hit it.

### Finding your call sites

There is no codemod, and that is deliberate rather than unfinished: the rewrite is structural, not a rename. `NewPatient(a, b, c)` has to become a method chain terminated by `Build()`, and a regex that renamed `WithPatientActive` to `SetActive` in place would leave code that does not compile. The mechanical part is choosing the new name, and the deprecation message already did that for every call site.

To get the whole list, each site paired with its exact replacement:

```shell
staticcheck ./... 2>&1 \
  | sed -nE 's/^(.+): ([A-Za-z0-9_.]+) is deprecated: use ([A-Za-z0-9_.]+) instead.*/\1\n      \2  ->  \3/p'
```

```text
main.go:23:3
      r4.WithPatientId  ->  PatientBuilder.SetId
main.go:24:3
      r4.WithPatientActive  ->  PatientBuilder.SetActive
main.go:25:3
      r4.WithPatientIdentifier  ->  PatientBuilder.AddIdentifier
```

Or, to size the job before starting it:

```shell
staticcheck ./... 2>&1 | grep -c SA1019
```

## Type changes

None of these can be flagged by a deprecation warning, because the fields keep their names and only change type — which is precisely why they are worth reading in advance. Each says whether it has landed or is still a forecast.

### `ResourceType` stops being a `string`

**Status: done, and the largest single change for most code.** `ResourceType` is now a zero-size marker type, so assigning it does not compile:

```go
p := &r4.Patient{ResourceType: "Patient", Id: r4.Ptr("p1")}  // v1
p := &r4.Patient{Id: r4.Ptr("p1")}                           // v2 — same output
```

Deleting the line is the whole migration. The output is byte-identical, `resourceType` included and still first.

To read it, use `GetResourceType()`, which works in v1 too — so code that already calls the method needs no change at all:

```go
p.GetResourceType()   // "Patient", both versions
```

Two corrections to what was forecast here earlier, now that it is measured:

- **The JSON key order does not change.** It was expected to; it does not.
- **`json.Marshal` was already not equivalent to `r4.Marshal`,** and still is not. The removed method called `SetEscapeHTML(false)`, but that never took effect: `encoding/json` re-compacts a `MarshalJSON` result using the outer encoder's setting. `Marshal` is what preserves `<` in narrative XHTML — that is unchanged, and it is why the docs tell you to prefer it.

What you gain: marshaling a 50-entry Bundle is **3.4× faster and allocates 4.7× less** (145850 → 43200 ns/op, 71900 → 15216 B/op), because the removed method built a second `bytes.Buffer` and `json.Encoder` for every resource.

And one silent bug is fixed. A method promoted from an embedded struct satisfies `json.Marshaler` for the outer type, so this used to serialize as a bare Patient, dropping `tenant` and `score` with no error:

```go
type WithMetadata struct {
    r4.Patient
    Tenant string `json:"tenant"`
    Score  int    `json:"score"`
}
```

### Repeating primitives become pointer slices

**Status: done.** `[]string` becomes `[]*string`, so that a `null` in the middle of a FHIR array survives:

```json
"given": ["A", null, "C"]
```

Today that middle `null` is read as `""` — an empty value that is clinically different from an absent one. Reading such an array back out re-emits invented data. The pointer slice is what makes absence representable.

`PtrSlice` converts a list of values into the pointer slice, and `Vals` converts back:

```go
Given: r4.PtrSlice("John", "Q")   // []*string
names := r4.Vals(hn.Given)        // []string, nil entries become ""
```

### R5 `integer64` travels as a string

**Status: done.** R5's `integer64` was generated as `int64` and serialized as a bare JSON number. The specification requires a string, because JSON numbers carry only 53 bits of integer precision and `integer64` is 64 — values above 2^53 were silently rounded.

The field type becomes `Integer64`, which marshals to a string and unmarshals from either form, so documents written by older versions still read:

```go
var att r5.Attachment
json.Unmarshal([]byte(`{"size":"9007199254740993"}`), &att)

att.Size.Int64()   // 9007199254740993 — exact, above 2^53
att.Size.String()  // "9007199254740993"
```

Reading still accepts the old bare-number form, so documents written by earlier versions are not rejected.

### Required complex fields become pointers

**Status: done.** 245 fields in r4, 250 in r4b and 315 in r5 were value structs, so they serialized even when nothing had been set:

```go
r4.Observation{Id: r4.Ptr("o1")}
// v1: {"resourceType":"Observation","id":"o1","code":{}}   ← invalid per ele-1
// v2: {"resourceType":"Observation","id":"o1"}
```

The reasoning for value fields was that the type should express the obligation. It could not — Go has no way to force a field to be set, so the only thing the value type achieved was emitting an empty object instead of nothing. Requiredness is a FHIR validator's job; what the type has to be able to say is "absent".

`Extension.url` went the same way, so an unset extension is now `{}` rather than `{"url":""}`.

**Migrating: add `&` to the struct literal.**

```go
Code: r4.CodeableConcept{Coding: ...}    // v1
Code: &r4.CodeableConcept{Coding: ...}   // v2
```

Reading is unaffected — Go dereferences automatically, so `obs.Code.Coding` still compiles. What changes is that it can now be nil, so guard it where the value might not have been set.

**Builder call sites do not change at all.** `SetCode` still takes a `CodeableConcept` by value and takes its address internally:

```go
r4.NewObservationBuilder().SetCode(r4.CodeableConcept{...}).Build()   // unchanged
```

### r4 helpers become functions

**Status: done.** `helpers.BodyWeight` and the other 58 LOINC and category constants were package-level vars; they are now functions returning a fresh `*r4.CodeableConcept`, and the 33 UCUM `Quantity*` helpers return `*r4.Quantity`.

```go
Code: helpers.BodyWeight,      // v1
Code: helpers.BodyWeight(),    // v2 — and the field is a pointer now, so this fits directly
```

With a builder, `Set*` still takes a value, so dereference:

```go
SetCode(*helpers.BodyWeight()).
AddCategory(*helpers.ObservationCategoryVitalSigns()).
```

This is a correctness fix, not tidying. A var handed the same value to every caller, and since `CodeableConcept` carries a `Coding` slice, even copying the struct shared the backing array — so adjusting a `display` on one resource changed it on every other resource built from that helper, and on the helper itself. It was already possible in v1; making required fields pointers widened it from the slices to the whole struct.

### `Contained` becomes a named slice type

**Status: forecast.** `[]Resource` becomes `ContainedList`. Assignment from `[]Resource`, `range`, `append`, and passing to a `func([]Resource)` all keep working; what changes is that `%T` prints `r4.ContainedList`.

### Code system type names change

**Status: forecast, and the least certain of these.** Enum type and constant names are expected to be derived from the FHIR `bindingName` extension, which renames roughly 657 types and 3,613 constants. A full old-to-new mapping table will accompany the release. This is the change most likely to be split out or staged, since the volume is large and the benefit is naming consistency rather than correctness.

## The import path

v2 modules carry the `/v2` suffix Go requires:

```shell
go get github.com/gofhir/models/r4/v2
```

```go
import "github.com/gofhir/models/r4/v2"

var p r4.Patient   // the package is still named r4
```

There is no `v2` directory in the repository — the suffix belongs to the module path, not the layout. Because the path changed, **v1 and v2 can coexist in one build**, which is what makes an incremental migration possible: move one package at a time while a dependency still pulls in v1.

## Staying on v1

v1 keeps receiving the deprecation warnings and any fix that does not require a breaking change. It is a valid place to sit while you migrate; the warnings are advisory and change no behavior.

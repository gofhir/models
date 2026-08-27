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

## Everything below is a forecast

**Status: not final.** These changes are planned for v2 and are described here so you can judge the impact early. Details, and in some cases whether they ship at all, may change. None of them can be flagged by a deprecation warning, because the fields keep their names and only change type — which is precisely why they are worth reading in advance.

### `ResourceType` stops being a `string`

To stop paying for a per-resource `MarshalJSON`, `ResourceType` is expected to become a zero-size marker type. Reading it as a string, or setting it in a struct literal, would no longer compile.

`GetResourceType()` keeps working and returns a `string`. **If you use the method rather than the field, this change costs you nothing** — that is the migration, and it works today in v1.

Two consequences worth checking now:

- **JSON key order changes.** If you compare payloads byte for byte, or hash them, that comparison will move.
- `json.Marshal(patient)` stops being equivalent to `r4.Marshal(patient)`.

### Repeating primitives become pointer slices

`[]string` becomes `[]*string`, so that a `null` in the middle of a FHIR array survives:

```json
"given": ["A", null, "C"]
```

Today that middle `null` is read as `""` — an empty value that is clinically different from an absent one. Reading such an array back out re-emits invented data. The pointer slice is what makes absence representable.

Conversion helpers (`PtrSlice` to go one way, `Vals` to come back) are expected to ship with the change. They are not in v1.6.0 — `Ptr`, `Val` and `First` are the helpers available today, and they cover scalars only.

### Required complex fields become pointers

Around 99 fields per version are non-pointer structs, so they serialize even when empty:

```go
r4.Observation{Id: r4.Ptr("o1")}
// {"resourceType":"Observation","id":"o1","code":{}}   ← code:{} is invalid FHIR
```

Making them pointers fixes the output and changes those field types.

### `Contained` becomes a named slice type

`[]Resource` becomes `ContainedList`. Assignment from `[]Resource`, `range`, `append`, and passing to a `func([]Resource)` all keep working; what changes is that `%T` prints `r4.ContainedList`.

### Code system type names change

Enum type and constant names are expected to be derived from the FHIR `bindingName` extension, which renames roughly 657 types and 3,613 constants. A full old-to-new mapping table will accompany the release. This is the change most likely to be split out or staged, since the volume is large and the benefit is naming consistency rather than correctness.

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

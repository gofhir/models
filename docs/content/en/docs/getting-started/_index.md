---
title: "Getting Started"
linkTitle: "Getting Started"
description: "Install gofhir/models, create your first FHIR resource, and learn about supported FHIR versions."
weight: 1
---

This section walks you through installing the library, building your first FHIR resource in Go, and understanding how the three supported FHIR versions are organized.

## Overview

**gofhir/models** provides type-safe Go structs for all FHIR resources, data types, and code systems. Each FHIR version (R4, R4B, R5) is published as an independent Go module, so you only import what you need.

The library is designed around three core principles:

1. **Type safety** -- Every FHIR field maps to a strongly typed Go struct field with proper pointer semantics for optional values.
2. **Multiple construction patterns** -- Choose between struct literals, fluent builders, or functional options depending on your coding style.
3. **FHIR-conformant serialization** -- JSON and XML marshaling follows the FHIR specification exactly, including narrative XHTML preservation and decimal precision.

## Guides

{{< cards >}}
  {{< card link="installation" title="Installation" subtitle="Install the Go module for your target FHIR version and configure your project." icon="download" >}}
  {{< card link="quick-start" title="Quick Start" subtitle="Create, serialize, and deserialize FHIR resources with working code examples." icon="play" >}}
  {{< card link="fhir-versions" title="FHIR Versions" subtitle="Understand the differences between R4, R4B, and R5 and how they are packaged." icon="book-open" >}}
{{< /cards >}}

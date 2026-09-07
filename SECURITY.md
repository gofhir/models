# Security Policy

## Supported versions

| Version | Branch | What it receives |
|---|---|---|
| **2.x** | `main` | fixes and features |
| 1.x | `maintenance-v1` | security fixes only |

The `/v2` modules are the current line. The v1 line stays open for security work
because the v2 import path changed, so upgrading is not always immediate.

## Reporting a vulnerability

**Use [private vulnerability reporting](https://github.com/gofhir/models/security/advisories/new)**
rather than a public issue. That opens a draft advisory only the maintainers can
see, and it is the right channel even if you are unsure whether what you found is
exploitable.

Useful in a report:

- the FHIR version and module version
- the smallest document or call that triggers it
- what you observed — a panic, a hang, memory growth, wrong data accepted

## What this library is responsible for

It parses documents from sources the caller may not control, which is where the
risk lives:

- **Resource exhaustion.** Deserializing nested resources re-reads their subtrees,
  so cost grows with size times nesting depth. `MaxResourceDepth` bounds that; the
  scan is a single linear pass and rejects a hostile document in microseconds.
- **Accepting what should be refused.** A document that decodes into the wrong
  type, or one whose declared type contradicts the target, is rejected rather than
  silently reinterpreted.

It is **not** a validator. The structs will hold a document FHIR would reject, and
that is by design — conformance is [`gofhir/validator`](https://github.com/gofhir/validator)'s
question. Do not treat successful decoding as evidence that a document is safe or
valid for your use.

## Past issues

**Depth limit bypassable by key casing** — fixed in `v1.7.1` and present in v2
from the start. `encoding/json` matches object keys case-insensitively, so
`{"RESOURCETYPE":"Patient"}` decoded as a Patient while the depth scan, comparing
bytes exactly, did not count it as a resource. A document nesting those was
accepted at any depth: 55 KB of nesting cost 466 ms of CPU and 179 MB of heap on
one core, and the cost is quadratic in depth.

If you are on a v1 release below 1.7.1, upgrade. Versions before 1.5.0 have no
depth limit at all and are affected for a different reason.

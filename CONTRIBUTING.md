# Contributing

## The Go files are generated — do not edit them

`r4/`, `r4b/` and `r5/` are output. A change made there disappears the next time
anyone regenerates, and CI checks for exactly that. The code lives in:

| Where | What |
|---|---|
| `internal/codegen/parser` | reads the FHIR StructureDefinition and ValueSet bundles |
| `internal/codegen/analyzer` | turns them into the types and properties to emit |
| `internal/codegen/generator/templates` | the templates that write the Go |

The hand-written files in the version directories are the tests (`*_test.go`).
Those you can edit.

## Getting set up

```bash
scripts/fetch-specs.sh        # the FHIR specification bundles (~143 MB, gitignored)
go run cmd/generator/main.go r4      # regenerate one version: r4, r4b or r5
```

The generator is deliberately strict: a missing spec file or an unresolved
ValueSet name collision fails the build rather than producing a package that is
quietly missing types.

## Running the checks

```bash
go test github.com/gofhir/models/internal/...   # parser, analyzer, generator, golden
go test -C r4 ./...                             # one version's own tests
```

The `internal/` packages are addressed by import path, not `./internal/...`. They
belong to the root module, which is not one of the workspace's modules, so a
relative pattern reports them as outside the module roots.

The conformance suite is a separate module and needs the corpus:

```bash
scripts/fetch-examples.sh                       # every example HL7 publishes
cd conformance && GOWORK=off go test ./...      # ~90s
```

`GOWORK=off` is required there: the suite must resolve the modules the way a
consumer does, not through the workspace.

### The corpus is a ratchet, not a snapshot

`TestRoundTrip` compares against recorded known-failure lists, and it fails **in
both directions** — on a regression, and on progress that has not been recorded.
If your change fixes examples, run `GOWORK=off go test ./... -update-known` from
`conformance/` and commit the diff: every line removed is a bug fixed, and the diff is the evidence.

### Golden files

`internal/codegen/generator` compares generated output against fixtures. If your
change is meant to alter what is emitted:

```bash
go test ./internal/codegen/generator -update
```

Then read the diff. It is the clearest description of what your change does to
the output.

## Commits

Commit subjects follow [conventional commits](https://www.conventionalcommits.org/)
because release-please builds the changelog and picks the version from them:

- `feat:` → minor, `fix:` → patch, `feat!:` or a `BREAKING CHANGE:` footer → major
- `docs:`, `test:`, `chore:`, `refactor:` do not appear in the changelog

Keep conventional-commit lines out of the commit *body*. release-please parses
those too, and a subject repeated inside the body appears twice in the changelog.

## What a good change looks like

**Measure before and after.** Several tasks in this project's history turned out
to be worth less than assumed, and a few were worth more; the difference was
always a measurement. A claim in a commit message should be something you ran.

**Verify a test by breaking the code.** A test that passes when the defect is
reintroduced is not testing the defect. Reintroduce it, watch the test fail, then
put it back.

**Say what a change does not cover.** The tests in `conformance/` include several
that record a known limitation rather than guarding an invariant — with the
measurement that made it a limitation instead of a bug. That is preferable to
leaving the boundary to be discovered.

## Scope

This library is FHIR types: parsing, serializing, and constructing them.

Validation is not here — it lives in
[`gofhir/validator`](https://github.com/gofhir/validator), which works on the raw
document. Two implementations of the same question can disagree, so the split is
deliberate rather than a gap waiting to be filled.

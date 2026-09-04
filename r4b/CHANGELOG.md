# Changelog

## [2.0.0](https://github.com/gofhir/models/compare/r4b/v1.7.0...r4b/v2.0.0) (2026-09-04)


### ⚠ BREAKING CHANGES

* reject documents that were being accepted and silently misread
* **codec:** carry extensions on primitives nested in backbone elements
* **codesystems:** name enums after the binding, not the ValueSet title
* remove the functional options, as v1.6.0 said it would ([#44](https://github.com/gofhir/models/issues/44))
* **json:** close a depth-guard bypass, and stop parsing twice to dispatch ([#40](https://github.com/gofhir/models/issues/40))
* **json:** dispatch contained from one type, not 437 methods ([#39](https://github.com/gofhir/models/issues/39))
* **json:** make required complex fields pointers ([#37](https://github.com/gofhir/models/issues/37))
* **json:** emit resourceType from a zero-size marker ([#36](https://github.com/gofhir/models/issues/36))
* **json:** represent the positional null in repeating primitives ([#27](https://github.com/gofhir/models/issues/27))

### Features

* add Ptr/Val/First helpers and resolve the ValueSet name collision ([#20](https://github.com/gofhir/models/issues/20)) ([7135046](https://github.com/gofhir/models/commit/7135046b0c391a84011bc518083a9064d3fc6c5a))
* **codesystems:** name enums after the binding, not the ValueSet title ([f155454](https://github.com/gofhir/models/commit/f155454d3943bfe15704836d5948c0f697a57b5d))
* expose FHIR version from the generated FHIRPath model ([#11](https://github.com/gofhir/models/issues/11)) ([6930c5f](https://github.com/gofhir/models/commit/6930c5f406a011f40933f01c6af840a1f3888b1f)), closes [#10](https://github.com/gofhir/models/issues/10)
* expose whether a name resolves to a type in the FHIRPath model ([#15](https://github.com/gofhir/models/issues/15)) ([d98450a](https://github.com/gofhir/models/commit/d98450ae1ede486c92bc3b61f9416b978c546661)), closes [#14](https://github.com/gofhir/models/issues/14)
* generate FHIRPath model metadata for version-aware expression evaluation ([b280322](https://github.com/gofhir/models/commit/b280322d3c8fe45e76989ed569a1478125986bdb)), closes [#8](https://github.com/gofhir/models/issues/8)
* remove the functional options, as v1.6.0 said it would ([#44](https://github.com/gofhir/models/issues/44)) ([319c6c5](https://github.com/gofhir/models/commit/319c6c5edb7b799fb8f2625986b18d2d70f2c195))
* replace *float64 with custom Decimal type for FHIR decimal precision preservation ([1a53aee](https://github.com/gofhir/models/commit/1a53aeecdecba440ea488984fb8c067f70732f41)), closes [#6](https://github.com/gofhir/models/issues/6)
* **xml:** carry the narrative through instead of dropping it ([#34](https://github.com/gofhir/models/issues/34)) ([cd5f841](https://github.com/gofhir/models/commit/cd5f841a8c2baca861e7145778e5e762ef5cbfd7))


### Bug Fixes

* **codec:** carry extensions on primitives nested in backbone elements ([868b6f1](https://github.com/gofhir/models/commit/868b6f11617623f37e2c0125f62ed8c915d1bcd7))
* **json:** bound resource nesting depth to stop a remote DoS ([#19](https://github.com/gofhir/models/issues/19)) ([42b4003](https://github.com/gofhir/models/commit/42b4003a066be00a0434a998d123eed3b7dcfc18))
* **json:** close a depth-guard bypass, and stop parsing twice to dispatch ([#40](https://github.com/gofhir/models/issues/40)) ([3b3e6c3](https://github.com/gofhir/models/commit/3b3e6c382441229226752b0aba131d4b6a5e2cf9))
* **json:** make R5 Bundle.issues round-trip, and accept explicit null ([#25](https://github.com/gofhir/models/issues/25)) ([42c5fc8](https://github.com/gofhir/models/commit/42c5fc8ff52578b797819515bd5e84257d2ef7dd))
* **json:** make required complex fields pointers ([#37](https://github.com/gofhir/models/issues/37)) ([359aa66](https://github.com/gofhir/models/commit/359aa66540069df3fad65e58c08df87c3c298708))
* **json:** represent the positional null in repeating primitives ([#27](https://github.com/gofhir/models/issues/27)) ([db8d311](https://github.com/gofhir/models/commit/db8d311e58b6e87c1c077befd70c4353c7e667be))
* prevent HTML escaping in FHIR narrative text.div fields ([b543156](https://github.com/gofhir/models/commit/b543156e327ce11d92dc398f779911ba75169b41)), closes [#2](https://github.com/gofhir/models/issues/2)
* reject documents that were being accepted and silently misread ([dcd92b7](https://github.com/gofhir/models/commit/dcd92b7e0aa24b66ead42f792483b1d2bdbfc5f1))
* update all import paths to github.com/gofhir/fhir ([f97bfb1](https://github.com/gofhir/models/commit/f97bfb100ceaf780ad28202e7d87d13d25037a44))
* update all import paths to github.com/gofhir/models ([eda97a2](https://github.com/gofhir/models/commit/eda97a26f1d9db154e5488c8c00671b14d9f5395))
* **xml:** wrap resource-valued elements and keep the decoder in sync ([#21](https://github.com/gofhir/models/issues/21)) ([2ca0868](https://github.com/gofhir/models/commit/2ca0868c54f8a6ebd466794e058eac8ca43555bd))


### Performance Improvements

* build the FHIRPath model on first use ([#43](https://github.com/gofhir/models/issues/43)) ([bff2d98](https://github.com/gofhir/models/commit/bff2d98c9f54417b625c2efb8b82eb2672cfb27f))
* **json:** dispatch contained from one type, not 437 methods ([#39](https://github.com/gofhir/models/issues/39)) ([2f59314](https://github.com/gofhir/models/commit/2f593143bd2bf9f42704eed739c14509f4db21f7))
* **json:** emit resourceType from a zero-size marker ([#36](https://github.com/gofhir/models/issues/36)) ([ed5836c](https://github.com/gofhir/models/commit/ed5836c87edfe7a4338c9683f85506aad4424c39))


### Code Refactoring

* consolidate generated files and add XML deserialization ([b0b217e](https://github.com/gofhir/models/commit/b0b217e9e3990b2f075ed6df98a2f1a578ce2e29))

## [1.5.1](https://github.com/gofhir/models/compare/r4b/v1.5.0...r4b/v1.5.1) (2026-08-27)


### Bug Fixes

* **json:** make R5 Bundle.issues round-trip, and accept explicit null ([#25](https://github.com/gofhir/models/issues/25)) ([42c5fc8](https://github.com/gofhir/models/commit/42c5fc8ff52578b797819515bd5e84257d2ef7dd))

## [1.5.0](https://github.com/gofhir/models/compare/r4b/v1.4.0...r4b/v1.5.0) (2026-08-21)


### Features

* add Ptr/Val/First helpers and resolve the ValueSet name collision ([#20](https://github.com/gofhir/models/issues/20)) ([7135046](https://github.com/gofhir/models/commit/7135046b0c391a84011bc518083a9064d3fc6c5a))


### Bug Fixes

* **json:** bound resource nesting depth to stop a remote DoS ([#19](https://github.com/gofhir/models/issues/19)) ([42b4003](https://github.com/gofhir/models/commit/42b4003a066be00a0434a998d123eed3b7dcfc18))
* **xml:** wrap resource-valued elements and keep the decoder in sync ([#21](https://github.com/gofhir/models/issues/21)) ([2ca0868](https://github.com/gofhir/models/commit/2ca0868c54f8a6ebd466794e058eac8ca43555bd))

## [1.4.0](https://github.com/gofhir/models/compare/r4b/v1.3.0...r4b/v1.4.0) (2026-08-02)


### Features

* expose whether a name resolves to a type in the FHIRPath model ([#15](https://github.com/gofhir/models/issues/15)) ([d98450a](https://github.com/gofhir/models/commit/d98450ae1ede486c92bc3b61f9416b978c546661)), closes [#14](https://github.com/gofhir/models/issues/14)

## [1.3.0](https://github.com/gofhir/models/compare/r4b/v1.2.0...r4b/v1.3.0) (2026-08-01)


### Features

* expose FHIR version from the generated FHIRPath model ([#11](https://github.com/gofhir/models/issues/11)) ([6930c5f](https://github.com/gofhir/models/commit/6930c5f406a011f40933f01c6af840a1f3888b1f)), closes [#10](https://github.com/gofhir/models/issues/10)

## [1.2.0](https://github.com/gofhir/models/compare/r4b/v1.1.0...r4b/v1.2.0) (2026-03-01)


### Features

* generate FHIRPath model metadata for version-aware expression evaluation ([b280322](https://github.com/gofhir/models/commit/b280322d3c8fe45e76989ed569a1478125986bdb)), closes [#8](https://github.com/gofhir/models/issues/8)

## [1.1.0](https://github.com/gofhir/models/compare/r4b/v1.0.3...r4b/v1.1.0) (2026-02-17)


### Features

* replace *float64 with custom Decimal type for FHIR decimal precision preservation ([1a53aee](https://github.com/gofhir/models/commit/1a53aeecdecba440ea488984fb8c067f70732f41)), closes [#6](https://github.com/gofhir/models/issues/6)

## [1.0.3](https://github.com/gofhir/models/compare/r4b/v1.0.2...r4b/v1.0.3) (2026-02-16)


### Code Refactoring

* consolidate generated files and add XML deserialization ([b0b217e](https://github.com/gofhir/models/commit/b0b217e9e3990b2f075ed6df98a2f1a578ce2e29))

## [1.0.2](https://github.com/gofhir/models/compare/r4b/v1.0.1...r4b/v1.0.2) (2026-02-14)


### Bug Fixes

* prevent HTML escaping in FHIR narrative text.div fields ([b543156](https://github.com/gofhir/models/commit/b543156e327ce11d92dc398f779911ba75169b41)), closes [#2](https://github.com/gofhir/models/issues/2)
* update all import paths to github.com/gofhir/models ([eda97a2](https://github.com/gofhir/models/commit/eda97a26f1d9db154e5488c8c00671b14d9f5395))

## [1.0.1](https://github.com/gofhir/models/compare/r4b/v1.0.0...r4b/v1.0.1) (2026-01-24)


### Bug Fixes

* update all import paths to github.com/gofhir/models ([f97bfb1](https://github.com/gofhir/models/commit/f97bfb100ceaf780ad28202e7d87d13d25037a44))

## [0.2.0](https://github.com/robertoAraneda/gofhir/compare/r4b/v0.1.0...r4b/v0.2.0) (2026-01-17)


### ⚠ BREAKING CHANGES

* Package import paths have changed.

### Features

* initial release ([82ec28c](https://github.com/robertoAraneda/gofhir/commit/82ec28c30a38afb26bbf7b2503945573606da517))


### Code Refactoring

* migrate to multi-module monorepo architecture ([42ae0de](https://github.com/robertoAraneda/gofhir/commit/42ae0de8aa2f98cbe6e94fcef4736a6a0184bfb7))

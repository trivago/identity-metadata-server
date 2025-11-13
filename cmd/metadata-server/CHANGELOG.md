# Changelog

## [3.7.1](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.7.0...metadata-server-3.7.1) (2025-11-13)


### Bug Fixes

* properly react on timeouts and HTTP 429 ([#6](https://github.com/trivago/identity-metadata-server/issues/6)) ([dd9ad86](https://github.com/trivago/identity-metadata-server/commit/dd9ad868fdfd4d3c30c1a8e740e07e1d10dbba38))
* use internal HTTP call instead of the default golang one ([dd9ad86](https://github.com/trivago/identity-metadata-server/commit/dd9ad868fdfd4d3c30c1a8e740e07e1d10dbba38))

## [3.7.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.6.4...metadata-server-3.7.0) (2025-10-09)


### Features

* support json output for serviceAccounts endpoint ([#1](https://github.com/trivago/identity-metadata-server/issues/1)) ([4706e9b](https://github.com/trivago/identity-metadata-server/commit/4706e9bc86e8317a825ced05b7d63fc6a7a4fecd))


### Miscellaneous

* make unit and integration tests pass if not relevant ([4706e9b](https://github.com/trivago/identity-metadata-server/commit/4706e9bc86e8317a825ced05b7d63fc6a7a4fecd))

## [3.6.4](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.6.3...metadata-server-3.6.4) (2025-06-23)


### Bug Fixes

* duplicate metric labels ([b9c5947](https://github.com/trivago/identity-metadata-server/commit/b9c594799c161175bfee4ef086f27a4cdedb7d2c))

## [3.6.3](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.6.2...metadata-server-3.6.3) (2025-06-23)


### Bug Fixes

* address inconsistencies in metrics ([982e535](https://github.com/trivago/identity-metadata-server/commit/982e535ab4f65e377dec0dae4c70c44fd92bfd66))
* make "mode" (host|kubernetes) available in all metrics ([982e535](https://github.com/trivago/identity-metadata-server/commit/982e535ab4f65e377dec0dae4c70c44fd92bfd66))
* make "mode" (host|kubernetes) available in all metrics ([e41c649](https://github.com/trivago/identity-metadata-server/commit/e41c649adbb74e23ba9409cf8eaa93b997779ec4))
* make "node_name" available in all metrics ([982e535](https://github.com/trivago/identity-metadata-server/commit/982e535ab4f65e377dec0dae4c70c44fd92bfd66))
* make "node_name" available in all metrics ([e41c649](https://github.com/trivago/identity-metadata-server/commit/e41c649adbb74e23ba9409cf8eaa93b997779ec4))
* use the default prom-client registry so we use the default metrics all other golang libraries have ([982e535](https://github.com/trivago/identity-metadata-server/commit/982e535ab4f65e377dec0dae4c70c44fd92bfd66))
* use the default prom-client registry so we use the default metrics all other golang libraries have ([e41c649](https://github.com/trivago/identity-metadata-server/commit/e41c649adbb74e23ba9409cf8eaa93b997779ec4))
* use the same clientIP lookup as the tokenhandler when generating the "client_ip" metric label ([982e535](https://github.com/trivago/identity-metadata-server/commit/982e535ab4f65e377dec0dae4c70c44fd92bfd66))
* use the same clientIP lookup as the tokenhandler when generating the "client_ip" metric label ([e41c649](https://github.com/trivago/identity-metadata-server/commit/e41c649adbb74e23ba9409cf8eaa93b997779ec4))

## [3.6.2](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.6.1...metadata-server-3.6.2) (2025-06-23)


### Bug Fixes

* add default go runtime metrics ([#77](https://github.com/trivago/identity-metadata-server/issues/77)) ([fdc8669](https://github.com/trivago/identity-metadata-server/commit/fdc8669670d58dba896cceae289db0a11d1a578f))

## [3.6.1](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.6.0...metadata-server-3.6.1) (2025-06-20)


### Bug Fixes

* add latency buckets for 5ms and 10ms ([0ff09ac](https://github.com/trivago/identity-metadata-server/commit/0ff09ac60334c5a6ea8524a907e458b5ce7c4ad4))
* use kubernetes API for fetching pods if `kubeletPath` is set to an empty string ([0ff09ac](https://github.com/trivago/identity-metadata-server/commit/0ff09ac60334c5a6ea8524a907e458b5ce7c4ad4))

## [3.6.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.5.0...metadata-server-3.6.0) (2025-06-18)


### Features

* lookup pods using the kubelet API ([#71](https://github.com/trivago/identity-metadata-server/issues/71)) ([8f1753c](https://github.com/trivago/identity-metadata-server/commit/8f1753ce922110dab026ec439ce7b91e1b654a99))


### Bug Fixes

* dependency updates ([8f1753c](https://github.com/trivago/identity-metadata-server/commit/8f1753ce922110dab026ec439ce7b91e1b654a99))
* more functions in the request chain are now context-aware ([8f1753c](https://github.com/trivago/identity-metadata-server/commit/8f1753ce922110dab026ec439ce7b91e1b654a99))
* non-mTLS requests are now using HTTP2 and accept all configured rootCAs ([8f1753c](https://github.com/trivago/identity-metadata-server/commit/8f1753ce922110dab026ec439ce7b91e1b654a99))
* the metrics endpoint is not generating access logs anymore ([8f1753c](https://github.com/trivago/identity-metadata-server/commit/8f1753ce922110dab026ec439ce7b91e1b654a99))

## [3.5.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.4.2...metadata-server-3.5.0) (2025-06-17)


### Features

* Add metrics for API endpoints being called (requests, duration) ([5e54c5b](https://github.com/trivago/identity-metadata-server/commit/5e54c5b19ec158d347e3749b4bd9c58451f7dd9f))
* Add metrics for service account cache (hit, miss) ([5e54c5b](https://github.com/trivago/identity-metadata-server/commit/5e54c5b19ec158d347e3749b4bd9c58451f7dd9f))
* Add metrics for token cache (hit, miss, set) ([5e54c5b](https://github.com/trivago/identity-metadata-server/commit/5e54c5b19ec158d347e3749b4bd9c58451f7dd9f))
* add more metrics for metadata-server ([#68](https://github.com/trivago/identity-metadata-server/issues/68)) ([5e54c5b](https://github.com/trivago/identity-metadata-server/commit/5e54c5b19ec158d347e3749b4bd9c58451f7dd9f))

## [3.4.2](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.4.1...metadata-server-3.4.2) (2025-06-14)


### Bug Fixes

* consider Running _and_ Pending pods ([#65](https://github.com/trivago/identity-metadata-server/issues/65)) ([87b3608](https://github.com/trivago/identity-metadata-server/commit/87b3608e5d09771e6585b3cf5cc0c2e9916e5704))

## [3.4.1](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.4.0...metadata-server-3.4.1) (2025-06-13)


### Bug Fixes

* **ci:** trigger release ([4b4c4d1](https://github.com/trivago/identity-metadata-server/commit/4b4c4d19e190df6843b9c90d1175337f35e41464))

## [3.4.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.3.0...metadata-server-3.4.0) (2025-06-13)


### Features

* allow changing token lifetimes ([#60](https://github.com/trivago/identity-metadata-server/issues/60)) ([39d6faf](https://github.com/trivago/identity-metadata-server/commit/39d6faf976560948073c0c46122b4428fc867dbf))
* **metadata-server:** token lifetimes can now be configured ([39d6faf](https://github.com/trivago/identity-metadata-server/commit/39d6faf976560948073c0c46122b4428fc867dbf))


### Bug Fixes

* **metadata-server:** the default lifetime for identity tokens has been lowered from 1h to 10m ([39d6faf](https://github.com/trivago/identity-metadata-server/commit/39d6faf976560948073c0c46122b4428fc867dbf))
* **metadata-server:** token request tokens will have the same lifetime as the tokens they request, but a minimum value of 10m ([39d6faf](https://github.com/trivago/identity-metadata-server/commit/39d6faf976560948073c0c46122b4428fc867dbf))

## [3.3.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.2.3...metadata-server-3.3.0) (2025-05-27)


### Features

* add scopes endpoint ([#48](https://github.com/trivago/identity-metadata-server/issues/48)) ([9922ec0](https://github.com/trivago/identity-metadata-server/commit/9922ec0a8d8d5c6062a25ca2fdba815e4e04800c))


### Bug Fixes

* Increase token min lifetime from 5s to 1m ([321def2](https://github.com/trivago/identity-metadata-server/commit/321def20eee80e65aa4366a9c333149592c344a7))
* lower kubernetes service account cache TTL from 10m to 2m ([321def2](https://github.com/trivago/identity-metadata-server/commit/321def20eee80e65aa4366a9c333149592c344a7))


### Miscellaneous

* add configuration to metadata-server readme ([321def2](https://github.com/trivago/identity-metadata-server/commit/321def20eee80e65aa4366a9c333149592c344a7))

## [3.2.3](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.2.2...metadata-server-3.2.3) (2025-05-20)


### Bug Fixes

* **ci:** trigger release ([a36c964](https://github.com/trivago/identity-metadata-server/commit/a36c964f9f8b78c963b5966ea51e47265f9dc2e8))

## [3.2.2](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.2.1...metadata-server-3.2.2) (2025-05-19)


### Bug Fixes

* **ci:** trigger release ([a30ea94](https://github.com/trivago/identity-metadata-server/commit/a30ea94e066efa0e60cb9356ea8ae22642cc1438))

## [3.2.1](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.2.0...metadata-server-3.2.1) (2025-05-19)


### Bug Fixes

* **ci:** testing release-please workflow ([18f31d8](https://github.com/trivago/identity-metadata-server/commit/18f31d8c1daedafcc9516551a9910c86c5037462))

## [3.2.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.1.0...metadata-server-3.2.0) (2025-05-19)


### Features

* basic metadata-server metrics ([#26](https://github.com/trivago/identity-metadata-server/issues/26)) ([ec8e8c8](https://github.com/trivago/identity-metadata-server/commit/ec8e8c8fb69a3d19538c14643c73c784c79bfad0))

## [3.1.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-3.0.0...metadata-server-3.1.0) (2025-05-19)


### Features

* **identity-server:** New endpoint /renew allowing to refresh a client certificate ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))
* **metadata-server:** Auto refresh TLS client certificates before expiry ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))
* New utility functions for file and certificate handling ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))
* PDX-1603 Auto-renew client certificates ([#25](https://github.com/trivago/identity-metadata-server/issues/25)) ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))


### Bug Fixes

* HttpGETJson and HttpPOSTJson will return a textual representation of the HTTP status code if there is no body ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))
* HttpGETJson and HttpPOSTJson will set an "Accept" header if not given ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))
* **identity-server:** Propagate context when signing the identity-server identity ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))


### Miscellaneous

* Add an (optional) integration test for certificate renewal ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))
* Added unit tests for identity server and shared modules ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))
* Make the integration test use a separate service account ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))
* Update of all go dependencies ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))
* Update to go 1.24 ([fda3a46](https://github.com/trivago/identity-metadata-server/commit/fda3a465edeb11ffa097df177fd205d463c5f153))

## [3.0.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-2.2.0...metadata-server-3.0.0) (2025-05-15)


### ⚠ BREAKING CHANGES

* **identity-server:** Introduced cloud-based certificate management with automated downloads and generation.
* **identity-server:** Removed legacy client, certificate, and database management functionality.
* Use Certificate Authority service as a backend ([#22](https://github.com/trivago/identity-metadata-server/issues/22))

### Features

* Enabled CRL and CA export to GCS for the Certificate Authority ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* **identity-server:** Enabled enhanced client identity verification and certificate revocation updates. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* **identity-server:** Enhanced HTTP helper functions for JSON handling in requests. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* **identity-server:** Implemented a Certificate Revocation List (CRL) for managing revoked certificates. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* **identity-server:** Introduced cloud-based certificate management with automated downloads and generation. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* **identity-server:** Streamlined configuration for workload identity and certificate authority settings. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* Use Certificate Authority service as a backend ([#22](https://github.com/trivago/identity-metadata-server/issues/22)) ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))


### Refactorings

* **identity-server:** Removed legacy client, certificate, and database management functionality. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* **metadata-server:** Moved token providers to a new internal module ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))


### Miscellaneous

* Centralized shared constants for more consistent behavior. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* Introduced new IAM roles and service accounts for improved access control. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* Modified certificate generation script for improved output management. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* Refined test coverage for token and identity scenarios. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* Updated GitHub Actions workflow for integration tests to utilize Google Cloud services. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))
* Updated infrastructure and IAM settings for certificate operations. ([50b5160](https://github.com/trivago/identity-metadata-server/commit/50b516054a35de15c369609d1dede0dcf2850249))

## [2.2.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-2.1.0...metadata-server-2.2.0) (2025-03-17)


### Features

* detect token collisions ([#20](https://github.com/trivago/identity-metadata-server/issues/20)) ([d7bc386](https://github.com/trivago/identity-metadata-server/commit/d7bc386e277cb6778bcb6bafd8c7896c761a09ee))


### Bug Fixes

* don't return tokens about to expire ([d7bc386](https://github.com/trivago/identity-metadata-server/commit/d7bc386e277cb6778bcb6bafd8c7896c761a09ee))


### Miscellaneous

* improve token test robustness against nil values ([d7bc386](https://github.com/trivago/identity-metadata-server/commit/d7bc386e277cb6778bcb6bafd8c7896c761a09ee))

## [2.1.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-2.0.0...metadata-server-2.1.0) (2025-02-26)


### Features

* **identity-server:** all required client data is now derived from the client TLS certificate ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))
* **identity-server:** support for different key formats ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))


### Bug Fixes

* better error handling in many cases ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))
* **identity-server:** Certificate loaded from the database accept deletion without a restart ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))
* **identity-server:** clients are now presented the required root CA ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))
* **identity-server:** KeyID added to JWKS ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))
* **identity-server:** Race condition when registering multiple client certificates in parallel ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))
* **metadata-server:** Improved handling of identities containing whitespace characters ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))
* **metadata-server:** The request context will be used in all outgoing calls ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))


### Miscellaneous

* added example setup scripts ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))
* added example systemd units ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))
* allow container builds to enable CGO ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))
* Setup a workload identity pool for machines ([52a6e16](https://github.com/trivago/identity-metadata-server/commit/52a6e168cb009c7715830dab0160f98a08df2a3d))

## [2.0.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-1.2.0...metadata-server-2.0.0) (2025-02-18)


### ⚠ BREAKING CHANGES

* **identity-server:** rename /client endpoint to /identity
* **identity-server:** rename /.well-known/jwks.json endpoint to /jwks.json
* **metadata-server:** renamed kubernetes-metadata-server to metadata-server
* **identity-server:** allow multiple audiences in token request

### Features

* add support for mTLS based http functions ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))
* **identity-server:** add endpoint to retrieve identity information ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))
* **metadata-server:** Add support for host metadata-server mode ([#13](https://github.com/trivago/identity-metadata-server/issues/13)) ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))
* **metadata-server:** implement host token provider ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))


### Refactorings

* **identity-server:** allow multiple audiences in token request ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))
* **identity-server:** rename /.well-known/jwks.json endpoint to /jwks.json ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))
* **identity-server:** rename /client endpoint to /identity ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))
* **metadata-server:** renamed kubernetes-metadata-server to metadata-server ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))


### Miscellaneous

* add debug output for token exchange request ([31d9afd](https://github.com/trivago/identity-metadata-server/commit/31d9afd4ad2e4641df4bed0fb0b5b5638fa5964b))
* add new/changed endpoints to integration test ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))
* use new metadata-server name in CI and docs ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))

## [1.2.0](https://github.com/trivago/identity-metadata-server/compare/metadata-server-v1.1.3...metadata-server-1.2.0) (2025-02-18)


### Features

* /identity now returns a JWT ([c031658](https://github.com/trivago/identity-metadata-server/commit/c03165892ff560f35b59107fa295d5cdb2746f80))
* /universe-domain now returns the expected string ([c031658](https://github.com/trivago/identity-metadata-server/commit/c03165892ff560f35b59107fa295d5cdb2746f80))
* implement list endpoints behavior ([bbb035b](https://github.com/trivago/identity-metadata-server/commit/bbb035b39ab46c6c077647fc8b93857c81346f62))
* Rework identity endpoint ([#1](https://github.com/trivago/identity-metadata-server/issues/1)) ([c031658](https://github.com/trivago/identity-metadata-server/commit/c03165892ff560f35b59107fa295d5cdb2746f80))
* The audience and scope URL parameters are now taken into account ([c031658](https://github.com/trivago/identity-metadata-server/commit/c03165892ff560f35b59107fa295d5cdb2746f80))
* The service account cache TTL is now configurable ([c031658](https://github.com/trivago/identity-metadata-server/commit/c03165892ff560f35b59107fa295d5cdb2746f80))


### Bug Fixes

* add missing response header to sa token requests ([a307f6e](https://github.com/trivago/identity-metadata-server/commit/a307f6ec18e954e0763b89797d7e76eef6f618f3))
* add timezone to tokenTimeFormat (proper RFC3339) ([8b5f4ed](https://github.com/trivago/identity-metadata-server/commit/8b5f4ed2d7009e33d5ed31f58db4244ae8464802))
* allow detection requests ([ca9b0a7](https://github.com/trivago/identity-metadata-server/commit/ca9b0a7d603c92d0bc0ab19ac22c3479f18648c4))
* allow multiple scopes for access tokens ([#4](https://github.com/trivago/identity-metadata-server/issues/4)) ([8f8b054](https://github.com/trivago/identity-metadata-server/commit/8f8b054c0d6fc6cffdf735afae6a2153ed0251d9))
* always set metadata-flavor response header ([11da8a4](https://github.com/trivago/identity-metadata-server/commit/11da8a4b714023c91a2797d7ab82c4d9be319142))
* golang client race condition on main endpoint fetch ([df5c1aa](https://github.com/trivago/identity-metadata-server/commit/df5c1aa7e12e7707a80a4ddb14b7b7d116b44e6e))
* If the given "scopes" do not set the platform or IAM scope, the IAM scope is added for the Token Request Token. ([8f8b054](https://github.com/trivago/identity-metadata-server/commit/8f8b054c0d6fc6cffdf735afae6a2153ed0251d9))
* The "scopes" URL parameter is now properly evaluated ([8f8b054](https://github.com/trivago/identity-metadata-server/commit/8f8b054c0d6fc6cffdf735afae6a2153ed0251d9))
* The PodIP is now derived from the network connection, not HTTP headers ([c031658](https://github.com/trivago/identity-metadata-server/commit/c03165892ff560f35b59107fa295d5cdb2746f80))
* The token cache now has a GC for expired tokens ([c031658](https://github.com/trivago/identity-metadata-server/commit/c03165892ff560f35b59107fa295d5cdb2746f80))
* Token cache hash collision ([56220a0](https://github.com/trivago/identity-metadata-server/commit/56220a01a83d9610f995bb0652de87df9a6ed6b7))


### Miscellaneous

* add empty changelog ([c760002](https://github.com/trivago/identity-metadata-server/commit/c76000272d552cc8fedac6d75d5cb9f5c4ee27c0))
* add more unittests ([1afab54](https://github.com/trivago/identity-metadata-server/commit/1afab54cc2e4c9834ecf1a3f85619edee99e6d6f))
* add some useful links to readme ([907c7d0](https://github.com/trivago/identity-metadata-server/commit/907c7d00bba74c56006ae3d2d2a5a67620a0cca1))
* add unit test for token cache ([2afa36a](https://github.com/trivago/identity-metadata-server/commit/2afa36a52a22ae79e8b1ecf26be6ae06ed897165))
* Added a script containing the "integration test" run in the demo container ([8f8b054](https://github.com/trivago/identity-metadata-server/commit/8f8b054c0d6fc6cffdf735afae6a2153ed0251d9))
* Added test targets to the justfile ([8f8b054](https://github.com/trivago/identity-metadata-server/commit/8f8b054c0d6fc6cffdf735afae6a2153ed0251d9))
* Changed the security settings in the demo container manifest (temporary fix until I find out why this thing need root) ([8f8b054](https://github.com/trivago/identity-metadata-server/commit/8f8b054c0d6fc6cffdf735afae6a2153ed0251d9))
* convert default scopes to constants ([9619bd3](https://github.com/trivago/identity-metadata-server/commit/9619bd3b86357645d9239b3126d5cba26eea6951))
* finalize handler tests for tokens ([0ace952](https://github.com/trivago/identity-metadata-server/commit/0ace95219942bebe2e03b33cad750b59bf0557b8))
* Lots of code cleanup for better readability and to get rid of global initializers ([c031658](https://github.com/trivago/identity-metadata-server/commit/c03165892ff560f35b59107fa295d5cdb2746f80))
* share code between server and client ([#8](https://github.com/trivago/identity-metadata-server/issues/8)) ([601048b](https://github.com/trivago/identity-metadata-server/commit/601048b7172317faf97c154d014ffa84f0759e83))

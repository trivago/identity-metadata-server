# Changelog

## [2.3.0](https://github.com/trivago/identity-metadata-server/compare/identity-server-2.2.1...identity-server-2.3.0) (2025-07-09)


### Features

* Add an optional go profiling endpoint ([731d253](https://github.com/trivago/identity-metadata-server/commit/731d253f4e12c50192e4b958a052965213ca9f38))


### Bug Fixes

* Allow setting an idle timeout (defaults to 620s) ([731d253](https://github.com/trivago/identity-metadata-server/commit/731d253f4e12c50192e4b958a052965213ca9f38))

## [2.2.1](https://github.com/trivago/identity-metadata-server/compare/identity-server-2.2.0...identity-server-2.2.1) (2025-07-09)


### Bug Fixes

* CRL update was not called regularly ([b43a4b3](https://github.com/trivago/identity-metadata-server/commit/b43a4b3d1048d7a2ab8c88de2484c3ed935ed0b6))
* goroutine leak in identity-server ([#89](https://github.com/trivago/identity-metadata-server/issues/89)) ([b43a4b3](https://github.com/trivago/identity-metadata-server/commit/b43a4b3d1048d7a2ab8c88de2484c3ed935ed0b6))

## [2.2.0](https://github.com/trivago/identity-metadata-server/compare/identity-server-2.1.0...identity-server-2.2.0) (2025-07-09)


### Features

* add basic metrics to identity server ([#88](https://github.com/trivago/identity-metadata-server/issues/88)) ([ada1704](https://github.com/trivago/identity-metadata-server/commit/ada17040df934e20ecd1f39befa8d3e8661ade14))
* basic metadata-server metrics ([#26](https://github.com/trivago/identity-metadata-server/issues/26)) ([ec8e8c8](https://github.com/trivago/identity-metadata-server/commit/ec8e8c8fb69a3d19538c14643c73c784c79bfad0))

## [2.1.0](https://github.com/trivago/identity-metadata-server/compare/identity-server-2.0.0...identity-server-2.1.0) (2025-05-19)


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

## [2.0.0](https://github.com/trivago/identity-metadata-server/compare/identity-server-1.1.0...identity-server-2.0.0) (2025-04-10)


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

## [1.1.0](https://github.com/trivago/identity-metadata-server/compare/identity-server-1.0.0...identity-server-1.1.0) (2025-02-26)


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

## [1.0.0](https://github.com/trivago/identity-metadata-server/compare/identity-server-0.1.1...identity-server-1.0.0) (2025-02-18)


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

* add new/changed endpoints to integration test ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))
* **main:** release identity-server 0.1.1 ([#12](https://github.com/trivago/identity-metadata-server/issues/12)) ([cb41cfd](https://github.com/trivago/identity-metadata-server/commit/cb41cfd0e022093f2b401ed17e3901cd9d73ce81))
* use new metadata-server name in CI and docs ([163db4e](https://github.com/trivago/identity-metadata-server/commit/163db4e61356ca81e6156db9eba046add39b7d5e))

## [0.1.1](https://github.com/trivago/identity-metadata-server/compare/identity-server-v0.1.0...identity-server-0.1.1) (2025-02-18)


### Miscellaneous

* **main:** release identity-server 0.1.1 ([#10](https://github.com/trivago/identity-metadata-server/issues/10)) ([ddafeaf](https://github.com/trivago/identity-metadata-server/commit/ddafeaf8446e107cdad6fbb2296774888c73e949))
* share code between server and client ([#8](https://github.com/trivago/identity-metadata-server/issues/8)) ([601048b](https://github.com/trivago/identity-metadata-server/commit/601048b7172317faf97c154d014ffa84f0759e83))

## [0.1.1](https://github.com/trivago/identity-metadata-server/compare/identity-server-v0.1.0...identity-server-0.1.1) (2025-02-18)


### Miscellaneous

* share code between server and client ([#8](https://github.com/trivago/identity-metadata-server/issues/8)) ([601048b](https://github.com/trivago/identity-metadata-server/commit/601048b7172317faf97c154d014ffa84f0759e83))

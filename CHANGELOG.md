# Changelog

## [0.2.3-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.2.2-alpha...v0.2.3-alpha) (2022-07-07)


### Bug Fixes

* change port from 8080 to 8084 ([364f155](https://github.com/instill-ai/mgmt-backend/commit/364f15579a8b55f142fb50611474cdaee8ca5e91))

## [0.2.2-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.2.1-alpha...v0.2.2-alpha) (2022-06-27)


### Miscellaneous Chores

* release v0.2.2-alpha ([1ccaadc](https://github.com/instill-ai/mgmt-backend/commit/1ccaadc09a28db880f8987b479a0522cee0a9309))

## [0.2.1-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.2.0-alpha...v0.2.1-alpha) (2022-06-24)


### Bug Fixes

* check existing user by id ([dd862a4](https://github.com/instill-ai/mgmt-backend/commit/dd862a455d022cd225b2ff90d33cfd2902b5c5c4))

## [0.2.0-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.1.2-alpha...v0.2.0-alpha) (2022-06-22)


### Features

* add reporter for usage collection ([320d9a4](https://github.com/instill-ai/mgmt-backend/commit/320d9a430df114187a7df2989db115bf06a8fda7))
* add retrieve all user function ([da78893](https://github.com/instill-ai/mgmt-backend/commit/da78893ef71611aaac248a9bcdd640066a6b9b55))


### Bug Fixes

* add `tlsenabled` in usage backend configuration ([dd8843a](https://github.com/instill-ai/mgmt-backend/commit/dd8843a0851a83cd52614e03a304f301e96a9fec))
* init config before logger ([455ef7a](https://github.com/instill-ai/mgmt-backend/commit/455ef7ab89434894dd9d0dbbb72332d43002daad))
* init config first ([4e3dea6](https://github.com/instill-ai/mgmt-backend/commit/4e3dea677c70ff66ac4ef0a285897d3fa471d52f))
* put client connection in main ([3ff4821](https://github.com/instill-ai/mgmt-backend/commit/3ff48210398bd6cdd4cfa99ec3beb3dadfb93027))
* refactor usage collection ([c6a41f6](https://github.com/instill-ai/mgmt-backend/commit/c6a41f63b9470455f54ea13ac48960e02fb2f56b))
* specify time duration unit ([27b77d2](https://github.com/instill-ai/mgmt-backend/commit/27b77d2ffc72d64da6101c64643e02c3af696286))

### [0.1.2-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.1.1-alpha...v0.1.2-alpha) (2022-05-31)


### Bug Fixes

* support CORS ([#3](https://github.com/instill-ai/mgmt-backend/issues/3)) ([5fb1439](https://github.com/instill-ai/mgmt-backend/commit/5fb14390b7821e5b1c08cd489d9262b69be6c4ee))
* use cors package to replace naive implementation ([ae49c1c](https://github.com/instill-ai/mgmt-backend/commit/ae49c1c541a2fb08908c068c40b5a7bc17c3be8d))


### Miscellaneous Chores

* release 0.1.2-alpha ([21526dd](https://github.com/instill-ai/mgmt-backend/commit/21526dd454e5151912e61972b2f2da1c2c56c893))

### [0.1.1-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.1.0-alpha...v0.1.1-alpha) (2022-05-18)


### Bug Fixes

* use dynamic uid for default user ([eac0d03](https://github.com/instill-ai/mgmt-backend/commit/eac0d035657bff89462b691a01ae3f3d58e58871))

## [0.1.0-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.0.0-alpha...v0.1.0-alpha) (2022-05-12)


### Features

* add json schema for user struct ([473e197](https://github.com/instill-ai/mgmt-backend/commit/473e197d60f9dd7f2140f9d9cd9bdbcb65ae2aa5))
* add server code ([b7c719a](https://github.com/instill-ai/mgmt-backend/commit/b7c719a662883f94b41f7f399abec3131233f577))


### Bug Fixes

* add custom UI properties in JSON schema ([fc372bc](https://github.com/instill-ai/mgmt-backend/commit/fc372bcc229146111b2571dcd5b5564e6b3dc94f))
* add lookup and use x library ([8cd0139](https://github.com/instill-ai/mgmt-backend/commit/8cd0139d9a2e1f5bb8a7dd1ef441ce660b74387f))
* add migration files in dockerfile ([fc74a5c](https://github.com/instill-ai/mgmt-backend/commit/fc74a5cb0bd08120031f2977349ad94f9e519859))
* add migration sql files ([e47f7dc](https://github.com/instill-ai/mgmt-backend/commit/e47f7dc0b79d81e636ede83fa355d1a6ac7a1a60))
* allow query by both ID and UUID ([0331647](https://github.com/instill-ai/mgmt-backend/commit/03316471fe541b0e0539684f553400190377dd7e))
* auto create/update time with gorm tag ([407a845](https://github.com/instill-ai/mgmt-backend/commit/407a84572eea3b0b41e033e370f582194697d063))
* handle nil pointer in convertor ([1569e1c](https://github.com/instill-ai/mgmt-backend/commit/1569e1ca5a2c8190acda88f4a38981576fc30c33))
* refactor backend based on new proto ([3bf00a9](https://github.com/instill-ai/mgmt-backend/commit/3bf00a99abc2e8b16460ff8577acd3d745c5499d))
* refactor to be consistent with proto ([52f9602](https://github.com/instill-ai/mgmt-backend/commit/52f96029b441752cae0c936849316d83505220a4))
* refactor to the latest proto ([d3e905b](https://github.com/instill-ai/mgmt-backend/commit/d3e905bde547fa21c5f4efad7b27cd171b1331ba))
* update json schema based on new proto ([94794ba](https://github.com/instill-ai/mgmt-backend/commit/94794ba58c3da87d013b6c0f82eb3d353a78601d))
* validate `id` is not in UUID format ([85946d8](https://github.com/instill-ai/mgmt-backend/commit/85946d846f5ab8fe61cf3622407926c03b9e7b76))

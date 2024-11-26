# Changelog

## [0.22.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.21.0-beta...v0.22.0-beta) (2024-11-26)


### Features

* **metric:** implement new pipeline dashboard endpoints ([#238](https://github.com/instill-ai/mgmt-backend/issues/238)) ([afc27e8](https://github.com/instill-ai/mgmt-backend/commit/afc27e8dd5e655a595e4573a067b663e8c0ac7b9))

## [0.21.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.20.2-beta...v0.21.0-beta) (2024-11-04)


### Features

* **metric:** model trigger chart record api ([#245](https://github.com/instill-ai/mgmt-backend/issues/245)) ([89cce72](https://github.com/instill-ai/mgmt-backend/commit/89cce7259833ddab40ffa8b8a7a8acb97d192695))
* **metric:** model trigger counts api ([#248](https://github.com/instill-ai/mgmt-backend/issues/248)) ([7294d49](https://github.com/instill-ai/mgmt-backend/commit/7294d49b653af3076a49f7c51a03273c735be45a))


### Bug Fixes

* **metric:** fix path conflicts ([#249](https://github.com/instill-ai/mgmt-backend/issues/249)) ([c6cebe6](https://github.com/instill-ai/mgmt-backend/commit/c6cebe6b88e80ee93f201207bc8164d73aabfe4d))

## [0.20.2-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.20.1-beta...v0.20.2-beta) (2024-10-23)


### Bug Fixes

* **metric:** reintroduce /metrics/vdp/pipeline/triggers endpoint ([#243](https://github.com/instill-ai/mgmt-backend/issues/243)) ([d595d16](https://github.com/instill-ai/mgmt-backend/commit/d595d167eda4c0cd9468dc8d6ec91066bda5dd9a))

## [0.20.1-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.20.0-beta...v0.20.1-beta) (2024-10-14)


### Bug Fixes

* **dashboard:** broken InfluxDB query on trigger count ([#241](https://github.com/instill-ai/mgmt-backend/issues/241)) ([15a3b2e](https://github.com/instill-ai/mgmt-backend/commit/15a3b2e54c7ad9d21a0ee631f3e69838f5331ce9))

## [0.20.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.19.2-beta...v0.20.0-beta) (2024-10-08)


### Miscellaneous Chores

* **release:** release v0.20.0-beta ([0d54015](https://github.com/instill-ai/mgmt-backend/commit/0d540157b86c2b1253f140053418174c6d179e67))

## [0.19.2-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.19.1-beta...v0.19.2-beta) (2024-08-13)


### Miscellaneous Chores

* release v0.19.2-beta ([f3641f9](https://github.com/instill-ai/mgmt-backend/commit/f3641f9819e3f9fb1ae9ad28eb381e23a81a15d7))

## [0.19.1-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.19.0-beta...v0.19.1-beta) (2024-08-01)


### Bug Fixes

* fix organization can not be updated ([#233](https://github.com/instill-ai/mgmt-backend/issues/233)) ([e811b0c](https://github.com/instill-ai/mgmt-backend/commit/e811b0cb5f879e3996c4e201a98f22c7259f8399))
* fix the issue preventing users from inviting organization members ([#231](https://github.com/instill-ai/mgmt-backend/issues/231)) ([1f0c499](https://github.com/instill-ai/mgmt-backend/commit/1f0c49946386e331067907d10a8cd376dbaca743))


### Miscellaneous Chores

* release v0.19.1-beta ([94e7b39](https://github.com/instill-ai/mgmt-backend/commit/94e7b39f7cbadb471b912f9f237d528afe713088))

## [0.19.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.18.1-beta...v0.19.0-beta) (2024-07-29)


### Features

* add private endpoint `CheckNamespaceByUIDAdmin` ([#227](https://github.com/instill-ai/mgmt-backend/issues/227)) ([274d365](https://github.com/instill-ai/mgmt-backend/commit/274d365b4aaf2e63119ea963bf3c0f52b7a7198d))
* **mgmt:** add knowledge base acl model ([#226](https://github.com/instill-ai/mgmt-backend/issues/226)) ([b4b813e](https://github.com/instill-ai/mgmt-backend/commit/b4b813ea27626be1c7ccaadc30497bf1e64ef06b))
* unify pipeline trigger chart endpoint with credit chart ([#228](https://github.com/instill-ai/mgmt-backend/issues/228)) ([c2ccc17](https://github.com/instill-ai/mgmt-backend/commit/c2ccc17d3d6bd84149c42fa3c1cc7000175e8146))
* use explicit user_id and organization_id in request params ([#224](https://github.com/instill-ai/mgmt-backend/issues/224)) ([3388190](https://github.com/instill-ai/mgmt-backend/commit/33881905b4a37d5f6a416470deea4210fa96ea0b))


### Bug Fixes

* restore pipeline dashboard enpoints ([#230](https://github.com/instill-ai/mgmt-backend/issues/230)) ([02e32df](https://github.com/instill-ai/mgmt-backend/commit/02e32dff8d174e8c9b4499fb05da9bdef51647b3))

## [0.18.1-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.18.0-beta...v0.18.1-beta) (2024-07-16)


### Miscellaneous Chores

* release v0.18.1-beta ([c858468](https://github.com/instill-ai/mgmt-backend/commit/c8584686650121107735ed5ea3cb34f85ea8dba8))

## [0.18.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.17.0-beta...v0.18.0-beta) (2024-06-29)


### Features

* add script to create `preset` namespace ([#219](https://github.com/instill-ai/mgmt-backend/issues/219)) ([a578ed9](https://github.com/instill-ai/mgmt-backend/commit/a578ed900aedb73c75b1d57dd3df3bbe0ca8085e))


### Bug Fixes

* fix metric endpoints bug ([#221](https://github.com/instill-ai/mgmt-backend/issues/221)) ([b24db57](https://github.com/instill-ai/mgmt-backend/commit/b24db572545fbf68a892241718a2a9e4ac461846))

## [0.17.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.16.0-beta...v0.17.0-beta) (2024-06-18)


### Features

* **endpoints:** use camelCase for `filter` and `orderBy` query strings ([#217](https://github.com/instill-ai/mgmt-backend/issues/217)) ([12242f8](https://github.com/instill-ai/mgmt-backend/commit/12242f8b2d8ccf06a4ba6fed8a0d38d35a4a805b))
* use camelCase for HTTP body ([#213](https://github.com/instill-ai/mgmt-backend/issues/213)) ([81c518e](https://github.com/instill-ai/mgmt-backend/commit/81c518eb5f32413bad4d81b49adfebdd62dda4dc))


### Bug Fixes

* receive config in InfluxDB constructor ([#215](https://github.com/instill-ai/mgmt-backend/issues/215)) ([2fc8f3e](https://github.com/instill-ai/mgmt-backend/commit/2fc8f3e5fdf9d73f748a694b046f4f0fef6f82fe))

## [0.16.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.15.2-beta...v0.16.0-beta) (2024-06-06)


### Features

* update last use time when api token is used ([#210](https://github.com/instill-ai/mgmt-backend/issues/210)) ([10fa9e2](https://github.com/instill-ai/mgmt-backend/commit/10fa9e238a2faab183d3f77157cde32b6f362689))

## [0.15.2-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.15.1-beta...v0.15.2-beta) (2024-05-16)


### Miscellaneous Chores

* release v0.15.2-beta ([bf945cc](https://github.com/instill-ai/mgmt-backend/commit/bf945ccd69af29fed23bbeeafd76e6ae332d0ab7))

## [0.15.1-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.15.0-beta...v0.15.1-beta) (2024-05-07)


### Bug Fixes

* fix organization converter bug ([15ed2a2](https://github.com/instill-ai/mgmt-backend/commit/15ed2a20e09cc95dd043532868f09e168f081651))

## [0.15.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.14.0-beta...v0.15.0-beta) (2024-04-23)


### Features

* add `admin` role for organization ([#203](https://github.com/instill-ai/mgmt-backend/issues/203)) ([ad161d8](https://github.com/instill-ai/mgmt-backend/commit/ad161d8660de6de037ac13a8f5765c770b585be4))
* update credit datamodel ([#198](https://github.com/instill-ai/mgmt-backend/issues/198)) ([fe3524f](https://github.com/instill-ai/mgmt-backend/commit/fe3524f5b71566185894fe57a73bac7d265a73a2))

## [0.14.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.13.0-beta...v0.14.0-beta) (2024-04-11)


### Features

* add credit datamodel and methods ([#194](https://github.com/instill-ai/mgmt-backend/issues/194)) ([a220469](https://github.com/instill-ai/mgmt-backend/commit/a2204690c479ebbf09b080d5a31771f4e1ea5a2d))
* implement endpoints for user and organization avatars ([#195](https://github.com/instill-ai/mgmt-backend/issues/195)) ([28769fa](https://github.com/instill-ai/mgmt-backend/commit/28769fa667dc8e5b5577cf10e0efb0fdca79ba8e))

## [0.13.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.12.0-beta...v0.13.0-beta) (2024-03-28)


### Features

* add configuration for read-replica database ([#189](https://github.com/instill-ai/mgmt-backend/issues/189)) ([afdcd5f](https://github.com/instill-ai/mgmt-backend/commit/afdcd5f656503d08734fc14cd965c31b4326d259))
* pin the user to read from the primary database for a certain time frame after mutating the data ([#192](https://github.com/instill-ai/mgmt-backend/issues/192)) ([0f3f707](https://github.com/instill-ai/mgmt-backend/commit/0f3f70731e3cde9b65738a7c1c909c71704ddcae))
* set TTL to owner cache ([#191](https://github.com/instill-ai/mgmt-backend/issues/191)) ([2222908](https://github.com/instill-ai/mgmt-backend/commit/2222908d893dfe6fe5cafa290012b8799252ed21))
* support OpenFGA read replica ([#193](https://github.com/instill-ai/mgmt-backend/issues/193)) ([7ea8814](https://github.com/instill-ai/mgmt-backend/commit/7ea881454ec23309fc1f40c27f3f7cd9e58fa90d))

## [0.12.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.11.0-beta...v0.12.0-beta) (2024-02-29)


### Bug Fixes

* **usage:** send usage report with `AuthenticatedUser` ([#185](https://github.com/instill-ai/mgmt-backend/issues/185)) ([9f54034](https://github.com/instill-ai/mgmt-backend/commit/9f5403479f4b88193751cf892688336e55f67713))


### Miscellaneous Chores

* release v0.12.0-beta ([cec252b](https://github.com/instill-ai/mgmt-backend/commit/cec252b024bb7fbd9bd95b2b281a7ff6780df757))

## [0.11.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.10.0-beta...v0.11.0-beta) (2024-02-16)


### Features

* delete related resources when an organization is deleted ([#180](https://github.com/instill-ai/mgmt-backend/issues/180)) ([313d942](https://github.com/instill-ai/mgmt-backend/commit/313d9427ed84113aa6d7074d2554619ed0c62ce8))
* refactor AuthenticatedUser endpoints ([#175](https://github.com/instill-ai/mgmt-backend/issues/175)) ([fbf08af](https://github.com/instill-ai/mgmt-backend/commit/fbf08af9dcc2ea11ae6abf7163de87e17f490aeb))
* refactor user and organization profile fields ([#177](https://github.com/instill-ai/mgmt-backend/issues/177)) ([94ff2df](https://github.com/instill-ai/mgmt-backend/commit/94ff2df41b4676d9b6e45c7e35cd2fe696d56e40))


### Bug Fixes

* fix the membership endpoint broken after organization deleted ([#178](https://github.com/instill-ai/mgmt-backend/issues/178)) ([59981c6](https://github.com/instill-ai/mgmt-backend/commit/59981c684ff2e05b6505ee9d7ff18c5372c54517))
* **worker:** fix temporal cloud namespace init ([#181](https://github.com/instill-ai/mgmt-backend/issues/181)) ([1f7ec74](https://github.com/instill-ai/mgmt-backend/commit/1f7ec74af6d410f18f19afb50ec78a7028e473e7))

## [0.10.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.9.0-beta...v0.10.0-beta) (2024-01-30)


### Features

* add FGA model for Model ([#172](https://github.com/instill-ai/mgmt-backend/issues/172)) ([c7dbf67](https://github.com/instill-ai/mgmt-backend/commit/c7dbf67ed69cc53b34085ef40c587d0a50ad198e))


### Bug Fixes

* the `instill-auth-type` header was missing when request the pipeline ([#174](https://github.com/instill-ai/mgmt-backend/issues/174)) ([29190c2](https://github.com/instill-ai/mgmt-backend/commit/29190c2a7f54ecd3266dabdd36da1ed3276da6d5))

## [0.9.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.8.0-beta...v0.9.0-beta) (2023-12-31)


### Features

* compress avatar image ([#167](https://github.com/instill-ai/mgmt-backend/issues/167)) ([fe9b3e7](https://github.com/instill-ai/mgmt-backend/commit/fe9b3e7fbee8fa1fd4e57c6706d946881924e60e))
* implement cache mechanism ([#164](https://github.com/instill-ai/mgmt-backend/issues/164)) ([f3ef2ab](https://github.com/instill-ai/mgmt-backend/commit/f3ef2abff65a194582a97be92bed41d1bebf9bba))


### Bug Fixes

* fix `api_token` cache miss bug ([#166](https://github.com/instill-ai/mgmt-backend/issues/166)) ([ebc9acd](https://github.com/instill-ai/mgmt-backend/commit/ebc9acd14026913bee81e07986f8f96208a3378f))
* remove user role enum ([#168](https://github.com/instill-ai/mgmt-backend/issues/168)) ([24187be](https://github.com/instill-ai/mgmt-backend/commit/24187be851fbdbd58ed0f0f7ba33e89e975ea6a9))

## [0.8.0-beta](https://github.com/instill-ai/mgmt-backend/compare/v0.7.0-alpha...v0.8.0-beta) (2023-12-16)


### Features

* **datamodel:** add profile_avatar and profile_data fields ([#144](https://github.com/instill-ai/mgmt-backend/issues/144)) ([fa8a18a](https://github.com/instill-ai/mgmt-backend/commit/fa8a18afb61e3c4614788c4bfad8a6e5036ac0da))
* **organization:** add organization and membership rules ([#146](https://github.com/instill-ai/mgmt-backend/issues/146)) ([a264412](https://github.com/instill-ai/mgmt-backend/commit/a26441225490c3f2f066a39933a6d8e3349fa2e3))


### Bug Fixes

* **handler:** fix organization can not revoke user membership ([#154](https://github.com/instill-ai/mgmt-backend/issues/154)) ([348f222](https://github.com/instill-ai/mgmt-backend/commit/348f222dd4eeb1ad344d5e26cd7313c087d4c2dd))


### Miscellaneous Chores

* release v0.8.0-beta ([b69a74a](https://github.com/instill-ai/mgmt-backend/commit/b69a74ae79c1cc24ab7f833b762c64962578feac))

## [0.7.0-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.6.2-alpha...v0.7.0-alpha) (2023-11-28)


### Features

* **organization:** implement organization endpoints ([#138](https://github.com/instill-ai/mgmt-backend/issues/138)) ([b6a960f](https://github.com/instill-ai/mgmt-backend/commit/b6a960f6d53fe2d1b8c0c06f96962066443c6ef7))


### Miscellaneous Chores

* release v0.7.0-alpha ([2adcc65](https://github.com/instill-ai/mgmt-backend/commit/2adcc653d8946c8884f121ef00c0308b68db7eb2))

## [0.6.2-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.6.1-alpha...v0.6.2-alpha) (2023-10-27)


### Miscellaneous Chores

* **release:** release v0.6.2-alpha ([750cff6](https://github.com/instill-ai/mgmt-backend/commit/750cff698c192cd72f2fa2e7da14eefa7439048c))

## [0.6.1-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.6.0-alpha...v0.6.1-alpha) (2023-10-13)


### Miscellaneous Chores

* **release:** release v0.6.1-alpha ([30e1e28](https://github.com/instill-ai/mgmt-backend/commit/30e1e28540a085e5abe78947bb69a7998abd5c44))

## [0.6.0-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.5.0-alpha...v0.6.0-alpha) (2023-09-30)


### Features

* **auth:** implement authentication apis ([#126](https://github.com/instill-ai/mgmt-backend/issues/126)) ([4ae7e73](https://github.com/instill-ai/mgmt-backend/commit/4ae7e735b1de67269e4e123b0e8ec1d1f6748a42))
* **auth:** support `api_token` authentication ([#128](https://github.com/instill-ai/mgmt-backend/issues/128)) ([73c113c](https://github.com/instill-ai/mgmt-backend/commit/73c113c538d97a41912595b974de26f6858cbb55))


### Bug Fixes

* **init:** fix default user creation bug ([#130](https://github.com/instill-ai/mgmt-backend/issues/130)) ([1ede4d8](https://github.com/instill-ai/mgmt-backend/commit/1ede4d836cceb0ac66cb55178daac039f193ab68))


### Miscellaneous Chores

* **release:** release v0.6.0-alpha ([ef65c9e](https://github.com/instill-ai/mgmt-backend/commit/ef65c9e372da98eff29075ba34123ba7859d989d))

## [0.5.0-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.4.0-alpha...v0.5.0-alpha) (2023-09-13)


### Miscellaneous Chores

* **release:** release v0.5.0-alpha ([e2ee696](https://github.com/instill-ai/mgmt-backend/commit/e2ee696b71a381857f014e39ab6a9380af835f94))

## [0.4.0-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.12-alpha...v0.4.0-alpha) (2023-08-03)


### Miscellaneous Chores

* **release:** release v0.4.0-alpha ([04cf3c8](https://github.com/instill-ai/mgmt-backend/commit/04cf3c8273eae99bc6d92bdc2364dbd2a54c3578))

## [0.3.12-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.11-alpha...v0.3.12-alpha) (2023-07-24)


### Bug Fixes

* **metric:** fix pagination with multiple key sorting ([#106](https://github.com/instill-ai/mgmt-backend/issues/106)) ([aed156f](https://github.com/instill-ai/mgmt-backend/commit/aed156fa0b73a077e633ae6677a8bef1cf75db53))


### Miscellaneous Chores

* **release:** release v0.3.12-alpha ([8315a0b](https://github.com/instill-ai/mgmt-backend/commit/8315a0ba7a3d0063d9426cc8b18ad920b5423a2d))

## [0.3.11-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.10-alpha...v0.3.11-alpha) (2023-07-09)


### Miscellaneous Chores

* **release:** release v0.3.11-alpha ([65990f3](https://github.com/instill-ai/mgmt-backend/commit/65990f39009e6dbca46f16d3ee9bf2b2c42d9816))

## [0.3.10-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.9-alpha...v0.3.10-alpha) (2023-06-20)


### Miscellaneous Chores

* **release:** release 0.3.10-alpha ([c49adf9](https://github.com/instill-ai/mgmt-backend/commit/c49adf9a387ef34812c6532e4bbab7764bc13315))

## [0.3.9-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.8-alpha...v0.3.9-alpha) (2023-06-11)


### Miscellaneous Chores

* **release:** release v0.3.9-alpha ([77f7138](https://github.com/instill-ai/mgmt-backend/commit/77f7138ccdf0f793b178c26dd1c44d6272999675))

## [0.3.8-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.7-alpha...v0.3.8-alpha) (2023-06-02)


### Miscellaneous Chores

* **release:** release v0.3.8-alpha ([31e6cf6](https://github.com/instill-ai/mgmt-backend/commit/31e6cf666441c907cc116d1984305f4402777626))

## [0.3.7-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.6-alpha...v0.3.7-alpha) (2023-05-11)


### Miscellaneous Chores

* **release:** release v0.3.7-alpha ([112c82f](https://github.com/instill-ai/mgmt-backend/commit/112c82f4f5d678e3f3cfef83717a009609d45655))

## [0.3.6-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.5-alpha...v0.3.6-alpha) (2023-05-05)


### Miscellaneous Chores

* **release:** release v0.3.6-alpha ([b84804c](https://github.com/instill-ai/mgmt-backend/commit/b84804c18b508155ba786af386eefb28004cf7ba))

## [0.3.5-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.4-alpha...v0.3.5-alpha) (2023-04-25)


### Bug Fixes

* ignore `create_time` and `update_time` when converting pb to db  ([#67](https://github.com/instill-ai/mgmt-backend/issues/67)) ([8d1fbcd](https://github.com/instill-ai/mgmt-backend/commit/8d1fbcdd8e03108532d3f71d1cbcce404cd00f1a))

## [0.3.4-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.3-alpha...v0.3.4-alpha) (2023-04-15)


### Miscellaneous Chores

* **release:** release v0.3.4-alpha ([817eed7](https://github.com/instill-ai/mgmt-backend/commit/817eed70d8c4a0fd588485bce870dc9feb09fb9c))

## [0.3.3-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.2-alpha...v0.3.3-alpha) (2023-04-07)


### Miscellaneous Chores

* release v0.3.3-alpha ([6c87fed](https://github.com/instill-ai/mgmt-backend/commit/6c87fed8cddb2f69baf7bd44b717d0a708f92959))

## [0.3.2-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.1-alpha...v0.3.2-alpha) (2023-03-26)


### Miscellaneous Chores

* release v0.3.2-alpha ([7a3d1bb](https://github.com/instill-ai/mgmt-backend/commit/7a3d1bb1474d6154e21fadf8b2841b5c3f8ad31f))

## [0.3.1-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.3.0-alpha...v0.3.1-alpha) (2023-02-20)


### Bug Fixes

* fix admin and public servers mixed issue ([8f8adef](https://github.com/instill-ai/mgmt-backend/commit/8f8adef36bab2212effe38a35ab918daaa628675))

## [0.3.0-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.2.9-alpha...v0.3.0-alpha) (2023-02-10)


### Features

* refactor to admin/public services ([#26](https://github.com/instill-ai/mgmt-backend/issues/26)) ([426ab1c](https://github.com/instill-ai/mgmt-backend/commit/426ab1c9ce53aec1d1d7cc43fa6a9507a71f2d87))


### Bug Fixes

* update roles to follow the protobuf ([208d95d](https://github.com/instill-ai/mgmt-backend/commit/208d95dc4e6d944b8d0e2ae15d4f692c218c3578))
* use error logs for usage ([88ab8d4](https://github.com/instill-ai/mgmt-backend/commit/88ab8d448cc6f009f39ab347b8c782bc1d411f87))

## [0.2.9-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.2.8-alpha...v0.2.9-alpha) (2023-01-30)


### Miscellaneous Chores

* release v0.2.9-alpha ([ccab1d5](https://github.com/instill-ai/mgmt-backend/commit/ccab1d594aaab6faecef76a24f08b411ea101f20))

## [0.2.8-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.2.7-alpha...v0.2.8-alpha) (2023-01-14)


### Miscellaneous Chores

* release v0.2.8-alpha ([3e7904b](https://github.com/instill-ai/mgmt-backend/commit/3e7904bde9679c6c901b602adabfb47f05ba18a9))

## [0.2.7-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.2.6-alpha...v0.2.7-alpha) (2022-11-30)


### Miscellaneous Chores

* release 0.2.7-alpha ([2ad918e](https://github.com/instill-ai/mgmt-backend/commit/2ad918ef28b1e9e5817ac0e1da1d1b8690159713))

## [0.2.6-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.2.5-alpha...v0.2.6-alpha) (2022-08-21)


### Bug Fixes

* update usage ([6d607a9](https://github.com/instill-ai/mgmt-backend/commit/6d607a9bd3287744b1bf7ba480720fffe92893fc))

## [0.2.5-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.2.4-alpha...v0.2.5-alpha) (2022-08-17)


### Miscellaneous Chores

* release 0.2.5-alpha ([f2b418f](https://github.com/instill-ai/mgmt-backend/commit/f2b418f803a5125183bf982eb8409f5be52df714))

## [0.2.4-alpha](https://github.com/instill-ai/mgmt-backend/compare/v0.2.3-alpha...v0.2.4-alpha) (2022-07-19)


### Bug Fixes

* change edition to local-ce:dev ([037d738](https://github.com/instill-ai/mgmt-backend/commit/037d73816b7ceb0aec9dd34d684c7cb8e7ee6d81))
* refactor error code ([#11](https://github.com/instill-ai/mgmt-backend/issues/11)) ([4114e8e](https://github.com/instill-ai/mgmt-backend/commit/4114e8e5494723fdec9a39a87e47f08316fab307))

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

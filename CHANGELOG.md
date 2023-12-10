# Changelog

## [0.4.0](https://github.com/go-rq/rq/compare/v0.3.1...v0.4.0) (2023-12-10)


### Features

* export GetEnvironment ([#20](https://github.com/go-rq/rq/issues/20)) ([65a20b1](https://github.com/go-rq/rq/commit/65a20b1f7018e7135bb8dcf64730c1b09833d0f4))
* request.String() outputs script text ([#21](https://github.com/go-rq/rq/issues/21)) ([c1bb210](https://github.com/go-rq/rq/commit/c1bb21082ecc61f262d33ced3f043cef39f2c21d))


### Bug Fixes

* resolve bug where script logs were  not output it tests ([#18](https://github.com/go-rq/rq/issues/18)) ([4a3be6e](https://github.com/go-rq/rq/commit/4a3be6e7961c01e76c69bc46afe3c0ec735ed2e2))
* string formatting of the request ([#22](https://github.com/go-rq/rq/issues/22)) ([75a0475](https://github.com/go-rq/rq/commit/75a047590e954880d90958700adcbb59ee76924a))


### Documentation

* add examples ([#23](https://github.com/go-rq/rq/issues/23)) ([2bf4275](https://github.com/go-rq/rq/commit/2bf427562fbe96dbf7ed344a3b09b24bd19c3496))

## [0.3.1](https://github.com/go-rq/rq/compare/v0.3.0...v0.3.1) (2023-11-30)


### Bug Fixes

* always capture the response body in treqs verbose mode ([#17](https://github.com/go-rq/rq/issues/17)) ([2737e0a](https://github.com/go-rq/rq/commit/2737e0a5c57406ef9f46eb5d4ca6ccac20a508b8))
* where response bodies are now visible in treqs verbose mode ([#15](https://github.com/go-rq/rq/issues/15)) ([37dd05d](https://github.com/go-rq/rq/commit/37dd05d6e7788b06c23ae86e6494e52f3928ab47))

## [0.3.0](https://github.com/go-rq/rq/compare/v0.2.0...v0.3.0) (2023-11-27)


### Features

* export Request funcs `ApplyEnv` and `ToHttpRequest` ([#10](https://github.com/go-rq/rq/issues/10)) ([5c5c5d9](https://github.com/go-rq/rq/commit/5c5c5d9afb512a3ede6a663b5dc86aaa99be041d))

## [0.2.0](https://github.com/go-rq/rq/compare/v0.1.0...v0.2.0) (2023-11-27)


### Features

* add rudimentary log collection from script runtime ([#9](https://github.com/go-rq/rq/issues/9)) ([116610c](https://github.com/go-rq/rq/commit/116610cffaf35fcc9843a5767c797cebff26c685))
* add support for skipping requests from the pre-request script ([#6](https://github.com/go-rq/rq/issues/6)) ([7df8dc2](https://github.com/go-rq/rq/commit/7df8dc28fe959efb7b0b6823085be40f3d25afd7))
* make the parsed json response available in post-request scripts ([#8](https://github.com/go-rq/rq/issues/8)) ([c9b36ae](https://github.com/go-rq/rq/commit/c9b36aeaf4ad8f45a164901754b6ffb475b711d3))

## [0.1.0](https://github.com/go-rq/rq/compare/v0.0.0...v0.1.0) (2023-11-26)


### Features

* add initial `rq` library ([#2](https://github.com/go-rq/rq/issues/2)) ([771aab1](https://github.com/go-rq/rq/commit/771aab1ce8128f09ff1ed72813f76b8d1ded0a34))
* add treqs package for use in go-tests ([#4](https://github.com/go-rq/rq/issues/4)) ([6bbfb8c](https://github.com/go-rq/rq/commit/6bbfb8cb169749aa6b139c1763fc6449b1115232))

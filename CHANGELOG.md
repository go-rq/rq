# Changelog

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

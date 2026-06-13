# Changelog

## [1.1.1](https://github.com/shanehull/obsidian-remote/compare/v1.1.0...v1.1.1) (2026-06-13)


### Bug Fixes

* use PATCH for targeted update_note operations instead of PUT/POST ([#37](https://github.com/shanehull/obsidian-remote/issues/37)) ([99b19cd](https://github.com/shanehull/obsidian-remote/commit/99b19cd5707bc6f50d3058299b2f04dd1603fdfc))

## [1.1.0](https://github.com/shanehull/obsidian-remote/compare/v1.0.0...v1.1.0) (2026-06-01)


### Features

* add move_note tool for moving/renaming vault files ([#32](https://github.com/shanehull/obsidian-remote/issues/32)) ([4c023cd](https://github.com/shanehull/obsidian-remote/commit/4c023cd2f2fcd47dba9855c41e2fdbe5bceaa165))

## [1.0.0](https://github.com/shanehull/obsidian-remote/compare/v0.3.0...v1.0.0) (2026-05-19)


### ⚠ BREAKING CHANGES

* append_note removed. Use update_note with operation=append.

### Features

* add section targeting to read_note and update_note, remove append_note ([#25](https://github.com/shanehull/obsidian-remote/issues/25)) ([6b75515](https://github.com/shanehull/obsidian-remote/commit/6b75515b37921742401d1d6827304471f5ac999a))


### Bug Fixes

* add PATCH to vault content-type detection ([#28](https://github.com/shanehull/obsidian-remote/issues/28)) ([020104d](https://github.com/shanehull/obsidian-remote/commit/020104d60a936409398aa2a999fb3accd33df229))

## [0.3.0](https://github.com/shanehull/obsidian-remote/compare/v0.2.0...v0.3.0) (2026-05-18)


### Features

* add count parameter to search_replace for occurrence control ([#23](https://github.com/shanehull/obsidian-remote/issues/23)) ([e06ee65](https://github.com/shanehull/obsidian-remote/commit/e06ee658fbc989d8655dacfaaa04a6d011fcecd8))

## [0.2.0](https://github.com/shanehull/obsidian-remote/compare/v0.1.0...v0.2.0) (2026-05-09)


### Features

* add graceful shutdown on SIGINT/SIGTERM ([#21](https://github.com/shanehull/obsidian-remote/issues/21)) ([f173844](https://github.com/shanehull/obsidian-remote/commit/f1738449c6f52f0f78741cde66b14ca8467585f7))

## [0.1.0](https://github.com/shanehull/obsidian-remote/compare/v0.0.3...v0.1.0) (2026-05-09)


### Features

* add /healthz endpoint to MCP bridge ([#16](https://github.com/shanehull/obsidian-remote/issues/16)) ([eda882d](https://github.com/shanehull/obsidian-remote/commit/eda882d3dd4856bf4ce448363b361dfcdf70b7a5))


### Bug Fixes

* pin base image to v1.12.7-ls127 and add e2e tests ([#15](https://github.com/shanehull/obsidian-remote/issues/15)) ([ea21b41](https://github.com/shanehull/obsidian-remote/commit/ea21b41f7c54b21a588a9d6915e9e89b59a63f47))
* resolve goreportcard issues, add tests, and bump CI actions ([#17](https://github.com/shanehull/obsidian-remote/issues/17)) ([9b4ff9f](https://github.com/shanehull/obsidian-remote/commit/9b4ff9ff4225430fe78f1df3bc1ce2e5eea9b4e8))
* set correct display port and prevent duplicate obsidian launch ([#11](https://github.com/shanehull/obsidian-remote/issues/11)) ([de1b19a](https://github.com/shanehull/obsidian-remote/commit/de1b19a7984b05ef248b1cdfa60c423bf92accf2))

## [0.0.3](https://github.com/shanehull/obsidian-remote/compare/v0.0.2...v0.0.3) (2026-05-06)


### Bug Fixes

* tool hints ([#9](https://github.com/shanehull/obsidian-remote/issues/9)) ([de64125](https://github.com/shanehull/obsidian-remote/commit/de64125b7713c851077981e0a59a8606bd85119e))

## [0.0.2](https://github.com/shanehull/obsidian-remote/compare/v0.0.1...v0.0.2) (2026-03-18)


### Bug Fixes

* tool success responses ([#5](https://github.com/shanehull/obsidian-remote/issues/5)) ([c1aa7d8](https://github.com/shanehull/obsidian-remote/commit/c1aa7d850ac68fe9bd268e11876d33e7fd8ac876))

## 0.0.1 (2026-03-18)


### Miscellaneous Chores

* release 0.0.1 ([e821ba7](https://github.com/shanehull/obsidian-remote/commit/e821ba71f1e8ff8da284788204847c4091cbc763))

## Changelog

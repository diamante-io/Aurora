# Changelog

All notable changes to this project will be documented in this
file.  This project adheres to [Semantic Versioning](http://semver.org/).

As this project is pre 1.0, breaking changes may happen for minor version
bumps.  A breaking change will get clearly notified in this log.

## [Unreleased]

### Changed

- BREAKING CHANGE: The `url` database configuration has been renamed to `dsn` to more accurately reflect its content.

### Fixed

- TLS support re-enabled.

### Added

- Reverse federation is now optional.
- Logging:  http requests will be logged at the "Info" log level

## [v0.2.0] - 2016-08-17

Initial release after import from https://github.com/hcnet/federation

[Unreleased]: https://github.com/hcnet/go/compare/federation-v0.2.0...master

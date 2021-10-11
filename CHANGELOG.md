# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Changed
- Move `cmd/whatsnew.go` to the project's root path, so that the tool can be
  installed more easily with `go install` command.
- Use `golang:1.17.1-alpine3.14` as base Builder image in the Dockerfile.
- Upgrade dependencies.

## [1.0.0-alpha.2] - 2021-10-01
### Changed
- Enable client-side round-robin DNS load balancing for all gRPC connections.

## [1.0.0-alpha.1] - 2021-09-28
### Added
- Allow setting reservation timeout and number of retries for each Faktory job
  from configuration.

### Changed
- Use the WebArticle translated title, when available, as preferred text data
  source in text-classifier, vectorizer, and zero-shot-classifier workers. 
- Upgrade dependencies.

## [1.0.0-alpha] - 2021-09-26
### Changed
- The whole project has been completely rewritten. Most notably, the simplistic
  way of handling workers' jobs with RabbitMQ has been replaced with more
  reliable jobs scheduling using [Faktory](https://contribsys.com/faktory/).

## [0.5.0] - 2021-09-26
### Added
- API server for managing sources.

### Changed
- Use go `1.16`.
- Simplify the Dockerfile.
- Update the README.

## [0.4.0] - 2021-04-23
### Changed
- Upgrade spaGO to `v0.5.2` and adapt the code.

## [0.3.3] - 2021-03-23
### Added
- Add `max_tweets_number` to tweets-fetching configuration.

## [0.3.2] - 2021-03-23
### Fixed
- Fix wrong RabbitMQ routing key for messages published from the tweets-fetching
  worker.

## [0.3.1] - 2021-03-23
### Fixed
- Fix missing CA-certificates when running in Docker container, using alpine 
  as base Docker image.

## [0.3.0] - 2021-03-23
### Added
- Twitter source integration.

## [0.2.0] - 2021-01-28
### Added
- Allow env vars expansion in config file.

### Changed
- Use root-level sample configuration file for configuration tests.

## [0.1.1] - 2021-01-26
### Changed
- Skip certificate verification during web scraping.

## [0.1.0] - 2021-01-25
First versioned release, ready to be tested.

[Unreleased]: https://github.com/SpecializedGeneralist/whatsnew/compare/v1.0.0-alpha.1...HEAD
[1.0.0-alpha.1]: https://github.com/SpecializedGeneralist/whatsnew/compare/v1.0.0-alpha...1.0.0-alpha.1
[1.0.0-alpha]: https://github.com/SpecializedGeneralist/whatsnew/compare/v0.5.0...v1.0.0-alpha
[0.5.0]: https://github.com/SpecializedGeneralist/whatsnew/compare/v0.4.0...v0.5.0
[0.4.0]: https://github.com/SpecializedGeneralist/whatsnew/compare/v0.3.3...v0.4.0
[0.3.3]: https://github.com/SpecializedGeneralist/whatsnew/compare/v0.3.2...v0.3.3
[0.3.2]: https://github.com/SpecializedGeneralist/whatsnew/compare/v0.3.1...v0.3.2
[0.3.1]: https://github.com/SpecializedGeneralist/whatsnew/compare/v0.3.0...v0.3.1
[0.3.0]: https://github.com/SpecializedGeneralist/whatsnew/compare/v0.2.0...v0.3.0
[0.2.0]: https://github.com/SpecializedGeneralist/whatsnew/compare/v0.1.1...v0.2.0
[0.1.1]: https://github.com/SpecializedGeneralist/whatsnew/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/SpecializedGeneralist/whatsnew/releases/tag/v0.1.0
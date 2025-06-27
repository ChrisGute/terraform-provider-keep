# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- GPG signing for all release artifacts
- Multi-architecture support (Linux, macOS, Windows on amd64 and arm64)
- Automated GitHub Actions workflow for releases
- Terraform Registry publishing
- GPG key management documentation
- Development guide with release process

### Changed
- Updated GoReleaser configuration for multi-platform builds
- Improved documentation and examples
- Enhanced error handling and logging
- Updated provider configuration options

### Fixed
- Resolved GPG signing issues in CI/CD pipeline
- Fixed environment variable handling in tests
- Corrected provider source in documentation

## [0.1.5] - 2025-06-27

### Added
- GPG signing for release artifacts
- Multi-architecture build support
- GitHub Actions workflow for automated releases
- Documentation for GPG verification
- Support for Terraform Registry publishing

### Fixed
- Resolved GPG signing path issues
- Fixed environment variable handling in GoReleaser
- Updated documentation for new release process

## [0.1.0] - 2025-06-26

### Added
- Initial release of the KeepHQ Terraform provider
- Support for managing mapping rules (create, read, update, delete, import)
- Comprehensive documentation and examples
- GitHub Actions workflow for testing and CI/CD
- Unit and acceptance tests for mapping rule resource
- Support for CSV data in mapping rules
- Matcher support for mapping rules

### Changed
- Updated provider schema to match KeepHQ API requirements
- Improved error handling and validation
- Enhanced test coverage for mapping rules
- Updated README with current status and usage examples

### Fixed
- Fixed issue with mapping rule import verification
- Resolved state management issues in the mapping rule resource
- Fixed type handling in API requests and responses

### Known Limitations
- The `disabled` field in mapping rules is not currently supported by the KeepHQ API (see #TBD)
- Alert and provider resources are marked as experimental and not fully functional yet

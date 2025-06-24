# Changelog

All notable changes to SDL will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Comprehensive documentation suite
  - SDL Language Reference with complete syntax guide
  - Component Library documentation for all built-in components
  - Contributing guidelines with code style and PR process
- Installation improvements
  - Makefile-based installation with prerequisite checking
  - Automated dependency installation
  - Version management system
- Release infrastructure
  - Version embedding in binary
  - Semantic versioning support
  - Release automation via Makefile

### Changed
- Removed hardcoded paths from .air.toml configuration
- Updated README with complete prerequisites including goyacc
- Improved installation process to be more user-friendly

### Fixed
- Configuration files now use environment variables instead of hardcoded paths

## [0.9.0] - 2025-06-24

This pre-release represents a major architectural overhaul. See RELEASE_NOTES.md for detailed changes.

### Major Features
- gRPC architecture migration with HTTP gateway
- Enhanced simulation engine with virtual time
- Interactive web dashboard with real-time visualization
- Flow analysis system with automatic rate calculation
- High-throughput traffic generation (1000+ RPS)
- Comprehensive example systems (Uber, Netflix, etc.)

### Technical Improvements
- Pluggable tracer interface design
- Ring buffer-based metric storage
- Method-level system visualization
- Multi-canvas support
- Hot reload development workflow
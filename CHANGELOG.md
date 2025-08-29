# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2025-08-29

### Added
- Initial release of Go Logstash Logger
- **Minimum Go version**: 1.25
- Support for DEBUG, INFO, WARN, ERROR, FATAL log levels
- TCP and UDP connection support to Logstash
- Asynchronous logging with configurable buffering
- JSON log formatting for Logstash compatibility
- Configurable timestamp formats with runtime modification
- Automatic function name capture using Go runtime introspection
- Automatic reconnection on connection loss (TCP only)
- Thread-safe operations with mutex protection
- Caller information display (filename:line)
- Configurable buffer size and flush intervals
- Batch processing for optimal network usage
- DefaultConfig() function with sensible defaults
- Comprehensive test suite
- Example application with Docker Compose setup
- Complete documentation and usage examples

### Features
- **High Performance**: Asynchronous logging with buffering (100,000+ logs/second)
- **Reliability**: Automatic reconnection with configurable retry logic
- **Flexibility**: Multiple timestamp formats and runtime configuration
- **Debugging**: Automatic function name and caller information capture
- **Production Ready**: Robust error handling and graceful degradation

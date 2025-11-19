# Go Logger for Logstash

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.25-blue)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![CI](https://github.com/SergeyDavidenko/logger/workflows/CI/badge.svg)](https://github.com/SergeyDavidenko/logger/actions)
[![Coverage](https://codecov.io/gh/SergeyDavidenko/logger/branch/main/graph/badge.svg)](https://codecov.io/gh/SergeyDavidenko/logger)
[![Go Report Card](https://goreportcard.com/badge/github.com/SergeyDavidenko/logger)](https://goreportcard.com/report/github.com/SergeyDavidenko/logger)
[![go.dev](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white)](https://pkg.go.dev/github.com/SergeyDavidenko/logger)

A high-performance, production-ready Go logging library with native Logstash integration. Designed for modern applications that need reliable, fast, and feature-rich logging capabilities.

## ✨ Key Features

- 🚀 **High Performance**: Asynchronous logging with buffering (100,000+ logs/second)
- 🔄 **Auto-Reconnection**: Robust TCP connection handling with configurable retry logic
- ⚡ **Function Detection**: Automatic function name capture using Go runtime introspection
- 🕒 **Flexible Timestamps**: Configurable timestamp formats with runtime modification
- 🌐 **Protocol Support**: Both TCP and UDP connections to Logstash
- 🛡️ **Production Ready**: Thread-safe operations with comprehensive error handling
- 📊 **ELK Stack**: Native JSON formatting for Elasticsearch/Logstash/Kibana integration

## 📦 Installation

**Requirements:** Go 1.25 or higher

```bash
go get github.com/SergeyDavidenko/logger
```

## 🚀 Quick Start

```go
package main

import (
    "github.com/SergeyDavidenko/logger"
)

func main() {
    // Use default configuration
    config := logger.DefaultConfig()
    config.LogstashHost = "localhost"
    config.LogstashPort = 5008
    config.LogstashEnabled = true
    config.AppName = "my-app"
    
    log := logger.New(config)
    defer log.Close()
    
    log.Info("Hello, World!")
    // Output: [2025-08-29T07:50:30.228Z] INFO [main] [main.go:15] Hello, World!
}
```

## Features

- Support for different logging levels (DEBUG, INFO, WARN, ERROR, FATAL)
- TCP and UDP connection support to Logstash
- **Asynchronous logging with buffering** for high performance
- JSON log formatting
- **Configurable timestamp formats** with runtime modification support
- **Automatic function name capture** - automatically identifies the calling function
- Ability to enable/disable sending to Logstash
- Automatic reconnection on connection loss (TCP only)
- Message formatting support
- Thread-safe operations
- Caller information display (filename:line)
- Configurable buffer size and flush intervals
- Batch processing for optimal network usage
- **Structured logging with `With()` method** - add context fields to loggers

## Usage

```go
package main

import (
    "logger"
)

func main() {
    // Option 1: Use default configuration and customize as needed
    config := logger.DefaultConfig()
    config.LogstashHost = "localhost"
    config.LogstashPort = 5008
    config.LogstashEnabled = true
    config.AppName = "my-app"
    config.MinLevel = logger.INFO
    
    // Option 2: Create custom configuration from scratch
    // config := logger.Config{
    //     TimestampFormat:   "2006-01-02 15:04:05.000", // custom timestamp format
    //     LogstashHost:      "localhost",
    //     LogstashPort:      5008,
    //     LogstashEnabled:   true,
    //     Protocol:          logger.TCP, // or logger.UDP
    //     AppName:           "my-app",
    //     MinLevel:          logger.INFO,
    //     ReconnectAttempts: 5,                   // max 5 reconnection attempts (0 = infinite)
    //     ReconnectDelay:    3 * time.Second,     // 3 seconds between attempts
    //     // Async logging for high performance
    //     AsyncEnabled:      true,                // enable async logging
    //     BufferSize:        2000,                // buffer up to 2000 log entries
    //     FlushInterval:     500 * time.Millisecond, // flush every 500ms
    // }

    // Create logger
    log := logger.New(config)
    defer log.Close()

    // Use different logging levels
    log.Debug("Debug message")
    log.Info("Information message")
    log.Warn("Warning message")
    log.Error("Error: %s", "error description")

    // Control Logstash logging
    log.SetLogstashEnabled(false) // disable
    log.SetLogstashEnabled(true)  // enable
    
    // Configure timestamp format at runtime
    log.SetTimestampFormat("15:04:05 02/01/2006")      // time first format
    log.SetTimestampFormat("2006-01-02T15:04:05Z07:00") // RFC3339 format
    currentFormat := log.GetTimestampFormat()           // get current format
    
    // Force flush of buffered logs (async mode only)
    log.Flush()
    
    // Structured logging with context fields
    loggerWithFields := log.With(map[string]interface{}{
        "user_id":    123,
        "request_id": "req-456",
        "service":    "api",
        "trace_id":    "abc123",
    })
    loggerWithFields.Info("User action completed")
    // Fields will be included in JSON output sent to Logstash
    
    // Nested context fields
    requestLogger := loggerWithFields.With(map[string]interface{}{
        "endpoint": "/api/users",
        "method":   "GET",
    })
    requestLogger.Info("Request processed")
}

// Example functions to demonstrate automatic function name capture
func businessLogic(log *logger.Logger) {
    log.Info("Processing business logic")
    // Output: [2025-08-29 07:50:30.228] INFO [businessLogic] [main.go:11] Processing business logic
}

func databaseOperation(log *logger.Logger) {
    log.Debug("Connecting to database")
    log.Info("Database operation completed successfully")
    // Output: [2025-08-29 07:50:30.228] DEBUG [databaseOperation] [main.go:16] Connecting to database
    // Output: [2025-08-29 07:50:30.228] INFO [databaseOperation] [main.go:17] Database operation completed successfully
}

func handleError(log *logger.Logger) {
    log.Error("An error occurred in error handler")
    // Output: [2025-08-29 07:50:30.228] ERROR [handleError] [main.go:21] An error occurred in error handler
}
```

## Default Configuration

The logger provides a `DefaultConfig()` function that returns a configuration with sensible defaults. This is the recommended starting point for most applications.

### Default Settings

```go
config := logger.DefaultConfig()
// Returns:
// {
//     TimestampFormat: "2006-01-02T15:04:05.000Z", // ISO format with milliseconds
//     LogstashEnabled: false,                       // disabled by default for safety
//     AsyncEnabled:    true,                        // async logging for performance
//     BufferSize:      1000,                        // moderate buffer size
//     FlushInterval:   1 * time.Second,             // flush every second
//     Protocol:        "",                          // will default to TCP when enabled
//     MinLevel:        0,                           // will log all levels (DEBUG and above)
//     ReconnectAttempts: 0,                         // infinite reconnect attempts
//     ReconnectDelay:    0,                         // will default to 5 seconds
// }
```

### Quick Start Examples

```go
// Minimal setup - just enable Logstash with defaults
config := logger.DefaultConfig()
config.LogstashHost = "localhost"
config.LogstashPort = 5008
config.LogstashEnabled = true
config.AppName = "my-app"
logger := logger.New(config)

// Console-only logging (no Logstash)
config := logger.DefaultConfig()
config.AppName = "console-app"
logger := logger.New(config)

// High-performance setup with custom timestamp
config := logger.DefaultConfig()
config.TimestampFormat = "15:04:05.000"  // time only
config.BufferSize = 5000                  // larger buffer
config.FlushInterval = 100 * time.Millisecond // faster flushing
config.AppName = "high-perf-app"
logger := logger.New(config)
```

### Benefits of Using DefaultConfig()

1. **Sensible Defaults**: Pre-configured with recommended settings for most use cases
2. **Performance Optimized**: Async logging enabled by default for better performance
3. **Safe Defaults**: Logstash disabled by default to prevent connection errors during development
4. **Easy Customization**: Modify only the fields you need to change
5. **Future-Proof**: New default settings will be automatically included in updates

## Running the example

```bash
cd example
go run main.go
```

## Testing

```bash
go test -v
go test -bench=.
```

## API Documentation

### Types

#### LogLevel
```go
type LogLevel int
```
Represents the logging level with constants: DEBUG, INFO, WARN, ERROR, FATAL.

**Functions:**
- `ParseLogLevel(level string) (LogLevel, bool)` - Parses a string and returns the corresponding LogLevel
- `MustParseLogLevel(level string) LogLevel` - Parses a string and returns the corresponding LogLevel, panics if invalid
- `(l LogLevel) String() string` - Returns the string representation of the log level

#### Protocol
```go
type Protocol string
```
Network protocol type with constants: TCP, UDP.

#### Config
```go
type Config struct {
    TimestampFormat   string        // Custom timestamp format (default: "2006-01-02T15:04:05.000Z")
    LogstashHost      string        // Logstash server address
    LogstashPort      int           // Logstash port
    LogstashEnabled   bool          // Enable/disable Logstash logging
    Protocol          Protocol      // Network protocol (TCP or UDP)
    AppName           string        // Application name
    MinLevel          LogLevel      // Minimum logging level
    ReconnectAttempts int           // Maximum reconnection attempts (0 = infinite)
    ReconnectDelay    time.Duration // Delay between reconnection attempts
    // Async logging configuration
    AsyncEnabled      bool          // Enable asynchronous logging
    BufferSize        int           // Size of the log buffer (default: 1000)
    FlushInterval     time.Duration // Interval to flush logs (default: 1 second)
}
```

#### Logger
```go
type Logger struct {
    // ... (internal fields)
}
```

### Functions

#### DefaultConfig() Config
Returns a default configuration with recommended settings:
- `TimestampFormat`: "2006-01-02T15:04:05.000Z" (ISO format)
- `LogstashEnabled`: false (disabled by default)
- `AsyncEnabled`: true (async logging enabled)
- `BufferSize`: 1000 (moderate buffer size)
- `FlushInterval`: 1 second

#### New(config Config) *Logger
Creates a new logger instance with the specified configuration.

#### (l *Logger) Debug(message string, args ...interface{})
Logs a message at DEBUG level.

#### (l *Logger) Info(message string, args ...interface{})
Logs a message at INFO level.

#### (l *Logger) Warn(message string, args ...interface{})
Logs a message at WARN level.

#### (l *Logger) Error(message string, args ...interface{})
Logs a message at ERROR level.

#### (l *Logger) Fatal(message string, args ...interface{})
Logs a message at FATAL level, flushes all buffered logs, and exits the program.

#### (l *Logger) SetLogstashEnabled(enabled bool)
Enables or disables logging to Logstash.

#### (l *Logger) IsLogstashEnabled() bool
Returns the current status of Logstash logging.

#### (l *Logger) SetTimestampFormat(format string)
Sets the timestamp format for log entries. Use Go's time format layout.

#### (l *Logger) GetTimestampFormat() string
Returns the current timestamp format.

#### (l *Logger) Close() error
Closes the connection to Logstash.

#### (l *Logger) Flush()
Forces immediate flush of all buffered logs (async mode only).

#### (l *Logger) With(fields map[string]interface{}) *Logger
Returns a new logger instance with the specified fields added to the context. The fields will be included in all subsequent log entries from this logger. This method is thread-safe and creates a new logger instance, so the original logger remains unchanged.

#### Standard log.Logger Interface Methods
The logger implements the standard `log.Logger` interface for compatibility:
- `(l *Logger) Print(v ...interface{})` - Logs at INFO level
- `(l *Logger) Printf(format string, v ...interface{})` - Logs at INFO level with formatting
- `(l *Logger) Println(v ...interface{})` - Logs at INFO level
- `(l *Logger) Fatalf(format string, v ...interface{})` - Logs at FATAL level and exits
- `(l *Logger) Fatalln(v ...interface{})` - Logs at FATAL level and exits
- `(l *Logger) Panic(v ...interface{})` - Logs at ERROR level and panics
- `(l *Logger) Panicf(format string, v ...interface{})` - Logs at ERROR level with formatting and panics
- `(l *Logger) Panicln(v ...interface{})` - Logs at ERROR level and panics
- `(l *Logger) SetOutput(w io.Writer)` - Sets the output writer
- `(l *Logger) Output() io.Writer` - Returns the current output writer
- `(l *Logger) SetFlags(flag int)` - Sets logging flags (for compatibility)
- `(l *Logger) Flags() int` - Returns current flags
- `(l *Logger) SetPrefix(prefix string)` - Sets a prefix for log messages
- `(l *Logger) Prefix() string` - Returns the current prefix

## Performance & Asynchronous Logging

### Synchronous vs Asynchronous Logging

**Synchronous Logging** (default):
- Each log call blocks until the message is sent to Logstash
- Guarantees immediate delivery but can slow down your application
- Network latency directly affects application performance

**Asynchronous Logging** (recommended for production):
- Log calls return immediately after adding to buffer
- Background goroutine processes logs in batches
- Dramatically improves application performance
- Configurable buffer size and flush intervals

### Performance Comparison

```go
// Synchronous: ~1000 logs/second (network dependent)
config := logger.Config{
    AsyncEnabled: false,
    // ... other settings
}

// Asynchronous: ~100,000+ logs/second
config := logger.Config{
    AsyncEnabled:  true,
    BufferSize:    2000,                   // larger buffer for high throughput
    FlushInterval: 100 * time.Millisecond, // frequent flushes for low latency
    // ... other settings
}
```

### Async Configuration Guidelines

**High Throughput Applications:**
```go
AsyncEnabled:  true,
BufferSize:    5000,                   // large buffer
FlushInterval: 1 * time.Second,        // less frequent flushes
```

**Low Latency Applications:**
```go
AsyncEnabled:  true,
BufferSize:    1000,                   // moderate buffer
FlushInterval: 100 * time.Millisecond, // frequent flushes
```

**Memory Constrained Applications:**
```go
AsyncEnabled:  true,
BufferSize:    500,                    // small buffer
FlushInterval: 500 * time.Millisecond, // moderate flushes
```

## Protocol Differences

### TCP (Transmission Control Protocol)
- **Reliable**: Guarantees message delivery and order
- **Connection-oriented**: Maintains persistent connection to Logstash
- **Automatic reconnection**: Reconnects automatically on connection loss with configurable retry logic
- **Connection monitoring**: Periodically checks connection health
- **Configurable retries**: Set maximum attempts and delay between reconnections
- **Higher overhead**: More network overhead due to connection management
- **Best for**: Production environments where log delivery is critical

### UDP (User Datagram Protocol)
- **Fast**: Lower latency, no connection overhead
- **Connectionless**: Creates new connection for each log message
- **No delivery guarantee**: Messages may be lost in network congestion
- **Lower overhead**: Minimal network overhead
- **Best for**: High-throughput scenarios where some log loss is acceptable

## Reconnection Configuration

The logger supports automatic reconnection for TCP connections with the following options:

### ReconnectAttempts
- **0**: Infinite reconnection attempts (default)
- **Positive number**: Maximum number of reconnection attempts
- **Example**: `ReconnectAttempts: 5` - try to reconnect up to 5 times

### ReconnectDelay
- **Default**: 5 seconds
- **Configurable**: Any `time.Duration` value
- **Example**: `ReconnectDelay: 3 * time.Second` - wait 3 seconds between attempts

### Example Configurations

```go
// Infinite reconnection attempts with 5-second delay
config := logger.Config{
    // ... other settings ...
    ReconnectAttempts: 0,                // infinite attempts
    ReconnectDelay:    5 * time.Second,  // 5 seconds between attempts
}

// Limited reconnection attempts with custom delay
config := logger.Config{
    // ... other settings ...
    ReconnectAttempts: 10,               // max 10 attempts
    ReconnectDelay:    2 * time.Second,  // 2 seconds between attempts
}

// Fast reconnection for high-availability scenarios
config := logger.Config{
    // ... other settings ...
    ReconnectAttempts: 50,                    // many attempts
    ReconnectDelay:    500 * time.Millisecond, // fast retry
}
```

## Timestamp Format Configuration

The logger supports configurable timestamp formats using Go's time format layout. You can set the format during initialization or change it at runtime.

### Default Format
The default timestamp format is: `"2006-01-02T15:04:05.000Z"` (ISO format with milliseconds)

### Common Timestamp Formats

```go
// ISO format with milliseconds (default)
TimestampFormat: "2006-01-02T15:04:05.000Z"
// Output: [2025-08-29T07:42:26.244Z]

// RFC3339 format
TimestampFormat: "2006-01-02T15:04:05Z07:00"
// Output: [2025-08-29T07:42:26Z]

// Simple date and time
TimestampFormat: "2006-01-02 15:04:05.000"
// Output: [2025-08-29 07:42:26.244]

// Time first format
TimestampFormat: "15:04:05 02/01/2006"
// Output: [07:42:26 29/08/2025]

// Human readable format
TimestampFormat: "Jan 2, 2006 15:04:05"
// Output: [Aug 29, 2025 07:42:26]

// Time only
TimestampFormat: "15:04:05"
// Output: [07:42:26]

// Date only
TimestampFormat: "02/01/2006"
// Output: [29/08/2025]
```

### Configuration Examples

```go
// Set format during initialization
config := logger.Config{
    TimestampFormat: "2006-01-02 15:04:05.000", // custom format
    // ... other settings
}
logger := logger.New(config)

// Change format at runtime
logger.SetTimestampFormat("15:04:05 02/01/2006")      // time first
logger.SetTimestampFormat("2006-01-02T15:04:05Z07:00") // RFC3339
logger.SetTimestampFormat("")                          // reset to default

// Get current format
currentFormat := logger.GetTimestampFormat()
fmt.Printf("Current format: %s", currentFormat)
```

### Go Time Format Reference

Go uses a unique approach to time formatting. Instead of using symbols like `%Y` or `%m`, it uses a reference time: **Mon Jan 2 15:04:05 MST 2006**, which is Unix time `1136239445`.

Common format components:
- `2006` - Year (4 digits)
- `06` - Year (2 digits)
- `01` - Month (2 digits)
- `1` - Month (1-2 digits)
- `Jan` - Month (abbreviated name)
- `January` - Month (full name)
- `02` - Day (2 digits)
- `2` - Day (1-2 digits)
- `15` - Hour (24-hour, 2 digits)
- `3` - Hour (12-hour, 1-2 digits)
- `03` - Hour (12-hour, 2 digits)
- `04` - Minute (2 digits)
- `4` - Minute (1-2 digits)
- `05` - Second (2 digits)
- `5` - Second (1-2 digits)
- `.000` - Milliseconds
- `.000000` - Microseconds
- `Z07:00` - Timezone offset
- `MST` - Timezone name

## Function Name Capture

The logger automatically captures and displays the name of the function that called the logging method. This provides better context and makes debugging easier by clearly showing the source of each log message.

### How It Works

The logger uses Go's `runtime.Caller()` and `runtime.FuncForPC()` to automatically determine the calling function name. This happens transparently without any additional configuration required.

### Console Output Format

The console output includes the function name in the following format:
```
[timestamp] LEVEL [function_name] [filename:line] message
```

### Examples

```go
func main() {
    log := logger.New(logger.DefaultConfig())
    log.Info("Application started")
    // Output: [2025-08-29T07:50:30.228Z] INFO [main] [main.go:10] Application started
    
    processData(log)
    handleError(log)
}

func processData(log *logger.Logger) {
    log.Debug("Starting data processing")
    log.Info("Data processed successfully")
    // Output: [2025-08-29T07:50:30.228Z] DEBUG [processData] [main.go:15] Starting data processing
    // Output: [2025-08-29T07:50:30.228Z] INFO [processData] [main.go:16] Data processed successfully
}

func handleError(log *logger.Logger) {
    log.Warn("Potential issue detected")
    log.Error("Error occurred during processing")
    // Output: [2025-08-29T07:50:30.228Z] WARN [handleError] [main.go:20] Potential issue detected
    // Output: [2025-08-29T07:50:30.228Z] ERROR [handleError] [main.go:21] Error occurred during processing
}
```

### Benefits

1. **Better Debugging**: Instantly see which function generated each log message
2. **Code Organization**: Logs are naturally organized by function context
3. **Zero Configuration**: Works automatically without any setup
4. **Performance**: Minimal overhead using Go's runtime introspection
5. **Reliability**: Graceful fallback to "unknown" or "main" if function name cannot be determined

### JSON Output

In JSON format (sent to Logstash), the function name is stored in the `logger_name` field:

```json
{
    "@timestamp": "2025-08-29T07:50:30.228Z",
    "level": "INFO",
    "message": "Processing business logic",
    "logger_name": "businessLogic",
    "thread_name": "main.go:11",
    "appname": "my-app",
    "hostip": "192.168.1.100",
    "containerId": "hostname",
    "type": "logback"
}
```

This allows for powerful filtering and analysis in your log aggregation system based on function names.

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details on how to contribute to this project.

### Development

```bash
# Clone the repository
git clone https://github.com/SergeyDavidenko/logger.git
cd logger

# Run tests
go test -v

# Run benchmarks
go test -bench=.

# Run example
cd example
go run main.go
```

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Inspired by the need for high-performance logging in Go applications
- Built for seamless integration with the ELK Stack (Elasticsearch, Logstash, Kibana)
- Designed with production reliability and developer experience in mind

## 📊 Benchmarks

Performance comparison with popular Go logging libraries. All benchmarks use `io.Discard` to suppress output and async logging is disabled for fair comparison.

### Log a message and 10 fields

Benchmark results for logging a message with 10 fields.

| Package | Time | Time % to zap | Objects Allocated |
|---------|------|---------------|-------------------|
| ⚡ zerolog | 44.23 ns/op | -77% | 0 allocs/op |
| ⚡ zap | 190.0 ns/op | +0% | 0 allocs/op |
| apex/log | 409.3 ns/op | +115% | 6 allocs/op |
| **logger (this package)** | **670.2 ns/op** | **+253%** | **4 allocs/op** |
| zap (sugared) | 704.9 ns/op | +271% | 1 allocs/op |
| slog (LogAttrs) | 742.4 ns/op | +291% | 1 allocs/op |
| slog | 773.3 ns/op | +307% | 1 allocs/op |
| go-kit | 1486 ns/op | +682% | 29 allocs/op |
| log15 | 2593 ns/op | +1265% | 42 allocs/op |
| logrus | 2700 ns/op | +1321% | 52 allocs/op |

**Analysis:**
- This logger ranks **4th** out of 10 tested loggers
- Execution time: 670.2 ns/op
- Memory allocations: 4 allocs/op
- Faster than slog, zap (sugared), go-kit, log15, and logrus
- Comparable performance to slog and zap (sugared)
- **20% fewer allocations** compared to original version

### Log a message with a logger that already has 10 fields of context

Benchmark results for logging a message when the logger already has 10 fields of context pre-configured using `With()`.

| Package | Time | Time % to zap | Objects Allocated |
|---------|------|---------------|-------------------|
| ⚡ zerolog | 45.05 ns/op | -77% | 0 allocs/op |
| ⚡ zap | 192.1 ns/op | +0% | 0 allocs/op |
| ⚡ zap (sugared) | 196.8 ns/op | +2% | 0 allocs/op |
| slog | 295.7 ns/op | +54% | 0 allocs/op |
| slog (LogAttrs) | 297.5 ns/op | +55% | 0 allocs/op |
| apex/log | 404.1 ns/op | +110% | 6 allocs/op |
| **logger (this package)** | **432.9 ns/op** | **+125%** | **3 allocs/op** |
| go-kit | 1374 ns/op | +616% | 29 allocs/op |
| logrus | 2214 ns/op | +1053% | 43 allocs/op |
| log15 | 2741 ns/op | +1327% | 41 allocs/op |

**Analysis:**
- This logger ranks **7th** out of 10 tested loggers when logging with pre-configured context
- Execution time: 432.9 ns/op
- Memory allocations: 3 allocs/op
- Faster than go-kit, log15, and logrus
- Comparable performance to apex/log (only 7% slower)
- **40% fewer allocations** compared to original version
- Full support for structured logging with `With()` method

### Running Benchmarks

```bash
cd benchmarks
go test -bench=. -benchmem -run=XXX
```

To run only this logger:
```bash
go test -bench=BenchmarkLogger -benchmem -run=XXX
```

For detailed benchmark results and analysis, see [benchmarks/README.md](benchmarks/README.md).

*Benchmarks run on Go 1.25.0, macOS, Apple M3 Pro (12 cores), 18 GB RAM*

---

**Made with ❤️ for the Go community**
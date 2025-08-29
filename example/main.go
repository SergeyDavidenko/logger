package main

import (
	"fmt"
	"time"

	"github.com/SergeyDavidenko/logger"
)

// Example functions to demonstrate function name capture
func businessLogic(log *logger.Logger) {
	log.Info("Processing business logic")
	log.Warn("Business logic warning")
}

func databaseOperation(log *logger.Logger) {
	log.Debug("Connecting to database")
	log.Info("Database operation completed successfully")
}

func handleError(log *logger.Logger) {
	log.Error("An error occurred in error handler")
}

func main() {
	// Demonstrate parsing log level from string
	logLevelStr := "INFO" // This could come from config file or environment variable
	minLevel, ok := logger.ParseLogLevel(logLevelStr)
	if !ok {
		// Fallback to default level if parsing fails
		minLevel = logger.INFO
		fmt.Printf("Warning: Invalid log level '%s', using INFO\n", logLevelStr)
	}
	fmt.Printf("Using log level: %s\n", minLevel.String())

	// Create logger configuration with async logging
	config := logger.Config{
		TimestampFormat:   "2006-01-02 15:04:05.000", // custom timestamp format
		LogstashHost:      "localhost",               // logstash server address
		LogstashPort:      5008,                      // port from logstash configuration
		LogstashEnabled:   true,                      // enable sending to logstash
		Protocol:          logger.TCP,                // use TCP protocol (can be logger.UDP)
		AppName:           "example-app",
		MinLevel:          minLevel,        // use parsed minimum logging level
		ReconnectAttempts: 5,               // max 5 reconnection attempts (0 = infinite)
		ReconnectDelay:    3 * time.Second, // 3 seconds between reconnection attempts
		// Async logging configuration
		AsyncEnabled:  true,                   // enable async logging for better performance
		BufferSize:    2000,                   // buffer up to 2000 log entries
		FlushInterval: 500 * time.Millisecond, // flush every 500ms
	}

	// Create logger
	log := logger.New(config)
	defer log.Close() // close connection on exit

	// Examples of using different logging levels
	log.Debug("This is a debug message")
	log.Info("Application started successfully with custom timestamp format")
	log.Warn("This is a warning")
	log.Error("An error occurred: %s", "example error")

	// Demonstrate function name capture
	log.Info("=== Function Name Capture Demonstration ===")
	businessLogic(log)
	databaseOperation(log)
	handleError(log)

	// Demonstrate timestamp format configuration
	log.Info("Current timestamp format: %s", log.GetTimestampFormat())

	// Change timestamp format during runtime
	log.SetTimestampFormat("15:04:05 02/01/2006") // time first, then date
	log.Info("Changed timestamp format to time-first format")

	// Try another format
	log.SetTimestampFormat("2006-01-02T15:04:05Z07:00") // RFC3339 format
	log.Info("Changed to RFC3339 timestamp format")

	// Back to custom format
	log.SetTimestampFormat("2006-01-02 15:04:05.000")
	log.Info("Back to original custom format")

	// Demonstration of enabling/disabling logstash
	log.Info("Message with logstash enabled")

	log.SetLogstashEnabled(false)
	log.Info("Message with logstash disabled (console only)")

	log.SetLogstashEnabled(true)
	log.Info("Message with logstash enabled again")

	// Examples with formatting
	userName := "John"
	userID := 12345
	log.Info("User %s (ID: %d) logged into the system", userName, userID)

	// Example with multiple parameters
	log.Debug("Debug information: status=%s, code=%d, time=%v",
		"active", 200, time.Now())

	// Check logstash status
	if log.IsLogstashEnabled() {
		log.Info("Logstash is enabled (TCP) with auto-reconnect and async logging")
	} else {
		log.Info("Logstash is disabled")
	}

	// Demonstrate connection resilience
	log.Info("Logger will automatically reconnect if connection is lost")
	log.Info("Reconnect attempts: max 5, delay: 3 seconds")

	// Demonstrate async performance
	log.Info("Async logging: buffer size 2000, flush interval 500ms")

	// High-performance logging demo
	log.Info("Starting high-performance logging demo...")
	start := time.Now()
	for i := 0; i < 1000; i++ {
		log.Debug("High-performance log entry #%d", i)
	}
	elapsed := time.Since(start)
	log.Info("Logged 1000 entries in %v (async buffering)", elapsed)

	// Demonstrate UDP protocol usage (no reconnection for UDP)
	log.Info("Switching to UDP protocol demonstration...")

	udpConfig := logger.Config{
		LogstashHost:    "localhost",
		LogstashPort:    5009,  // different port for UDP (would need UDP input in logstash)
		LogstashEnabled: false, // disable for demo since we don't have UDP input configured
		Protocol:        logger.UDP,
		AppName:         "example-app-udp",
		MinLevel:        logger.INFO,
	}

	udpLog := logger.New(udpConfig)
	defer udpLog.Close()

	udpLog.Info("UDP protocol: no persistent connection, no auto-reconnect")

	// Demonstrate manual flush
	log.Info("Forcing flush of all buffered logs...")
	log.Flush() // Force immediate flush
	log.Info("All logs flushed!")

	// Demonstrate log level parsing functionality
	log.Info("=== Log Level Parsing Demonstration ===")

	// Test various log level strings
	testLevels := []string{"DEBUG", "info", "WARN", "error", "FATAL", "warning", "invalid"}
	for _, levelStr := range testLevels {
		if parsedLevel, ok := logger.ParseLogLevel(levelStr); ok {
			log.Info("Successfully parsed '%s' -> %s", levelStr, parsedLevel.String())
		} else {
			log.Warn("Failed to parse log level: '%s', defaulted to INFO", levelStr)
		}
	}

	// Demonstrate MustParseLogLevel (safe usage)
	log.Info("=== MustParseLogLevel Demonstration ===")
	safeLevel := logger.MustParseLogLevel("ERROR")
	log.Info("MustParseLogLevel('ERROR') = %s", safeLevel.String())

	// Small pause to send all logs
	time.Sleep(100 * time.Millisecond)
}

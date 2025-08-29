package logger

import (
	"testing"
	"time"
)

func TestLogger_New(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false, // disable for test
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	if logger.appName != "test-app" {
		t.Errorf("Expected app name 'test-app', got '%s'", logger.appName)
	}

	if logger.logstashEnabled {
		t.Error("Logstash should be disabled")
	}

	if logger.protocol != TCP {
		t.Errorf("Expected protocol TCP, got %s", logger.protocol)
	}

	logger.Close()
}

func TestLogger_NewWithUDP(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5009,
		LogstashEnabled: false, // disable for test
		Protocol:        UDP,
		AppName:         "test-app-udp",
		MinLevel:        INFO,
	}

	logger := New(config)
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	if logger.protocol != UDP {
		t.Errorf("Expected protocol UDP, got %s", logger.protocol)
	}

	logger.Close()
}

func TestLogger_DefaultProtocol(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false,
		// Protocol not specified - should default to TCP
		AppName:  "test-app",
		MinLevel: DEBUG,
	}

	logger := New(config)
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	if logger.protocol != TCP {
		t.Errorf("Expected default protocol TCP, got %s", logger.protocol)
	}

	logger.Close()
}

func TestLogger_ReconnectConfig(t *testing.T) {
	config := Config{
		LogstashHost:      "localhost",
		LogstashPort:      5008,
		LogstashEnabled:   false,
		Protocol:          TCP,
		AppName:           "test-app",
		MinLevel:          DEBUG,
		ReconnectAttempts: 3,
		ReconnectDelay:    2 * time.Second,
	}

	logger := New(config)
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	if logger.reconnectAttempts != 3 {
		t.Errorf("Expected reconnect attempts 3, got %d", logger.reconnectAttempts)
	}

	if logger.reconnectDelay != 2*time.Second {
		t.Errorf("Expected reconnect delay 2s, got %v", logger.reconnectDelay)
	}

	logger.Close()
}

func TestLogger_DefaultReconnectConfig(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false,
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		// ReconnectAttempts and ReconnectDelay not specified - should use defaults
	}

	logger := New(config)
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	if logger.reconnectAttempts != -1 { // -1 means infinite
		t.Errorf("Expected default reconnect attempts -1 (infinite), got %d", logger.reconnectAttempts)
	}

	if logger.reconnectDelay != 5*time.Second {
		t.Errorf("Expected default reconnect delay 5s, got %v", logger.reconnectDelay)
	}

	logger.Close()
}

func TestLogger_AsyncConfig(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false, // disable for test
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    true,
		BufferSize:      500,
		FlushInterval:   200 * time.Millisecond,
	}

	logger := New(config)
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	if !logger.asyncEnabled {
		t.Error("Async logging should be enabled")
	}

	if logger.bufferSize != 500 {
		t.Errorf("Expected buffer size 500, got %d", logger.bufferSize)
	}

	if logger.flushInterval != 200*time.Millisecond {
		t.Errorf("Expected flush interval 200ms, got %v", logger.flushInterval)
	}

	logger.Close()
}

func TestLogger_AsyncDefaults(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false,
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    true,
		// BufferSize and FlushInterval not specified - should use defaults
	}

	logger := New(config)
	if logger == nil {
		t.Fatal("Logger should not be nil")
	}

	if logger.bufferSize != 1000 {
		t.Errorf("Expected default buffer size 1000, got %d", logger.bufferSize)
	}

	if logger.flushInterval != 1*time.Second {
		t.Errorf("Expected default flush interval 1s, got %v", logger.flushInterval)
	}

	logger.Close()
}

func TestLogger_AsyncPerformance(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false, // disable actual sending for test
		Protocol:        TCP,
		AppName:         "perf-test",
		MinLevel:        DEBUG,
		AsyncEnabled:    true,
		BufferSize:      1000,
		FlushInterval:   100 * time.Millisecond,
	}

	logger := New(config)
	defer logger.Close()

	// Test that async logging is much faster
	start := time.Now()
	for i := 0; i < 100; i++ {
		logger.Info("Performance test message %d", i)
	}
	elapsed := time.Since(start)

	// Async logging should be very fast (under 10ms for 100 messages)
	if elapsed > 10*time.Millisecond {
		t.Errorf("Async logging too slow: %v (expected < 10ms)", elapsed)
	}

	// Wait a bit and then flush to ensure all logs are processed
	time.Sleep(200 * time.Millisecond)
	logger.Flush()
}

func TestLogger_SetLogstashEnabled(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false,
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// Check initial state
	if logger.IsLogstashEnabled() {
		t.Error("Logstash should be initially disabled")
	}

	// Enable logstash
	logger.SetLogstashEnabled(true)
	if !logger.IsLogstashEnabled() {
		t.Error("Logstash should be enabled")
	}

	// Disable logstash
	logger.SetLogstashEnabled(false)
	if logger.IsLogstashEnabled() {
		t.Error("Logstash should be disabled")
	}

	// Give goroutines time to cleanup
	time.Sleep(10 * time.Millisecond)
}

func TestLogger_LogLevels(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        INFO, // set minimum level to INFO
	}

	logger := New(config)
	defer logger.Close()

	// These methods should not cause panic
	logger.Debug("Debug message") // should not be logged due to minimum level
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")
}

func TestLogger_MessageFormatting(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// Test message formatting
	logger.Info("Simple message")
	logger.Info("Formatted message: %s", "test")
	logger.Info("Multiple params: %s, %d, %v", "string", 42, time.Now())
}

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
		{LogLevel(999), "UNKNOWN"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.level.String())
		}
	}
}

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		input       string
		expectedLvl LogLevel
		expectedOk  bool
	}{
		{"DEBUG", DEBUG, true},
		{"debug", DEBUG, true},
		{"  DEBUG  ", DEBUG, true},
		{"INFO", INFO, true},
		{"info", INFO, true},
		{"WARN", WARN, true},
		{"warn", WARN, true},
		{"WARNING", WARN, true},
		{"warning", WARN, true},
		{"ERROR", ERROR, true},
		{"error", ERROR, true},
		{"FATAL", FATAL, true},
		{"fatal", FATAL, true},
		{"INVALID", INFO, false},
		{"", INFO, false},
		{"trace", INFO, false},
	}

	for _, test := range tests {
		level, ok := ParseLogLevel(test.input)
		if level != test.expectedLvl || ok != test.expectedOk {
			t.Errorf("ParseLogLevel(%q) = (%v, %v), expected (%v, %v)",
				test.input, level, ok, test.expectedLvl, test.expectedOk)
		}
	}
}

func TestMustParseLogLevel(t *testing.T) {
	// Test valid levels
	validTests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", DEBUG},
		{"info", INFO},
		{"WARN", WARN},
		{"error", ERROR},
		{"FATAL", FATAL},
	}

	for _, test := range validTests {
		level := MustParseLogLevel(test.input)
		if level != test.expected {
			t.Errorf("MustParseLogLevel(%q) = %v, expected %v", test.input, level, test.expected)
		}
	}

	// Test invalid level (should panic)
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustParseLogLevel with invalid level should panic")
		}
	}()
	MustParseLogLevel("INVALID")
}

func BenchmarkLogger_Info(b *testing.B) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false, // disable for benchmark
		AppName:         "bench-app",
		MinLevel:        INFO,
	}

	logger := New(config)
	defer logger.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("Benchmark message %d", i)
	}
}

func TestLogger_TimestampFormat(t *testing.T) {
	// Test default timestamp format
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// Default format should be ISO format
	expectedDefault := "2006-01-02T15:04:05.000Z"
	if logger.GetTimestampFormat() != expectedDefault {
		t.Errorf("Expected default timestamp format '%s', got '%s'", expectedDefault, logger.GetTimestampFormat())
	}

	// Test custom timestamp format in config
	customFormat := "2006-01-02 15:04:05.000"
	configWithCustom := Config{
		TimestampFormat: customFormat,
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	loggerCustom := New(configWithCustom)
	defer loggerCustom.Close()

	if loggerCustom.GetTimestampFormat() != customFormat {
		t.Errorf("Expected custom timestamp format '%s', got '%s'", customFormat, loggerCustom.GetTimestampFormat())
	}

	// Test setting timestamp format at runtime
	newFormat := "15:04:05 02/01/2006"
	logger.SetTimestampFormat(newFormat)

	if logger.GetTimestampFormat() != newFormat {
		t.Errorf("Expected updated timestamp format '%s', got '%s'", newFormat, logger.GetTimestampFormat())
	}

	// Test setting empty format (should use default)
	logger.SetTimestampFormat("")
	if logger.GetTimestampFormat() != expectedDefault {
		t.Errorf("Expected default timestamp format after setting empty '%s', got '%s'", expectedDefault, logger.GetTimestampFormat())
	}

	// Test various timestamp formats
	testFormats := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02 15:04:05",       // Simple datetime
		"15:04:05",                  // Time only
		"02/01/2006",                // Date only
		"Jan 2, 2006 15:04:05",      // Human readable
	}

	for _, format := range testFormats {
		logger.SetTimestampFormat(format)
		if logger.GetTimestampFormat() != format {
			t.Errorf("Expected timestamp format '%s', got '%s'", format, logger.GetTimestampFormat())
		}
	}
}

func TestLogger_FunctionNameCapture(t *testing.T) {
	// We'll capture the log entry to verify the LoggerName field
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	log := New(config)
	defer log.Close()

	// This test function should have its name captured
	log.Info("Testing function name capture")

	// Note: This is more of an integration test to ensure the code doesn't crash
	// The actual function name capture is demonstrated in the example
}

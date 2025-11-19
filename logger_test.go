package logger

import (
	"bytes"
	"io"
	"log"
	"os"
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

// Standard logging interface compatibility tests

func TestLogger_StandardLogInterface(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// Test that our logger implements the basic interface methods
	// Note: Fatal method has different signature than standard log, so we test separately
	var _ interface {
		Print(v ...interface{})
		Printf(format string, v ...interface{})
		Println(v ...interface{})
		Fatalf(format string, v ...interface{})
		Fatalln(v ...interface{})
		Panic(v ...interface{})
		Panicf(format string, v ...interface{})
		Panicln(v ...interface{})
		SetOutput(w io.Writer)
		SetFlags(flag int)
		Flags() int
		SetPrefix(prefix string)
		Prefix() string
	} = logger

	// Test Print methods (these should not panic)
	logger.Print("test message")
	logger.Printf("test %s", "formatted")
	logger.Println("test", "println")
}

func TestLogger_SetOutput(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// Test default output (compare types rather than exact instances)
	defaultOutput := logger.Output()
	if defaultOutput == nil {
		t.Error("Default output should not be nil")
	}

	// Test setting custom output
	var buf bytes.Buffer
	logger.SetOutput(&buf)

	currentOutput := logger.Output()
	if currentOutput != &buf {
		t.Errorf("Output should be set to custom buffer, got %T", currentOutput)
	}

	// Test that we can set back to os.Stderr
	logger.SetOutput(os.Stderr)
	if logger.Output() != os.Stderr {
		t.Error("Should be able to set output back to os.Stderr")
	}
}

func TestLogger_PrefixAndFlags(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// Test default prefix and flags
	if logger.Prefix() != "" {
		t.Errorf("Default prefix should be empty, got '%s'", logger.Prefix())
	}

	if logger.Flags() != 0 {
		t.Errorf("Default flags should be 0, got %d", logger.Flags())
	}

	// Test setting prefix
	testPrefix := "TEST: "
	logger.SetPrefix(testPrefix)
	if logger.Prefix() != testPrefix {
		t.Errorf("Expected prefix '%s', got '%s'", testPrefix, logger.Prefix())
	}

	// Test setting flags
	testFlags := log.LstdFlags | log.Lshortfile
	logger.SetFlags(testFlags)
	if logger.Flags() != testFlags {
		t.Errorf("Expected flags %d, got %d", testFlags, logger.Flags())
	}
}

func TestLogger_PanicMethods(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// Test Panic method
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Panic() should have panicked")
			} else {
				expected := "test panic message"
				if r != expected {
					t.Errorf("Expected panic message '%s', got '%v'", expected, r)
				}
			}
		}()
		logger.Panic("test panic message")
	}()

	// Test Panicf method
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Panicf() should have panicked")
			} else {
				expected := "formatted panic: test"
				if r != expected {
					t.Errorf("Expected panic message '%s', got '%v'", expected, r)
				}
			}
		}()
		logger.Panicf("formatted panic: %s", "test")
	}()

	// Test Panicln method
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Panicln() should have panicked")
			} else {
				expected := "panic line test"
				if r != expected {
					t.Errorf("Expected panic message '%s', got '%v'", expected, r)
				}
			}
		}()
		logger.Panicln("panic", "line", "test")
	}()
}

func TestLogger_AsStandardLoggerReplacement(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// Test that we can use our logger where a standard logger is expected
	useStandardLogger := func(l interface {
		Printf(format string, v ...interface{})
		Println(v ...interface{})
	}) {
		l.Printf("Standard logger test: %s", "success")
		l.Println("Standard logger println test")
	}

	// This should not cause any compilation errors or runtime panics
	useStandardLogger(logger)
}

func TestLogger_PrintMethods_MessageFormatting(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// Test that Print methods format messages correctly
	// These should not panic and should handle various input types
	logger.Print("simple")
	logger.Print("multiple", "arguments", 123, true)
	logger.Printf("formatted %s %d %v", "string", 42, []int{1, 2, 3})
	logger.Println("println", "with", "multiple", "args")

	// Test edge cases
	logger.Print()          // empty args
	logger.Printf("")       // empty format
	logger.Println()        // empty args
	logger.Printf("%s", "") // empty string
}

func TestLogger_FatalMethods_DoNotExit(t *testing.T) {
	// Note: We can't easily test Fatal methods because they call os.Exit(1)
	// In a real scenario, you would need to test these in a separate process
	// or use dependency injection to mock os.Exit

	// This test just documents the expected behavior
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// These methods exist and have the correct signatures
	_ = logger.Fatal
	_ = logger.Fatalf
	_ = logger.Fatalln

	// In real usage, these would call os.Exit(1):
	// logger.Fatal("fatal message")
	// logger.Fatalf("fatal %s", "formatted")
	// logger.Fatalln("fatal", "line")
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.TimestampFormat != "2006-01-02T15:04:05.000Z" {
		t.Errorf("Expected default timestamp format, got %s", config.TimestampFormat)
	}

	if config.LogstashEnabled != false {
		t.Error("Expected LogstashEnabled to be false by default")
	}

	if config.AsyncEnabled != true {
		t.Error("Expected AsyncEnabled to be true by default")
	}

	if config.BufferSize != 1000 {
		t.Errorf("Expected default buffer size 1000, got %d", config.BufferSize)
	}

	if config.FlushInterval != 1*time.Second {
		t.Errorf("Expected default flush interval 1s, got %v", config.FlushInterval)
	}
}

func TestLogger_LogstashUDP(t *testing.T) {
	// Test UDP logging (even if connection fails, the code path should be executed)
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5009,
		LogstashEnabled: true,
		Protocol:        UDP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    false, // Use sync to test UDP path
	}

	logger := New(config)
	defer logger.Close()

	// This should attempt to send via UDP (will fail but code path is executed)
	logger.Info("Test UDP message")
	time.Sleep(10 * time.Millisecond) // Give time for async operations
}

func TestLogger_LogstashAsync(t *testing.T) {
	// Test async logging with Logstash enabled
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: true,
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    true,
		BufferSize:      100,
		FlushInterval:   50 * time.Millisecond,
	}

	logger := New(config)
	defer logger.Close()

	// Send multiple messages to test async buffer
	for i := 0; i < 10; i++ {
		logger.Info("Async test message %d", i)
	}

	// Wait for async processing
	time.Sleep(100 * time.Millisecond)
	logger.Flush()
}

func TestLogger_Flush(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: true,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    true,
		BufferSize:      100,
		FlushInterval:   1 * time.Second,
	}

	logger := New(config)
	defer logger.Close()

	// Flush when buffer is empty
	logger.Flush()

	// Send messages to fill buffer
	for i := 0; i < 5; i++ {
		logger.Info("Flush test message %d", i)
	}

	// Flush should wait for buffer to be empty
	logger.Flush()
}

func TestLogger_Close(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: true,
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    true,
		BufferSize:      100,
		FlushInterval:   50 * time.Millisecond,
	}

	logger := New(config)

	// Send some messages
	logger.Info("Message before close")
	logger.Warn("Another message")

	// Close should flush and cleanup
	err := logger.Close()
	if err != nil {
		t.Errorf("Close should not return error when connection is nil or closed, got: %v", err)
	}

	// Test closing already closed logger
	err = logger.Close()
	if err != nil {
		t.Errorf("Close should not return error on second call, got: %v", err)
	}
}

func TestLogger_ReconnectLogic(t *testing.T) {
	config := Config{
		LogstashHost:      "localhost",
		LogstashPort:      5008,
		LogstashEnabled:   true,
		Protocol:          TCP,
		AppName:           "test-app",
		MinLevel:          DEBUG,
		ReconnectAttempts: 2,
		ReconnectDelay:    100 * time.Millisecond,
		AsyncEnabled:      false,
	}

	logger := New(config)
	defer logger.Close()

	// Reconnect monitor is started automatically when LogstashEnabled is true
	// Give it time to start and attempt reconnection
	time.Sleep(200 * time.Millisecond)

	// Send a log message which will trigger reconnection logic if connection is broken
	logger.Info("Test message to trigger reconnect logic")
	time.Sleep(200 * time.Millisecond)
}

func TestLogger_LogstashSync(t *testing.T) {
	// Test sync logging with Logstash enabled
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: true,
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    false, // Use sync mode
	}

	logger := New(config)
	defer logger.Close()

	// This should attempt to send via TCP (will fail but code path is executed)
	logger.Info("Test sync TCP message")
	time.Sleep(10 * time.Millisecond)
}

func TestLogger_LogLevelFiltering(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        WARN, // Only WARN and above
	}

	logger := New(config)
	defer logger.Close()

	// These should not be logged
	logger.Debug("Debug message")
	logger.Info("Info message")

	// These should be logged
	logger.Warn("Warning message")
	logger.Error("Error message")
}

func TestLogger_EmptyMessage(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// Test empty messages
	logger.Info("")
	logger.Info("%s", "")
	logger.Debug("")
	logger.Warn("")
	logger.Error("")
}

func TestLogger_ConnectToLogstash_EmptyHost(t *testing.T) {
	config := Config{
		LogstashHost:    "",
		LogstashPort:    0,
		LogstashEnabled: true,
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// connectToLogstash is called internally, but with empty host/port it should fail
	// We test this by enabling logstash which will try to connect
	logger.SetLogstashEnabled(true)
	time.Sleep(10 * time.Millisecond)

	// Connection should fail silently (no panic)
	logger.Info("Test message")
	time.Sleep(10 * time.Millisecond)
}

func TestLogger_SetLogstashEnabled_WithConnection(t *testing.T) {
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

	// Enable logstash - should attempt connection
	logger.SetLogstashEnabled(true)
	time.Sleep(10 * time.Millisecond)

	// Disable logstash - should close connection
	logger.SetLogstashEnabled(false)
	time.Sleep(10 * time.Millisecond)

	// Re-enable to test reconnection path
	logger.SetLogstashEnabled(true)
	time.Sleep(10 * time.Millisecond)
}

func TestLogger_AsyncBufferFull(t *testing.T) {
	// Test async buffer full scenario
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: true,
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    true,
		BufferSize:      5, // Small buffer to test full scenario
		FlushInterval:   100 * time.Millisecond,
	}

	logger := New(config)
	defer logger.Close()

	// Fill buffer quickly to trigger fallback to sync
	for i := 0; i < 10; i++ {
		logger.Info("Buffer test %d", i)
	}

	time.Sleep(200 * time.Millisecond)
	logger.Flush()
}

func TestLogger_Flush_WhenAsyncDisabled(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    false, // Async disabled
	}

	logger := New(config)
	defer logger.Close()

	// Flush should return immediately when async is disabled
	logger.Flush()
}

func TestLogger_CheckAndReconnect_WhenDisabled(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: false, // Disabled
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// checkAndReconnect should return early when logstash is disabled
	// We can't call it directly, but we can test the path through reconnect monitor
	// by enabling and then disabling
	logger.SetLogstashEnabled(true)
	time.Sleep(50 * time.Millisecond)
	logger.SetLogstashEnabled(false)
	time.Sleep(50 * time.Millisecond)
}

func TestLogger_CheckAndReconnect_UDP(t *testing.T) {
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5009,
		LogstashEnabled: true,
		Protocol:        UDP, // UDP protocol
		AppName:         "test-app",
		MinLevel:        DEBUG,
	}

	logger := New(config)
	defer logger.Close()

	// checkAndReconnect should return early for UDP protocol
	// Test by enabling/disabling to trigger reconnect monitor
	time.Sleep(100 * time.Millisecond)
	logger.Info("Test message")
	time.Sleep(50 * time.Millisecond)
}

func TestLogger_LogWithLogstashEnabled(t *testing.T) {
	// Test the full log path with logstash enabled
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: true,
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    false, // Use sync to test direct path
	}

	logger := New(config)
	defer logger.Close()

	// This will test logToLogstash -> logToLogstashSync -> logToLogstashTCP path
	logger.Debug("Debug with logstash")
	logger.Info("Info with logstash")
	logger.Warn("Warn with logstash")
	logger.Error("Error with logstash")
	time.Sleep(10 * time.Millisecond)
}

func TestLogger_LogWithLogstashAsync(t *testing.T) {
	// Test async log path
	config := Config{
		LogstashHost:    "localhost",
		LogstashPort:    5008,
		LogstashEnabled: true,
		Protocol:        TCP,
		AppName:         "test-app",
		MinLevel:        DEBUG,
		AsyncEnabled:    true,
		BufferSize:      100,
		FlushInterval:   50 * time.Millisecond,
	}

	logger := New(config)
	defer logger.Close()

	// This will test logToLogstash -> logToLogstashAsync path
	logger.Info("Async log message 1")
	logger.Warn("Async log message 2")
	logger.Error("Async log message 3")
	time.Sleep(100 * time.Millisecond)
	logger.Flush()
}

func TestLogger_LogLevelFiltering_EdgeCases(t *testing.T) {
	// Test various log level filtering scenarios
	tests := []struct {
		name      string
		minLevel  LogLevel
		shouldLog map[LogLevel]bool
	}{
		{
			name:     "DEBUG level",
			minLevel: DEBUG,
			shouldLog: map[LogLevel]bool{
				DEBUG: true,
				INFO:  true,
				WARN:  true,
				ERROR: true,
			},
		},
		{
			name:     "INFO level",
			minLevel: INFO,
			shouldLog: map[LogLevel]bool{
				DEBUG: false,
				INFO:  true,
				WARN:  true,
				ERROR: true,
			},
		},
		{
			name:     "WARN level",
			minLevel: WARN,
			shouldLog: map[LogLevel]bool{
				DEBUG: false,
				INFO:  false,
				WARN:  true,
				ERROR: true,
			},
		},
		{
			name:     "ERROR level",
			minLevel: ERROR,
			shouldLog: map[LogLevel]bool{
				DEBUG: false,
				INFO:  false,
				WARN:  false,
				ERROR: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				LogstashEnabled: false,
				AppName:         "test-app",
				MinLevel:        tt.minLevel,
			}

			logger := New(config)
			defer logger.Close()

			// Test each log level
			logger.Debug("debug message")
			logger.Info("info message")
			logger.Warn("warn message")
			logger.Error("error message")
		})
	}
}

func TestLogger_With(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        INFO,
	}

	baseLogger := New(config)
	defer baseLogger.Close()

	// Create logger with fields
	loggerWithFields := baseLogger.With(map[string]interface{}{
		"user_id":    123,
		"request_id": "req-456",
		"service":    "test-service",
	})

	// Capture output
	var buf bytes.Buffer
	loggerWithFields.SetOutput(&buf)

	// Log a message
	loggerWithFields.Info("test message")

	output := buf.String()
	if output == "" {
		t.Fatal("Expected output, got empty string")
	}

	// Check that output contains the message
	if !bytes.Contains([]byte(output), []byte("test message")) {
		t.Errorf("Expected output to contain 'test message', got: %s", output)
	}

	// Test that original logger is unchanged
	var buf2 bytes.Buffer
	baseLogger.SetOutput(&buf2)
	baseLogger.Info("original message")

	output2 := buf2.String()
	if bytes.Contains([]byte(output2), []byte("user_id")) {
		t.Error("Original logger should not have fields from With()")
	}
}

func TestLogger_WithNested(t *testing.T) {
	config := Config{
		LogstashEnabled: false,
		AppName:         "test-app",
		MinLevel:        INFO,
	}

	baseLogger := New(config)
	defer baseLogger.Close()

	// Create nested loggers with fields
	logger1 := baseLogger.With(map[string]interface{}{
		"level1": "value1",
	})

	logger2 := logger1.With(map[string]interface{}{
		"level2": "value2",
	})

	// Capture output
	var buf bytes.Buffer
	logger2.SetOutput(&buf)

	// Log a message
	logger2.Info("nested message")

	output := buf.String()
	if output == "" {
		t.Fatal("Expected output, got empty string")
	}

	// Check that output contains the message
	if !bytes.Contains([]byte(output), []byte("nested message")) {
		t.Errorf("Expected output to contain 'nested message', got: %s", output)
	}
}

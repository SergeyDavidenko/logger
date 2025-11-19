package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// ParseLogLevel parses a string and returns the corresponding LogLevel
// Accepts: "DEBUG", "INFO", "WARN", "ERROR", "FATAL" (case insensitive)
// Returns the LogLevel and a boolean indicating if parsing was successful
func ParseLogLevel(level string) (LogLevel, bool) {
	switch strings.ToUpper(strings.TrimSpace(level)) {
	case "DEBUG":
		return DEBUG, true
	case "INFO":
		return INFO, true
	case "WARN", "WARNING":
		return WARN, true
	case "ERROR":
		return ERROR, true
	case "FATAL":
		return FATAL, true
	default:
		return INFO, false // default to INFO if parsing fails
	}
}

// MustParseLogLevel parses a string and returns the corresponding LogLevel
// Panics if the level string is invalid
func MustParseLogLevel(level string) LogLevel {
	if logLevel, ok := ParseLogLevel(level); ok {
		return logLevel
	}
	panic(fmt.Sprintf("invalid log level: %s", level))
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp   string                 `json:"@timestamp"`
	Level       string                 `json:"level"`
	Message     string                 `json:"message"`
	LoggerName  string                 `json:"logger_name"`
	ThreadName  string                 `json:"thread_name"`
	AppName     string                 `json:"appname"`
	HostIP      string                 `json:"hostip"`
	ContainerID string                 `json:"containerId"`
	Type        string                 `json:"type"`
	Fields      map[string]interface{} `json:"fields,omitempty"`
}

// Logger is the main logger structure
type Logger struct {
	mu                sync.RWMutex
	logstashEnabled   bool
	logstashHost      string
	logstashPort      int
	protocol          Protocol
	conn              net.Conn
	appName           string
	hostIP            string
	containerID       string
	minLevel          LogLevel
	reconnectAttempts int
	reconnectDelay    time.Duration
	ctx               context.Context
	cancel            context.CancelFunc
	reconnectTicker   *time.Ticker
	timestampFormat   string
	// Async logging
	logBuffer     chan LogEntry
	bufferSize    int
	flushInterval time.Duration
	asyncEnabled  bool
	// Standard log interface compatibility
	output io.Writer
	prefix string
	flags  int
	// Buffer pool for optimization
	bufPool sync.Pool
	// Context fields for structured logging
	fields map[string]interface{}
}

// Protocol represents the network protocol type
type Protocol string

const (
	TCP Protocol = "tcp"
	UDP Protocol = "udp"
)

// Config represents logger configuration
type Config struct {
	TimestampFormat   string
	LogstashHost      string
	LogstashPort      int
	LogstashEnabled   bool
	Protocol          Protocol // TCP or UDP
	AppName           string
	MinLevel          LogLevel
	ReconnectAttempts int           // Maximum reconnection attempts (0 = infinite)
	ReconnectDelay    time.Duration // Delay between reconnection attempts
	// Async logging configuration
	AsyncEnabled  bool          // Enable asynchronous logging
	BufferSize    int           // Size of the log buffer (default: 1000)
	FlushInterval time.Duration // Interval to flush logs (default: 1 second)
}

// ConfigDefaults returns the default configuration
func DefaultConfig() Config {
	return Config{
		TimestampFormat: "2006-01-02T15:04:05.000Z",
		LogstashEnabled: false,
		AsyncEnabled:    true,
		BufferSize:      1000,
		FlushInterval:   1 * time.Second,
	}
}

// New creates a new logger instance
func New(config Config) *Logger {
	// Set default protocol to TCP if not specified
	protocol := config.Protocol
	if protocol == "" {
		protocol = TCP
	}

	// Set default reconnection parameters
	reconnectAttempts := config.ReconnectAttempts
	if reconnectAttempts == 0 {
		reconnectAttempts = -1 // infinite attempts
	}

	reconnectDelay := config.ReconnectDelay
	if reconnectDelay == 0 {
		reconnectDelay = 5 * time.Second // default 5 seconds
	}

	// Set default async parameters
	bufferSize := config.BufferSize
	if bufferSize == 0 {
		bufferSize = 1000 // default buffer size
	}

	flushInterval := config.FlushInterval
	if flushInterval == 0 {
		flushInterval = 1 * time.Second // default flush interval
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Set default timestamp format if not specified
	timestampFormat := config.TimestampFormat
	if timestampFormat == "" {
		timestampFormat = "2006-01-02T15:04:05.000Z"
	}

	logger := &Logger{
		logstashEnabled:   config.LogstashEnabled,
		logstashHost:      config.LogstashHost,
		logstashPort:      config.LogstashPort,
		protocol:          protocol,
		appName:           config.AppName,
		minLevel:          config.MinLevel,
		reconnectAttempts: reconnectAttempts,
		reconnectDelay:    reconnectDelay,
		ctx:               ctx,
		cancel:            cancel,
		timestampFormat:   timestampFormat,
		// Async configuration
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		asyncEnabled:  config.AsyncEnabled,
		// Standard log interface compatibility
		output: os.Stderr,
		prefix: "",
		flags:  0,
		// Initialize buffer pool
		bufPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
		// Initialize fields map
		fields: make(map[string]interface{}),
	}

	// Initialize async logging if enabled
	if logger.asyncEnabled && logger.logstashEnabled {
		logger.logBuffer = make(chan LogEntry, bufferSize)
		logger.startAsyncProcessor()
	}

	// Get host IP address
	logger.hostIP = getHostIP()

	// Get container ID (can be extended for Docker)
	logger.containerID = getContainerID()

	if logger.logstashEnabled && logger.protocol == TCP {
		logger.connectToLogstash()
		logger.startReconnectMonitor()
	}

	return logger
}

// SetLogstashEnabled enables/disables logging to logstash
func (l *Logger) SetLogstashEnabled(enabled bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.logstashEnabled = enabled

	if enabled && l.protocol == TCP {
		if l.conn == nil {
			l.connectToLogstash()
		}
		if l.reconnectTicker == nil {
			l.startReconnectMonitorUnsafe()
		}
	} else if !enabled {
		if l.conn != nil {
			l.conn.Close()
			l.conn = nil
		}
		if l.reconnectTicker != nil {
			l.reconnectTicker.Stop()
			l.reconnectTicker = nil
		}
	}
}

// IsLogstashEnabled returns the status of logstash logging
func (l *Logger) IsLogstashEnabled() bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.logstashEnabled
}

// SetTimestampFormat sets the timestamp format for log entries
func (l *Logger) SetTimestampFormat(format string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if format == "" {
		format = "2006-01-02T15:04:05.000Z" // default format
	}
	l.timestampFormat = format
}

// GetTimestampFormat returns the current timestamp format
func (l *Logger) GetTimestampFormat() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.timestampFormat
}

// connectToLogstash establishes connection to logstash
func (l *Logger) connectToLogstash() bool {
	if l.logstashHost == "" || l.logstashPort == 0 {
		return false
	}

	address := net.JoinHostPort(l.logstashHost, fmt.Sprintf("%d", l.logstashPort))
	conn, err := net.Dial(string(l.protocol), address)
	if err != nil {
		fmt.Printf("Failed to connect to logstash at %s via %s: %v\n", address, l.protocol, err)
		return false
	}

	l.conn = conn
	return true
}

// startReconnectMonitor starts a background goroutine to monitor connection health
func (l *Logger) startReconnectMonitor() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.startReconnectMonitorUnsafe()
}

// startReconnectMonitorUnsafe starts reconnect monitor without locking (must be called with mutex held)
func (l *Logger) startReconnectMonitorUnsafe() {
	if l.protocol != TCP || l.reconnectTicker != nil {
		return // Only for TCP connections and if not already started
	}

	l.reconnectTicker = time.NewTicker(l.reconnectDelay)
	ticker := l.reconnectTicker // Capture ticker to avoid race condition

	go func() {
		defer func() {
			if ticker != nil {
				ticker.Stop()
			}
		}()

		for {
			select {
			case <-l.ctx.Done():
				return
			case <-ticker.C:
				l.checkAndReconnect()
			}
		}
	}()
}

// checkAndReconnect checks connection health and reconnects if necessary
func (l *Logger) checkAndReconnect() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if !l.logstashEnabled || l.protocol != TCP {
		return
	}

	// Check if connection is alive by trying to set a deadline
	if l.conn != nil {
		if tcpConn, ok := l.conn.(*net.TCPConn); ok {
			err := tcpConn.SetKeepAlive(true)
			if err != nil {
				// Connection is broken, close it
				l.conn.Close()
				l.conn = nil
			}
		}
	}

	// If connection is nil, try to reconnect
	if l.conn == nil {
		l.attemptReconnect()
	}
}

// attemptReconnect attempts to reconnect with retry logic
func (l *Logger) attemptReconnect() {
	attempts := 0
	maxAttempts := l.reconnectAttempts

	for {
		if maxAttempts > 0 && attempts >= maxAttempts {
			fmt.Printf("Max reconnection attempts (%d) reached for logstash\n", maxAttempts)
			return
		}

		attempts++
		fmt.Printf("Attempting to reconnect to logstash (attempt %d)...\n", attempts)

		if l.connectToLogstash() {
			fmt.Printf("Successfully reconnected to logstash after %d attempts\n", attempts)
			return
		}

		// Wait before next attempt
		select {
		case <-l.ctx.Done():
			return
		case <-time.After(l.reconnectDelay):
			continue
		}
	}
}

// startAsyncProcessor starts the background goroutine for async log processing
func (l *Logger) startAsyncProcessor() {
	go func() {
		ticker := time.NewTicker(l.flushInterval)
		defer ticker.Stop()

		batch := make([]LogEntry, 0, 100) // Process in batches for better performance

		for {
			select {
			case <-l.ctx.Done():
				// Flush remaining logs before exit
				l.flushBatch(batch)
				for len(l.logBuffer) > 0 {
					entry := <-l.logBuffer
					l.logToLogstashSync(entry)
				}
				return

			case entry := <-l.logBuffer:
				batch = append(batch, entry)
				// Process batch if it's full
				if len(batch) >= 100 {
					l.processBatch(batch)
					batch = batch[:0] // Reset batch
				}

			case <-ticker.C:
				// Periodic flush
				if len(batch) > 0 {
					l.processBatch(batch)
					batch = batch[:0] // Reset batch
				}
			}
		}
	}()
}

// processBatch processes a batch of log entries
func (l *Logger) processBatch(batch []LogEntry) {
	for _, entry := range batch {
		l.logToLogstashSync(entry)
	}
}

// flushBatch flushes remaining entries in the batch
func (l *Logger) flushBatch(batch []LogEntry) {
	if len(batch) > 0 {
		l.processBatch(batch)
	}
}

// Flush forces immediate flush of all buffered logs
func (l *Logger) Flush() {
	if !l.asyncEnabled {
		return
	}

	// Wait for buffer to be empty with timeout
	timeout := time.After(5 * time.Second)
	for {
		select {
		case <-timeout:
			fmt.Printf("Flush timeout: %d logs remaining in buffer\n", len(l.logBuffer))
			return
		default:
			if len(l.logBuffer) == 0 {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}
}

// log is the main logging method
func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	// Format the message
	// Note: fmt.Sprintf is acceptable here as the format is user-provided
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	// Get information about the calling function
	pc, file, line, ok := runtime.Caller(2)
	var threadName, loggerName string
	if ok {
		// Extract only the filename without the full path - optimized
		var fileName string
		if lastSlash := strings.LastIndexByte(file, '/'); lastSlash >= 0 {
			fileName = file[lastSlash+1:]
		} else {
			fileName = file
		}
		// Use strconv for number formatting to reduce allocations
		threadName = fileName + ":" + strconv.Itoa(line)

		// Get function name
		if fn := runtime.FuncForPC(pc); fn != nil {
			fullFuncName := fn.Name()
			// Extract just the function name without package path - optimized
			if lastDot := strings.LastIndexByte(fullFuncName, '.'); lastDot >= 0 {
				loggerName = fullFuncName[lastDot+1:]
			} else {
				loggerName = fullFuncName
			}
		} else {
			loggerName = "unknown"
		}
	} else {
		threadName = "main"
		loggerName = "main"
	}

	// Get fields from logger (thread-safe, but we don't copy to avoid allocations)
	// Fields are read-only after logger creation, so it's safe to read without copying
	l.mu.RLock()
	fields := l.fields
	l.mu.RUnlock()

	entry := LogEntry{
		Timestamp:   time.Now().UTC().Format(l.timestampFormat),
		Level:       level.String(),
		Message:     message,
		LoggerName:  loggerName,
		ThreadName:  threadName,
		AppName:     l.appName,
		HostIP:      l.hostIP,
		ContainerID: l.containerID,
		Type:        "logback",
	}
	// Only set Fields if there are any (to avoid empty map allocation)
	if len(fields) > 0 {
		entry.Fields = fields
	}

	// Output to console
	l.logToConsole(entry)

	// Send to logstash if enabled
	if l.logstashEnabled {
		l.logToLogstash(entry)
	}
}

// logToConsole outputs log to console
func (l *Logger) logToConsole(entry LogEntry) {
	// Use buffer pool to reduce allocations
	buf := l.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer l.bufPool.Put(buf)

	// Build message using buffer (more efficient than fmt.Sprintf)
	buf.WriteString("[")
	buf.WriteString(entry.Timestamp)
	buf.WriteString("] ")
	buf.WriteString(entry.Level)
	buf.WriteString(" [")
	buf.WriteString(entry.LoggerName)
	buf.WriteString("] [")
	buf.WriteString(entry.ThreadName)
	buf.WriteString("] ")
	buf.WriteString(entry.Message)
	buf.WriteByte('\n')

	l.mu.RLock()
	output := l.output
	l.mu.RUnlock()

	if output != nil {
		output.Write(buf.Bytes())
	} else {
		os.Stderr.Write(buf.Bytes())
	}
}

// logToLogstash sends log to logstash
func (l *Logger) logToLogstash(entry LogEntry) {
	if l.asyncEnabled {
		l.logToLogstashAsync(entry)
	} else {
		l.logToLogstashSync(entry)
	}
}

// logToLogstashAsync sends log to buffer for async processing
func (l *Logger) logToLogstashAsync(entry LogEntry) {
	select {
	case l.logBuffer <- entry:
		// Successfully added to buffer
	default:
		// Buffer is full, fallback to sync logging to avoid blocking
		fmt.Printf("Log buffer full, falling back to sync logging\n")
		l.logToLogstashSync(entry)
	}
}

// logToLogstashSync sends log synchronously
func (l *Logger) logToLogstashSync(entry LogEntry) {
	if l.protocol == UDP {
		l.logToLogstashUDP(entry)
	} else {
		l.logToLogstashTCP(entry)
	}
}

// logToLogstashTCP sends log to logstash via TCP
func (l *Logger) logToLogstashTCP(entry LogEntry) {
	l.mu.RLock()
	conn := l.conn
	l.mu.RUnlock()

	if conn == nil {
		// Try to reconnect immediately
		l.mu.Lock()
		if l.conn == nil {
			l.connectToLogstash()
		}
		conn = l.conn
		l.mu.Unlock()

		if conn == nil {
			return // Still no connection
		}
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Failed to marshal log entry: %v\n", err)
		return
	}

	// Add newline for json_lines codec
	jsonData = append(jsonData, '\n')

	// Set write timeout to detect broken connections
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	}

	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Printf("Failed to write to logstash via TCP: %v\n", err)
		// Mark connection as broken
		l.mu.Lock()
		if l.conn != nil {
			l.conn.Close()
			l.conn = nil
		}
		l.mu.Unlock()
	} else {
		// Reset write deadline on successful write
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetWriteDeadline(time.Time{})
		}
	}
}

// logToLogstashUDP sends log to logstash via UDP
func (l *Logger) logToLogstashUDP(entry LogEntry) {
	address := net.JoinHostPort(l.logstashHost, fmt.Sprintf("%d", l.logstashPort))

	conn, err := net.Dial("udp", address)
	if err != nil {
		fmt.Printf("Failed to connect to logstash via UDP at %s: %v\n", address, err)
		return
	}
	defer conn.Close()

	jsonData, err := json.Marshal(entry)
	if err != nil {
		fmt.Printf("Failed to marshal log entry: %v\n", err)
		return
	}

	// Add newline for json_lines codec
	jsonData = append(jsonData, '\n')

	_, err = conn.Write(jsonData)
	if err != nil {
		fmt.Printf("Failed to write to logstash via UDP: %v\n", err)
	}
}

// Debug logs a message at DEBUG level
func (l *Logger) Debug(message string, args ...interface{}) {
	l.log(DEBUG, message, args...)
}

// Info logs a message at INFO level
func (l *Logger) Info(message string, args ...interface{}) {
	l.log(INFO, message, args...)
}

// Warn logs a message at WARN level
func (l *Logger) Warn(message string, args ...interface{}) {
	l.log(WARN, message, args...)
}

// Error logs a message at ERROR level
func (l *Logger) Error(message string, args ...interface{}) {
	l.log(ERROR, message, args...)
}

// Fatal logs a message at FATAL level and exits the program
func (l *Logger) Fatal(message string, args ...interface{}) {
	l.log(FATAL, message, args...)
	os.Exit(1)
}

// With returns a new logger instance with the specified fields added to the context.
// The fields will be included in all subsequent log entries from this logger.
// This method is thread-safe and creates a new logger instance, so the original logger remains unchanged.
func (l *Logger) With(fields map[string]interface{}) *Logger {
	// Create a new logger with copied configuration
	l.mu.RLock()
	newLogger := &Logger{
		logstashEnabled:   l.logstashEnabled,
		logstashHost:      l.logstashHost,
		logstashPort:      l.logstashPort,
		protocol:          l.protocol,
		appName:           l.appName,
		hostIP:            l.hostIP,
		containerID:       l.containerID,
		minLevel:          l.minLevel,
		reconnectAttempts: l.reconnectAttempts,
		reconnectDelay:    l.reconnectDelay,
		ctx:               l.ctx,
		cancel:            l.cancel, // Share the same cancel function
		timestampFormat:   l.timestampFormat,
		bufferSize:        l.bufferSize,
		flushInterval:     l.flushInterval,
		asyncEnabled:      l.asyncEnabled,
		output:            l.output,
		prefix:            l.prefix,
		flags:             l.flags,
		bufPool: sync.Pool{
			New: func() interface{} {
				return &bytes.Buffer{}
			},
		},
		fields: make(map[string]interface{}),
	}
	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}
	l.mu.RUnlock()

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	// Share async buffer if enabled
	if l.asyncEnabled && l.logstashEnabled {
		newLogger.logBuffer = l.logBuffer
	}

	return newLogger
}

// Standard log interface methods for compatibility with log package

// Print calls l.Info for the default logger
func (l *Logger) Print(v ...interface{}) {
	message := fmt.Sprint(v...)
	l.Info(message)
}

// Printf calls l.Info for the default logger
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

// Println calls l.Info for the default logger
func (l *Logger) Println(v ...interface{}) {
	message := fmt.Sprintln(v...)
	// Remove trailing newline since our logger adds its own formatting
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[:len(message)-1]
	}
	l.Info(message)
}

// Fatalf calls l.Fatal with formatted message
func (l *Logger) Fatalf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.log(FATAL, message)
	os.Exit(1)
}

// Fatalln calls l.Fatal with sprintln-formatted message
func (l *Logger) Fatalln(v ...interface{}) {
	message := fmt.Sprintln(v...)
	// Remove trailing newline since our logger adds its own formatting
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[:len(message)-1]
	}
	l.log(FATAL, message)
	os.Exit(1)
}

// Panic logs a message at ERROR level and panics
func (l *Logger) Panic(v ...interface{}) {
	message := fmt.Sprint(v...)
	l.Error(message)
	panic(message)
}

// Panicf logs a formatted message at ERROR level and panics
func (l *Logger) Panicf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.Error(message)
	panic(message)
}

// Panicln logs a sprintln-formatted message at ERROR level and panics
func (l *Logger) Panicln(v ...interface{}) {
	message := fmt.Sprintln(v...)
	// Remove trailing newline since our logger adds its own formatting
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[:len(message)-1]
	}
	l.Error(message)
	panic(message)
}

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

// Output returns the current output destination
func (l *Logger) Output() io.Writer {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.output
}

// SetFlags sets the output flags for the logger (for compatibility)
func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flags = flag
}

// Flags returns the output flags for the logger
func (l *Logger) Flags() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.flags
}

// SetPrefix sets the output prefix for the logger
func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

// Prefix returns the output prefix for the logger
func (l *Logger) Prefix() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.prefix
}

// Close closes the connection to logstash and stops monitoring
func (l *Logger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Flush remaining logs if async is enabled
	if l.asyncEnabled {
		l.Flush()
	}

	// Stop the async processor and reconnect monitor
	if l.cancel != nil {
		l.cancel()
	}

	if l.reconnectTicker != nil {
		l.reconnectTicker.Stop()
		l.reconnectTicker = nil
	}

	// Close the log buffer channel
	if l.logBuffer != nil {
		close(l.logBuffer)
		l.logBuffer = nil
	}

	// Close the connection
	if l.conn != nil {
		err := l.conn.Close()
		l.conn = nil
		return err
	}

	return nil
}

// getHostIP gets the host IP address
func getHostIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "unknown"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// getContainerID gets the container ID (basic implementation)
func getContainerID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

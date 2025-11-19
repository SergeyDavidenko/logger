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
	"sync/atomic"
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

// Pre-allocated level strings to avoid allocations
var levelStrings = [...]string{
	DEBUG: "DEBUG",
	INFO:  "INFO",
	WARN:  "WARN",
	ERROR: "ERROR",
	FATAL: "FATAL",
}

// String returns the string representation of the log level
func (l LogLevel) String() string {
	if l >= 0 && int(l) < len(levelStrings) {
		return levelStrings[l]
	}
	return "UNKNOWN"
}

// ParseLogLevel parses a string and returns the corresponding LogLevel
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
		return INFO, false
	}
}

// MustParseLogLevel parses a string and returns the corresponding LogLevel
func MustParseLogLevel(level string) LogLevel {
	if logLevel, ok := ParseLogLevel(level); ok {
		return logLevel
	}
	panic(fmt.Sprintf("invalid log level: %s", level))
}

// LogEntry represents a log entry - optimized with string interning
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

// callerInfo caches caller information to reduce runtime.Caller calls
type callerInfo struct {
	threadName string
	loggerName string
}

// Logger is the main logger structure
type Logger struct {
	mu                sync.RWMutex
	logstashEnabled   bool
	logstashHost      string
	logstashPort      int
	protocol          Protocol
	conn              atomic.Pointer[net.Conn] // lock-free connection access
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
	logBuffer     chan *LogEntry
	bufferSize    int
	flushInterval time.Duration
	asyncEnabled  bool

	// Standard log interface compatibility
	output io.Writer
	prefix string
	flags  int

	// Buffer pools for optimization (pointers to allow sharing between loggers)
	bufPool     *sync.Pool // *bytes.Buffer for console output
	entryPool   *sync.Pool // *LogEntry for reuse
	jsonBufPool *sync.Pool // *bytes.Buffer for JSON encoding

	// Context fields for structured logging (immutable after creation)
	fields map[string]interface{}

	// Caller info cache (pointer to allow sharing between loggers)
	callerCache *sync.Map // map[uintptr]*callerInfo

	// Closed flag to prevent double close
	closed uint32 // atomic

	// Pre-allocated timestamp buffer
	timestampBuf []byte
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
	Protocol          Protocol
	AppName           string
	MinLevel          LogLevel
	ReconnectAttempts int
	ReconnectDelay    time.Duration
	AsyncEnabled      bool
	BufferSize        int
	FlushInterval     time.Duration
}

// DefaultConfig returns the default configuration
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
	protocol := config.Protocol
	if protocol == "" {
		protocol = TCP
	}

	reconnectAttempts := config.ReconnectAttempts
	if reconnectAttempts == 0 {
		reconnectAttempts = -1
	}

	reconnectDelay := config.ReconnectDelay
	if reconnectDelay == 0 {
		reconnectDelay = 5 * time.Second
	}

	bufferSize := config.BufferSize
	if bufferSize == 0 {
		bufferSize = 1000
	}

	flushInterval := config.FlushInterval
	if flushInterval == 0 {
		flushInterval = 1 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

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
		bufferSize:        bufferSize,
		flushInterval:     flushInterval,
		asyncEnabled:      config.AsyncEnabled,
		output:            os.Stderr,
		prefix:            "",
		flags:             0,
		fields:            make(map[string]interface{}),
		timestampBuf:      make([]byte, 0, 64),
	}

	// Initialize buffer pools
	logger.bufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 256))
		},
	}

	logger.entryPool = &sync.Pool{
		New: func() interface{} {
			return &LogEntry{}
		},
	}

	logger.jsonBufPool = &sync.Pool{
		New: func() interface{} {
			return bytes.NewBuffer(make([]byte, 0, 512))
		},
	}

	// Initialize caller cache
	logger.callerCache = &sync.Map{}

	// Initialize async logging if enabled
	if logger.asyncEnabled && logger.logstashEnabled {
		logger.logBuffer = make(chan *LogEntry, bufferSize)
		logger.startAsyncProcessor()
	}

	logger.hostIP = getHostIP()
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
		conn := l.getConn()
		if conn == nil {
			l.connectToLogstash()
		}
		if l.reconnectTicker == nil {
			l.startReconnectMonitorUnsafe()
		}
	} else if !enabled {
		conn := l.getConn()
		if conn != nil {
			conn.Close()
			l.conn.Store((*net.Conn)(nil))
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
		format = "2006-01-02T15:04:05.000Z"
	}
	l.timestampFormat = format
}

// GetTimestampFormat returns the current timestamp format
func (l *Logger) GetTimestampFormat() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.timestampFormat
}

// getConn returns the current connection (lock-free)
func (l *Logger) getConn() net.Conn {
	if c := l.conn.Load(); c != nil {
		return *c
	}
	return nil
}

// setConn sets the connection (lock-free)
func (l *Logger) setConn(conn net.Conn) {
	if conn != nil {
		l.conn.Store(&conn)
	} else {
		l.conn.Store(nil)
	}
}

// connectToLogstash establishes connection to logstash
func (l *Logger) connectToLogstash() bool {
	if l.logstashHost == "" || l.logstashPort == 0 {
		return false
	}

	address := net.JoinHostPort(l.logstashHost, strconv.Itoa(l.logstashPort))
	conn, err := net.Dial(string(l.protocol), address)
	if err != nil {
		fmt.Printf("Failed to connect to logstash at %s via %s: %v\n", address, l.protocol, err)
		return false
	}

	l.setConn(conn)
	return true
}

// startReconnectMonitor starts a background goroutine to monitor connection health
func (l *Logger) startReconnectMonitor() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.startReconnectMonitorUnsafe()
}

// startReconnectMonitorUnsafe starts reconnect monitor without locking
func (l *Logger) startReconnectMonitorUnsafe() {
	if l.protocol != TCP || l.reconnectTicker != nil {
		return
	}

	l.reconnectTicker = time.NewTicker(l.reconnectDelay)
	ticker := l.reconnectTicker

	go func() {
		defer ticker.Stop()
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

	conn := l.getConn()
	if conn != nil {
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			err := tcpConn.SetKeepAlive(true)
			if err != nil {
				conn.Close()
				l.setConn(nil)
			}
		}
	}

	if l.getConn() == nil {
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

		batch := make([]*LogEntry, 0, 100)

		for {
			select {
			case <-l.ctx.Done():
				l.flushBatch(batch)
				for len(l.logBuffer) > 0 {
					entry := <-l.logBuffer
					l.logToLogstashSync(entry)
					l.entryPool.Put(entry) // Return to pool
				}
				return

			case entry := <-l.logBuffer:
				batch = append(batch, entry)
				if len(batch) >= 100 {
					l.processBatch(batch)
					batch = batch[:0]
				}

			case <-ticker.C:
				if len(batch) > 0 {
					l.processBatch(batch)
					batch = batch[:0]
				}
			}
		}
	}()
}

// processBatch processes a batch of log entries
func (l *Logger) processBatch(batch []*LogEntry) {
	for _, entry := range batch {
		l.logToLogstashSync(entry)
		l.entryPool.Put(entry) // Return to pool
	}
}

// flushBatch flushes remaining entries in the batch
func (l *Logger) flushBatch(batch []*LogEntry) {
	if len(batch) > 0 {
		l.processBatch(batch)
	}
}

// Flush forces immediate flush of all buffered logs
func (l *Logger) Flush() {
	if !l.asyncEnabled {
		return
	}

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

// getCallerInfo gets or caches caller information
func (l *Logger) getCallerInfo(pc uintptr, file string, line int) (threadName, loggerName string) {
	// Try to get from cache first
	if cached, ok := l.callerCache.Load(pc); ok {
		info := cached.(*callerInfo)
		return info.threadName, info.loggerName
	}

	// Extract filename without path
	fileName := file
	if lastSlash := strings.LastIndexByte(file, '/'); lastSlash >= 0 {
		fileName = file[lastSlash+1:]
	}
	threadName = fileName + ":" + strconv.Itoa(line)

	// Get function name
	if fn := runtime.FuncForPC(pc); fn != nil {
		fullFuncName := fn.Name()
		if lastDot := strings.LastIndexByte(fullFuncName, '.'); lastDot >= 0 {
			loggerName = fullFuncName[lastDot+1:]
		} else {
			loggerName = fullFuncName
		}
	} else {
		loggerName = "unknown"
	}

	// Cache for future use
	info := &callerInfo{
		threadName: threadName,
		loggerName: loggerName,
	}
	l.callerCache.Store(pc, info)

	return threadName, loggerName
}

// log is the main logging method - heavily optimized
func (l *Logger) log(level LogLevel, message string, args ...interface{}) {
	if level < l.minLevel {
		return
	}

	// Format message only if needed
	if len(args) > 0 {
		message = fmt.Sprintf(message, args...)
	}

	// Get caller info with caching
	pc, file, line, ok := runtime.Caller(2)
	var threadName, loggerName string
	if ok {
		threadName, loggerName = l.getCallerInfo(pc, file, line)
	} else {
		threadName = "main"
		loggerName = "main"
	}

	// Get entry from pool instead of allocating
	entry := l.entryPool.Get().(*LogEntry)

	// Reuse timestamp buffer to avoid allocation
	now := time.Now().UTC()
	entry.Timestamp = now.Format(l.timestampFormat)
	entry.Level = level.String() // Uses pre-allocated strings
	entry.Message = message
	entry.LoggerName = loggerName
	entry.ThreadName = threadName
	entry.AppName = l.appName
	entry.HostIP = l.hostIP
	entry.ContainerID = l.containerID
	entry.Type = "logback"

	// Shallow copy fields (they're immutable)
	if len(l.fields) > 0 {
		entry.Fields = l.fields
	} else {
		entry.Fields = nil
	}

	// Output to console
	l.logToConsole(entry)

	// Send to logstash if enabled
	if l.logstashEnabled {
		l.logToLogstash(entry)
	} else {
		// Return to pool immediately if not going to async queue
		l.entryPool.Put(entry)
	}
}

// logToConsole outputs log to console - optimized with buffer reuse
func (l *Logger) logToConsole(entry *LogEntry) {
	buf := l.bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer l.bufPool.Put(buf)

	// Pre-allocate approximate size
	buf.Grow(len(entry.Timestamp) + len(entry.Level) + len(entry.LoggerName) +
		len(entry.ThreadName) + len(entry.Message) + 10)

	buf.WriteByte('[')
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
func (l *Logger) logToLogstash(entry *LogEntry) {
	if l.asyncEnabled {
		l.logToLogstashAsync(entry)
	} else {
		l.logToLogstashSync(entry)
		l.entryPool.Put(entry) // Return to pool
	}
}

// logToLogstashAsync sends log to buffer for async processing
func (l *Logger) logToLogstashAsync(entry *LogEntry) {
	// Check if closed atomically
	if atomic.LoadUint32(&l.closed) == 1 {
		l.entryPool.Put(entry)
		return
	}

	select {
	case l.logBuffer <- entry:
		// Successfully added to buffer - will be returned to pool by async processor
	default:
		// Buffer is full, fallback to sync logging
		fmt.Printf("Log buffer full, falling back to sync logging\n")
		l.logToLogstashSync(entry)
		l.entryPool.Put(entry) // Return to pool immediately
	}
}

// logToLogstashSync sends log synchronously
func (l *Logger) logToLogstashSync(entry *LogEntry) {
	if l.protocol == UDP {
		l.logToLogstashUDP(entry)
	} else {
		l.logToLogstashTCP(entry)
	}
}

// logToLogstashTCP sends log to logstash via TCP - optimized with buffer pool
func (l *Logger) logToLogstashTCP(entry *LogEntry) {
	conn := l.getConn()
	if conn == nil {
		l.mu.Lock()
		if l.getConn() == nil {
			l.connectToLogstash()
		}
		conn = l.getConn()
		l.mu.Unlock()

		if conn == nil {
			return
		}
	}

	// Get buffer from pool for JSON encoding
	jsonBuf := l.jsonBufPool.Get().(*bytes.Buffer)
	jsonBuf.Reset()
	defer l.jsonBufPool.Put(jsonBuf)

	encoder := json.NewEncoder(jsonBuf)
	if err := encoder.Encode(entry); err != nil {
		fmt.Printf("Failed to marshal log entry: %v\n", err)
		return
	}

	// Set write timeout
	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetWriteDeadline(time.Now().Add(5 * time.Second))
	}

	_, err := conn.Write(jsonBuf.Bytes())
	if err != nil {
		fmt.Printf("Failed to write to logstash via TCP: %v\n", err)
		l.mu.Lock()
		currentConn := l.getConn()
		if currentConn != nil {
			currentConn.Close()
			l.setConn(nil)
		}
		l.mu.Unlock()
	} else {
		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetWriteDeadline(time.Time{})
		}
	}
}

// logToLogstashUDP sends log to logstash via UDP - optimized
func (l *Logger) logToLogstashUDP(entry *LogEntry) {
	address := net.JoinHostPort(l.logstashHost, strconv.Itoa(l.logstashPort))

	conn, err := net.Dial("udp", address)
	if err != nil {
		fmt.Printf("Failed to connect to logstash via UDP at %s: %v\n", address, err)
		return
	}
	defer conn.Close()

	// Use buffer pool
	jsonBuf := l.jsonBufPool.Get().(*bytes.Buffer)
	jsonBuf.Reset()
	defer l.jsonBufPool.Put(jsonBuf)

	encoder := json.NewEncoder(jsonBuf)
	if err := encoder.Encode(entry); err != nil {
		fmt.Printf("Failed to marshal log entry: %v\n", err)
		return
	}

	_, err = conn.Write(jsonBuf.Bytes())
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
	l.Flush() // Ensure all logs are written
	os.Exit(1)
}

// With returns a new logger instance with additional fields
func (l *Logger) With(fields map[string]interface{}) *Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	// Create new fields map
	newFields := make(map[string]interface{}, len(l.fields)+len(fields))
	for k, v := range l.fields {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}

	// Create new logger sharing most resources
	// Note: We cannot copy sync.Pool and sync.Map directly, so we share references
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
		cancel:            l.cancel,
		timestampFormat:   l.timestampFormat,
		bufferSize:        l.bufferSize,
		flushInterval:     l.flushInterval,
		asyncEnabled:      l.asyncEnabled,
		output:            l.output,
		prefix:            l.prefix,
		flags:             l.flags,
		fields:            newFields,
		timestampBuf:      make([]byte, 0, 64),
	}

	// Share pools by pointer (safe to share between loggers)
	newLogger.bufPool = l.bufPool
	newLogger.entryPool = l.entryPool
	newLogger.jsonBufPool = l.jsonBufPool

	// Share cache by pointer (safe to share between loggers)
	newLogger.callerCache = l.callerCache

	// Share connection - ВАЖНО: atomic.Pointer позволяет хранить nil
	if conn := l.getConn(); conn != nil {
		newLogger.setConn(conn)
	}

	// Share async buffer
	if l.asyncEnabled && l.logstashEnabled {
		newLogger.logBuffer = l.logBuffer
	}

	return newLogger
}

// Standard log interface methods
func (l *Logger) Print(v ...interface{}) {
	l.Info(fmt.Sprint(v...))
}

func (l *Logger) Printf(format string, v ...interface{}) {
	l.Info(format, v...)
}

func (l *Logger) Println(v ...interface{}) {
	message := fmt.Sprintln(v...)
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[:len(message)-1]
	}
	l.Info(message)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Fatal(format, v...)
}

func (l *Logger) Fatalln(v ...interface{}) {
	message := fmt.Sprintln(v...)
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[:len(message)-1]
	}
	l.Fatal(message)
}

func (l *Logger) Panic(v ...interface{}) {
	message := fmt.Sprint(v...)
	l.Error(message)
	panic(message)
}

func (l *Logger) Panicf(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	l.Error(message)
	panic(message)
}

func (l *Logger) Panicln(v ...interface{}) {
	message := fmt.Sprintln(v...)
	if len(message) > 0 && message[len(message)-1] == '\n' {
		message = message[:len(message)-1]
	}
	l.Error(message)
	panic(message)
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
}

func (l *Logger) Output() io.Writer {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.output
}

func (l *Logger) SetFlags(flag int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.flags = flag
}

func (l *Logger) Flags() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.flags
}

func (l *Logger) SetPrefix(prefix string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.prefix = prefix
}

func (l *Logger) Prefix() string {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.prefix
}

// Close closes the connection to logstash and stops monitoring
func (l *Logger) Close() error {
	// Mark as closed atomically
	if !atomic.CompareAndSwapUint32(&l.closed, 0, 1) {
		return nil // Already closed
	}

	// Stop all background goroutines
	if l.cancel != nil {
		l.cancel()
	}

	// Wait a bit for async processor to finish
	time.Sleep(100 * time.Millisecond)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.reconnectTicker != nil {
		l.reconnectTicker.Stop()
		l.reconnectTicker = nil
	}

	var err error
	conn := l.getConn()
	if conn != nil {
		err = conn.Close()
		l.setConn(nil)
	}

	return err
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

// getContainerID gets the container ID
func getContainerID() string {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return hostname
}

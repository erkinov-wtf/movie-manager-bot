package logger

import (
	"bytes"
	"encoding/json"
	"go.uber.org/zap/zapcore"
	"net/http"
	"sync"
	"time"
)

const (
	maxBatchSize    = 100
	flushInterval   = 5 * time.Second
	contentTypeJSON = "application/json"
)

var betterStackURL string
var authToken string

// BetterStackLogger implements log forwarding to BetterStack
type BetterStackLogger struct {
	underlying internalLogger // Wrap the existing logger
	logChan    chan logEntry
	batchMutex sync.Mutex
	batch      []logEntry
	wg         sync.WaitGroup
	stopChan   chan struct{}
	httpClient *http.Client
	logLevel   zapcore.Level // Store the configured log level
}

// logEntry represents a single log entry to be sent to BetterStack
type logEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

// newBetterStackLogger creates a new BetterStackLogger that wraps an existing logger
func newBetterStackLogger(underlying internalLogger, level zapcore.Level, host, token string) *BetterStackLogger {
	logger := &BetterStackLogger{
		underlying: underlying,
		logChan:    make(chan logEntry, 1000), // Buffer size
		batch:      make([]logEntry, 0, maxBatchSize),
		stopChan:   make(chan struct{}),
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logLevel: level,
	}

	betterStackURL = host
	authToken = token

	// Start the background worker
	logger.wg.Add(1)
	go logger.processLogEntries()

	return logger
}

// processLogEntries handles log entries in the background
func (b *BetterStackLogger) processLogEntries() {
	defer b.wg.Done()

	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()

	for {
		select {
		case entry := <-b.logChan:
			b.addToBatch(entry)
		case <-ticker.C:
			b.flush()
		case <-b.stopChan:
			b.flush() // Final flush before stopping
			return
		}
	}
}

// addToBatch adds a log entry to the current batch and flushes if needed
func (b *BetterStackLogger) addToBatch(entry logEntry) {
	b.batchMutex.Lock()
	defer b.batchMutex.Unlock()

	b.batch = append(b.batch, entry)
	if len(b.batch) >= maxBatchSize {
		go b.sendBatch(b.batch)
		b.batch = make([]logEntry, 0, maxBatchSize)
	}
}

// flush sends any pending log entries
func (b *BetterStackLogger) flush() {
	b.batchMutex.Lock()
	defer b.batchMutex.Unlock()

	if len(b.batch) > 0 {
		go b.sendBatch(b.batch)
		b.batch = make([]logEntry, 0, maxBatchSize)
	}
}

// sendBatch sends a batch of logs to BetterStack
func (b *BetterStackLogger) sendBatch(batch []logEntry) {
	body, err := json.Marshal(batch)
	if err != nil {
		// Log locally that marshaling failed
		b.underlying.error("Failed to marshal logs for BetterStack", "error", err.Error())
		return
	}

	req, err := http.NewRequest(http.MethodPost, betterStackURL, bytes.NewBuffer(body))
	if err != nil {
		b.underlying.error("Failed to create BetterStack request", "error", err.Error())
		return
	}

	req.Header.Set("Content-Type", contentTypeJSON)
	req.Header.Set("Authorization", authToken)

	resp, err := b.httpClient.Do(req)
	if err != nil {
		b.underlying.error("Failed to send logs to BetterStack", "error", err.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		b.underlying.error("BetterStack API returned non-202 status",
			"status", resp.StatusCode,
			"status_text", resp.Status)
	}
}

// Stop shuts down the logger gracefully
func (b *BetterStackLogger) Stop() {
	close(b.stopChan)
	b.wg.Wait()
}

// shouldLog checks if the given level should be logged based on configuration
func (b *BetterStackLogger) shouldLog(level zapcore.Level) bool {
	return level >= b.logLevel
}

// info logs an info level message
func (b *BetterStackLogger) info(msg string, args ...interface{}) {
	// First log locally
	b.underlying.info(msg, args...)

	// Then queue for async sending
	if b.shouldLog(zapcore.InfoLevel) {
		b.queueLog("INFO", msg, args...)
	}
}

// error logs an error level message
func (b *BetterStackLogger) error(msg string, args ...interface{}) {
	// First log locally
	b.underlying.error(msg, args...)

	// Then queue for async sending
	if b.shouldLog(zapcore.ErrorLevel) {
		b.queueLog("ERROR", msg, args...)
	}
}

// debug logs a debug level message
func (b *BetterStackLogger) debug(msg string, args ...interface{}) {
	// First log locally
	b.underlying.debug(msg, args...)

	// Then queue for async sending
	if b.shouldLog(zapcore.DebugLevel) {
		b.queueLog("DEBUG", msg, args...)
	}
}

// warn logs a warning level message
func (b *BetterStackLogger) warn(msg string, args ...interface{}) {
	// First log locally
	b.underlying.warn(msg, args...)

	// Then queue for async sending
	if b.shouldLog(zapcore.WarnLevel) {
		b.queueLog("WARN", msg, args...)
	}
}

// sync flushes buffered logs
func (b *BetterStackLogger) sync() error {
	b.flush()
	return b.underlying.sync()
}

// queueLog converts args to a structured log entry and queues it
func (b *BetterStackLogger) queueLog(level, msg string, args ...interface{}) {
	// Create a structured field map that preserves the same structure as Zap
	fields := make(map[string]interface{})

	// Process key-value pairs while preserving nested structures
	for i := 0; i < len(args); i += 2 {
		if i+1 < len(args) {
			if key, ok := args[i].(string); ok {
				fields[key] = args[i+1]
			}
		}
	}

	entry := logEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   msg,
		Fields:    fields,
	}

	// Try to send to channel without blocking
	select {
	case b.logChan <- entry:
		// Successfully queued
	default:
		// Channel is full, log locally that we're dropping messages
		b.underlying.warn("BetterStack log channel full, dropping message")
	}
}

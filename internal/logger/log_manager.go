package logger

import (
	"os"
	"path"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Global logger instance
var (
	logInstance zerolog.Logger
	consoleLog  zerolog.Logger
	logQueue    chan func()
	once        sync.Once
)

var (
	totalRequests int64
	successfulOps int64
	errors        int64
	memoryUsage   uint64
)

func logMetrics() {
	for {
		time.Sleep(5 * time.Second)
		runtime.GC()
		var memStats runtime.MemStats
		runtime.ReadMemStats(&memStats)
		atomic.StoreUint64(&memoryUsage, memStats.Alloc)
		log.Printf("[METRICS] Requests: %d | Success: %d | Errors: %d | Memory Usage: %d KB\n",
			atomic.LoadInt64(&totalRequests),
			atomic.LoadInt64(&successfulOps),
			atomic.LoadInt64(&errors),
			atomic.LoadUint64(&memoryUsage)/1024)
	}
}

func logRequest(id string, status string, duration time.Duration, err error) {
	if err != nil {
		atomic.AddInt64(&errors, 1)
		log.Printf("[ERROR] Request ID: %s | Duration: %s | Error: %v\n", id, duration, err)
	} else {
		atomic.AddInt64(&successfulOps, 1)
		log.Printf("[SUCCESS] Request ID: %s | Duration: %s | Status: %s\n", id, duration, status)
	}
	atomic.AddInt64(&totalRequests, 1)
}

// func main() {
// 	logFile, err := os.OpenFile("db_logs.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
// 	if err != nil {
// 		log.Fatalf("Failed to open log file: %v", err)
// 	}
// 	defer logFile.Close()
// 	log.SetOutput(logFile)

// 	go logMetrics()

// 	// Simulate database requests
// 	for i := 0; i < 10; i++ {
// 		start := time.Now()
// 		id := "REQ-" + time.Now().Format("150405.000")
// 		logRequest(id, "OK", time.Since(start), nil)
// 		time.Sleep(1 * time.Second)
// 	}
// }

// GetLogLevelFromFlag reads the log level from the command-line flag
func GetLogLevelFromFlag(logLevel *string) zerolog.Level {
	switch *logLevel {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel // Default to INFO if invalid level is provided
	}
}

// InitLogger initializes the logger (thread-safe & non-blocking)
func InitLogger(logFile string, logLevel *string) {

	os.MkdirAll(path.Dir(logFile), 0777)
	once.Do(func() {
		// Set time format for logs
		zerolog.TimeFieldFormat = time.RFC3339

		// Set up log rotation with lumberjack
		rotatingFile := &lumberjack.Logger{
			Filename:   logFile, // Log file path
			MaxSize:    20,      // Max file size (MB) before rotation
			MaxBackups: 3,       // Keep last 3 logs
			MaxAge:     7,       // Keep logs for 7 days
			Compress:   true,    // Compress rotated logs
		}

		// // Open log file with append mode
		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			panic("Failed to open log file: " + err.Error()) // Failing fast
		}

		defer file.Close()

		level := GetLogLevelFromFlag(logLevel)

		// Console logger (pretty print output)
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
		consoleLog = zerolog.New(consoleWriter).With().Timestamp().Logger()
		// Set up logger
		logInstance = zerolog.New(zerolog.MultiLevelWriter(
			rotatingFile,
			consoleWriter,
			// Add additional writers here if needed
		)).With().Timestamp().Logger().Level(level)

		// Create a buffered channel for async logging
		logQueue = make(chan func(), 1000) // Buffered channel (1000 logs in queue)

		// Start background goroutine for async processing
		go func() {
			for logFunc := range logQueue {
				logFunc() // Execute log function
			}
		}()

	})
}

// asyncLog sends a log function to the queue (non-blocking)
func asyncLog(logFunc func()) {
	select {
	case logQueue <- logFunc:
	default: // Drop logs if queue is full (prevents blocking)
	}
}

// Info logs an informational message asynchronously
func Info(msg string, fields map[string]interface{}) {
	asyncLog(func() {
		event := logInstance.Info()
		if fields != nil {
			for k, v := range fields {
				event = event.Interface(k, v)
			}
		}
		event.Msg(msg)
	})
}

// Debug logs a debug message asynchronously
func Debug(msg string, fields map[string]interface{}) {
	asyncLog(func() {
		event := logInstance.Debug()
		if fields != nil {
			for k, v := range fields {
				event = event.Interface(k, v)
			}
		}
		event.Msg(msg)
	})
}

// Warn logs a warning message asynchronously
func Warn(msg string, fields map[string]interface{}) {
	asyncLog(func() {
		event := logInstance.Warn()
		if fields != nil {
			for k, v := range fields {
				event = event.Interface(k, v)
			}
		}
		event.Msg(msg)
	})
}

// Error logs an error message asynchronously
func Error(msg string, err error, fields map[string]interface{}) {
	asyncLog(func() {
		event := logInstance.Error().Err(err)
		if fields != nil {
			for k, v := range fields {
				event = event.Interface(k, v)
			}
		}
		event.Msg(msg)
	})
}

// Console logs a message to the console only
func Console(msg string, level zerolog.Level, fields map[string]interface{}) {
	event := consoleLog.WithLevel(level)
	if fields != nil {
		for k, v := range fields {
			event = event.Interface(k, v)
		}
	}
	event.Msg(msg)
}

// CloseLogger gracefully shuts down the logger
func CloseLogger() {
	close(logQueue) // Closes the channel, ensuring all logs are processed
}

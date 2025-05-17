package logger

import (
	"encoding/csv"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/*
Logger represents a thread safe logger object that can be
used by individual components to display output and write messages
to a single log file.

Log messages are stored as .csv files using the following columns:
Time, Component, Level, Message, ID
*/
type Logger struct {
	mu          sync.Mutex   // lock so loggers don't over write each other
	component   string       // name of the component this logger is attached to
	componentID string       // ID of the component this logger is attached to
	logfile     string       // absolute path to the csv log file
	log         *slog.Logger // slog instance
	csvWriter   *csv.Writer  // csv writer instance
}

// Log levels
const (
	INFO  string = "INFO"
	DEBUG string = "DEBUG"
	WARN  string = "WARN"
	ERROR string = "ERROR"
	FATAL string = "FATAL"
)

// Logger configs
// instantiate a new logger
func NewLogger(component string, id string) *Logger {
	// place log file in an designated directory, or the current
	// one if LOG_DIR is not set
	logDir, set := os.LookupEnv("LOG_DIR")
	if !set {
		logDir, _ = os.Getwd()
	}
	// create the log file if it doesn't already exist
	// log files have the name format: log-dd-mm-yyyy.csv, so
	// one new log file should be created per day.
	logFile := filepath.Join(logDir, fmt.Sprintf("log-%s.csv", getCurrentDate()))

	// make sure the log directory exists. if not, create it.
	if err := createLogDir(logDir); err != nil {
		log.Fatalf("failed to create log directory: %v", err)
	}

	// create the log file if it doesn't already exist
	if err := createLogFile(logFile); err != nil {
		log.Fatalf("failed to create log file: %v", err)
	}

	// open for use by the CSV writer.
	csvFile, err := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("failed to open log file: %v", err)
	}
	return &Logger{
		component:   component,
		componentID: id,
		logfile:     logFile,
		csvWriter:   csv.NewWriter(csvFile),
		log:         slog.New(slog.NewTextHandler(os.Stdout, nil)),
	}
}

// return todays date as dd-mm-yyyy
func getCurrentDate() string {
	now := time.Now()
	return fmt.Sprintf("%02d-%02d-%d", now.Day(), now.Month(), now.Year())
}

// make sure the log directory exists. if not, create it.
func createLogDir(logDirPath string) error {
	if _, err := os.Stat(logDirPath); errors.Is(err, os.ErrNotExist) {
		if err := os.Mkdir(logDirPath, 0666); err != nil {
			return fmt.Errorf("failed to create log directory: %v", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to get log dir stats: %v", err)
	}
	return nil
}

// create a log file if it doesn't exist
func createLogFile(lfpath string) error {
	if _, err := os.Stat(lfpath); errors.Is(err, os.ErrNotExist) {
		csvFile, err := os.Create(lfpath)
		if err != nil {
			return err
		}
		defer csvFile.Close()
		if err := csvFile.Chmod(0777); err != nil {
			return err
		}
		// add initial column names
		writer := csv.NewWriter(csvFile)
		writer.Write([]string{"Time", "Component", "Level", "Message", "ID"})
		writer.Flush()
	}
	return nil
}

// Info logs at LevelInfo and displays the message.
func (l *Logger) Info(msg string, v ...any) {
	l.log.Info(fmt.Sprintf(msg, v...))
	l.Log(INFO, fmt.Sprintf(msg, v...))
}

// Debug logs at LevelDebug and displays the message.
func (l *Logger) Debug(msg string, v ...any) {
	l.log.Debug(fmt.Sprintf(msg, v...))
	l.Log(DEBUG, fmt.Sprintf(msg, v...))
}

// Warn logs at LevelWarn and displays the message.
func (l *Logger) Warn(msg string, v ...any) {
	l.log.Warn(fmt.Sprintf(msg, v...))
	l.Log(WARN, fmt.Sprintf(msg, v...))
}

// Error logs at LevelError and displays the error message
func (l *Logger) Error(msg string, v ...any) {
	l.log.Error(fmt.Sprintf(msg, v...))
	l.Log(ERROR, fmt.Sprintf(msg, v...))
}

// Log writes a log entry to the CSV file. Does not display the message.
// All logging csv files use the columns: timestamp, component, level, message, and ID.
// The component and timestamp are provided by Log(), assuming
// Logger was instantiated correctly.
func (l *Logger) Log(level string, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().UTC()
	l.csvWriter.Write([]string{timestamp.Format(time.RFC3339), l.component, level, msg, l.componentID})
	l.csvWriter.Flush()
	if err := l.csvWriter.Error(); err != nil {
		log.Fatalf("error writing to log file: %v", err)
	}
}

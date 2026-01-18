package ui

import "soloterm/domain/log"

// LogViewHelper coordinates log-related UI operations
type LogViewHelper struct {
	logService *log.Service
}

// NewLogHandler creates a new log handler
func NewLogViewHelper(logService *log.Service) *LogViewHelper {
	return &LogViewHelper{
		logService: logService,
	}
}

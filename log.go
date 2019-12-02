package util

import "log"

var logger = log.Printf

// SetLogger set personal logger 
func SetLogger(logFn func(format string, v ...interface{})) {
	logger = logFn
}

// DeferLog print log with err handler
func DeferLog(action string, err error) {
	if err != nil {
		logger("finish %s with error: %s", action, err)
	} else {
		logger("finish %s succed", action)
	}
}

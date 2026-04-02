package bootstrap

import (
	"io"
	"log"
	"os"
	"path/filepath"
)

func InitLogging() (*os.File, error) {
	logDir := filepath.Clean("logs")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	logPath := filepath.Join(logDir, "server.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	log.SetOutput(io.MultiWriter(os.Stdout, file))
	log.SetFlags(log.Ldate | log.Ltime | log.Lmicroseconds)
	log.Printf("[bootstrap] logging initialized: %s", logPath)
	return file, nil
}

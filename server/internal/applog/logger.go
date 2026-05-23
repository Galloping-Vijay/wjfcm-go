package applog

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"wjfcm-go/internal/config"

	"github.com/gin-gonic/gin"
)

func Configure(cfg config.Config) (func() error, error) {
	writer, closeLog, err := logWriter(cfg.Log)
	if err != nil {
		return func() error { return nil }, err
	}

	log.SetOutput(writer)
	log.SetFlags(log.LstdFlags)
	gin.DefaultWriter = writer
	gin.DefaultErrorWriter = writer

	return closeLog, nil
}

func logWriter(cfg config.LogConfig) (io.Writer, func() error, error) {
	switch strings.ToLower(strings.TrimSpace(cfg.Channel)) {
	case "", "stack", "stdout", "console":
		return os.Stdout, func() error { return nil }, nil
	case "stderr":
		return os.Stderr, func() error { return nil }, nil
	case "single", "file":
		return openLogFile(filepath.Join(logPath(cfg), "wjfcm-go.log"), cfg.MaxSizeBytes)
	case "daily":
		name := "wjfcm-go-" + time.Now().Format("2006-01-02") + ".log"
		return openLogFile(filepath.Join(logPath(cfg), name), cfg.MaxSizeBytes)
	case "null", "discard", "none":
		return io.Discard, func() error { return nil }, nil
	default:
		return os.Stdout, func() error { return nil }, nil
	}
}

func logPath(cfg config.LogConfig) string {
	if path := strings.TrimSpace(cfg.Path); path != "" {
		return path
	}
	return filepath.Join("storage", "logs")
}

func openLogFile(path string, maxSize int64) (io.Writer, func() error, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return os.Stdout, func() error { return nil }, err
	}
	writer := &rotatingFileWriter{path: path, maxSize: maxSize}
	if err := writer.open(); err != nil {
		return os.Stdout, func() error { return nil }, err
	}
	return writer, writer.Close, nil
}

type rotatingFileWriter struct {
	mu      sync.Mutex
	path    string
	maxSize int64
	file    *os.File
	size    int64
}

func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}
	if w.maxSize > 0 && w.size > 0 && w.size+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

func (w *rotatingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	w.size = 0
	return err
}

func (w *rotatingFileWriter) open() error {
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}
	w.file = file
	w.size = info.Size()
	return nil
}

func (w *rotatingFileWriter) rotate() error {
	if w.file != nil {
		if err := w.file.Close(); err != nil {
			return err
		}
		w.file = nil
		w.size = 0
	}
	if _, err := os.Stat(w.path); err == nil {
		if err := os.Rename(w.path, nextRotatedPath(w.path)); err != nil {
			return err
		}
	}
	return w.open()
}

func nextRotatedPath(path string) string {
	dir := filepath.Dir(path)
	ext := filepath.Ext(path)
	base := strings.TrimSuffix(filepath.Base(path), ext)
	stamp := time.Now().Format("20060102150405")
	for i := 1; ; i++ {
		name := fmt.Sprintf("%s-%s.%03d%s", base, stamp, i, ext)
		candidate := filepath.Join(dir, name)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

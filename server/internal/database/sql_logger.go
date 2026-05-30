package database

import (
	"context"
	"errors"
	"fmt"
	"log"
	"runtime"
	"strings"
	"time"

	"wjfcms-go/internal/config"
	"wjfcms-go/internal/requestlog"

	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQLLogger struct {
	level         logger.LogLevel
	logSQL        bool
	logSlowSQL    bool
	logErrorSQL   bool
	requestLogSQL bool
	slowThreshold time.Duration
}

func NewSQLLogger(cfg config.Config) logger.Interface {
	dbCfg := cfg.DB
	if !dbCfg.LogSQL && !dbCfg.LogSlowSQL && !dbCfg.LogErrorSQL && !cfg.Log.RequestEnabled {
		return logger.Default.LogMode(logger.Silent)
	}

	return SQLLogger{
		level:         parseLogLevel(dbCfg.LogLevel),
		logSQL:        dbCfg.LogSQL,
		logSlowSQL:    dbCfg.LogSlowSQL,
		logErrorSQL:   dbCfg.LogErrorSQL,
		requestLogSQL: cfg.Log.RequestEnabled,
		slowThreshold: time.Duration(dbCfg.SlowThresholdMS) * time.Millisecond,
	}
}

func (l SQLLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.level = level
	return l
}

func (l SQLLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Info {
		log.Printf("[gorm] [info] "+msg, data...)
	}
}

func (l SQLLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Warn {
		log.Printf("[gorm] [warn] "+msg, data...)
	}
}

func (l SQLLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Error {
		log.Printf("[gorm] [error] "+msg, data...)
	}
}

func (l SQLLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.level <= logger.Silent && !l.requestLogSQL {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()
	source := callerSource()
	rowsText := "-"
	if rows >= 0 {
		rowsText = fmt.Sprintf("%d", rows)
	}

	recordType := "sql"
	errText := ""
	hasError := err != nil && !errors.Is(err, gorm.ErrRecordNotFound)
	isSlow := l.slowThreshold > 0 && elapsed > l.slowThreshold
	if hasError {
		recordType = "error"
		errText = err.Error()
	} else if isSlow {
		recordType = "slow"
	}
	if l.requestLogSQL && !errors.Is(err, gorm.ErrRecordNotFound) {
		requestlog.AddSQL(sqlRecord(recordType, source, elapsed, rowsText, sql, errText))
	}

	switch {
	case hasError && l.logErrorSQL:
		log.Printf("[gorm] [error] [%s] [%.2fms] [rows:%s] %s | %v", source, float64(elapsed.Nanoseconds())/1e6, rowsText, sql, err)
	case l.logSlowSQL && isSlow:
		log.Printf("[gorm] [slow] [%s] [%.2fms] [rows:%s] %s", source, float64(elapsed.Nanoseconds())/1e6, rowsText, sql)
	case l.logSQL && l.level >= logger.Info:
		log.Printf("[gorm] [sql] [%s] [%.2fms] [rows:%s] %s", source, float64(elapsed.Nanoseconds())/1e6, rowsText, sql)
	}
}

func sqlRecord(recordType string, source string, elapsed time.Duration, rows string, sql string, errText string) requestlog.SQLRecord {
	return requestlog.SQLRecord{
		Type:     recordType,
		Source:   source,
		Elapsed:  float64(elapsed.Nanoseconds()) / 1e6,
		Rows:     rows,
		SQL:      sql,
		Error:    errText,
		CreateAt: time.Now().Format(time.RFC3339Nano),
	}
}

func parseLogLevel(value string) logger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "silent":
		return logger.Silent
	case "error":
		return logger.Error
	case "warn", "warning":
		return logger.Warn
	default:
		return logger.Info
	}
}

func callerSource() string {
	pcs := make([]uintptr, 32)
	n := runtime.Callers(4, pcs)
	frames := runtime.CallersFrames(pcs[:n])
	for {
		frame, more := frames.Next()
		if isApplicationFrame(frame.File) {
			return fmt.Sprintf("%s:%d", normalizePath(frame.File), frame.Line)
		}
		if !more {
			break
		}
	}
	return "unknown:0"
}

func isApplicationFrame(file string) bool {
	if file == "" {
		return false
	}
	file = strings.ReplaceAll(file, "\\", "/")
	if strings.Contains(file, "gorm.io/") ||
		strings.Contains(file, "database/sql") ||
		strings.Contains(file, "internal/database/sql_logger.go") ||
		strings.Contains(file, "runtime/") {
		return false
	}
	return strings.Contains(file, "/wjfcms-go/server/") ||
		strings.Contains(file, "/internal/handler/") ||
		strings.Contains(file, "/internal/service/") ||
		strings.Contains(file, "/internal/router/") ||
		strings.Contains(file, "/cmd/api/")
}

func normalizePath(file string) string {
	file = strings.ReplaceAll(file, "\\", "/")
	if index := strings.Index(file, "/wjfcms-go/server/"); index >= 0 {
		return file[index+len("/wjfcms-go/server/"):]
	}
	return file
}

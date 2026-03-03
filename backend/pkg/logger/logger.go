package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ModuleWeight defines the relative importance weight for each module.
// Higher values indicate more critical modules in the system.
var ModuleWeight = map[string]int{
	"server":      100,
	"database":    95,
	"redis":       90,
	"auth":        90,
	"gateway":     85,
	"websocket":   85,
	"bot":         80,
	"interaction": 80,
	"asset":       75,
	"avatar":      75,
	"storage":     70,
	"queue":       70,
	"event":       65,
	"middleware":   60,
	"cache":       55,
	"handler":     50,
}

// Config holds logging configuration.
type Config struct {
	Level    string
	Format   string
	Output   string
	FilePath string
}

// New creates a structured zap logger with the specified configuration.
// Log output format includes: datetime, module, level (priority), weight.
func New(cfg Config) *zap.Logger {
	level := parseLevel(cfg.Level)

	encoderCfg := zapcore.EncoderConfig{
		TimeKey:       "datetime",
		LevelKey:      "priority",
		NameKey:       "module",
		CallerKey:     "caller",
		MessageKey:    "message",
		StacktraceKey: "stacktrace",
		LineEnding:    zapcore.DefaultLineEnding,
		EncodeLevel:   zapcore.CapitalLevelEncoder,
		EncodeTime:    zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000"),
		EncodeCaller:  zapcore.ShortCallerEncoder,
		EncodeName:    zapcore.FullNameEncoder,
	}

	if cfg.Format == "console" {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderCfg.ConsoleSeparator = " | "
	}

	var encoder zapcore.Encoder
	if cfg.Format == "json" {
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	}

	var writeSyncer zapcore.WriteSyncer
	switch cfg.Output {
	case "file":
		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(fmt.Sprintf("failed to open log file %s: %v", cfg.FilePath, err))
		}
		writeSyncer = zapcore.AddSync(file)
	case "both":
		file, err := os.OpenFile(cfg.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(fmt.Sprintf("failed to open log file %s: %v", cfg.FilePath, err))
		}
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), zapcore.AddSync(file))
	default:
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)

	return zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
}

// Named returns a child logger scoped to the given module name.
// The module weight is automatically attached based on the module registry.
func Named(parent *zap.Logger, module string) *zap.Logger {
	weight := ModuleWeight[module]
	if weight == 0 {
		weight = 10
	}
	return parent.Named(module).With(zap.Int("weight", weight))
}

func parseLevel(level string) zapcore.Level {
	switch level {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}

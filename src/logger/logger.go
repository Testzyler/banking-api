package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.SugaredLogger
)

type LoggerConfig struct {
	Level       string `mapstructure:"level"`       // debug, info, warn, error
	Environment string `mapstructure:"environment"` // development, production
	LogColor    bool   `mapstructure:"log_color"`   // true or false
	LogJson     bool   `mapstructure:"log_json"`    // true or false
}

func getLogLevel(level string) zapcore.Level {
	switch level {
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "debug":
		return zapcore.DebugLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func InitLogger(config *LoggerConfig) error {
	var logEncoderConfig zapcore.EncoderConfig
	var logEncoder zapcore.Encoder

	logLevel := getLogLevel(config.Level)

	// Set encoder config based on environment
	if config.Environment == "production" {
		logEncoderConfig = zap.NewProductionEncoderConfig()
		logEncoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
		logEncoder = zapcore.NewJSONEncoder(logEncoderConfig)
	} else {
		logEncoderConfig = zap.NewDevelopmentEncoderConfig()
		logEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		if config.LogColor {
			logEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		} else {
			logEncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		}

		if config.LogJson {
			logEncoder = zapcore.NewJSONEncoder(logEncoderConfig)
		} else {
			logEncoder = zapcore.NewConsoleEncoder(logEncoderConfig)
		}
	}

	core := zapcore.NewCore(
		logEncoder,
		os.Stdout,
		zap.NewAtomicLevelAt(logLevel),
	)

	zapLogger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zap.ErrorLevel))
	Logger = zapLogger.Sugar()
	return nil
}

func SyncLogger() {
	Logger.Infof("Flush logger")
	Logger.Sync()
}

// Convenience methods to use the global logger instance
func Info(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Infow(msg, keysAndValues...)
	}
}

func Infof(template string, args ...interface{}) {
	if Logger != nil {
		Logger.Infof(template, args...)
	}
}

func Debug(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Debugw(msg, keysAndValues...)
	}
}

func Debugf(template string, args ...interface{}) {
	if Logger != nil {
		Logger.Debugf(template, args...)
	}
}

func Warn(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Warnw(msg, keysAndValues...)
	}
}

func Warnf(template string, args ...interface{}) {
	if Logger != nil {
		Logger.Warnf(template, args...)
	}
}

func Error(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Errorw(msg, keysAndValues...)
	}
}

func Errorf(template string, args ...interface{}) {
	if Logger != nil {
		Logger.Errorf(template, args...)
	}
}

func Fatal(msg string, keysAndValues ...interface{}) {
	if Logger != nil {
		Logger.Fatalw(msg, keysAndValues...)
	} else {
		// Fallback if logger is not initialized
		os.Exit(1)
	}
}

func Fatalf(template string, args ...interface{}) {
	if Logger != nil {
		Logger.Fatalf(template, args...)
	} else {
		os.Exit(1)
	}
}

func With(args ...interface{}) *zap.SugaredLogger {
	if Logger != nil {
		return Logger.With(args...)
	}
	return nil
}

func GetLogger() *zap.SugaredLogger {
	return Logger
}

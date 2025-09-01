/*
 * Use of this software is governed by the Business Source License
 * included in the LICENSE file. Production use is permitted, but
 * offering this software as a managed service requires a separate
 * commercial license.
 */

package serve

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.Logger
var LogLevel zapcore.Level = zapcore.DebugLevel
var LogJsonFmt bool = false

func InitLogger() {

	// defaults
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(LogLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "Logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder, // color when console
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// json overrides
	if !LogJsonFmt {
		cfg.Encoding = "console"

		// use color logs
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

		// disable caller filepath
		cfg.EncoderConfig.EncodeCaller = nil
		cfg.EncoderConfig.CallerKey = ""
		cfg.DisableCaller = true

		// disable stack trace
		cfg.EncoderConfig.StacktraceKey = ""
		cfg.DisableStacktrace = true
	}

	lgr, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to create Logger: %v", err))
	} else {
		Logger = lgr
	}
}

func mustLogger() {
	if Logger == nil {
		panic("Logger not initialized")
	}
}

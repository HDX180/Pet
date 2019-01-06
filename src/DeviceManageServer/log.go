package DeviceManageServer

import (
	"go.uber.org/zap"
	// "go.uber.org/zap/zapcore"
)

var Logger *zap.Logger

func InitLogger() error {
	level := zap.NewAtomicLevelAt(zap.DebugLevel)
	custom_cfg := zap.Config{
		Level:            level,
		Development:      true,
		Encoding:         "console",
		EncoderConfig:    zap.NewDevelopmentEncoderConfig(),
		OutputPaths:      []string{"stderr", "dms.log"},
		ErrorOutputPaths: []string{"stderr"},
	}

	// cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	var err error
	Logger, err = custom_cfg.Build()
	return err
}

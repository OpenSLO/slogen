package libs

import (
	"github.com/mattn/go-colorable"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

func init() {

	base := zap.NewDevelopmentEncoderConfig()
	base.EncodeLevel = zapcore.CapitalColorLevelEncoder
	nosugar := zap.New(zapcore.NewCore(
		zapcore.NewConsoleEncoder(base),
		zapcore.AddSync(colorable.NewColorableStdout()),
		zapcore.DebugLevel,
	))

	log = nosugar.Sugar()
}

func Log() *zap.SugaredLogger {
	return log
}

type InfoLogger struct {
}

func (t InfoLogger) Printf(format string, vals ...interface{}) {
	log.Infof(format, vals...)
}

package libs

import "go.uber.org/zap"

var log *zap.SugaredLogger

func init() {
	plainLogger, err := zap.NewDevelopment()

	if err != nil {
		panic(err)
	}

	log = plainLogger.Sugar()
}

func Log() *zap.SugaredLogger {
	return log
}

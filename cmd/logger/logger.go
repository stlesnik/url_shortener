package logger

import "go.uber.org/zap"

var Sugaarz *zap.SugaredLogger

func InitLogger() {
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	Sugaarz = logger.Sugar()
}

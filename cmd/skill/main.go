package main

import (
	"log"
	"math/rand"
	"time"

	"github.com/petergtz/alexa-journal/cmd/skill/factory"

	"github.com/petergtz/go-alexa/lambda"

	"go.uber.org/zap"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	logger := createLoggerWith(zap.NewAtomicLevelAt(zap.DebugLevel))
	defer logger.Sync()

	lambda.StartLambdaSkill(factory.CreateSkill(logger), logger)
}

func createLoggerWith(logLevel zap.AtomicLevel) *zap.SugaredLogger {
	loggerConfig := zap.NewProductionConfig()
	loggerConfig.Level = logLevel
	loggerConfig.DisableStacktrace = true
	logger, e := loggerConfig.Build()
	if e != nil {
		log.Panic(e)
	}
	return logger.Sugar().With("function-instance-id", rand.Int63())
}
